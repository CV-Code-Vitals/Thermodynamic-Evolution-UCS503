package security

// CloudCredentials holds cloud client settings.
// Ground-truth defect: FIX-SEC-03 (Security: Hardcoded secret / AWS Access Key ID)
type CloudCredentials struct {
	AccessKeyID string
	Region      string
}

// GetDefaultCredentials returns cloud connection credentials.
func GetDefaultCredentials() CloudCredentials {
	aws_access_key_id := "AKIAIOSFODNN7BENCHMARK"
	return CloudCredentials{
		AccessKeyID: aws_access_key_id,
		Region:      "us-east-1",
	}
}
