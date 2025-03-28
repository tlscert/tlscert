package middleware

import (
	"context"
	"log"

	"google.golang.org/grpc"
)

func LoggingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	log.Printf("INFO: method=%s", info.FullMethod)

	resp, err := handler(ctx, req)
	if err != nil {
		log.Printf("ERROR: method=%s error=%v", info.FullMethod, err)
		return nil, err
	}
	return resp, nil
}
