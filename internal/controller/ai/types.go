package ai

import "mime/multipart"

// --- Request Structs for AI Operations ---

type GenerateContextRequest struct {
	Question string   `json:"question"`
	Keywords []string `json:"keywords"`
	Language string   `json:"language"`
}

type DetectVariablesRequest struct {
	Question string `json:"question"`
}

type SegmentQuestionRequest struct {
	Question string `json:"question"`
}

type GenerateMCQVariationsRequest struct {
	Question    string   `json:"question"`
	Options     []string `json:"options"`
	AnswerIndex int32    `json:"answerIndex"`
}

type GenerateMSQVariationsRequest struct {
	Question      string   `json:"question"`
	Options       []string `json:"options"`
	AnswerIndices []int32  `json:"answerIndices"`
}

type FilterAndRandomizeRequest struct {
	Question   string `json:"question"`
	UserPrompt string `json:"userPrompt"`
}

type CreateCorpusRequest struct {
	CorpusName string `json:"corpusName" binding:"required"`
}

type DeleteCorpusDocumentRequest struct {
	CorpusName string `json:"corpusName" binding:"required"`
	FileID     string `json:"fileId" binding:"required"` // Can be fileId, RAG file ID, or display name
}

type AddCorpusDocumentFormRequest struct {
	CorpusName string                `form:"corpusName" binding:"required"`
	File       *multipart.FileHeader `form:"file" binding:"required"`
}

type ViewDocumentRequest struct {
	DocumentID string `uri:"id" binding:"required"`
}

type AgentRequest struct {
	TeacherId string                `form:"teacherId" binding:"required"`
	Role      string                `form:"role" binding:"required"`
	Message   string                `form:"message" binding:"required"`
	File      *multipart.FileHeader `form:"file"`
	FileType  string                `form:"fileType"`
	CreatedAt string                `form:"createdAt"`
	UpdatedAt string                `form:"updatedAt"`
}

type RAGAgentRequest struct {
	CorpusName string `json:"corpusName" binding:"required"`
	Role       string `json:"role" binding:"required"`
	Message    string `json:"message" binding:"required"`
	File       string `json:"file"`
	CreatedAt  string `json:"createdAt"`
	UpdatedAt  string `json:"updatedAt"`
}

// --- Response Types ---

type AIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

type CorpusResponse struct {
	CorpusName string                 `json:"corpusName"`
	Status     string                 `json:"status"`
	Message    string                 `json:"message"`
	Details    map[string]interface{} `json:"details,omitempty"`
}

type DocumentResponse struct {
	DocumentID  string `json:"documentId"`
	DisplayName string `json:"displayName"`
	CorpusName  string `json:"corpusName"`
	Status      string `json:"status"`
	SignedURL   string `json:"signedUrl,omitempty"`
	ContentType string `json:"contentType,omitempty"`
	DownloadURL string `json:"downloadUrl,omitempty"`
}
