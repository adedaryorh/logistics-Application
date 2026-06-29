package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
	"sync"
	"time"
)

type SchemaRegistryClient struct {
	baseURL string
	client  *http.Client
	mu      sync.RWMutex
	cache   map[string]struct{}
}

func NewSchemaRegistryClient(baseURL string) *SchemaRegistryClient {
	return &SchemaRegistryClient{
		baseURL: strings.TrimSpace(baseURL),
		client: &http.Client{
			Timeout: 3 * time.Second,
		},
		cache: map[string]struct{}{},
	}
}

func (c *SchemaRegistryClient) ValidateEvent(ctx context.Context, event Event) error {
	if c == nil || c.baseURL == "" {
		return nil
	}

	cacheKey := fmt.Sprintf("%s:%d", event.Schema, event.SchemaVersion)
	c.mu.RLock()
	_, ok := c.cache[cacheKey]
	c.mu.RUnlock()
	if ok {
		return nil
	}

	base, err := url.Parse(c.baseURL)
	if err != nil {
		return fmt.Errorf("parse schema registry url: %w", err)
	}
	base.Path = path.Join(base.Path, "/subjects/", event.Schema, "/versions/", fmt.Sprintf("%d", event.SchemaVersion))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base.String(), nil)
	if err != nil {
		return fmt.Errorf("build schema registry request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("query schema registry: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("schema registry missing subject %s version %d: status %d", event.Schema, event.SchemaVersion, resp.StatusCode)
	}

	var payload struct {
		Subject string `json:"subject"`
		Version int    `json:"version"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return fmt.Errorf("decode schema registry response: %w", err)
	}
	if payload.Subject != "" && payload.Subject != event.Schema {
		return fmt.Errorf("schema registry returned subject %q for event schema %q", payload.Subject, event.Schema)
	}
	if payload.Version != 0 && payload.Version != event.SchemaVersion {
		return fmt.Errorf("schema registry returned version %d for event schema version %d", payload.Version, event.SchemaVersion)
	}

	c.mu.Lock()
	c.cache[cacheKey] = struct{}{}
	c.mu.Unlock()
	return nil
}
