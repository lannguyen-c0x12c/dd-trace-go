// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2023 Datadog, Inc.

package sharedsec

import (
	"context"
	"errors"
	"reflect"

	"github.com/lannguyen-c0x12c/dd-trace-go/internal/appsec/dyngo"
	"github.com/lannguyen-c0x12c/dd-trace-go/internal/appsec/dyngo/instrumentation"
	"github.com/lannguyen-c0x12c/dd-trace-go/internal/log"
)

type (
	// UserIDOperation type representing a call to appsec.SetUser(). It gets both created and destroyed in a single
	// call to ExecuteUserIDOperation
	UserIDOperation struct {
		dyngo.Operation
		Error error
	}
	// UserIDOperationArgs is the user ID operation arguments.
	UserIDOperationArgs struct {
		UserID string
	}
	// UserIDOperationRes is the user ID operation results.
	UserIDOperationRes struct{}

	// OnUserIDOperationStart function type, called when a user ID
	// operation starts.
	OnUserIDOperationStart func(operation *UserIDOperation, args UserIDOperationArgs)

	// UserMonitoringError wraps an error interface to decorate it with additional appsec data, if needed
	UserMonitoringError struct {
		error
	}
)

// NewUserMonitoringError creates a new user monitoring error that returns `msg` upon calling `Error()`
func NewUserMonitoringError(msg string) *UserMonitoringError {
	return &UserMonitoringError{
		errors.New(msg),
	}
}

var userIDOperationArgsType = reflect.TypeOf((*UserIDOperationArgs)(nil)).Elem()

// ExecuteUserIDOperation starts and finishes the UserID operation by emitting a dyngo start and finish events
// An error is returned if the user associated to that operation must be blocked
func ExecuteUserIDOperation(parent dyngo.Operation, args UserIDOperationArgs) error {
	op := &UserIDOperation{Operation: dyngo.NewOperation(parent)}
	dyngo.StartOperation(op, args)
	dyngo.FinishOperation(op, UserIDOperationRes{})
	return op.Error
}

// ListenedType returns the type a OnUserIDOperationStart event listener
// listens to, which is the UserIDOperationStartArgs type.
func (OnUserIDOperationStart) ListenedType() reflect.Type { return userIDOperationArgsType }

// Call the underlying event listener function by performing the type-assertion
// on v whose type is the one returned by ListenedType().
func (f OnUserIDOperationStart) Call(op dyngo.Operation, v interface{}) {
	f(op.(*UserIDOperation), v.(UserIDOperationArgs))
}

// MonitorUser starts and finishes a UserID operation.
// A call to the WAF is made to check the user ID and an error is returned if the
// user should be blocked. The return value is nil otherwise.
func MonitorUser(ctx context.Context, userID string) error {
	if parent, ok := ctx.Value(instrumentation.ContextKey{}).(dyngo.Operation); ok {
		return ExecuteUserIDOperation(parent, UserIDOperationArgs{UserID: userID})
	}
	log.Error("appsec: user ID monitoring ignored: could not find the http handler instrumentation metadata in the request context: the request handler is not being monitored by a middleware function or the provided context is not the expected request context")
	return nil

}
