package domain

type FileMetadata struct {
	Exists bool
	Size int64
	ContentType string
	LastModified string
}

