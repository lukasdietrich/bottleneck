package middleware

import "github.com/lukasdietrich/bottleneck"

// StandardMiddleware is the standard function signature for middleware using the base context type.
type StandardMiddleware func(*bottleneck.Context, bottleneck.Next) error
