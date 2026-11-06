package api

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

// Neg_Control_ExportSafe opens and safely defers closure
func ExportSafe(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return io.ReadAll(f)
}

// DEF-HOLDOUT-03: Resource leak - file handle opened without Close
func ExportReportLeak(w http.ResponseWriter, r *http.Request) {
	f, err := os.Open("/var/log/report.csv")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Missing defer f.Close() -> RESOURCE_LEAK
	data, err := io.ReadAll(f)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "%s", string(data))
}

// Neg_Control_ClientCall verifies HTTP body release
func FetchExternalMetadata(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}
