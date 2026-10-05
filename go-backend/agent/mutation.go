package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type MutationAgent struct {
	Client  *http.Client
	APIBase string
	APIKey  string
	Model   string
}

type llmMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type llmRequest struct {
	Model    string       `json:"model"`
	Messages []llmMessage `json:"messages"`
}

type llmResponse struct {
	Choices []struct {
		Message llmMessage `json:"message"`
	} `json:"choices"`
}

func (m *MutationAgent) ProposeMutation(ctx context.Context, currentCode, funcName string) (string, error) {
	prompt := fmt.Sprintf(`You are an expert system engineer that refactors code to optimize its "thermodynamic energy" (cyclomatic complexity, nesting depth, allocation in hot loops).

You will be provided with a target function: "%s".

Your task:
- Rewrite the function body to reduce nesting, minimize complexity, and optimize loop allocations.
- strictly preserve function signatures, parameter names, and return types.
- Require the response to be ONLY the replacement function body enclosed in a single fenced code block.

Function to mutate:
%s
`, funcName, currentCode)

	reqBody := llmRequest{
		Model: m.Model,
		Messages: []llmMessage{
			{Role: "system", Content: "You are an expert code optimizer. Only return code inside a single fenced code block."},
			{Role: "user", Content: prompt},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", m.APIBase+"/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	if m.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+m.APIKey)
	}

	resp, err := m.Client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf("LLM API error (status %d): %s", resp.StatusCode, string(body))
	}

	var respData llmResponse
	if err := json.NewDecoder(resp.Body).Decode(&respData); err != nil {
		return "", err
	}

	if len(respData.Choices) == 0 {
		return "", fmt.Errorf("no response from LLM")
	}

	content := respData.Choices[0].Message.Content
	return extractCodeBlock(content), nil
}

func extractCodeBlock(content string) string {
	lines := strings.Split(content, "\n")
	var code []string
	inBlock := false

	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			if inBlock {
				break
			}
			inBlock = true
			continue
		}
		if inBlock {
			code = append(code, line)
		}
	}

	if !inBlock {
		return strings.TrimSpace(content)
	}
	return strings.Join(code, "\n")
}
