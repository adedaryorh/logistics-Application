package redis

import platformconfig "github.com/adedaryorh/logistics-platform/pkg/config"

func NewStoreFromConfig(cfg *platformconfig.Config) Store {
	if cfg == nil {
		return NewMemoryStore()
	}
	return NewFallbackStore(
		NewClient(Options{
			Addr:     cfg.Redis.Addr,
			Password: cfg.Redis.Password,
			DB:       cfg.Redis.DB,
		}),
		NewMemoryStore(),
	)
}
