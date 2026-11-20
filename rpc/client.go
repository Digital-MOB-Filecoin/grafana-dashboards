package rpc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync/atomic"
)

// Client is the interface for making JSON-RPC calls to a Lotus node.
// Params and result are left as any so callers control serialization.
type Client interface {
	Call(ctx context.Context, method string, params []any, result any) error
}

// HTTPClient implements Client over HTTP using the Lotus JSON-RPC 2.0 wire format.
type HTTPClient struct {
	endpoint string
	token    string
	http     *http.Client
	seq      atomic.Uint64
}

func NewHTTPClient(endpoint, token string) *HTTPClient {
	return &HTTPClient{
		endpoint: endpoint,
		token:    token,
		http:     &http.Client{},
	}
}

type rpcRequest struct {
	JSONRPC string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  []any  `json:"params"`
	ID      uint64 `json:"id"`
}

type rpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      uint64          `json:"id"`
	Result  json.RawMessage `json:"result"`
	Error   *rpcError       `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *rpcError) Error() string {
	return fmt.Sprintf("rpc error %d: %s", e.Code, e.Message)
}

func (c *HTTPClient) Call(ctx context.Context, method string, params []any, result any) error {
	if params == nil {
		params = []any{}
	}
	reqBody, err := json.Marshal(rpcRequest{
		JSONRPC: "2.0",
		Method:  "Filecoin." + method,
		Params:  params,
		ID:      c.seq.Add(1),
	})
	if err != nil {
		return fmt.Errorf("marshal rpc request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(reqBody))
	if err != nil {
		return fmt.Errorf("build http request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.token)
	}

	httpResp, err := c.http.Do(httpReq)
	if err != nil {
		return fmt.Errorf("http call %s: %w", method, err)
	}
	defer httpResp.Body.Close()

	raw, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	var resp rpcResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return fmt.Errorf("unmarshal rpc response: %w", err)
	}
	if resp.Error != nil {
		return resp.Error
	}
	if result != nil {
		if err := json.Unmarshal(resp.Result, result); err != nil {
			return fmt.Errorf("unmarshal result for %s: %w", method, err)
		}
	}
	return nil
}
