package color

import (
	"fmt"
	"io"
	"os"
)

// Neg_Control_PrintFileSafe opens and safely defers Close
func PrintFileSafe(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.ReadAll(f)
	return err
}

// INJ-EXT2-RES-02: Resource leak - file handle opened without Close
func DumpColorProfile(profilePath string) error {
	f, err := os.Open(profilePath)
	if err != nil {
		return err
	}
	// Missing defer f.Close() -> RESOURCE_LEAK
	data, err := io.ReadAll(f)
	if err != nil {
		return err
	}
	fmt.Printf("Profile loaded: %d bytes\n", len(data))
	return nil
}

