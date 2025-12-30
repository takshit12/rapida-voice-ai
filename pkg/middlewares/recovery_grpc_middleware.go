// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.
package middlewares

import (
	"context"
	"runtime/debug"

	"github.com/rapidaai/pkg/commons"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func NewRecoveryUnaryServerMiddleware(logger commons.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				stackTrace := string(debug.Stack())
				logger.Error("panic recovered in stream",
					zap.Any("panic", r),
					zap.String("method", info.FullMethod),
					zap.String("stackTrace", stackTrace),
				)
				err = status.Errorf(codes.Internal, "internal server error")
			}
		}()

		return handler(ctx, req)
	}
}

// StreamServerInterceptor returns a new unary server interceptors that perform recovery from panics.
// NOTE(bwplotka): For more complex auth interceptor see https://github.com/grpc/grpc-go/blob/master/authz/grpc_authz_server_interceptors.go.
func NewRecoveryStreamServerMiddleware(logger commons.Logger) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler,
	) (err error) {
		defer func() {
			if r := recover(); r != nil {
				stackTrace := string(debug.Stack())
				logger.Error("panic recovered in stream",
					zap.Any("panic", r),
					zap.String("method", info.FullMethod),
					zap.String("stackTrace", stackTrace),
				)
				err = status.Errorf(codes.Internal, "internal server error")
			}
		}()
		return handler(srv, ss)
	}
}
