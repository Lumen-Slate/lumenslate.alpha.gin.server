package routes

import (
	"lumenslate/internal/controller/ai"

	"github.com/gin-gonic/gin"
)

func RegisterAIRoutes(r *gin.Engine) {
	aiGroup := r.Group("/ai")
	{
		// Question-related AI services (from question_controller.go)
		aiGroup.POST("/generate-context", ai.GenerateContextHandler)
		aiGroup.POST("/detect-variables", ai.DetectVariablesHandler)
		aiGroup.POST("/segment-question", ai.SegmentQuestionHandler)
		aiGroup.POST("/generate-mcq", ai.GenerateMCQVariationsHandler)
		aiGroup.POST("/generate-msq", ai.GenerateMSQVariationsHandler)
		aiGroup.POST("/filter-and-randomize", ai.FilterAndRandomizeHandler)

		// Agent services (from agent_controller.go)
		aiGroup.POST("/agent", ai.AgentHandler)

		// RAG agent services (from rag_controller.go)
		aiGroup.POST("/rag-agent", ai.RAGAgentHandler)

		// Corpus management (from corpus_controller.go and rag_controller.go)
		aiGroup.POST("/rag-agent/create-corpus", ai.CreateCorpusHandler)
		aiGroup.POST("/rag-agent/list-corpus-content", ai.ListCorpusContentHandler)
		aiGroup.POST("/rag-agent/list-all-corpora", ai.ListAllCorporaHandler)

		// Document management (from document_controller.go)
		aiGroup.POST("/rag-agent/add-corpus-document", ai.AddCorpusDocumentHandler)
		aiGroup.POST("/rag-agent/delete-corpus-document", ai.DeleteCorpusDocumentHandler)
		aiGroup.GET("/rag-agent/:corpusName/documents", ai.ListCorpusDocumentsHandler)
		aiGroup.GET("/documents/view/:id", ai.ViewDocumentHandler)
		aiGroup.DELETE("/documents/:id", ai.DeleteCorpusDocumentByIDHandler)
	}
}
