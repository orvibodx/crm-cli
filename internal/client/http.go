package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/orvibodx/crm-cli/internal/api"
)

var envURLs = map[string]string{
	"prod":  "https://crm.orvibo.com/api/",
	"test":  "https://crm-test.orvibo.com/test/api/",
	"local": "http://localhost:8090/api/",
}

func DoRequest(method, path string, body interface{}, token, envName string) (*api.APIResponse, error) {
	baseURL, ok := envURLs[envName]
	if !ok {
		return nil, fmt.Errorf("unknown env: %s", envName)
	}

	fullURL := baseURL + path
	var req *http.Request
	var err error

	normalized := api.NormalizePath(path)

	if api.IsRequestParam(normalized) && body != nil {
		params := toQueryString(body)
		if params != "" {
			fullURL += "?" + params
		}
		req, err = http.NewRequest(method, fullURL, nil)
	} else {
		var bodyReader io.Reader
		if body != nil {
			jsonData, marshalErr := json.Marshal(body)
			if marshalErr != nil {
				return nil, fmt.Errorf("marshal body: %w", marshalErr)
			}
			bodyReader = bytes.NewReader(jsonData)
		}
		req, err = http.NewRequest(method, fullURL, bodyReader)
	}

	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	if token != "" {
		req.Header.Set("Admin-Token", token)
	}
	if body != nil && !api.IsRequestParam(normalized) {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var apiResp api.APIResponse
	if err := json.Unmarshal(data, &apiResp); err != nil {
		return nil, fmt.Errorf("parse response: %s", string(data))
	}

	return &apiResp, nil
}

func toQueryString(v interface{}) string {
	if v == nil {
		return ""
	}
	jsonData, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	var m map[string]interface{}
	if err := json.Unmarshal(jsonData, &m); err != nil {
		return ""
	}
	params := url.Values{}
	for key, val := range m {
		if val == nil {
			continue
		}
		switch v := val.(type) {
		case string:
			if v != "" {
				params.Set(key, v)
			}
		case float64:
			params.Set(key, fmt.Sprintf("%.0f", v))
		default:
			params.Set(key, fmt.Sprintf("%v", v))
		}
	}
	return params.Encode()
}

func CheckResponse(resp *api.APIResponse) error {
	switch resp.Code {
	case 0, 200:
		return nil
	case 302:
		return fmt.Errorf("token expired, run: crm-cli auth login")
	case 401:
		return fmt.Errorf("unauthorized: %s", resp.Msg)
	default:
		return fmt.Errorf("API error (code %d): %s", resp.Code, resp.Msg)
	}
}
