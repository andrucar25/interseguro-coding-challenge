# Architecture

Intended request flow:

```text
Client → Go/Fiber API → QR factorization → HTTP request → Node/Express API → statistics → Go API → Client
```

The Go service will produce the matrices from QR factorization. The Node service will calculate maximum, minimum, average, total sum, and whether either result matrix is diagonal.

## Project assumption

The challenge statement inconsistently mentions matrix rotation in some places, while its explicit functional requirement requests QR factorization. Treat QR factorization as the intended operation unless a task explicitly resolves the contradiction differently. Do not implement either operation based on this documentation alone.

## Intended deployment direction

Both APIs are intended to run as Docker images in Google Cloud Run, stored in Artifact Registry and provisioned with Terraform. The Go service is public; the Node service should be private when practical and invoked by Go through Cloud Run IAM service-to-service authentication. GitHub Actions may later run tests, build images, and deploy. This is future direction, not current infrastructure.
