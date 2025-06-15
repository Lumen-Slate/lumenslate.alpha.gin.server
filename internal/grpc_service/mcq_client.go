package service

import (
	"context"
	"time"
	pb "lumenslate/internal/proto/ai_service"
)

func GenerateMCQVariations(question string, options []string, answerIndex int32) ([]*pb.MCQQuestion, error) {
	client, conn, err := DialGRPC()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req := &pb.MCQRequest{
		Question:    question,
		Options:     options,
		AnswerIndex: answerIndex,
	}
	res, err := client.GenerateMCQVariations(ctx, req)
	if err != nil {
		return nil, err
	}

	return res.GetVariations(), nil
}
