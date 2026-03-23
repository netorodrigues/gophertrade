package interceptors

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type contextKey string

const UserIDKey contextKey = "user_id"

// UnaryServerMockAuth interceptor to extract user-id from metadata
func UnaryServerMockAuth() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		userID := "anonymous"
		if ok {
			if values := md.Get("user-id"); len(values) > 0 {
				userID = values[0]
			}
		}

		ctx = context.WithValue(ctx, UserIDKey, userID)
		return handler(ctx, req)
	}
}
