package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
)

const (
	tokenHMACKeyParameter     = "/ucs503/admin/token-hmac-key"
	analysisAPITokenParameter = "/ucs503/analysis/api-token"
)

type AuthorizerRequest struct {
	Headers map[string]string `json:"headers"`
}

type AuthorizerResponse struct {
	IsAuthorized bool              `json:"isAuthorized"`
	Context      map[string]string `json:"context,omitempty"`
}

type SecretStore interface {
	Get(context.Context, string) (string, error)
}

func main() {
	awsConfig, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		panic(fmt.Errorf("load AWS configuration: %w", err))
	}

	store := &ssmSecretStore{client: ssm.NewFromConfig(awsConfig)}
	lambda.Start(NewAuthorizerHandler(store, time.Now))
}

func VerifyArchiveBearer(token, signingKey string, now time.Time) bool {
	if token == "" || len(signingKey) < 32 {
		return false
	}

	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return false
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil || len(payload) != 40 {
		return false
	}
	providedMAC, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil || len(providedMAC) != sha256.Size {
		return false
	}

	mac := hmac.New(sha256.New, []byte(signingKey))
	_, _ = mac.Write(payload)
	if !hmac.Equal(providedMAC, mac.Sum(nil)) {
		return false
	}

	expiresAt := time.Unix(0, int64(binary.BigEndian.Uint64(payload[:8])))
	return expiresAt.After(now)
}

func NewAuthorizerHandler(store SecretStore, now func() time.Time) func(context.Context, AuthorizerRequest) (AuthorizerResponse, error) {
	return func(ctx context.Context, request AuthorizerRequest) (AuthorizerResponse, error) {
		bearer, ok := authorizationBearer(request.Headers)
		if !ok {
			return AuthorizerResponse{IsAuthorized: false}, nil
		}

		signingKey, err := store.Get(ctx, tokenHMACKeyParameter)
		if err != nil {
			return AuthorizerResponse{}, fmt.Errorf("read token signing key: %w", err)
		}
		if !VerifyArchiveBearer(bearer, signingKey, now()) {
			return AuthorizerResponse{IsAuthorized: false}, nil
		}

		analysisToken, err := store.Get(ctx, analysisAPITokenParameter)
		if err != nil {
			return AuthorizerResponse{}, fmt.Errorf("read analysis API token: %w", err)
		}
		if analysisToken == "" {
			return AuthorizerResponse{}, errors.New("analysis API token is empty")
		}

		return AuthorizerResponse{
			IsAuthorized: true,
			Context:      map[string]string{"analysisApiToken": analysisToken},
		}, nil
	}
}

func authorizationBearer(headers map[string]string) (string, bool) {
	var authorization string
	foundAuthorization := false
	for name, value := range headers {
		if strings.EqualFold(name, "Authorization") {
			if foundAuthorization {
				return "", false
			}
			foundAuthorization = true
			authorization = value
		}
	}

	parts := strings.Fields(authorization)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", false
	}
	return parts[1], true
}

type parameterClient interface {
	GetParameter(context.Context, *ssm.GetParameterInput, ...func(*ssm.Options)) (*ssm.GetParameterOutput, error)
}

type ssmSecretStore struct {
	client parameterClient
}

func (s *ssmSecretStore) Get(ctx context.Context, name string) (string, error) {
	output, err := s.client.GetParameter(ctx, &ssm.GetParameterInput{
		Name:           aws.String(name),
		WithDecryption: aws.Bool(true),
	})
	if err != nil {
		return "", fmt.Errorf("get SSM parameter %s: %w", name, err)
	}
	if output == nil || output.Parameter == nil || output.Parameter.Type != types.ParameterTypeSecureString ||
		output.Parameter.Value == nil || *output.Parameter.Value == "" {
		return "", fmt.Errorf("SSM parameter %s is missing, empty, or not a SecureString", name)
	}
	return *output.Parameter.Value, nil
}
