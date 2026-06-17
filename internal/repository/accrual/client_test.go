package accrual

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/paulwwyvern/gophermart/internal/model/errs"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestNewClientURL(t *testing.T) {
	tests := []struct {
		name    string
		baseUrl string
		want    string
	}{
		{
			name:    "Test #1",
			baseUrl: "http://example.com",
			want:    "http://example.com/api/orders/",
		},
		{
			name:    "Test #2",
			baseUrl: "http://example.com/",
			want:    "http://example.com/api/orders/",
		},
		{
			name:    "Test #3",
			baseUrl: "     http://example.com  ",
			want:    "http://example.com/api/orders/",
		},
		{
			name:    "Test #4",
			baseUrl: "lecom",
			want:    "lecom/api/orders/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(tt.baseUrl)
			assert.Equal(t, tt.want, client.baseUrl)
		})
	}
}

func TestClient_GetOrderStatus_OK(t *testing.T) {
	client := NewClient("http://example.com")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		assert.Equal(t, http.MethodGet, r.Method)

		w.Header().Set("Content-Type", "application/json")

		w.WriteHeader(http.StatusOK)

		w.Write([]byte(`{"order":"12345678903","status": "PROCESSED","accrual": 314.271}`))
	}))
	defer server.Close()
	client.client = server.Client()

	status, accrual, err := client.GetOrderStatus(context.Background(), "12345678903")

	assert.Equal(t, "PROCESSED", status)
	assert.Equal(t, decimal.NewFromFloat(314.271), accrual)
	assert.NoError(t, err)
}

func TestClient_GetOrderStatus_NoContent(t *testing.T) {
	client := NewClient("http://example.com")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)

		w.WriteHeader(http.StatusNoContent)

	}))
	defer server.Close()
	client.client = server.Client()

	status, accrual, err := client.GetOrderStatus(context.Background(), "12345678903")

	assert.Equal(t, "", status)
	assert.Equal(t, decimal.Decimal{}, accrual)
	assert.ErrorIs(t, errs.ErrAccrualNotRegistered, err)
}

func TestClient_GetOrderStatus_InternalError(t *testing.T) {
	client := NewClient("http://example.com")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)

		w.WriteHeader(http.StatusInternalServerError)

	}))
	defer server.Close()
	client.client = server.Client()

	status, accrual, err := client.GetOrderStatus(context.Background(), "12345678903")

	assert.Equal(t, "", status)
	assert.Equal(t, decimal.Decimal{}, accrual)
	assert.Error(t, err)
}

func TestClient_GetOrderStatus_TooManyRequests(t *testing.T) {
	client := NewClient("http://example.com")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("Retry-After", "1234")

		w.WriteHeader(http.StatusTooManyRequests)

		w.Write([]byte("too many requests"))

	}))
	defer server.Close()
	client.client = server.Client()

	status, accrual, err := client.GetOrderStatus(context.Background(), "12345678903")

	assert.Equal(t, "", status)
	assert.Equal(t, decimal.Decimal{}, accrual)

	var errTooManyRequests *errs.ErrAccrualTooManyRequests
	assert.ErrorAs(t, err, &errTooManyRequests)

	assert.Equal(t, errTooManyRequests.RetryAfter, 1234)
	assert.Equal(t, errTooManyRequests.Body, "too many requests")
}
