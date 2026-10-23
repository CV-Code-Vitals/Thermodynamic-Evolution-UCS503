build-ArchiveApiFunction:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -tags lambda.norpc -ldflags="-s -w" -o "$(ARTIFACTS_DIR)/bootstrap" .
