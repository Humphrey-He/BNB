package rpcgateway

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/asset-platform/multi-chain-asset-platform/internal/repository"
	"github.com/ethereum/go-ethereum/ethclient"
	gethrpc "github.com/ethereum/go-ethereum/rpc"
)

// CallOptions configures retry and timeout behavior for resilient RPC calls.
type CallOptions struct {
	Timeout    time.Duration
	MaxRetries int
}

// DefaultCallOptions returns conservative defaults for chain RPC access.
func DefaultCallOptions() CallOptions {
	return CallOptions{
		Timeout:    12 * time.Second,
		MaxRetries: 2,
	}
}

// ResilientClient executes RPC calls against multiple providers with timeout,
// retry, provider failover, and circuit-breaker awareness.
type ResilientClient struct {
	manager     *ProviderManager
	callOptions CallOptions
	httpClient  *http.Client
}

// NewResilientClient builds a resilient client from providers loaded in DB.
func NewResilientClient(chainID int64, repo repository.RPCProviderRepository, opts CallOptions) (*ResilientClient, error) {
	if opts.Timeout <= 0 {
		opts.Timeout = DefaultCallOptions().Timeout
	}
	if opts.MaxRetries < 0 {
		opts.MaxRetries = 0
	}

	factory := NewProviderManagerFactory(repo, DefaultProviderConfig())
	manager, err := factory.GetManager(context.Background(), chainID)
	if err != nil {
		return nil, err
	}

	return &ResilientClient{
		manager:     manager,
		callOptions: opts,
		httpClient: &http.Client{
			Timeout: opts.Timeout,
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
				DialContext: (&net.Dialer{
					Timeout:   5 * time.Second,
					KeepAlive: 30 * time.Second,
				}).DialContext,
				TLSHandshakeTimeout:   5 * time.Second,
				ResponseHeaderTimeout: opts.Timeout,
				IdleConnTimeout:       90 * time.Second,
				MaxIdleConns:          20,
			},
		},
	}, nil
}

// HealthSnapshot captures current provider health for diagnostics.
type HealthSnapshot struct {
	ProviderID    int64     `json:"provider_id"`
	ProviderName  string    `json:"provider_name"`
	URL           string    `json:"url"`
	IsActive      bool      `json:"is_active"`
	CircuitState  string    `json:"circuit_state"`
	FailureCount  int       `json:"failure_count"`
	LastError     string    `json:"last_error,omitempty"`
	LastCheckedAt time.Time `json:"last_checked_at,omitempty"`
}

// InspectProviders returns the current provider state for health endpoints.
func (c *ResilientClient) InspectProviders() []HealthSnapshot {
	providers := c.manager.GetProviders()
	snapshots := make([]HealthSnapshot, 0, len(providers))
	for _, p := range providers {
		lastErr := ""
		if err := p.GetLastError(); err != nil {
			lastErr = err.Error()
		}
		snapshots = append(snapshots, HealthSnapshot{
			ProviderID:    p.ID,
			ProviderName:  p.Name,
			URL:           p.URL,
			IsActive:      p.IsActive,
			CircuitState:  p.circuitBreaker.State().String(),
			FailureCount:  p.GetFailureCount(),
			LastError:     lastErr,
			LastCheckedAt: time.Now(),
		})
	}
	return snapshots
}

// Call executes fn against providers until one succeeds or all candidates fail.
func (c *ResilientClient) Call(ctx context.Context, fn func(context.Context, *ethclient.Client) error) error {
	var lastErr error
	used := make(map[int64]struct{})
	attempts := len(c.manager.GetProviders())
	if attempts == 0 {
		return ErrNoAvailableProvider
	}

	for i := 0; i < attempts; i++ {
		provider, err := c.selectNextProvider(used)
		if err != nil {
			if lastErr != nil {
				return fmt.Errorf("%w: %v", ErrAllProvidersFailed, lastErr)
			}
			return err
		}
		used[provider.ID] = struct{}{}

		err = c.callProvider(ctx, provider, fn)
		if err == nil {
			return nil
		}
		lastErr = err
	}

	if lastErr == nil {
		lastErr = ErrNoAvailableProvider
	}
	return fmt.Errorf("%w: %v", ErrAllProvidersFailed, lastErr)
}

func (c *ResilientClient) selectNextProvider(excluded map[int64]struct{}) (*Provider, error) {
	providers := c.manager.GetProviders()
	for range providers {
		p, err := c.manager.SelectProvider()
		if err != nil {
			return nil, err
		}
		if _, ok := excluded[p.ID]; ok {
			continue
		}
		return p, nil
	}
	return nil, ErrNoAvailableProvider
}

func (c *ResilientClient) callProvider(
	ctx context.Context,
	provider *Provider,
	fn func(context.Context, *ethclient.Client) error,
) error {
	var lastErr error
	for attempt := 0; attempt <= c.callOptions.MaxRetries; attempt++ {
		callCtx, cancel := context.WithTimeout(ctx, c.callOptions.Timeout)
		client, err := c.dialProvider(provider)
		if err != nil {
			cancel()
			provider.RecordFailure(err)
			return NewProviderError(provider.ID, provider.Name, err)
		}

		err = fn(callCtx, client)
		cancel()
		client.Close()

		if err == nil {
			provider.RecordSuccess()
			return nil
		}

		wrapped := classifyRPCError(err)
		provider.RecordFailure(wrapped)
		lastErr = NewProviderError(provider.ID, provider.Name, wrapped)

		if !IsRetryable(wrapped) {
			break
		}
	}
	return lastErr
}

func (c *ResilientClient) dialProvider(provider *Provider) (*ethclient.Client, error) {
	raw, err := gethrpc.DialOptions(
		context.Background(),
		provider.URL,
		gethrpc.WithHTTPClient(c.httpClient),
	)
	if err != nil {
		return nil, err
	}
	return ethclient.NewClient(raw), nil
}

func classifyRPCError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, context.DeadlineExceeded) || strings.Contains(strings.ToLower(err.Error()), "timeout") {
		return &RetryableError{Err: ErrRequestTimeout}
	}

	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return &RetryableError{Err: ErrRequestTimeout}
	}

	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		if urlErr.Timeout() {
			return &RetryableError{Err: ErrRequestTimeout}
		}
		if strings.Contains(strings.ToLower(urlErr.Err.Error()), "tls") ||
			strings.Contains(strings.ToLower(urlErr.Err.Error()), "connection refused") ||
			strings.Contains(strings.ToLower(urlErr.Err.Error()), "no such host") {
			return &RetryableError{Err: err}
		}
	}

	lower := strings.ToLower(err.Error())
	if strings.Contains(lower, "connection refused") ||
		strings.Contains(lower, "tls handshake timeout") ||
		strings.Contains(lower, "eof") ||
		strings.Contains(lower, "503") ||
		strings.Contains(lower, "502") {
		return &RetryableError{Err: err}
	}

	return err
}
