package actions

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"

	"unibee/internal/logic/scenario"
)

func init() {
	scenario.RegisterAction(scenario.StepHTTPRequest, &HTTPAction{})
}

// HTTPAction performs an HTTP request to an external API.
type HTTPAction struct{}

func (a *HTTPAction) Execute(ctx context.Context, execCtx *scenario.ExecutionContext, step *scenario.StepDSL) (map[string]interface{}, error) {
	method, _ := step.Params["method"].(string)
	url, _ := step.Params["url"].(string)

	if url == "" {
		return nil, fmt.Errorf("http_request: url is required")
	}
	if method == "" {
		method = "GET"
	}
	method = strings.ToUpper(method)

	// Build request body
	var bodyReader io.Reader
	if bodyRaw, ok := step.Params["body"]; ok && bodyRaw != nil {
		bodyBytes, err := json.Marshal(bodyRaw)
		if err != nil {
			return nil, fmt.Errorf("http_request: failed to marshal body: %w", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("http_request: failed to create request: %w", err)
	}

	// Set headers
	if headersRaw, ok := step.Params["headers"]; ok {
		if headers, ok := headersRaw.(map[string]interface{}); ok {
			for k, v := range headers {
				if s, ok := v.(string); ok {
					req.Header.Set(k, s)
				}
			}
		}
	}

	// Default content-type for non-GET
	if method != "GET" && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	// Execute with timeout
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http_request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1024*1024)) // 1MB limit
	if err != nil {
		return nil, fmt.Errorf("http_request: failed to read response: %w", err)
	}

	output := map[string]interface{}{
		"status_code": resp.StatusCode,
		"body":        string(respBody),
	}

	// Try to parse JSON response into variables
	var respJSON map[string]interface{}
	if err := json.Unmarshal(respBody, &respJSON); err == nil {
		for k, v := range respJSON {
			if s, ok := v.(string); ok {
				execCtx.Variables["http_"+k] = s
			}
		}
		output["json"] = respJSON
	}

	g.Log().Infof(ctx, "scenario exec %d: HTTP %s %s → %d", execCtx.ExecutionID, method, url, resp.StatusCode)

	if resp.StatusCode >= 400 {
		return output, fmt.Errorf("http_request: server returned %d", resp.StatusCode)
	}

	return output, nil
}
