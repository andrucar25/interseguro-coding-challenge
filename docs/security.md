# Security

- Never commit credentials, tokens, or secrets; `.env` files must remain uncommitted.
- Validate untrusted external input and avoid logging sensitive values.
- Use environment variables for service configuration.
- Apply least privilege to runtime identities and service access.
- The intended production posture is a private Node Cloud Run service invoked by Go through Cloud Run IAM authentication.

Keep security measures proportionate to this technical challenge; do not add enterprise controls without a concrete requirement.
