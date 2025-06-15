package service

import (
	"context"
	"time"
	pb "lumenslate/internal/proto/ai_service"
)

func GenerateMSQVariations(question string, options []string, answerIndices []int32) ([]*pb.MSQQuestion, error) {
	client, conn, err := DialGRPC()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req := &pb.MSQRequest{
		Question:      question,
		Options:       options,
		AnswerIndices: answerIndices,
	}
	res, err := client.GenerateMSQVariations(ctx, req)
	if err != nil {
		return nil, err
	}

	return res.GetVariations(), nil
}
