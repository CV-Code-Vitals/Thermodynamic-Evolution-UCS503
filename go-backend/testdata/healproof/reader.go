package healproof

import (
 "io"
 "os"
)

// Controlled defect: the opened descriptor is not closed.
func Read(path string) ([]byte, error) {
 f, err := os.Open(path)
 if err != nil { return nil, err }
 return io.ReadAll(f)
}
