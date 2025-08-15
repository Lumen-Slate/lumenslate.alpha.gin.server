# Implementation Plan

- [x] 1. Update Document model with async processing fields





  - Add Status, UpdatedAt, and ErrorMsg fields to Document struct
  - Update BSON and JSON tags for proper serialization
  - Set default values for new fields in constructor
  - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.5_

- [x] 2. Enhance Document Repository with status management methods





  - [x] 2.1 Implement CreateDocument method with proper error handling


    - Write CreateDocument function that inserts document with default "pending" status
    - Add proper context handling and MongoDB error wrapping
    - Include automatic timestamp setting for CreatedAt and UpdatedAt
    - _Requirements: 5.1, 4.4_



  - [x] 2.2 Implement UpdateStatus method for atomic status updates






    - Write UpdateStatus function that updates status, errorMsg, and updatedAt atomically
    - Use MongoDB's $set operator for atomic field updates
    - Include proper error handling and document existence validation


    - _Requirements: 5.2, 5.4_

  - [x] 2.3 Implement UpdateFields method for flexible document updates

    - Write UpdateFields function that accepts bson.M for flexible field updates
    - Automatically update updatedAt timestamp on any field change
    - Add validation to prevent updating immutable fields like fileId
    - _Requirements: 5.3, 5.5_


- [x] 3. Enhance GCS Service with async processing support methods




  - [x] 3.1 Implement UploadFileWithCustomName method


    - Write function to upload files with specified object names instead of auto-generated names
    - Support custom content-type and original filename metadata
    - Return file size and handle upload errors properly
    - _Requirements: 6.1_

  - [x] 3.2 Implement RenameObject method for atomic file moves


    - Write function to atomically move GCS objects from temporary to final names
    - Use GCS copy-and-delete pattern for rename operation
    - Handle errors gracefully and provide descriptive error messages
    - _Requirements: 6.2, 6.4_


  - [x] 3.3 Implement DeleteObject method for cleanup operations

    - Write function to delete GCS objects with proper error handling
    - Support both temporary and final object deletion
    - Log deletion operations for audit trail
    - _Requirements: 6.3, 6.5_

- [x] 4. Create background task implementation for document processing





  - [x] 4.1 Define task constants and payload structures


    - Define TypeAddDocumentToCorpus constant for task identification
    - Create DocumentTaskPayload struct with all required fields
    - Add JSON serialization tags for proper payload marshaling
    - _Requirements: 8.3_

  - [x] 4.2 Implement NewAddDocumentToCorpusTask function


    - Write function to create Asynq tasks with proper payload serialization
    - Add error handling for JSON marshaling failures
    - Set appropriate task options (retry policy, timeout)
    - _Requirements: 8.3_

  - [x] 4.3 Implement HandleAddDocumentToCorpusTask function


    - Write main task handler that processes document ingestion
    - Add document to Vertex AI RAG corpus using existing addVertexAICorpusDocument
    - Rename GCS object from temporary to final name on success
    - Update MongoDB status to "completed" when all steps succeed
    - _Requirements: 2.2, 2.3, 2.4_

  - [x] 4.4 Add comprehensive error handling and rollback logic


    - Implement cleanup on RAG ingestion failure (delete GCS file, update status to "failed")
    - Handle partial failures gracefully (log warnings, continue where possible)
    - Add detailed error logging with context for debugging
    - _Requirements: 2.5, 2.6, 7.1, 7.2, 7.3, 7.4, 7.5_

- [x] 5. Create Asynq server configuration and setup





  - [x] 5.1 Implement AsynqServer struct and constructor


    - Create AsynqServer struct with server and mux fields
    - Write NewAsynqServer constructor with Redis configuration
    - Set up configurable concurrency and retry policies
    - _Requirements: 8.1, 8.2_


  - [x] 5.2 Implement task handler registration

    - Write RegisterHandlers method to map task types to handler functions
    - Register HandleAddDocumentToCorpusTask for TypeAddDocumentToCorpus
    - Add proper error handling for registration failures
    - _Requirements: 8.3_

  - [x] 5.3 Implement server lifecycle management


    - Write Start method to begin processing background tasks
    - Write Stop method for graceful server shutdown
    - Add proper error handling and logging for server operations
    - _Requirements: 8.4, 8.5_

- [x] 6. Refactor AddCorpusDocumentHandler for async processing





  - [x] 6.1 Update handler to use temporary file upload workflow


    - Modify file upload to use temporary object names with UUID prefixes
    - Remove synchronous RAG ingestion from main handler flow
    - Keep existing file validation and GCS upload logic
    - _Requirements: 1.2, 1.3_



  - [x] 6.2 Integrate document metadata storage with pending status





    - Create Document struct with status="pending" and current timestamp
    - Use enhanced CreateDocument repository method for database storage
    - Handle database errors and provide appropriate error responses


    - _Requirements: 1.3, 4.4, 4.5_

  - [x] 6.3 Add Asynq task enqueuing to handler





    - Initialize Asynq client for task enqueuing


    - Create and enqueue DocumentTaskPayload with all required fields
    - Handle task enqueue errors and provide fallback responses
    - _Requirements: 1.4, 8.3_

  - [x] 6.4 Update response format for immediate feedback





    - Return HTTP 200 with fileId and status="pending" immediately
    - Remove long-running RAG processing from response path
    - Add response time logging to verify performance improvement
    - _Requirements: 1.1, 1.5_

- [x] 7. Implement document status checking endpoint




  - [x] 7.1 Create GetDocumentStatusHandler function


    - Write new handler for GET /ai/rag-agent/document-status/:fileId
    - Extract fileId from URL parameters with validation
    - Query MongoDB using GetDocumentByFileID repository method
    - _Requirements: 3.1, 3.3_

  - [x] 7.2 Format status response with comprehensive information


    - Return JSON response with fileId, status, errorMsg, and updatedAt
    - Handle document not found cases with HTTP 404 responses
    - Add proper error handling for database query failures
    - _Requirements: 3.2, 3.4_

- [x] 8. Add new route registration for status endpoint





  - Update ai_routes.go to include GET /ai/rag-agent/document-status/:fileId route
  - Map route to GetDocumentStatusHandler function
  - Ensure proper route ordering and parameter handling
  - _Requirements: 3.1_

- [x] 9. Update main application to start Asynq server





  - Initialize AsynqServer in main.go or appropriate startup file
  - Start background task processing alongside web server
  - Add graceful shutdown handling for both web and task servers
  - _Requirements: 8.4_

- [x] 10. Add comprehensive error handling and logging





  - [x] 10.1 Implement structured logging throughout async workflow


    - Add correlation IDs to track requests through async processing
    - Log task lifecycle events (enqueue, start, complete, fail)
    - Include performance metrics in log entries
    - _Requirements: 7.4, 7.5_


  - [x] 10.2 Add monitoring and alerting capabilities

    - Implement metrics collection for success/failure rates
    - Add queue depth monitoring and processing lag metrics
    - Create health check endpoints for background processing
    - _Requirements: 7.4, 7.5_