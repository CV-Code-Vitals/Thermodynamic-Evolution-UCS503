# Deployment metadata query fix

## Intent

Allow the image deployment workflow to read the existing `bundle-sha256` S3 metadata key so it can safely recognize an already-uploaded bundle.

## Scope

Change only the AWS CLI JMESPath expression in `.github/workflows/deploy-analysis.yml` to quote the hyphenated metadata key. Preserve upload metadata, checksum comparison, and deployment behavior.

## Approaches

- **Quote the key in JMESPath (selected):** use `Metadata."bundle-sha256"`. This is the smallest change and keeps stored metadata compatible.
- **Rename the metadata key:** update the upload and query together. This adds unnecessary changes and may complicate compatibility with existing S3 objects.

## Verification

Run the targeted workflow tests and validate the workflow syntax. No AWS resources or GitHub Actions runs will be modified by this local change.
