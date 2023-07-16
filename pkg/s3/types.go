package s3

import (
	"time"
)

type ObjectDetails struct {
	Name         string    `json:"name"`
	IsFolder     bool      `json:"isFolder"`
	Size         int64     `json:"size"`
	LastModified time.Time `json:"lastModified"`
	DownloadLink string    `json:"downloadLink,omitempty"`
}

type ListFilesResponse struct {
	Files               *[]ObjectDetails `json:"data"`
	NextPageToken       string           `json:"nextPageToken,omitempty"` // for pagination purposes only if
	IsLastPage          bool             `json:"isLastPage,omitempty"`    // for pagination purposes only if
	NoOfRecordsReturned int32            `json:"noOfRecordsReturned,omitempty"`
	FilesCount          int32            `json:"filesCount,omitempty"`
	FoldersCount        int32            `json:"foldersCount,omitempty"`
}

type SuccessResponse struct {
	Status       string      `json:"status"`
	ResponseCode int         `json:"response_code"`
	Data         interface{} `json:"data"`
}

type FailureResponse struct {
	Status       string `json:"status"`
	ResponseCode int    `json:"response_code"`
	ErrorMessage string `json:"error_message"`
}

type S3UploadPayload struct {
	Bucket     string `json:"bucket"`
	FolderPath string `json:"folderPath"`
}

// FileInfo represents the details of a file or folder
type FileInfo struct {
	Name     string `json:"name"`
	IsFolder bool   `json:"isFolder"`
	// Add more fields as per your requirements
}

type FilterOptions struct {
	SizeRange          string
	TimeRange          string
	FileTypes          []string
	FilenameQuery      string
	FilenameFilterType string
	FileSize           int64
	FileSizeFilterType string
}

type FilterSizeRange struct {
	MinSize int64
	MaxSize int64
}
