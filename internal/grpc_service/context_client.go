package service

import (
	"context"
	pb "lumenslate/internal/proto/ai_service"
	"time"
)

func GenerateContext(question string, keywords []string, language string) (string, error) {
	client, conn, err := DialGRPC()
	if err != nil {
		return "", err
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req := &pb.GenerateContextRequest{
		Question: question,
		Keywords: keywords,
		Language: language,
	}

	res, err := client.GenerateContext(ctx, req)
	if err != nil {
		return "", err
	}

	return res.GetContent(), nil
}
