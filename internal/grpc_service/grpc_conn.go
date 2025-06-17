package service

import (
	"crypto/tls"
	"log"
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
	log.Printf("[gRPC] Target address: %s", target)

	var conn *grpc.ClientConn
	var err error

	// For local development (localhost), use insecure connection
	if target == "localhost:50051" {
		log.Println("[gRPC] Using insecure credentials for localhost")
		conn, err = grpc.NewClient(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		log.Println("[gRPC] Using TLS credentials for remote connection")
		creds := credentials.NewTLS(&tls.Config{})
		conn, err = grpc.NewClient(target, grpc.WithTransportCredentials(creds))
	}

	if err != nil {
		log.Printf("[gRPC] Failed to connect to %s: %v", target, err)
		return nil, nil, err
	}

	log.Printf("[gRPC] Successfully connected to %s", target)
	client := pb.NewAIServiceClient(conn)
	return client, conn, nil
}
