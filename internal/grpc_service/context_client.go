package service

import (
	"context"
	"log"
	pb "lumenslate/internal/proto/ai_service"
	"time"
)

func GenerateContext(question string, keywords []string, language string) (string, error) {
	log.Printf("[gRPC] GenerateContext called with Question='%s', Keywords=%v, Language='%s'", question, keywords, language)
	client, conn, err := DialGRPC()
	if err != nil {
		log.Printf("[gRPC] DialGRPC failed: %v", err)
		return "", err
	}
	defer func() {
		log.Println("[gRPC] Closing gRPC connection")
		conn.Close()
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req := &pb.GenerateContextRequest{
		Question: question,
		Keywords: keywords,
		Language: language,
	}
	log.Printf("[gRPC] Sending GenerateContextRequest: %+v", req)

	res, err := client.GenerateContext(ctx, req)
	if err != nil {
		log.Printf("[gRPC] GenerateContext RPC error: %v", err)
		return "", err
	}

	log.Printf("[gRPC] GenerateContext RPC success, content length: %d", len(res.GetContent()))
	return res.GetContent(), nil
}
