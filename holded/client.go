package holded

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/luisra51/mcp-holded/internal"
)

type Client struct {
	BaseURL string
	Client  *http.Client
	APIKey  string
	Limiter *internal.MultiLimiter
}

var retryableStatusCodes = map[int]struct{}{
	http.StatusTooManyRequests:    {},
	http.StatusBadGateway:         {},
	http.StatusServiceUnavailable: {},
	http.StatusGatewayTimeout:     {},
}

func NewClient(cfg Config) *Client {
	transport := http.DefaultTransport
	client := &http.Client{Timeout: cfg.Timeout, Transport: transport}
	return &Client{
		BaseURL: cfg.URL,
		Client:  client,
		APIKey:  cfg.APIKey,
		Limiter: limiterFromConfig(cfg),
	}
}

func limiterFromConfig(cfg Config) *internal.MultiLimiter {
	if cfg.DisableRateLimit {
		return nil
	}
	return internal.NewDefaultLimiter()
}

func (c *Client) do(req *http.Request) (*http.Response, error) {
	if c.APIKey == "" {
		return nil, fmt.Errorf("missing HOLDED_API_KEY")
	}
	var bodyBytes []byte
	if req.Body != nil {
		var err error
		bodyBytes, err = io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		req.Body.Close()
	}
	delays := []time.Duration{0, time.Second, 2 * time.Second, 4 * time.Second}
	var lastErr error
	for attempt, delay := range delays {
		if delay > 0 {
			select {
			case <-req.Context().Done():
				return nil, req.Context().Err()
			case <-time.After(delay):
			}
		}
		if c.Limiter != nil {
			if err := c.Limiter.Wait(req.Context()); err != nil {
				return nil, err
			}
		}
		attemptReq := req.Clone(req.Context())
		if bodyBytes != nil {
			attemptReq.Body = io.NopCloser(bytes.NewReader(bodyBytes))
			attemptReq.ContentLength = int64(len(bodyBytes))
		}
		attemptReq.Header = req.Header.Clone()
		attemptReq.Header.Set("key", c.APIKey)
		resp, err := c.Client.Do(attemptReq)
		if err != nil {
			lastErr = err
			if attempt < len(delays)-1 {
				continue
			}
			return nil, err
		}
		if _, retry := retryableStatusCodes[resp.StatusCode]; retry && attempt < len(delays)-1 {
			_, _ = io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
			lastErr = fmt.Errorf("upstream api retryable status: %s", resp.Status)
			continue
		}
		return resp, nil
	}
	return nil, lastErr
}

func (c *Client) NewRequest(method, path string, q url.Values, body any) (*http.Request, error) {
	base := strings.TrimRight(c.BaseURL, "/") + "/" + strings.TrimLeft(path, "/")
	if len(q) > 0 {
		base += "?" + q.Encode()
	}
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, base, reader)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return req, nil
}

func (c *Client) DoJSON(req *http.Request, out any) error {
	resp, err := c.do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
		if len(b) == 0 {
			return fmt.Errorf("upstream api error: %s", resp.Status)
		}
		return fmt.Errorf("upstream api error: %s: %s", resp.Status, strings.TrimSpace(string(b)))
	}
	if out == nil {
		return nil
	}
	if resp.StatusCode == http.StatusNoContent {
		return nil
	}
	err = json.NewDecoder(resp.Body).Decode(out)
	if err == io.EOF {
		return nil
	}
	return err
}

func (c *Client) DoRaw(req *http.Request) ([]byte, string, error) {
	resp, err := c.do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
		if len(b) == 0 {
			return nil, "", fmt.Errorf("upstream api error: %s", resp.Status)
		}
		return nil, "", fmt.Errorf("upstream api error: %s: %s", resp.Status, strings.TrimSpace(string(b)))
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}
	return b, resp.Header.Get("Content-Type"), nil
}

func (c *Client) UploadFile(ctx context.Context, path string, file []byte, filename string) (any, error) {
	base := strings.TrimRight(c.BaseURL, "/") + "/" + strings.TrimLeft(path, "/")
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, err := writer.CreateFormFile("file", filepath.Base(filename))
	if err != nil {
		return nil, err
	}
	if _, err := part.Write(file); err != nil {
		return nil, err
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, base, &buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	var payload any
	if err := c.DoJSON(req, &payload); err != nil {
		return nil, err
	}
	return payload, nil
}
