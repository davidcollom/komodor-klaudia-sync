package klaudia

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/hashicorp/go-retryablehttp"
)

type apiErrorEnvelope struct {
	Error struct {
		Code      int    `json:"code"`
		Message   string `json:"message"`
		RequestID string `json:"requestID"`
	} `json:"error"`
}

type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

func NewClient(baseURL, apiKey string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{}
	}
	if baseURL == "" {
		baseURL = DefaultAPIBaseURL
	}
	return &Client{baseURL: strings.TrimRight(baseURL, "/"), apiKey: apiKey, httpClient: httpClient}
}

func (c *Client) ListFiles(ctx context.Context, fileType string) ([]RemoteFile, error) {
	var response ListFilesResponse
	if err := c.doJSON(ctx, http.MethodGet, path.Join("/api/v2/klaudia/files", fileType), nil, &response); err != nil {
		return nil, err
	}
	return response.Files, nil
}

func (c *Client) DownloadFile(ctx context.Context, fileType, fileID string) ([]byte, error) {
	requestPath := path.Join("/api/v2/klaudia/files", fileType, fileID)
	req, err := c.newRequest(ctx, http.MethodGet, requestPath, nil, "")
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, formatAPIError(methodLabel(req), requestPath, resp.StatusCode, resp.Status, body)
	}
	return io.ReadAll(resp.Body)
}

func (c *Client) UploadFile(ctx context.Context, fileType, filename string, content []byte) error {
	return c.uploadOrUpdate(ctx, http.MethodPost, path.Join("/api/v2/klaudia/files", fileType), "files", filename, content)
}

func (c *Client) UpdateFile(ctx context.Context, fileType, fileID, filename string, content []byte) error {
	return c.uploadOrUpdate(ctx, http.MethodPut, path.Join("/api/v2/klaudia/files", fileType, fileID), "file", filename, content)
}

func (c *Client) DeleteFiles(ctx context.Context, fileType string, ids []string) error {
	body := map[string][]string{"fileIDs": ids}
	return c.doJSON(ctx, http.MethodDelete, path.Join("/api/v2/klaudia/files", fileType), body, nil)
}

func (c *Client) uploadOrUpdate(ctx context.Context, method, requestPath, fieldName, filename string, content []byte) error {
	var buffer bytes.Buffer
	writer := multipart.NewWriter(&buffer)
	part, err := writer.CreateFormFile(fieldName, filename)
	if err != nil {
		return err
	}
	if _, err := part.Write(content); err != nil {
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}

	req, err := c.newRequest(ctx, method, requestPath, &buffer, writer.FormDataContentType())
	if err != nil {
		return err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return formatAPIError(method, requestPath, resp.StatusCode, resp.Status, body)
	}
	return nil
}

func (c *Client) doJSON(ctx context.Context, method, requestPath string, body any, out any) error {
	var reader io.Reader
	contentType := "application/json"
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(payload)
	}
	req, err := c.newRequest(ctx, method, requestPath, reader, contentType)
	if err != nil {
		return err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode >= 400 {
		responseBody, _ := io.ReadAll(resp.Body)
		return formatAPIError(method, requestPath, resp.StatusCode, resp.Status, responseBody)
	}
	if out == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

func (c *Client) newRequest(ctx context.Context, method, requestPath string, body io.Reader, contentType string) (*http.Request, error) {
	requestURL := c.baseURL + requestPath
	req, err := http.NewRequestWithContext(ctx, method, requestURL, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-API-KEY", c.apiKey)
	req.Header.Set("User-Agent", "klaudia-sync-action/1.0")
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	return req, nil
}

func formatAPIError(method, requestPath string, statusCode int, status string, body []byte) error {
	trimmedBody := strings.TrimSpace(string(body))
	var envelope apiErrorEnvelope
	if err := json.Unmarshal(body, &envelope); err == nil && envelope.Error.Message != "" {
		if envelope.Error.RequestID != "" {
			return fmt.Errorf("api %s %s failed with %s (%d): %s (request_id=%s)", method, requestPath, status, statusCode, envelope.Error.Message, envelope.Error.RequestID)
		}
		return fmt.Errorf("api %s %s failed with %s (%d): %s", method, requestPath, status, statusCode, envelope.Error.Message)
	}
	if trimmedBody == "" {
		return fmt.Errorf("api %s %s failed with %s (%d)", method, requestPath, status, statusCode)
	}
	return fmt.Errorf("api %s %s failed with %s (%d): %s", method, requestPath, status, statusCode, trimmedBody)
}

func methodLabel(req *http.Request) string {
	if req == nil {
		return http.MethodGet
	}
	return req.Method
}

func NewRetryableHTTPClient(logger Logger, logLevel string) *http.Client {
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = 3
	retryClient.RetryWaitMin = 250 * time.Millisecond
	retryClient.RetryWaitMax = 2 * time.Second
	retryClient.Logger = nil
	retryClient.CheckRetry = retryablehttp.ErrorPropagatedRetryPolicy
	retryClient.Backoff = retryablehttp.DefaultBackoff
	retryClient.ResponseLogHook = func(log retryablehttp.Logger, resp *http.Response) {
		if logger == nil || resp == nil || logLevel != "debug" {
			return
		}
		if resp.StatusCode >= 500 || resp.StatusCode == http.StatusTooManyRequests {
			logger.Warnf("Retryable response from %s %s: %s", resp.Request.Method, resp.Request.URL.Path, resp.Status)
		}
	}
	retryClient.ErrorHandler = retryablehttp.PassthroughErrorHandler
	return retryClient.StandardClient()
}
