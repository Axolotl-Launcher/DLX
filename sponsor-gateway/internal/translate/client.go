package translate

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Request struct {
	Text       string `json:"text"`
	SourceLang string `json:"source_lang"`
	TargetLang string `json:"target_lang"`
}
type Response struct {
	Translations []Translation `json:"translations"`
}
type Translation struct {
	Text                   string `json:"text"`
	DetectedSourceLanguage string `json:"detected_source_language,omitempty"`
}
type Client struct {
	BaseURL, InternalToken string
	HTTP                   *http.Client
}

type ErrorKind string

const (
	ErrorTimeout         ErrorKind = "timeout"
	ErrorBusy            ErrorKind = "busy"
	ErrorUnavailable     ErrorKind = "unavailable"
	ErrorInvalidResponse ErrorKind = "invalid_response"
)

type UpstreamError struct {
	Kind   ErrorKind
	Status int
}

func (e *UpstreamError) Error() string { return fmt.Sprintf("translation upstream %s", e.Kind) }

func (c *Client) Translate(ctx context.Context, payload Request) (Response, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return Response{}, err
	}
	client := c.HTTP
	if client == nil {
		client = &http.Client{Timeout: 25 * time.Second}
	}
	// DLX /translate is the non-session endpoint protected by the internal TOKEN.
	endpoint := strings.TrimRight(c.BaseURL, "/") + "/translate"
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return Response{}, err
	}
	request.Header.Set("Content-Type", "application/json")
	if c.InternalToken != "" {
		request.Header.Set("Authorization", "Bearer "+c.InternalToken)
	}
	response, err := client.Do(request)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return Response{}, &UpstreamError{Kind: ErrorTimeout}
		}
		return Response{}, &UpstreamError{Kind: ErrorUnavailable}
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusRequestTimeout || response.StatusCode == http.StatusGatewayTimeout {
		return Response{}, &UpstreamError{Kind: ErrorTimeout, Status: response.StatusCode}
	}
	if response.StatusCode == http.StatusTooManyRequests || response.StatusCode >= 500 {
		io.Copy(io.Discard, io.LimitReader(response.Body, 4096))
		return Response{}, &UpstreamError{Kind: ErrorBusy, Status: response.StatusCode}
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		io.Copy(io.Discard, io.LimitReader(response.Body, 4096))
		return Response{}, &UpstreamError{Kind: ErrorUnavailable, Status: response.StatusCode}
	}
	var raw struct {
		Translations []Translation `json:"translations"`
		Data         string        `json:"data"`
	}
	if err = json.NewDecoder(io.LimitReader(response.Body, 2<<20)).Decode(&raw); err != nil {
		return Response{}, &UpstreamError{Kind: ErrorInvalidResponse}
	}
	if len(raw.Translations) == 0 && raw.Data != "" {
		raw.Translations = []Translation{{Text: raw.Data, DetectedSourceLanguage: payload.SourceLang}}
	}
	if len(raw.Translations) == 0 {
		return Response{}, &UpstreamError{Kind: ErrorInvalidResponse}
	}
	return Response{Translations: raw.Translations}, nil
}

func (c *Client) Health(ctx context.Context) error {
	client := c.HTTP
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(c.BaseURL, "/")+"/", nil)
	if err != nil {
		return err
	}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	io.Copy(io.Discard, io.LimitReader(response.Body, 4096))
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return &UpstreamError{Kind: ErrorUnavailable, Status: response.StatusCode}
	}
	return nil
}
