package ingest

import (
	"io"
	"net/http"
)

// Neg_Control_StreamSafe fetches stream and properly closes body
func FetchStreamSafe(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

// DEF-HOLDOUT-05: Resource leak - response body never closed
func IngestEventStream(targetURL string) ([]byte, error) {
	resp, err := http.Get(targetURL)
	if err != nil {
		return nil, err
	}
	// Missing defer resp.Body.Close() -> RESOURCE_LEAK
	return io.ReadAll(resp.Body)
}

