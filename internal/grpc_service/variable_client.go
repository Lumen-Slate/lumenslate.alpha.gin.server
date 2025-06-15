package service

import (
	"context"
	"time"
	pb "lumenslate/internal/proto/ai_service"
)

func DetectVariables(question string) ([]*pb.DetectedVariable, error) {
	client, conn, err := DialGRPC()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req := &pb.VariableDetectorRequest{Question: question}
	res, err := client.DetectVariables(ctx, req)
	if err != nil {
		return nil, err
	}

	return res.GetVariables(), nil
}
