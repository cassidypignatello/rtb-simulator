// Package httpclient provides a high-performance HTTP client for DSP communication.
// It uses fasthttp for connection pooling and sonic for fast JSON serialization.
package httpclient

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/bytedance/sonic"
	"github.com/valyala/fasthttp"

	"github.com/cass/rtb-simulator/pkg/openrtb"
)

// Client is a high-performance HTTP client for OpenRTB bid requests.
type Client struct {
	client          *fasthttp.Client
	timeout         time.Duration
	maxConnsPerHost int
	maxIdleConns    int
}

// Option configures the client.
type Option func(*Client)

// WithTimeout sets the request timeout.
func WithTimeout(d time.Duration) Option {
	return func(c *Client) {
		c.timeout = d
	}
}

// WithMaxConnsPerHost sets the maximum connections per host.
func WithMaxConnsPerHost(n int) Option {
	return func(c *Client) {
		c.maxConnsPerHost = n
	}
}

// WithMaxIdleConns sets the maximum idle connections.
func WithMaxIdleConns(n int) Option {
	return func(c *Client) {
		c.maxIdleConns = n
	}
}

// New creates a new HTTP client with the given options.
func New(opts ...Option) *Client {
	c := &Client{
		timeout:         100 * time.Millisecond,
		maxConnsPerHost: 100,
		maxIdleConns:    100,
	}

	for _, opt := range opts {
		opt(c)
	}

	c.client = &fasthttp.Client{
		MaxConnsPerHost:               c.maxConnsPerHost,
		MaxIdleConnDuration:           30 * time.Second,
		ReadTimeout:                   c.timeout,
		WriteTimeout:                  c.timeout,
		MaxConnWaitTimeout:            c.timeout,
		DisableHeaderNamesNormalizing: true, // Skip header normalization for performance
		DisablePathNormalizing:        true, // Skip path normalization for performance
		MaxResponseBodySize:           64 * 1024, // Limit to 64KB for RTB responses
	}

	return c
}

// Post sends a bid request and returns the response.
func (c *Client) Post(url string, req *openrtb.BidRequest) (*openrtb.BidResponse, error) {
	body, err := sonic.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	request := fasthttp.AcquireRequest()
	response := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(request)
	defer fasthttp.ReleaseResponse(response)

	request.SetRequestURI(url)
	request.Header.SetMethod(fasthttp.MethodPost)
	request.Header.SetContentType("application/json")
	request.SetBody(body)

	err = c.client.DoTimeout(request, response, c.timeout)
	if err != nil {
		if errors.Is(err, fasthttp.ErrTimeout) {
			return nil, &TimeoutError{err: err}
		}
		return nil, fmt.Errorf("do request: %w", err)
	}

	statusCode := response.StatusCode()

	// 204 No Content = no bid
	if statusCode == http.StatusNoContent {
		return &openrtb.BidResponse{ID: req.ID}, nil
	}

	if statusCode >= 400 {
		return nil, fmt.Errorf("server error: status %d", statusCode)
	}

	var resp openrtb.BidResponse
	if err := sonic.Unmarshal(response.Body(), &resp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return &resp, nil
}

// Close releases resources held by the client.
func (c *Client) Close() {
	// fasthttp.Client doesn't require explicit close
}

// TimeoutError indicates a request timeout.
type TimeoutError struct {
	err error
}

func (e *TimeoutError) Error() string {
	return fmt.Sprintf("request timeout: %v", e.err)
}

func (e *TimeoutError) Unwrap() error {
	return e.err
}

// IsTimeout returns true if the error is a timeout error.
func IsTimeout(err error) bool {
	var te *TimeoutError
	return errors.As(err, &te)
}
