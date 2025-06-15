package service

import (
	"context"
	"time"
	pb "lumenslate/internal/proto/ai_service"
)

func FilterAndRandomize(question string, userPrompt string) ([]*pb.RandomizedVariable, error) {
	client, conn, err := DialGRPC()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req := &pb.FilterAndRandomizerRequest{
		Question:   question,
		UserPrompt: userPrompt,
	}
	res, err := client.FilterAndRandomize(ctx, req)
	if err != nil {
		return nil, err
	}

	return res.GetVariables(), nil
}
