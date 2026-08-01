package openai

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestGenerateReturnsQuotaErrorForCreditBalanceExhausted(t *testing.T) {
	provider := &OpenAIProvider{
		apiKey: "test-key",
		model:  "gpt-test",
		client: &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if req.URL.String() != "https://api.openai.com/v1/chat/completions" {
				t.Fatalf("unexpected URL: %s", req.URL.String())
			}
			return &http.Response{
				StatusCode: http.StatusTooManyRequests,
				Body:       io.NopCloser(strings.NewReader(`{"error":{"code":"credit_balance_exhausted"}}`)),
				Header:     make(http.Header),
			}, nil
		})},
	}

	_, err := provider.Generate("test prompt")
	if err == nil {
		t.Fatal("expected an error")
	}

	var quotaErr *QuotaError
	if !errors.As(err, &quotaErr) {
		t.Fatalf("expected QuotaError, got %T", err)
	}
	if quotaErr.Code != "credit_balance_exhausted" {
		t.Fatalf("expected credit_balance_exhausted, got %q", quotaErr.Code)
	}
}
