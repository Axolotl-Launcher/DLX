package afdian

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	UserID, Token, BaseURL string
	HTTP                   *http.Client
	Now                    func() time.Time
}
type orderResponse struct {
	EC   int             `json:"ec"`
	Data json.RawMessage `json:"data"`
}
type remoteOrder struct {
	OutTradeNo  string          `json:"out_trade_no"`
	UserID      string          `json:"user_id"`
	TotalAmount json.RawMessage `json:"total_amount"`
	Status      json.RawMessage `json:"status"`
}

type orderPage struct {
	List []remoteOrder `json:"list"`
}

func (c *Client) VerifyOrder(ctx context.Context, outTradeNo string) (VerifiedOrder, error) {
	if strings.TrimSpace(outTradeNo) == "" || c.UserID == "" || c.Token == "" {
		return VerifiedOrder{}, ErrUnavailable
	}
	params, err := json.Marshal(map[string]string{"out_trade_no": outTradeNo})
	if err != nil {
		return VerifiedOrder{}, err
	}
	now := time.Now
	if c.Now != nil {
		now = c.Now
	}
	ts := now().Unix()
	signSource := c.Token + "params" + string(params) + "ts" + strconv.FormatInt(ts, 10) + "user_id" + c.UserID
	sum := md5.Sum([]byte(signSource))
	values := url.Values{"user_id": {c.UserID}, "params": {string(params)}, "ts": {strconv.FormatInt(ts, 10)}, "sign": {hex.EncodeToString(sum[:])}}
	base := c.BaseURL
	if base == "" {
		base = "https://afdian.com/api/open"
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(base, "/")+"/query-order", strings.NewReader(values.Encode()))
	if err != nil {
		return VerifiedOrder{}, err
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := c.HTTP
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}
	response, err := client.Do(request)
	if err != nil {
		return VerifiedOrder{}, err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return VerifiedOrder{}, ErrUnavailable
	}
	var body orderResponse
	if err = json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(&body); err != nil {
		return VerifiedOrder{}, err
	}
	if body.EC != 200 {
		return VerifiedOrder{}, ErrUnavailable
	}
	var order remoteOrder
	if err = json.Unmarshal(body.Data, &order); err != nil {
		return VerifiedOrder{}, err
	}
	// Afdian's query-order endpoint returns data.list, while older compatible
	// fixtures may return the order object directly. Accept both shapes, but
	// select only the exact requested order from the provider response.
	if order.OutTradeNo == "" {
		var page orderPage
		if err = json.Unmarshal(body.Data, &page); err != nil {
			return VerifiedOrder{}, err
		}
		for _, candidate := range page.List {
			if candidate.OutTradeNo == outTradeNo {
				order = candidate
				break
			}
		}
	}
	amount, err := fen(order.TotalAmount)
	if err != nil {
		return VerifiedOrder{}, err
	}
	status, err := statusName(order.Status)
	if err != nil {
		return VerifiedOrder{}, err
	}
	if order.OutTradeNo != outTradeNo {
		return VerifiedOrder{}, ErrUnavailable
	}
	return VerifiedOrder{OutTradeNo: order.OutTradeNo, AfdianUserID: order.UserID, ActualPaidFen: amount, Status: status}, nil
}
func fen(raw json.RawMessage) (int64, error) {
	var text string
	if err := json.Unmarshal(raw, &text); err != nil {
		var n float64
		if err = json.Unmarshal(raw, &n); err != nil {
			return 0, err
		}
		text = fmt.Sprintf("%.2f", n)
	}
	parts := strings.Split(strings.TrimSpace(text), ".")
	if len(parts) > 2 {
		return 0, fmt.Errorf("invalid amount")
	}
	whole, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, err
	}
	fraction := int64(0)
	if len(parts) == 2 {
		v := parts[1] + "00"
		fraction, err = strconv.ParseInt(v[:2], 10, 64)
		if err != nil {
			return 0, err
		}
	}
	return whole*100 + fraction, nil
}
func statusName(raw json.RawMessage) (string, error) {
	var n int
	if err := json.Unmarshal(raw, &n); err == nil {
		if n == 2 {
			return "paid", nil
		}
		return "unknown", nil
	}
	var text string
	if err := json.Unmarshal(raw, &text); err != nil {
		return "", err
	}
	if text == "2" || text == "paid" || text == "success" {
		return "paid", nil
	}
	return "unknown", nil
}
