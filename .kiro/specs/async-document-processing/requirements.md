# Requirements Document

## Introduction

This feature refactors the existing synchronous AddCorpusDocumentHandler into an asynchronous background job system using Asynq. The current implementation causes long response times (3-4 minutes) due to the slow Vertex AI RAG corpus ingestion process. The new asynchronous approach will provide immediate response to users while processing the document ingestion in the background, significantly improving user experience and system responsiveness.

## Requirements

### Requirement 1

**User Story:** As a user uploading documents to the RAG corpus, I want to receive an immediate response confirming my upload request, so that I don't have to wait 3-4 minutes for the slow RAG ingestion process to complete.

#### Acceptance Criteria

1. WHEN a user uploads a document via POST /ai/rag-agent/add-corpus-document THEN the system SHALL return HTTP 200 within 30 seconds with status "pending"
2. WHEN the upload request is processed THEN the system SHALL upload the file to GCS with a temporary name immediately
3. WHEN the file is uploaded to GCS THEN the system SHALL save document metadata to MongoDB with status = "pending"
4. WHEN the metadata is saved THEN the system SHALL enqueue an Asynq background task for RAG processing
5. WHEN the response is returned THEN it SHALL include fileId and status fields

### Requirement 2

**User Story:** As a system administrator, I want document processing to happen asynchronously in the background, so that the web server remains responsive and can handle multiple concurrent requests.

#### Acceptance Criteria

1. WHEN a document upload is initiated THEN the system SHALL create an Asynq task with document processing details
2. WHEN the background worker processes the task THEN it SHALL add the document to the Vertex AI RAG corpus
3. WHEN the RAG ingestion succeeds THEN the system SHALL rename the GCS object from temporary to final name
4. WHEN the processing completes successfully THEN the system SHALL update MongoDB status to "completed"
5. WHEN any step fails THEN the system SHALL update MongoDB status to "failed" with error message
6. WHEN processing fails THEN the system SHALL clean up by deleting the GCS file and any partial RAG ingestion

### Requirement 3

**User Story:** As a user, I want to check the status of my document upload, so that I know when the processing is complete or if there were any errors.

#### Acceptance Criteria

1. WHEN a user requests GET /ai/rag-agent/document-status/:fileId THEN the system SHALL return the current document status from MongoDB
2. WHEN the status is requested THEN the response SHALL include fileId, status, errorMsg (if any), and updatedAt timestamp
3. WHEN the document exists THEN the system SHALL return HTTP 200 with status information
4. WHEN the document doesn't exist THEN the system SHALL return HTTP 404 with appropriate error message

### Requirement 4

**User Story:** As a developer, I want the Document model to track processing status and errors, so that the system can provide accurate status information to users.

#### Acceptance Criteria

1. WHEN the Document model is defined THEN it SHALL include a Status field with values "pending", "completed", or "failed"
2. WHEN the Document model is defined THEN it SHALL include an UpdatedAt timestamp field
3. WHEN the Document model is defined THEN it SHALL include an optional ErrorMsg field for failure details
4. WHEN a document is created THEN the Status SHALL default to "pending"
5. WHEN the status changes THEN the UpdatedAt field SHALL be updated to the current timestamp

### Requirement 5

**User Story:** As a system, I need enhanced repository methods to manage document status updates, so that the background processing can properly track and update document states.

#### Acceptance Criteria

1. WHEN the document repository is implemented THEN it SHALL provide CreateDocument method for initial document creation
2. WHEN the document repository is implemented THEN it SHALL provide UpdateStatus method to change status and error message
3. WHEN the document repository is implemented THEN it SHALL provide UpdateFields method for flexible field updates
4. WHEN UpdateStatus is called THEN it SHALL update both status and updatedAt fields atomically
5. WHEN UpdateFields is called THEN it SHALL update the specified fields and updatedAt timestamp

### Requirement 6

**User Story:** As a system, I need enhanced GCS service methods to support the temporary upload and rename workflow, so that files can be safely processed asynchronously.

#### Acceptance Criteria

1. WHEN the GCS service is implemented THEN it SHALL provide UploadFileWithCustomName method for uploading with specific object names
2. WHEN the GCS service is implemented THEN it SHALL provide RenameObject method for moving objects from temporary to final names
3. WHEN the GCS service is implemented THEN it SHALL provide DeleteObject method for cleanup operations
4. WHEN RenameObject is called THEN it SHALL atomically move the object from source to destination name
5. WHEN any GCS operation fails THEN it SHALL return descriptive error messages for proper error handling

### Requirement 7

**User Story:** As a system administrator, I want proper error handling and rollback mechanisms, so that failed document processing doesn't leave the system in an inconsistent state.

#### Acceptance Criteria

1. WHEN RAG ingestion fails THEN the system SHALL delete the temporary GCS file
2. WHEN RAG ingestion fails THEN the system SHALL update document status to "failed" with error details
3. WHEN GCS rename fails THEN the system SHALL log the error but continue with temporary name
4. WHEN database updates fail THEN the system SHALL log errors for monitoring and debugging
5. WHEN any critical error occurs THEN the system SHALL attempt cleanup of created resources

### Requirement 8

**User Story:** As a system, I need proper Asynq task registration and server configuration, so that background tasks can be processed reliably with appropriate concurrency.

#### Acceptance Criteria

1. WHEN the Asynq server is configured THEN it SHALL use Redis as the message broker
2. WHEN the Asynq server is configured THEN it SHALL support configurable concurrency (default 10)
3. WHEN the task handler is registered THEN it SHALL be mapped to the correct task type constant
4. WHEN the server starts THEN it SHALL begin processing enqueued tasks immediately
5. WHEN task processing fails THEN Asynq SHALL handle retries according to configured policy