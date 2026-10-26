package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
)

const testSigningKey = "test-signing-key-with-at-least-32-bytes"

var testNow = time.Date(2026, time.October, 24, 12, 0, 0, 0, time.UTC)

func TestVerifyArchiveBearer_AcceptsValidToken(t *testing.T) {
	token := makeTestToken(t, testSigningKey, testNow.Add(time.Minute))

	if !VerifyArchiveBearer(token, testSigningKey, testNow) {
		t.Fatal("expected a valid, unexpired token to be accepted")
	}
}

func TestVerifyArchiveBearer_AcceptsVersionedToken(t *testing.T) {
	token := makeVersionedTestToken(t, testSigningKey, testNow.Add(-time.Minute), testNow.Add(time.Minute))

	if !VerifyArchiveBearer(token, testSigningKey, testNow) {
		t.Fatal("expected a valid versioned token to be accepted")
	}
}

func TestVerifyArchiveBearer_RejectsVersionedTokenIssuedInFuture(t *testing.T) {
	token := makeVersionedTestToken(t, testSigningKey, testNow.Add(time.Second), testNow.Add(time.Minute))

	if VerifyArchiveBearer(token, testSigningKey, testNow) {
		t.Fatal("expected a future-issued versioned token to be rejected")
	}
}

func TestVerifyArchiveBearer_RejectsVersionedTokenWithInvalidTimes(t *testing.T) {
	token := makeVersionedTestToken(t, testSigningKey, testNow.Add(-time.Minute), testNow.Add(-2*time.Minute))

	if VerifyArchiveBearer(token, testSigningKey, testNow) {
		t.Fatal("expected a versioned token expiring before issuance to be rejected")
	}
}

func TestVerifyArchiveBearer_RejectsMalformedToken(t *testing.T) {
	for _, token := range []string{"", "not-a-token", "a.b", "a.b.c"} {
		t.Run(token, func(t *testing.T) {
			if VerifyArchiveBearer(token, testSigningKey, testNow) {
				t.Fatal("expected malformed token to be rejected")
			}
		})
	}
}

func TestVerifyArchiveBearer_RejectsTamperedToken(t *testing.T) {
	token := makeTestToken(t, testSigningKey, testNow.Add(time.Minute))
	tampered := []byte(token)
	signatureStart := strings.IndexByte(token, '.') + 1
	if tampered[signatureStart] == 'A' {
		tampered[signatureStart] = 'B'
	} else {
		tampered[signatureStart] = 'A'
	}
	token = string(tampered)

	if VerifyArchiveBearer(token, testSigningKey, testNow) {
		t.Fatal("expected tampered token to be rejected")
	}
}

func TestVerifyArchiveBearer_RejectsShortSigningKey(t *testing.T) {
	token := makeTestToken(t, testSigningKey, testNow.Add(time.Minute))

	if VerifyArchiveBearer(token, "short", testNow) {
		t.Fatal("expected token to be rejected with a short signing key")
	}
}

func TestVerifyArchiveBearer_RejectsExpiredToken(t *testing.T) {
	token := makeTestToken(t, testSigningKey, testNow)

	if VerifyArchiveBearer(token, testSigningKey, testNow) {
		t.Fatal("expected token expiring at the current time to be rejected")
	}
}

func TestAuthorizerHandler_ReturnsInternalTokenOnlyForValidBearer(t *testing.T) {
	bearer := makeTestToken(t, testSigningKey, testNow.Add(time.Minute))
	const internalToken = "internal-analysis-api-token"
	secrets := &fakeSecretStore{
		values: map[string]string{
			tokenHMACKeyParameter:     testSigningKey,
			analysisAPITokenParameter: internalToken,
		},
	}
	handler := NewAuthorizerHandler(secrets, func() time.Time { return testNow })

	response, err := handler(context.Background(), AuthorizerRequest{
		Headers:   map[string]string{"authorization": "Bearer " + bearer},
		MethodArn: testMethodARN,
	})
	if err != nil {
		t.Fatalf("handler returned an error: %v", err)
	}
	if len(response.PolicyDocument.Statement) != 1 || response.PolicyDocument.Statement[0].Effect != "Allow" {
		t.Fatalf("expected valid bearer token to receive an Allow policy, got %#v", response.PolicyDocument)
	}
	if len(response.Context) != 1 || response.Context["analysisApiToken"] != internalToken {
		t.Fatalf("expected only the internal API token in context, got %#v", response.Context)
	}
	if response.Context["tokenHMACKey"] != "" || response.Context["bearerToken"] != "" {
		t.Fatalf("signing key or bearer token leaked into context: %#v", response.Context)
	}
	if got := strings.Join(secrets.requested, ","); got != tokenHMACKeyParameter+","+analysisAPITokenParameter {
		t.Fatalf("unexpected parameter reads: %s", got)
	}

	encoded, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("marshal response: %v", err)
	}
	if strings.Contains(string(encoded), bearer) || strings.Contains(string(encoded), testSigningKey) {
		t.Fatalf("response leaked bearer token or signing key: %s", encoded)
	}
}

func TestAuthorizerDoesNotLogBearerOrContextSecrets(t *testing.T) {
	bearer := makeTestToken(t, testSigningKey, testNow.Add(time.Minute))
	const internalToken = "internal-token-must-not-appear-in-logs"
	var output strings.Builder
	previousLogger := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&output, nil)))
	defer slog.SetDefault(previousLogger)

	store := &fakeSecretStore{values: map[string]string{
		tokenHMACKeyParameter:     testSigningKey,
		analysisAPITokenParameter: internalToken,
	}}
	handler := NewAuthorizerHandler(store, func() time.Time { return testNow })
	response, err := handler(context.Background(), AuthorizerRequest{
		Headers:   map[string]string{"Authorization": "Bearer " + bearer},
		MethodArn: testMethodARN,
	})
	if err != nil {
		t.Fatalf("handler returned an error: %v", err)
	}
	if response.Context["analysisApiToken"] != internalToken {
		t.Fatal("expected the integration token to be returned only in authorizer context")
	}
	for _, sensitiveValue := range []string{bearer, internalToken, testSigningKey, "Authorization"} {
		if strings.Contains(output.String(), sensitiveValue) {
			t.Fatalf("authorizer logs exposed %q: %s", sensitiveValue, output.String())
		}
	}
}

func TestAuthorizerHandler_RejectsInvalidBearerWithoutReadingInternalToken(t *testing.T) {
	secrets := &fakeSecretStore{
		values: map[string]string{
			tokenHMACKeyParameter:     testSigningKey,
			analysisAPITokenParameter: "must-not-be-read",
		},
	}
	handler := NewAuthorizerHandler(secrets, func() time.Time { return testNow })

	response, err := handler(context.Background(), AuthorizerRequest{
		MethodArn: testMethodARN,
		Headers:   map[string]string{"Authorization": "Bearer invalid"},
	})
	if err != nil {
		t.Fatalf("handler returned an error: %v", err)
	}
	if len(response.PolicyDocument.Statement) != 1 ||
		response.PolicyDocument.Statement[0].Effect != "Deny" || len(response.Context) != 0 {
		t.Fatalf("expected unauthenticated response without context, got %#v", response)
	}
	if len(secrets.requested) != 1 || secrets.requested[0] != tokenHMACKeyParameter {
		t.Fatalf("expected only the signing key to be read, got %#v", secrets.requested)
	}
}

func TestAuthorizerHandler_RejectsAmbiguousAuthorizationHeaders(t *testing.T) {
	secrets := &fakeSecretStore{values: map[string]string{tokenHMACKeyParameter: testSigningKey}}
	handler := NewAuthorizerHandler(secrets, func() time.Time { return testNow })

	response, err := handler(context.Background(), AuthorizerRequest{
		Headers: map[string]string{
			"Authorization": "",
			"authorization": "Bearer " + makeTestToken(t, testSigningKey, testNow.Add(time.Minute)),
		},
		MethodArn: testMethodARN,
	})
	if err != nil {
		t.Fatalf("handler returned an error: %v", err)
	}
	if len(response.PolicyDocument.Statement) != 1 ||
		response.PolicyDocument.Statement[0].Effect != "Deny" || len(secrets.requested) != 0 {
		t.Fatalf("expected ambiguous headers to be rejected before secret reads, got response %#v and reads %#v", response, secrets.requested)
	}
}

func TestAuthorizerHandler_PropagatesSecretStoreErrors(t *testing.T) {
	secrets := &fakeSecretStore{err: errors.New("parameter store unavailable")}
	handler := NewAuthorizerHandler(secrets, func() time.Time { return testNow })

	_, err := handler(context.Background(), AuthorizerRequest{
		Headers:   map[string]string{"Authorization": "Bearer " + makeTestToken(t, testSigningKey, testNow.Add(time.Minute))},
		MethodArn: testMethodARN,
	})
	if err == nil || !strings.Contains(err.Error(), "parameter store unavailable") {
		t.Fatalf("expected parameter-store failure to be surfaced, got %v", err)
	}
}

func TestSSMSecretStore_RequiresDecryptedSecureString(t *testing.T) {
	for _, testCase := range []struct {
		name   string
		output *ssm.GetParameterOutput
		want   string
	}{
		{
			name: "secure string",
			output: &ssm.GetParameterOutput{
				Parameter: &types.Parameter{
					Type:  types.ParameterTypeSecureString,
					Value: aws.String("secret"),
				},
			},
			want: "secret",
		},
		{name: "nil output"},
		{name: "missing parameter", output: &ssm.GetParameterOutput{}},
		{
			name: "ordinary string",
			output: &ssm.GetParameterOutput{
				Parameter: &types.Parameter{
					Type:  types.ParameterTypeString,
					Value: aws.String("secret"),
				},
			},
		},
		{
			name: "empty secure string",
			output: &ssm.GetParameterOutput{
				Parameter: &types.Parameter{
					Type:  types.ParameterTypeSecureString,
					Value: aws.String(""),
				},
			},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			client := &fakeParameterClient{output: testCase.output}
			store := &ssmSecretStore{client: client}

			value, err := store.Get(context.Background(), tokenHMACKeyParameter)
			if testCase.want == "" {
				if err == nil {
					t.Fatalf("expected invalid parameter result to fail, got value %q", value)
				}
				return
			}
			if err != nil {
				t.Fatalf("expected SecureString to be returned, got error: %v", err)
			}
			if value != testCase.want {
				t.Fatalf("expected %q, got %q", testCase.want, value)
			}
			if client.input == nil || client.input.Name == nil || *client.input.Name != tokenHMACKeyParameter ||
				client.input.WithDecryption == nil || !*client.input.WithDecryption {
				t.Fatalf("expected named parameter to be requested with decryption: %#v", client.input)
			}
		})
	}
}

type fakeSecretStore struct {
	values    map[string]string
	requested []string
	err       error
}

func (s *fakeSecretStore) Get(_ context.Context, name string) (string, error) {
	s.requested = append(s.requested, name)
	if s.err != nil {
		return "", s.err
	}
	value, ok := s.values[name]
	if !ok {
		return "", errors.New("parameter not found")
	}
	return value, nil
}

type fakeParameterClient struct {
	input  *ssm.GetParameterInput
	output *ssm.GetParameterOutput
}

func (c *fakeParameterClient) GetParameter(_ context.Context, input *ssm.GetParameterInput, _ ...func(*ssm.Options)) (*ssm.GetParameterOutput, error) {
	c.input = input
	return c.output, nil
}

func makeTestToken(t *testing.T, signingKey string, expiresAt time.Time) string {
	t.Helper()

	payload := make([]byte, 40)
	binary.BigEndian.PutUint64(payload[:8], uint64(expiresAt.UnixNano()))
	copy(payload[8:], strings.Repeat("n", 32))

	mac := hmac.New(sha256.New, []byte(signingKey))
	if _, err := mac.Write(payload); err != nil {
		t.Fatalf("write HMAC payload: %v", err)
	}

	return base64.RawURLEncoding.EncodeToString(payload) + "." +
		base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func makeVersionedTestToken(t *testing.T, signingKey string, issuedAt, expiresAt time.Time) string {
	t.Helper()

	payload := make([]byte, 49)
	payload[0] = 1
	binary.BigEndian.PutUint64(payload[1:9], uint64(issuedAt.UnixNano()))
	binary.BigEndian.PutUint64(payload[9:17], uint64(expiresAt.UnixNano()))
	copy(payload[17:], strings.Repeat("n", len(payload)-17))

	mac := hmac.New(sha256.New, []byte(signingKey))
	if _, err := mac.Write(payload); err != nil {
		t.Fatalf("write HMAC: %v", err)
	}
	return base64.RawURLEncoding.EncodeToString(payload) + "." +
		base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
