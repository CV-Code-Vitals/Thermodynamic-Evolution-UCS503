package benchmark

import (
	"bufio"
	"fmt"
	"net/http"
	"strings"
	"testing"
)

// TestNAT03_ErrorPathPanicReproducer demonstrates the catastrophic failure of
// placing `defer resp.Body.Close()` before checking `err != nil` or without verifying `resp != nil`.
func TestNAT03_ErrorPathPanicReproducer(t *testing.T) {
	// Scenario: A dial or read call fails immediately (e.g. invalid response / connection refused)
	// Upstream contract: returns (*http.Response == nil, err != nil).

	unsafeCaller := func() (errRet error) {
		// Simulate network call returning err != nil and resp == nil
		readResponseFail := func() (*http.Response, error) {
			// Simulates http.ReadResponse on broken connection
			return nil, fmt.Errorf("read tcp 127.0.0.1:ws: connection reset by peer")
		}

		resp, err := readResponseFail()

		// BUGGY REPAIR PLACEMENT:
		defer func() {
			if r := recover(); r != nil {
				errRet = fmt.Errorf("PANIC_OCCURRED: %v", r)
			}
		}()

		// The unsafe patch executes: defer resp.Body.Close()
		// If placed before error check or without resp != nil guard:
		defer resp.Body.Close()

		if err != nil {
			return err
		}
		return nil
	}

	err := unsafeCaller()
	if err == nil {
		t.Fatalf("Expected error or panic, got nil")
	}

	if !strings.Contains(err.Error(), "PANIC_OCCURRED") {
		t.Fatalf("Expected nil pointer dereference panic, got: %v", err)
	}

	t.Logf("CONFIRMED: Unsafe cleanup placement caused panic on error path: %v", err)
}

// TestNAT03_HttpReadResponseNilReturnContract verifies the standard Go contract:
// When http.ReadResponse fails, resp is strictly NIL.
func TestNAT03_HttpReadResponseNilReturnContract(t *testing.T) {
	// Malformed HTTP response header
	brokenReader := bufio.NewReader(strings.NewReader("INVALID_HTTP_DATA\r\n\r\n"))
	req, _ := http.NewRequest("GET", "http://example.com", nil)

	resp, err := http.ReadResponse(brokenReader, req)
	if err == nil {
		t.Fatalf("Expected error for invalid HTTP data")
	}
	if resp != nil {
		t.Fatalf("Expected resp == nil when ReadResponse fails, got: %v", resp)
	}

	t.Logf("CONFIRMED: http.ReadResponse returns (nil, err) on parse failure (err: %v)", err)
}
