package reliability

import (
	"io"
	"os"
)

// ReadFilesUnsafely opens multiple files in a loop without closing them, leaking file descriptors.
// Ground-truth defect: FIX-REL-02 (Reliability: Unsafe resource lifecycle / Leaked file handles)
func ReadFilesUnsafely(paths []string) ([]int64, error) {
	sizes := make([]int64, len(paths))
	for i, p := range paths {
		f, err := os.Open(p)
		if err != nil {
			return nil, err
		}
		// Resource leak: file handle not closed in loop
		pos, _ := f.Seek(0, io.SeekEnd)
		sizes[i] = pos
	}
	return sizes, nil
}

