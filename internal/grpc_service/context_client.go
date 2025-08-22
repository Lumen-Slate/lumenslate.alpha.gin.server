package service

import (
	"context"
	"log"
	pb "lumenslate/internal/proto/ai_service"
	"time"
)

func GenerateContext(question string, keywords []string, language string) (string, error) {
	log.Printf("[GenerateContext] Dialing gRPC server...")
	client, conn, err := DialGRPC()
	if err != nil {
		log.Printf("[GenerateContext] Failed to dial gRPC: %v", err)
		return "", err
	}
	defer func() {
		log.Printf("[GenerateContext] Closing gRPC connection")
		conn.Close()
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	log.Printf("[GenerateContext] Preparing request: question=%q, keywords=%v, language=%q", question, keywords, language)
	req := &pb.GenerateContextRequest{
		Question: question,
		Keywords: keywords,
		Language: language,
	}

	log.Printf("[GenerateContext] Sending request to gRPC service")
	res, err := client.GenerateContext(ctx, req)
	if err != nil {
		log.Printf("[GenerateContext] Error from gRPC service: %v", err)
		return "", err
	}

	log.Printf("[GenerateContext] Received response: content=%q", res.GetContent())
	return res.GetContent(), nil
}
