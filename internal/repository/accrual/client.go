package accrual

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/paulwwyvern/gophermart/internal/model/errs"
	"github.com/shopspring/decimal"
)

type Client struct {
	baseUrl string
	client  *http.Client
}

func NewClient(baseUrl string) *Client {

	baseUrl = strings.TrimRight(strings.TrimSpace(baseUrl), "/")
	baseUrl = baseUrl + "/api/orders/"

	client := &http.Client{
		Timeout: time.Second * 5,
	}
	return &Client{baseUrl: baseUrl, client: client}
}

type GetOrderStatusResponse struct {
	Order   string          `json:"order"`
	Status  string          `json:"status"`
	Accrual decimal.Decimal `json:"accrual,omitempty"`
}

func (c *Client) GetOrderStatus(ctx context.Context, number string) (string, decimal.Decimal, error) {
	url := c.baseUrl + number

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", decimal.Decimal{}, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return "", decimal.Decimal{}, err
	}

	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()

	if err != nil {
		return "", decimal.Decimal{}, err
	}

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNoContent {
			return "", decimal.Decimal{}, errs.ErrAccrualNotRegistered
		}
		if resp.StatusCode == http.StatusInternalServerError {
			return "", decimal.Decimal{}, errors.New("internal server error")
		}
		if resp.StatusCode == http.StatusTooManyRequests {
			retryAfter, err := strconv.Atoi(resp.Header.Get("Retry-After"))
			if err != nil {
				retryAfter = 0
			}

			retErr := &errs.ErrAccrualTooManyRequests{
				RetryAfter: retryAfter,
				Body:       string(body),
			}

			return "", decimal.Decimal{}, retErr
		}
		return "", decimal.Decimal{}, fmt.Errorf("unknown status code: %d", resp.StatusCode)
	}

	orderResp := GetOrderStatusResponse{}
	err = json.Unmarshal(body, &orderResp)
	if err != nil {
		return "", decimal.Decimal{}, err
	}

	return orderResp.Status, orderResp.Accrual, nil
}
