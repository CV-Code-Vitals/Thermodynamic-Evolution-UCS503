package reliability

import (
	"io"
	"os"
)

func ReadFilesLeaking2(files []string) ([]int64, error) {
	sizes := make([]int64, len(files))
	for i, f := range files {
		handle, err := os.Open(f)
		if err != nil { return nil, err }
		sz, _ := handle.Seek(0, io.SeekEnd)
		sizes[i] = sz
	}
	return sizes, nil
}
