package structural

// MapHTTPStatusToCategory is a clean, benign function that uses a simple, flat switch table.
// Ground-truth negative control: CLEAN-STR-01 (Not a defect)
func MapHTTPStatusToCategory(statusCode int) string {
	switch {
	case statusCode >= 200 && statusCode < 300:
		return "SUCCESS"
	case statusCode >= 300 && statusCode < 400:
		return "REDIRECTION"
	case statusCode >= 400 && statusCode < 500:
		return "CLIENT_ERROR"
	case statusCode >= 500 && statusCode < 600:
		return "SERVER_ERROR"
	default:
		return "UNKNOWN"
	}
}

