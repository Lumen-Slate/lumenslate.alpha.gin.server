package service

import (
	"context"
	"crypto/tls"
	"time"

	pb "lumenslate/internal/proto/ai_service"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func DialGRPC() (pb.AIServiceClient, *grpc.ClientConn, error) {
	target := "lumenslate-microservice-756147067348.asia-south1.run.app:443"

	// TLS credentials for secure connection
	creds := credentials.NewTLS(&tls.Config{})

	// Context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Dial the Cloud Run gRPC service using HTTPS/HTTP2
	conn, err := grpc.DialContext(ctx, target, grpc.WithTransportCredentials(creds))
	if err != nil {
		return nil, nil, err
	}

	client := pb.NewAIServiceClient(conn)
	return client, conn, nil
}
