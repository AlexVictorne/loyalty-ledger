package accrual

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testCase struct {
	name          string
	statusCode    int
	body          any
	retryAfter    string
	expectErr     bool
	expectTooMany bool
}

func TestHTTPClient_GetOrderInfo(t *testing.T) {
	tests := []testCase{
		{
			name:       "200 OK",
			statusCode: http.StatusOK,
			body:       OrderInfo{Order: "123", Status: "PROCESSED", Accrual: floatPtr(10.5)},
		},
		{
			name:       "204 No Content",
			statusCode: http.StatusNoContent,
			body:       nil,
			expectErr:  true,
		},
		{
			name:          "429 Too Many Requests",
			statusCode:    http.StatusTooManyRequests,
			body:          nil,
			retryAfter:    "3",
			expectErr:     true,
			expectTooMany: true,
		},
		{
			name:       "500 Internal Server Error",
			statusCode: http.StatusInternalServerError,
			body:       "fail",
			expectErr:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mux := http.NewServeMux()
			mux.HandleFunc("/api/orders/", func(w http.ResponseWriter, r *http.Request) {
				if tc.retryAfter != "" {
					w.Header().Set("Retry-After", tc.retryAfter)
				}
				w.WriteHeader(tc.statusCode)
				if tc.body != nil {
					switch b := tc.body.(type) {
					case string:
						fmt.Fprint(w, b)
					default:
						_ = json.NewEncoder(w).Encode(b)
					}
				}
			})
			ts := httptest.NewServer(mux)
			defer ts.Close()

			client := NewHTTPClient(ts.URL)
			ctx := context.Background()
			info, err := client.GetOrderInfo(ctx, "123")

			if tc.expectErr {
				assert.Error(t, err)
				if tc.expectTooMany {
					var tooMany *TooManyRequestsError
					require.ErrorAs(t, err, &tooMany)
					exp, _ := strconv.Atoi(tc.retryAfter)
					assert.Equal(t, exp, tooMany.RetryAfter)
				}
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, "123", info.Order)
			assert.Equal(t, "PROCESSED", info.Status)
			assert.NotNil(t, info.Accrual)
			assert.Equal(t, 10.5, *info.Accrual)
		})
	}
}

func floatPtr(f float64) *float64 {
	return &f
}
