package ui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type httpAPIClient struct {
	baseURL string
	apiKey  string
	http    *http.Client
}

func newAPIClient(cfg Config) *httpAPIClient {
	return &httpAPIClient{
		baseURL: cfg.httpURL(""),
		apiKey:  cfg.APIKey,
		http:    &http.Client{},
	}
}

func (c *httpAPIClient) do(method, path string, reqBody, dst any, wantStatus int) error {
	var body io.Reader
	if reqBody != nil {
		data, err := json.Marshal(reqBody)
		if err != nil {
			return err
		}
		body = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, c.baseURL+path, body)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", c.apiKey)
	if reqBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != wantStatus {
		return fmt.Errorf("server returned %d", resp.StatusCode)
	}

	if dst != nil {
		return json.NewDecoder(resp.Body).Decode(dst)
	}
	return nil
}
