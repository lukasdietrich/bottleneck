package bottleneck

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
)

var (
	// ErrInvalidContext indicates the use of a context that does not match the required type.
	ErrInvalidContext = errors.New("invalid context type")
)

var (
	baseContextType    = reflect.TypeOf(Context{})
	baseContextPtrType = reflect.PtrTo(baseContextType)
)

// A contextHolder is a container to hold a request context.
// The context is always a struct, that embeds the bottleneck.Context.
type contextHolder struct {
	contextType      reflect.Type  // TypeOf(&CustomContext{})
	contextValue     reflect.Value // ValueOf(&CustomContext{})
	baseContext      *Context
	baseContextValue reflect.Value // ValueOf(&Context{})
}

// unwrap returns either the custom context or the base context depending on the requested type.
func (h *contextHolder) unwrap(targetType reflect.Type) reflect.Value {
	switch targetType {
	case baseContextPtrType:
		return h.baseContextValue

	case h.contextType:
		return h.contextValue

	default:
		panic(fmt.Errorf("%w: %v instead of %v", ErrInvalidContext, targetType, h.contextType))
	}
}

// A contextCreator is a factory for new instances of a custom context.
type contextCreator struct {
	contextType    reflect.Type // TypeOf(Context{})
	contextPtrType reflect.Type // TypeOf(&Context{})
}

// newContextCreator creates a new contextCreator for a given instance of a custom context.
// The custom context must be a struct that embeds the base context. If the provided instance is not a valid custom
// context, the creation panics.
func newContextCreator(v interface{}) *contextCreator {
	t := reflect.TypeOf(v)

	if err := assertKind(reflect.Struct, t); err != nil {
		panic(err)
	}

	embeddedContext, ok := t.FieldByName("Context")
	if ok && embeddedContext.Anonymous {
		if err := assertType(baseContextType, embeddedContext.Type); err != nil {
			panic(err)
		}

		return &contextCreator{
			contextType:    t,
			contextPtrType: reflect.PtrTo(t),
		}
	}

	panic(ErrInvalidContext)
}

// create constructs a new contextHolder for given http handler parameters. The new contextHolder is then used to unwrap
// into handler functions.
func (c *contextCreator) create(res http.ResponseWriter, req *http.Request, params map[string]string) *contextHolder {
	var (
		contextValue = reflect.New(c.contextType)
		baseContext  = contextValue.Elem().FieldByName("Context").Addr().Interface().(*Context)
	)

	baseContext.request = req
	baseContext.response = res
	baseContext.params = params

	return &contextHolder{
		contextType:      c.contextPtrType,
		contextValue:     contextValue,
		baseContext:      baseContext,
		baseContextValue: reflect.ValueOf(baseContext),
	}
}

func (c *contextCreator) validateTarget(targetType reflect.Type) error {
	if targetType != baseContextPtrType && targetType != c.contextPtrType {
		return fmt.Errorf("%v is not a valid context type. must be %v or %v",
			targetType, baseContextPtrType, c.contextPtrType)
	}

	return nil
}

type wrappedHandler func(*contextHolder, *http.Request) error

func validateHandler(c *contextCreator, t reflect.Type) error {
	if err := assertKind(reflect.Func, t); err != nil {
		return errors.New("handler must be a func")
	}

	if t.NumIn() < 1 || t.NumIn() > 2 {
		return errors.New("handler must have at least one and at most two arguments")
	}

	if err := c.validateTarget(t.In(0)); err != nil {
		return err
	}

	if t.NumIn() > 1 {
		if err := assertKind(reflect.Ptr, t.In(1)); err != nil {
			return err
		}

		if err := assertKind(reflect.Struct, t.In(1).Elem()); err != nil {
			return err
		}
	}

	if t.NumOut() != 1 || t.Out(0) != reflect.TypeOf((*error)(nil)).Elem() {
		return errors.New("handler must return exactly one value of type error")
	}

	return nil
}

func wrapHandler(router *Router, handler Handler) wrappedHandler {
	var (
		handlerType  = reflect.TypeOf(handler)
		handlerValue = reflect.ValueOf(handler)

		payloadType reflect.Type
	)

	if err := validateHandler(router.contextCreator, handlerType); err != nil {
		panic(err)
	}

	if handlerType.NumIn() > 1 {
		payloadType = handlerType.In(1).Elem()
	}

	return func(ctx *contextHolder, req *http.Request) error {
		input := make([]reflect.Value, 0, 2)
		input = append(input, ctx.unwrap(handlerType.In(0)))

		if payloadType != nil {
			var (
				payloadValue     = reflect.New(payloadType)
				payloadInterface = payloadValue.Interface()
			)

			if err := router.Binder.Bind(req, payloadInterface); err != nil {
				return err
			}

			if err := router.Validator.Validate(req, payloadInterface); err != nil {
				return err
			}

			input = append(input, payloadValue)
		}

		if output := handlerValue.Call(input)[0]; !output.IsNil() {
			return output.Interface().(error)
		}

		return nil
	}
}

type wrappedMiddleware func(*contextHolder, Next) error

func validateMiddleware(c *contextCreator, t reflect.Type) error {
	if err := assertKind(reflect.Func, t); err != nil {
		return errors.New("middleware must be a func")
	}

	if t.NumIn() != 2 {
		return errors.New("middleware must have exactly two arguments")
	}

	if err := c.validateTarget(t.In(0)); err != nil {
		return err
	}

	if err := assertType(reflect.TypeOf((Next)(nil)), t.In(1)); err != nil {
		return err
	}

	if t.NumOut() != 1 || t.Out(0) != reflect.TypeOf((*error)(nil)).Elem() {
		return errors.New("middleware must return exactly one value of type error")
	}

	return nil
}

func wrapMiddleware(router *Router, middleware Middleware) wrappedMiddleware {
	var (
		middlewareValue = reflect.ValueOf(middleware)
		middlewareType  = reflect.TypeOf(middleware)
	)

	if err := validateMiddleware(router.contextCreator, middlewareType); err != nil {
		panic(err)
	}

	return func(ctx *contextHolder, next Next) error {
		input := []reflect.Value{ctx.unwrap(middlewareType.In(0)), reflect.ValueOf(next)}

		if output := middlewareValue.Call(input)[0]; !output.IsNil() {
			return output.Interface().(error)
		}

		return nil
	}
}

func wrapMiddlewareList(router *Router, middlewareList []Middleware) []wrappedMiddleware {
	wrapped := make([]wrappedMiddleware, len(middlewareList))

	for i := 0; i < len(wrapped); i++ {
		wrapped[i] = wrapMiddleware(router, middlewareList[i])
	}

	return wrapped
}

func assertType(expect, actual reflect.Type) error {
	if expect != actual {
		return fmt.Errorf("expected type (%s) but got (%s)", expect, actual)
	}

	return nil
}

func assertKind(expect reflect.Kind, actual reflect.Type) error {
	if expect != actual.Kind() {
		return fmt.Errorf("expected type of kind (%s) but got (%s)", expect, actual)
	}

	return nil
}
