package rpcgateway

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/asset-platform/multi-chain-asset-platform/internal/repository"
)

// Provider represents an RPC provider with management capabilities
type Provider struct {
	ID       int64
	ChainID  int64
	Name     string
	URL      string
	Weight   int
	IsActive bool
	// Internal state
	mu             sync.RWMutex
	failures       int
	lastError      error
	lastChecked    time.Time
	circuitBreaker *CircuitBreaker
	cooldown       time.Duration
	maxFailures    int
}

// ProviderConfig holds configuration for a provider
type ProviderConfig struct {
	MaxFailures     int
	CircuitCooldown time.Duration
}

// DefaultProviderConfig returns default provider configuration
func DefaultProviderConfig() ProviderConfig {
	return ProviderConfig{
		MaxFailures:     5,
		CircuitCooldown: 30 * time.Second,
	}
}

// NewProvider creates a new Provider from a repository model
func NewProvider(rp *repository.RPCProvider, config ProviderConfig) *Provider {
	cb := NewCircuitBreaker(rp.Name, config.MaxFailures, config.CircuitCooldown)
	return &Provider{
		ID:             rp.ID,
		ChainID:        rp.ChainID,
		Name:           rp.Name,
		URL:            rp.URL,
		Weight:         rp.Weight,
		IsActive:       rp.IsActive,
		circuitBreaker: cb,
		cooldown:       config.CircuitCooldown,
		maxFailures:    config.MaxFailures,
	}
}

// IsHealthy returns true if the provider is healthy (circuit closed)
func (p *Provider) IsHealthy() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.circuitBreaker.State() == StateClosed && p.IsActive
}

// RecordFailure records a failure for this provider
func (p *Provider) RecordFailure(err error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.failures++
	p.lastError = err
	p.lastChecked = time.Now()
	p.circuitBreaker.RecordFailure()
}

// RecordSuccess records a success for this provider
func (p *Provider) RecordSuccess() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.failures = 0
	p.lastError = nil
	p.lastChecked = time.Now()
	p.circuitBreaker.RecordSuccess()
}

// AllowRequest checks if requests can be made to this provider
func (p *Provider) AllowRequest() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.circuitBreaker.Allow() && p.IsActive
}

// GetFailureCount returns the current failure count
func (p *Provider) GetFailureCount() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.failures
}

// GetLastError returns the last error encountered
func (p *Provider) GetLastError() error {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.lastError
}

// ProviderManager manages multiple RPC providers
type ProviderManager struct {
	chainID     int64
	providers   []*Provider
	weights     []int // Cumulative weights for weighted selection
	totalWeight int
	mu          sync.RWMutex
}

// NewProviderManager creates a new ProviderManager
func NewProviderManager(chainID int64, providers []*Provider) *ProviderManager {
	pm := &ProviderManager{
		chainID:     chainID,
		providers:   providers,
		totalWeight: 0,
	}

	// Calculate cumulative weights
	pm.weights = make([]int, len(providers))
	for i, p := range providers {
		pm.totalWeight += p.Weight
		pm.weights[i] = pm.totalWeight
	}

	return pm
}

// SelectProvider selects a provider based on weighted random selection
func (pm *ProviderManager) SelectProvider() (*Provider, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	if len(pm.providers) == 0 {
		return nil, ErrNoAvailableProvider
	}

	// Filter active and healthy providers
	var candidates []*Provider
	candidateWeights := make([]int, 0, len(pm.providers))
	totalWeight := 0

	for _, p := range pm.providers {
		if p.IsActive && p.AllowRequest() {
			candidates = append(candidates, p)
			totalWeight += p.Weight
			candidateWeights = append(candidateWeights, totalWeight)
		}
	}

	if len(candidates) == 0 {
		return nil, ErrNoAvailableProvider
	}

	if len(candidates) == 1 {
		return candidates[0], nil
	}

	// Weighted random selection: pick random in [0, totalWeight) and find first candidate where cumulative > random
	random := rand.Intn(totalWeight)
	for i, weight := range candidateWeights {
		if random < weight {
			return candidates[i], nil
		}
	}

	// Fallback (shouldn't reach here)
	return candidates[0], nil
}

// GetProviders returns all providers
func (pm *ProviderManager) GetProviders() []*Provider {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.providers
}

// RefreshProviders reloads providers from the repository
func (pm *ProviderManager) RefreshProviders(providers []*Provider) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.providers = providers
	pm.totalWeight = 0
	pm.weights = make([]int, len(providers))
	for i, p := range providers {
		pm.totalWeight += p.Weight
		pm.weights[i] = pm.totalWeight
	}
}

// ProviderManagerFactory creates ProviderManager instances
type ProviderManagerFactory struct {
	repo   repository.RPCProviderRepository
	config ProviderConfig
}

// NewProviderManagerFactory creates a new factory
func NewProviderManagerFactory(repo repository.RPCProviderRepository, config ProviderConfig) *ProviderManagerFactory {
	return &ProviderManagerFactory{
		repo:   repo,
		config: config,
	}
}

// GetManager gets or creates a ProviderManager for a chain
func (f *ProviderManagerFactory) GetManager(ctx context.Context, chainID int64) (*ProviderManager, error) {
	providers, err := f.repo.GetActiveByChainID(chainID)
	if err != nil {
		return nil, fmt.Errorf("failed to load providers for chain %d: %w", chainID, err)
	}

	if len(providers) == 0 {
		return nil, ErrNoAvailableProvider
	}

	// Convert to internal provider type
	internalProviders := make([]*Provider, len(providers))
	for i, rp := range providers {
		internalProviders[i] = NewProvider(rp, f.config)
	}

	return NewProviderManager(chainID, internalProviders), nil
}
