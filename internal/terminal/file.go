package terminal

import "github.com/dustin/go-humanize"

type File struct {
	FileName string `json:"fileName"`
	FileSize uint64 `json:"fileSize"`
}

func (target *File) GetHumanReadableFileSize() string {
	return humanize.Bytes(target.FileSize)
}
