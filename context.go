package hierr

import (
	"fmt"
)

// ErrorContext is a element of key-value linked list of error contexts.
type ErrorContext struct {
	Key      string
	Value    interface{}
	Previous *ErrorContext
}

// Context adds new key-value context pair to current context list and return
// current context list.
func (context ErrorContext) Context(
	key string,
	value interface{},
) *ErrorContext {
	previous := &ErrorContext{
		Key:      key,
		Value:    value,
		Previous: context.Previous,
	}

	context.Previous = previous

	return &context
}

// Errorf produces context-rich hierarchical error, which will include all
// previously declared context key-value pairs.
func (context ErrorContext) Errorf(
	reason Reason,
	message string,
	args ...interface{},
) error {
	return Error{
		Message: fmt.Sprintf(message, args...),
		Reason:  reason,
		Context: &context,
	}
}

// Reason adds current context to the specified error. If error is not
// hierarchical error, it will be converted to such.
func (context ErrorContext) Reason(reason Reason) error {
	if previous, ok := reason.(Error); ok {
		context.Walk(func(key string, value interface{}) {
			previous.Context = previous.Context.Context(key, value)
		})

		return previous
	} else {
		return Error{
			Reason:  reason,
			Context: &context,
		}
	}
}

// Walk iterates over all key-value context pairs and calls specified
// callback for each.
func (context *ErrorContext) Walk(callback func(string, interface{})) {
	if context == nil {
		return
	}

	callback(context.Key, context.Value)

	if context.Previous != nil {
		context.Previous.Walk(callback)
	}
}

// GetKeyValuePairs returns slice of key-value context pairs, which will
// be always even, each even index is key and each odd index is value.
func (context *ErrorContext) GetKeyValuePairs() []interface{} {
	pairs := []interface{}{}

	context.Walk(func(name string, value interface{}) {
		pairs = append(pairs, name, value)
	})

	return pairs
}
