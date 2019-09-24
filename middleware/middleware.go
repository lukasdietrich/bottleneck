package middleware

import "github.com/lukasdietrich/bottleneck"

// MiddlewareFunc is the standard function signature for middleware using the base context type.
type MiddlewareFunc func(*bottleneck.Context, bottleneck.Next) error
