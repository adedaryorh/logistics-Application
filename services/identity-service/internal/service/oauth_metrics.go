package service

import "sync"

type OAuthProviderMetrics struct {
	Successes        int64 `json:"successes"`
	Failures         int64 `json:"failures"`
	ProfileFallbacks int64 `json:"profile_fallbacks"`
	SessionStarts    int64 `json:"session_starts"`
	RateLimited      int64 `json:"rate_limited"`
}

type OAuthMetricsSnapshot struct {
	Providers       map[string]OAuthProviderMetrics `json:"providers"`
	CleanupRuns     int64                           `json:"cleanup_runs"`
	CleanupDeleted  int64                           `json:"cleanup_deleted"`
	CleanupFailures int64                           `json:"cleanup_failures"`
}

type OAuthRuntimeMetrics struct {
	mu             sync.Mutex
	providers      map[string]*OAuthProviderMetrics
	cleanupRuns    int64
	cleanupDeleted int64
	cleanupFails   int64
}

func NewOAuthRuntimeMetrics() *OAuthRuntimeMetrics {
	return &OAuthRuntimeMetrics{
		providers: map[string]*OAuthProviderMetrics{},
	}
}

func (m *OAuthRuntimeMetrics) provider(name string) *OAuthProviderMetrics {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.providers[name] == nil {
		m.providers[name] = &OAuthProviderMetrics{}
	}
	return m.providers[name]
}

func (m *OAuthRuntimeMetrics) IncSuccess(provider string) {
	entry := m.provider(provider)
	m.mu.Lock()
	entry.Successes++
	m.mu.Unlock()
}

func (m *OAuthRuntimeMetrics) IncFailure(provider string) {
	entry := m.provider(provider)
	m.mu.Lock()
	entry.Failures++
	m.mu.Unlock()
}

func (m *OAuthRuntimeMetrics) IncProfileFallback(provider string) {
	entry := m.provider(provider)
	m.mu.Lock()
	entry.ProfileFallbacks++
	m.mu.Unlock()
}

func (m *OAuthRuntimeMetrics) IncSessionStart(provider string) {
	entry := m.provider(provider)
	m.mu.Lock()
	entry.SessionStarts++
	m.mu.Unlock()
}

func (m *OAuthRuntimeMetrics) IncRateLimited(provider string) {
	entry := m.provider(provider)
	m.mu.Lock()
	entry.RateLimited++
	m.mu.Unlock()
}

func (m *OAuthRuntimeMetrics) AddCleanupResult(deleted int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cleanupRuns++
	m.cleanupDeleted += deleted
}

func (m *OAuthRuntimeMetrics) IncCleanupFailure() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cleanupFails++
}

func (m *OAuthRuntimeMetrics) Snapshot() OAuthMetricsSnapshot {
	m.mu.Lock()
	defer m.mu.Unlock()
	providers := make(map[string]OAuthProviderMetrics, len(m.providers))
	for name, entry := range m.providers {
		providers[name] = *entry
	}
	return OAuthMetricsSnapshot{
		Providers:       providers,
		CleanupRuns:     m.cleanupRuns,
		CleanupDeleted:  m.cleanupDeleted,
		CleanupFailures: m.cleanupFails,
	}
}
