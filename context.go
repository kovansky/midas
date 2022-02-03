package midas

import "context"

// contextKey represents an internal key for adding context fields.
// This practice is considered best as it prevents other packages
// from interfering with our context keys.
type contextKey int

const (
	userConfigContextKey = contextKey(iota + 1)
)

// NewContextWithSiteConfig returns a new context with given config.
func NewContextWithSiteConfig(ctx context.Context, site Site) context.Context {
	return context.WithValue(ctx, userConfigContextKey, site)
}

// SiteConfigFromContext returns current config from context.
func SiteConfigFromContext(ctx context.Context) Site {
	config, _ := ctx.Value(userConfigContextKey).(Site)
	return config
}
