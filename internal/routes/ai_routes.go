package routes

import (
	"lumenslate/internal/controller"

	"github.com/gin-gonic/gin"
)

func RegisterAIRoutes(r *gin.Engine) {
	ai := r.Group("/ai")
	{
		// Core AI services
		ai.POST("/generate-context", controller.GenerateContextHandler)
		ai.POST("/detect-variables", controller.DetectVariablesHandler)
		ai.POST("/segment-question", controller.SegmentQuestionHandler)
		ai.POST("/generate-mcq", controller.GenerateMCQVariationsHandler)
		ai.POST("/generate-msq", controller.GenerateMSQVariationsHandler)
		ai.POST("/filter-and-randomize", controller.FilterAndRandomizeHandler)
		ai.POST("/agent", controller.AgentHandler)
		ai.POST("/rag-agent", controller.RAGAgentHandler)

		// RAG Corpus management
		ai.POST("/rag-agent/create-corpus", controller.CreateCorpusHandler)
		ai.POST("/rag-agent/list-corpus-content", controller.ListCorpusContentHandler)
		ai.POST("/rag-agent/list-all-corpora", controller.ListAllCorporaHandler)
		ai.POST("/rag-agent/add-corpus-document", controller.AddCorpusDocumentHandler)
		ai.POST("/rag-agent/delete-corpus-document", controller.DeleteCorpusDocumentHandler)
		ai.GET("/rag-agent/:corpusName/documents", controller.ListCorpusDocumentsHandler)

		// Individual document management
		ai.GET("/documents/view/:id", controller.ViewDocumentHandler)
		ai.DELETE("/documents/:id", controller.DeleteCorpusDocumentByIDHandler)
	}
}
