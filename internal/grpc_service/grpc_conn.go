package service

import (
	"crypto/tls"
	"os"

	pb "lumenslate/internal/proto/ai_service"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

func DialGRPC() (pb.AIServiceClient, *grpc.ClientConn, error) {
	// Use environment variable or default to Cloud Run service
	target := os.Getenv("GRPC_SERVICE_URL")
	if target == "" {
		target = "lumenslate-microservice-756147067348.asia-south1.run.app:443"
	}

	var conn *grpc.ClientConn
	var err error

	// For local development (localhost), use insecure connection
	if target == "localhost:50051" {
		conn, err = grpc.NewClient(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		// TLS credentials for secure connection to Cloud Run
		creds := credentials.NewTLS(&tls.Config{})
		conn, err = grpc.NewClient(target, grpc.WithTransportCredentials(creds))
	}

	if err != nil {
		return nil, nil, err
	}

	client := pb.NewAIServiceClient(conn)
	return client, conn, nil
}
