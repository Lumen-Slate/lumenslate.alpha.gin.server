package service

import (
	pb "lumenslate/internal/proto/ai_service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func DialGRPC() (pb.AIServiceClient, *grpc.ClientConn, error) {
	conn, err := grpc.Dial("0.0.0.0:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}
	client := pb.NewAIServiceClient(conn)
	return client, conn, nil
}
