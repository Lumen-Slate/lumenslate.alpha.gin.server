package service

import (
	"context"
	pb "lumenslate/internal/proto/ai_service"
	"time"
)

func Agent(file, fileType, userId, role, message, createdAt, updatedAt string) (map[string]interface{}, error) {
	client, conn, err := DialGRPC()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req := &pb.AgentRequest{
		File:      file,
		FileType:  fileType,
		UserId:    userId,
		Role:      role,
		Message:   message,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
	res, err := client.Agent(ctx, req)
	if err != nil {
		return nil, err
	}

	// Return as a map for generic JSON response
	return map[string]interface{}{
		"message":        res.GetMessage(),
		"user_id":        res.GetUserId(),
		"agent_name":     res.GetAgentName(),
		"agent_response": res.GetAgentResponse(),
		"session_id":     res.GetSessionId(),
		"createdAt":      res.GetCreatedAt(),
		"updatedAt":      res.GetUpdatedAt(),
		"response_time":  res.GetResponseTime(),
		"role":           res.GetRole(),
		"feedback":       res.GetFeedback(),
	}, nil
}
