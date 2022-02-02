package midas

import "context"

// contextKey represents an internal key for adding context fields.
// This practice is considered best as it prevents other packages
// from interfering with our context keys.
type contextKey int

const (
	userConfigContextKey = contextKey(iota + 1)
)

// NewContextWithUserConfig returns a new context with given config.
func NewContextWithUserConfig(ctx context.Context, config interface{}) context.Context {
	return context.WithValue(ctx, userConfigContextKey, config)
}

// UserConfigFromContext returns current config from context.
func UserConfigFromContext(ctx context.Context) interface{} {
	config, _ := ctx.Value(userConfigContextKey).(interface{})
	return config
}
