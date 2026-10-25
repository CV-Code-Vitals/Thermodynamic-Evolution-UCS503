package main

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

const testMethodARN = "arn:aws:execute-api:ap-south-1:123456789012:api123/prod/POST/api/evolve"

func TestAuthorizerHandler_ReturnsScopedRESTPolicyForValidBearer(t *testing.T) {
	bearer := makeTestToken(t, testSigningKey, testNow.Add(time.Minute))
	const internalToken = "internal-analysis-api-token"
	secrets := &fakeSecretStore{
		values: map[string]string{
			tokenHMACKeyParameter:     testSigningKey,
			analysisAPITokenParameter: internalToken,
		},
	}
	handler := NewAuthorizerHandler(secrets, func() time.Time { return testNow })
	response := invokeAuthorizer(t, handler, map[string]string{"Authorization": "Bearer " + bearer})

	if _, ok := response["principalId"].(string); !ok {
		t.Fatalf("expected a REST API principal ID, got %#v", response)
	}
	policy, ok := response["policyDocument"].(map[string]interface{})
	if !ok || policy["Version"] != "2012-10-17" {
		t.Fatalf("expected an IAM policy document, got %#v", response["policyDocument"])
	}
	statements, ok := policy["Statement"].([]interface{})
	if !ok || len(statements) != 1 {
		t.Fatalf("expected one policy statement, got %#v", policy["Statement"])
	}
	statement, ok := statements[0].(map[string]interface{})
	if !ok || statement["Action"] != "execute-api:Invoke" ||
		statement["Effect"] != "Allow" || statement["Resource"] != testMethodARN {
		t.Fatalf("expected method-scoped Allow policy, got %#v", statements[0])
	}
	contextValues, ok := response["context"].(map[string]interface{})
	if !ok || len(contextValues) != 1 || contextValues["analysisApiToken"] != internalToken {
		t.Fatalf("expected only the internal API token in context, got %#v", response["context"])
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

func TestAuthorizerHandler_DeniesInvalidBearer(t *testing.T) {
	valid := makeTestToken(t, testSigningKey, testNow.Add(time.Minute))
	tampered := []byte(valid)
	signatureStart := strings.IndexByte(valid, '.') + 1
	if tampered[signatureStart] == 'A' {
		tampered[signatureStart] = 'B'
	} else {
		tampered[signatureStart] = 'A'
	}

	for _, testCase := range []struct {
		name    string
		headers map[string]string
	}{
		{name: "missing"},
		{name: "malformed", headers: map[string]string{"Authorization": "Bearer not-a-token"}},
		{name: "expired", headers: map[string]string{"Authorization": "Bearer " + makeTestToken(t, testSigningKey, testNow)}},
		{name: "tampered", headers: map[string]string{"Authorization": "Bearer " + string(tampered)}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			secrets := &fakeSecretStore{
				values: map[string]string{
					tokenHMACKeyParameter:     testSigningKey,
					analysisAPITokenParameter: "must-not-be-read",
				},
			}
			handler := NewAuthorizerHandler(secrets, func() time.Time { return testNow })
			response := invokeAuthorizer(t, handler, testCase.headers)
			statement := firstPolicyStatement(t, response)
			if statement["Effect"] != "Deny" || statement["Resource"] != testMethodARN {
				t.Fatalf("expected method-scoped Deny policy, got %#v", statement)
			}
			if _, ok := response["context"]; ok {
				t.Fatalf("invalid bearer received authorizer context: %#v", response["context"])
			}
			for _, name := range secrets.requested {
				if name == analysisAPITokenParameter {
					t.Fatal("invalid bearer caused the internal API token to be read")
				}
			}
		})
	}
}

func TestRESTAuthorizer_RejectsAmbiguousAuthorizationHeaders(t *testing.T) {
	secrets := &fakeSecretStore{values: map[string]string{tokenHMACKeyParameter: testSigningKey}}
	handler := NewAuthorizerHandler(secrets, func() time.Time { return testNow })
	response := invokeAuthorizer(t, handler, map[string]string{
		"Authorization": "",
		"authorization": "Bearer " + makeTestToken(t, testSigningKey, testNow.Add(time.Minute)),
	})
	statement := firstPolicyStatement(t, response)
	if statement["Effect"] != "Deny" || len(secrets.requested) != 0 {
		t.Fatalf("expected an early Deny before secret reads, got policy %#v and reads %#v", statement, secrets.requested)
	}
}

func invokeAuthorizer(
	t *testing.T,
	handler func(context.Context, AuthorizerRequest) (AuthorizerResponse, error),
	headers map[string]string,
) map[string]interface{} {
	t.Helper()
	event, err := json.Marshal(map[string]interface{}{
		"headers":   headers,
		"methodArn": testMethodARN,
	})
	if err != nil {
		t.Fatalf("marshal REST authorizer event: %v", err)
	}
	var request AuthorizerRequest
	if err := json.Unmarshal(event, &request); err != nil {
		t.Fatalf("decode REST authorizer event: %v", err)
	}
	output, err := handler(context.Background(), request)
	if err != nil {
		t.Fatalf("handler returned an error: %v", err)
	}
	encoded, err := json.Marshal(output)
	if err != nil {
		t.Fatalf("marshal REST authorizer response: %v", err)
	}
	var response map[string]interface{}
	if err := json.Unmarshal(encoded, &response); err != nil {
		t.Fatalf("decode REST authorizer response: %v", err)
	}
	return response
}

func firstPolicyStatement(t *testing.T, response map[string]interface{}) map[string]interface{} {
	t.Helper()
	policy, ok := response["policyDocument"].(map[string]interface{})
	if !ok || policy["Version"] != "2012-10-17" {
		t.Fatalf("expected an IAM policy document, got %#v", response["policyDocument"])
	}
	statements, ok := policy["Statement"].([]interface{})
	if !ok || len(statements) != 1 {
		t.Fatalf("expected one policy statement, got %#v", policy["Statement"])
	}
	statement, ok := statements[0].(map[string]interface{})
	if !ok {
		t.Fatalf("expected policy statement object, got %#v", statements[0])
	}
	return statement
}
