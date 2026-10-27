package main

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
)

const startupTimeout = 10 * time.Second

type apigatewayV2Handler func(context.Context, events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error)

type ssmParameterAPI interface {
	DescribeParameters(context.Context, *ssm.DescribeParametersInput, ...func(*ssm.Options)) (*ssm.DescribeParametersOutput, error)
	GetParameter(context.Context, *ssm.GetParameterInput, ...func(*ssm.Options)) (*ssm.GetParameterOutput, error)
}

type lambdaConfig struct {
	region              string
	bucket              string
	table               string
	passkeyParameter    string
	signingKeyParameter string
}

func newAPIGatewayV2Handler(handler http.Handler) apigatewayV2Handler {
	adapter := httpadapter.NewV2(handler)
	return adapter.ProxyWithContext
}

func isLambdaRuntime() bool {
	return strings.TrimSpace(os.Getenv("AWS_LAMBDA_RUNTIME_API")) != ""
}

func readLambdaConfig(getenv func(string) string) (lambdaConfig, error) {
	values := lambdaConfig{
		region:              strings.TrimSpace(getenv("AWS_REGION")),
		bucket:              strings.TrimSpace(getenv("S3_BUCKET")),
		table:               strings.TrimSpace(getenv("DYNAMODB_TABLE")),
		passkeyParameter:    strings.TrimSpace(getenv("ADMIN_PASSKEY_PARAMETER")),
		signingKeyParameter: strings.TrimSpace(getenv("TOKEN_HMAC_KEY_PARAMETER")),
	}
	for name, value := range map[string]string{
		"AWS_REGION":               values.region,
		"S3_BUCKET":                values.bucket,
		"DYNAMODB_TABLE":           values.table,
		"ADMIN_PASSKEY_PARAMETER":  values.passkeyParameter,
		"TOKEN_HMAC_KEY_PARAMETER": values.signingKeyParameter,
	} {
		if value == "" {
			return lambdaConfig{}, fmt.Errorf("required Lambda configuration %s is missing", name)
		}
	}
	if err := validateCORSAllowedOrigins(getenv("CORS_ALLOWED_ORIGINS")); err != nil {
		return lambdaConfig{}, fmt.Errorf("required Lambda configuration CORS_ALLOWED_ORIGINS: %w", err)
	}
	return values, nil
}

func loadSSMSecureString(ctx context.Context, client ssmParameterAPI, name string) (string, error) {
	if strings.TrimSpace(name) == "" {
		return "", errors.New("SSM parameter name is required")
	}
	describeCtx, cancel := context.WithTimeout(ctx, startupTimeout)
	defer cancel()
	metadata, err := client.DescribeParameters(describeCtx, &ssm.DescribeParametersInput{
		ParameterFilters: []types.ParameterStringFilter{
			{Key: aws.String("Name"), Option: aws.String("Equals"), Values: []string{name}},
		},
	})
	if err != nil {
		return "", fmt.Errorf("describe SSM parameter: %w", err)
	}
	if metadata == nil || len(metadata.Parameters) != 1 ||
		aws.ToString(metadata.Parameters[0].Name) != name {
		return "", errors.New("SSM parameter metadata is missing or ambiguous")
	}
	parameterMetadata := metadata.Parameters[0]
	if parameterMetadata.Type != types.ParameterTypeSecureString ||
		parameterMetadata.Tier != types.ParameterTierStandard ||
		!isAWSManagedSSMKey(aws.ToString(parameterMetadata.KeyId)) {
		return "", errors.New("SSM parameter must be a standard SecureString encrypted with the AWS-managed SSM key")
	}

	valueCtx, cancel := context.WithTimeout(ctx, startupTimeout)
	defer cancel()
	output, err := client.GetParameter(valueCtx, &ssm.GetParameterInput{
		Name:           aws.String(name),
		WithDecryption: aws.Bool(true),
	})
	if err != nil {
		return "", fmt.Errorf("read SSM SecureString: %w", err)
	}
	if output == nil || output.Parameter == nil ||
		output.Parameter.Type != types.ParameterTypeSecureString ||
		aws.ToString(output.Parameter.Name) != name {
		return "", errors.New("SSM SecureString response is invalid")
	}
	value := aws.ToString(output.Parameter.Value)
	if strings.TrimSpace(value) == "" {
		return "", errors.New("SSM SecureString value is empty")
	}
	return value, nil
}

func isAWSManagedSSMKey(keyID string) bool {
	keyID = strings.ToLower(strings.TrimSpace(keyID))
	return keyID == "alias/aws/ssm" || strings.HasSuffix(keyID, ":alias/aws/ssm")
}

func newProductionAuthenticator(ctx context.Context, client ssmParameterAPI, passkeyParameter, signingKeyParameter string) (Authenticator, error) {
	passkey, err := loadSSMSecureString(ctx, client, passkeyParameter)
	if err != nil {
		return nil, fmt.Errorf("load admin passkey: %w", err)
	}
	signingKey, err := loadSSMSecureString(ctx, client, signingKeyParameter)
	if err != nil {
		return nil, fmt.Errorf("load token signing key: %w", err)
	}
	auth, err := NewHMACTokenAuthenticator(passkey, signingKey)
	if err != nil {
		return nil, fmt.Errorf("configure token authenticator: %w", err)
	}
	return auth, nil
}

func newLocalAuthenticator(passkey, signingKey string) (Authenticator, error) {
	if strings.TrimSpace(passkey) == "" {
		return nil, errors.New("ADMIN_PASSKEY is required for local mode")
	}
	if signingKey == "" {
		key := make([]byte, 32)
		if _, err := rand.Read(key); err != nil {
			return nil, fmt.Errorf("generate ephemeral local token key: %w", err)
		}
		signingKey = string(key)
	}
	return NewHMACTokenAuthenticator(passkey, signingKey)
}

func newLocalDependencies(passkey, signingKey, archivePath, uploadRoot string) (Dependencies, error) {
	auth, err := newLocalAuthenticator(passkey, signingKey)
	if err != nil {
		return Dependencies{}, err
	}
	if strings.TrimSpace(archivePath) == "" || strings.TrimSpace(uploadRoot) == "" {
		return Dependencies{}, errors.New("local archive paths are required")
	}
	objects := newLocalObjectStore(uploadRoot)
	repository := newLocalArchiveRepository(archivePath)
	accountPath := os.Getenv("USER_DATA_FILE")
	if accountPath == "" {
		accountPath = filepath.Join(filepath.Dir(archivePath), "accounts.json")
	}
	accounts, err := accountServiceFromEnvironment(&localAccountStore{path: accountPath}, os.Getenv("GITHUB_CLIENT_SECRET"))
	if err != nil {
		return Dependencies{}, err
	}
	return Dependencies{
		Auth:       auth,
		Archives:   newArchiveServiceWithLogger(repository, objects, slog.Default()),
		LocalFiles: objects,
		Logger:     slog.Default(),
		Accounts:   accounts,
	}, nil
}

func newAWSDependencies(ctx context.Context) (Dependencies, error) {
	values, err := readLambdaConfig(os.Getenv)
	if err != nil {
		return Dependencies{}, err
	}
	awsConfig, err := config.LoadDefaultConfig(ctx, config.WithRegion(values.region))
	if err != nil {
		return Dependencies{}, fmt.Errorf("load AWS role configuration: %w", err)
	}
	ssmClient := ssm.NewFromConfig(awsConfig)
	auth, err := newProductionAuthenticator(ctx, ssmClient, values.passkeyParameter, values.signingKeyParameter)
	if err != nil {
		return Dependencies{}, err
	}
	dynamoClient := dynamodb.NewFromConfig(awsConfig)
	s3Client := s3.NewFromConfig(awsConfig)
	repository := NewDynamoDBArchiveRepository(dynamoClient, values.table)
	objects := NewS3ObjectStore(s3Client, s3.NewPresignClient(s3Client), values.bucket)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	var accounts *githubAccountService
	if os.Getenv("GITHUB_CLIENT_ID") != "" {
		table := strings.TrimSpace(os.Getenv("ACCOUNTS_TABLE"))
		if table == "" {
			return Dependencies{}, errors.New("ACCOUNTS_TABLE is required for GitHub sign-in")
		}
		secret, err := loadSSMSecureString(ctx, ssmClient, os.Getenv("GITHUB_CLIENT_SECRET_PARAMETER"))
		if err != nil {
			return Dependencies{}, fmt.Errorf("load GitHub OAuth secret: %w", err)
		}
		accounts, err = accountServiceFromEnvironment(&dynamoAccountStore{client: dynamoClient, table: table}, secret)
		if err != nil {
			return Dependencies{}, err
		}
	} else if os.Getenv("GITHUB_CLIENT_SECRET_PARAMETER") != "" {
		return Dependencies{}, errors.New("GITHUB_CLIENT_ID is required when GitHub authentication is configured")
	}
	return Dependencies{
		Auth:     auth,
		Archives: newArchiveServiceWithLogger(repository, objects, logger),
		Logger:   logger,
		Accounts: accounts,
	}, nil
}

func startLambda() {
	ctx, cancel := context.WithTimeout(context.Background(), startupTimeout)
	defer cancel()
	deps, err := newAWSDependencies(ctx)
	if err != nil {
		slog.Error("Lambda startup failed", "error_type", fmt.Sprintf("%T", err))
		os.Exit(1)
	}
	handler := newRouter(deps)
	lambda.Start(newAPIGatewayV2Handler(handler))
}
