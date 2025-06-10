package service

import (
	"context"
	"time"

	pb "lumenslate/internal/proto/ai_service" // ✅ Replace with your actual go.mod module path

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// DialGRPC establishes a connection to the gRPC server.
func DialGRPC() (pb.AIServiceClient, *grpc.ClientConn, error) {
	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}
	client := pb.NewAIServiceClient(conn)
	return client, conn, nil
}

// --- GenerateContext ---
func GenerateContext(question string, keywords []string, language string) (string, error) {
	client, conn, err := DialGRPC()
	if err != nil {
		return "", err
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
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

// --- DetectVariables ---
func DetectVariables(question string) ([]*pb.DetectedVariable, error) {
	client, conn, err := DialGRPC()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &pb.VariableDetectorRequest{Question: question}
	res, err := client.DetectVariables(ctx, req)
	if err != nil {
		return nil, err
	}

	return res.GetVariables(), nil
}

// --- SegmentQuestion ---
func SegmentQuestion(question string) (string, error) {
	client, conn, err := DialGRPC()
	if err != nil {
		return "", err
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &pb.QuestionSegmentationRequest{Question: question}
	res, err := client.SegmentQuestion(ctx, req)
	if err != nil {
		return "", err
	}

	return res.GetSegmentedQuestion(), nil
}

// --- GenerateMCQVariations ---
func GenerateMCQVariations(question string, options []string, answerIndex int32) ([]*pb.MCQQuestion, error) {
	client, conn, err := DialGRPC()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
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

// --- GenerateMSQVariations ---
func GenerateMSQVariations(question string, options []string, answerIndices []int32) ([]*pb.MSQQuestion, error) {
	client, conn, err := DialGRPC()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
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

// --- FilterAndRandomize ---
func FilterAndRandomize(question string, userPrompt string) ([]*pb.RandomizedVariable, error) {
	client, conn, err := DialGRPC()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
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
