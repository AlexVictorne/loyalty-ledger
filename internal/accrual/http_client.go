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
)

type HTTPClient struct {
	addr      string
	client    *http.Client
	userAgent string
}

func NewHTTPClient(addr string) *HTTPClient {
	return &HTTPClient{
		addr:      addr,
		client:    &http.Client{Timeout: 10 * time.Second},
		userAgent: "gophermart-accrual-client/1.0",
	}
}

var (
	ErrTooManyRequests = errors.New("accrual: too many requests (429)")
	ErrNotFound        = errors.New("accrual: order not found (204)")
)

func (c *HTTPClient) GetOrderInfo(ctx context.Context, orderNumber string) (OrderInfo, error) {
	url := fmt.Sprintf("%s/api/orders/%s", strings.TrimRight(c.addr, "/"), orderNumber)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return OrderInfo{}, err
	}
	req.Header.Set("User-Agent", c.userAgent)
	resp, err := c.client.Do(req)
	if err != nil {
		return OrderInfo{}, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var info OrderInfo
		if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
			return OrderInfo{}, err
		}
		return info, nil
	case http.StatusNoContent:
		return OrderInfo{}, ErrNotFound
	case http.StatusTooManyRequests:
		retryAfter := 0
		if v := resp.Header.Get("Retry-After"); v != "" {
			if n, err := strconv.Atoi(v); err == nil {
				retryAfter = n
			}
		}
		return OrderInfo{}, &TooManyRequestsError{RetryAfter: retryAfter}
	default:
		b, _ := io.ReadAll(resp.Body)
		return OrderInfo{}, fmt.Errorf("accrual: unexpected status %d: %s", resp.StatusCode, string(b))
	}
}

type TooManyRequestsError struct {
	RetryAfter int // секунд
}

func (e *TooManyRequestsError) Error() string {
	return fmt.Sprintf("accrual: too many requests, retry after %d seconds", e.RetryAfter)
}
