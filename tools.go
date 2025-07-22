package go_as

import (
	"time"
)

// GenericSuccessFailureResult can be used for tools that only return a success/failure message.
type GenericSuccessFailureResult struct {
	Message string `json:"message"`
}

// ListDirectoryArgs defines arguments for list_directory tool
type ListDirectoryArgs struct {
	Path          string `json:"path"`
	Recursive     bool   `json:"recursive,omitempty"`
	MaxDepth      int    `json:"max_depth,omitempty"`
	IncludeHidden bool   `json:"include_hidden,omitempty"`
}

// FileSystemItem represents a file or directory in a listing
type FileSystemItem struct {
	Name         string    `json:"name"`
	Type         string    `json:"type"`           // "file" or "directory"
	Size         int64     `json:"size,omitempty"` // For files
	LastModified time.Time `json:"last_modified"`
	Path         string    `json:"path"`
}

// ListDirectoryResult defines the result for list_directory tool
type ListDirectoryResult struct {
	Items []FileSystemItem `json:"items"`
}

// MoveItemArgs defines arguments for move_item tool
type MoveItemArgs struct {
	SourcePath      string `json:"source_path"`
	DestinationPath string `json:"destination_path"`
}

// CopyItemArgs defines arguments for copy_item tool
type CopyItemArgs struct {
	SourcePath      string `json:"source_path"`
	DestinationPath string `json:"destination_path"`
	Overwrite       bool   `json:"overwrite,omitempty"`
}

// DeleteItemArgs defines arguments for delete_item tool
type DeleteItemArgs struct {
	Path      string `json:"path"`
	Recursive bool   `json:"recursive,omitempty"`
}

// ReadFileArgs defines arguments for read_file tool
type ReadFileArgs struct {
	Path     string `json:"path"`
	Encoding string `json:"encoding,omitempty"` // "utf-8", "base64"
}

// ReadFileResult defines the result for read_file tool
type ReadFileResult struct {
	Content  string `json:"content"`
	MimeType string `json:"mime_type"`
}

func (ReadFileResult) isContent() {}

// WriteFileArgs defines arguments for write_file tool
type WriteFileArgs struct {
	Path     string `json:"path"`
	Content  string `json:"content"`
	Append   bool   `json:"append,omitempty"`
	Encoding string `json:"encoding,omitempty"` // For content interpretation if needed
}

// GetItemPropertiesArgs defines arguments for get_item_properties tool
type GetItemPropertiesArgs struct {
	Path string `json:"path"`
}

// ItemProperties defines the detailed properties of a file or directory
type ItemProperties struct {
	Name         string    `json:"name"`
	Path         string    `json:"path"`
	Type         string    `json:"type"` // "file" or "directory"
	Size         int64     `json:"size"`
	LastModified time.Time `json:"last_modified"`
	CreatedAt    time.Time `json:"created_at"`
	Permissions  string    `json:"permissions"` // e.g., "rwxr-xr-x"
	IsReadOnly   bool      `json:"is_readonly"`
}

// ItemExistsArgs defines arguments for item_exists tool
type ItemExistsArgs struct {
	Path string `json:"path"`
}

// ItemExistsResult defines the result for item_exists tool
type ItemExistsResult struct {
	Exists bool   `json:"exists"`
	Type   string `json:"type"` // "file", "directory", or "not_found"
}

// CreateArchiveArgs defines arguments for create_archive tool
type CreateArchiveArgs struct {
	SourcePaths []string `json:"source_paths"`
	ArchivePath string   `json:"archive_path"`
	Format      string   `json:"format"` // "zip", "tar.gz"
}

// CreateArchiveResult defines the result for create_archive tool
type CreateArchiveResult struct {
	PathToArchive string `json:"path_to_archive"`
}

// ExtractArchiveArgs defines arguments for extract_archive tool
type ExtractArchiveArgs struct {
	ArchivePath     string `json:"archive_path"`
	DestinationPath string `json:"destination_path"`
	Format          string `json:"format,omitempty"` // Optional, auto-detect if possible
}
