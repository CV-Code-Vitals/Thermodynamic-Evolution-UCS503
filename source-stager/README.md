# GitHub Source Stager

The source stager is a private Lambda invoked by the analysis API proxy. It
downloads a public GitHub repository archive, validates it, stores the ZIP in
the private source bucket, and returns an opaque snapshot ID. It is not a
public HTTP endpoint, and private EC2 instances do not fetch repositories
directly.

## Request and response

Invoke the Lambda with:

```json
{"repository_url":"https://github.com/owner/repository"}
```

The URL may optionally end in `.git`. A successful invocation returns:

```json
{"snapshot_id":"<32 lowercase hexadecimal characters>","expires_at":"<UTC timestamp>"}
```

Callers provide only the repository URL and opaque ID. The object key is
derived by the stager as `snapshots/<snapshot_id>/source.zip`.

## Validation and limits

- Only public `https://github.com/{owner}/{repository}` URLs are accepted.
- Metadata and archive requests are restricted to `api.github.com`,
  `github.com`, and `codeload.github.com`, resolved to public IP addresses.
- Redirects must stay on those HTTPS hosts, use port 443, and remain within the
  redirect limit. Network connections are pinned to the validated DNS address.
- Compressed archives are limited to 50 MiB; expanded contents to 256 MiB;
  entries to 10,000; and individual files to 16 MiB.
- Traversal, absolute paths, duplicate paths, symlinks, invalid ZIP entries,
  and oversized responses are rejected before storage.
- S3 writes use server-side encryption. Configure a bucket lifecycle rule to
  expire `snapshots/` objects after one day; the response expiry timestamp is
  informational and does not replace that lifecycle policy.

## Runtime configuration and permissions

Set `SOURCE_BUCKET` to the dedicated private source bucket. The Lambda role
needs only `s3:PutObject` for `arn:aws:s3:::<source-bucket>/snapshots/*` and
the standard logging permissions required by the deployment. Keep the
function outside the VPC so it can reach GitHub without adding NAT to the
private analysis network. Restrict `lambda:InvokeFunction` to the analysis
proxy role; do not attach a public API trigger.

## Local verification

From this directory:

```powershell
go test ./...
go vet ./...
```
