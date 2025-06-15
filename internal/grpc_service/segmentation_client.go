package service

import (
	"context"
	"time"
	pb "lumenslate/internal/proto/ai_service"
)

func SegmentQuestion(question string) (string, error) {
	client, conn, err := DialGRPC()
	if err != nil {
		return "", err
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req := &pb.QuestionSegmentationRequest{Question: question}
	res, err := client.SegmentQuestion(ctx, req)
	if err != nil {
		return "", err
	}

	return res.GetSegmentedQuestion(), nil
}
