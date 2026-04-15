package domain

type UploadFilesResponse struct {
	FileID   string `json:"file_id"`
	UploadID string `json:"upload_id"`
}

type UploadStatusResponse struct {
	ProgressPercent int32 `json:"progress_percent"`
	Completed       bool  `json:"completed"`
}

type FileMetaResponse struct {
	FileID      string `json:"file_id"`
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	Size        int64  `json:"size"`
	Bucket      string `json:"bucket"`
}

type CreateBucketRequest struct {
	BucketName string `json:"bucket_name" binding:"required"`
}

type ListFilesResponse struct {
	Files []FileMetaResponse `json:"files"`
}