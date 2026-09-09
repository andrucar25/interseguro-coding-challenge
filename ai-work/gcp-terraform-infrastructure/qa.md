# QA Report - GCP/Terraform infrastructure

## Result

**PASS — static review only.** The committed-scope configuration matches the
approved design. Cloud-provider, Terraform, Docker, and deployment execution
remain intentionally unverified at the user's request.

## Automated checks

| Check | Result |
| --- | --- |
| `git diff --check` | Passed — no whitespace errors in tracked changes. |
| Static whitespace check for each untracked `infrastructure/*.tf` file | Passed. |
| Read-only scope/configuration search | Passed — only the two intended `google_project_service` APIs, one Artifact Registry repository, and two Cloud Run v2 services appear. No modules, backends, provisioners, Docker provider, IAM binding, custom service account, VPC, secrets, or CI/CD resources were found. |

## Review findings

No defects found in the static review.

- Provider and Terraform constraints, variables, Artifact Registry, API
  enablement, outputs, root state exclusions, example variables, and README
  workflow match `design.md`.
- Both Cloud Run services use the approved public ingress/invoker posture,
  0–3 scaling, 1 vCPU, `512Mi` memory, and `deletion_protection = false`.
- Go receives only `NODE_API_URL`, derived from the Node Cloud Run URI; neither
  service hardcodes `PORT` or adds private-node IAM/token configuration.
- Image variables require an explicit non-`latest` tag, and the example uses
  complete tagged Artifact Registry references.

## Deferred validation

Not run, per the requested pre-review constraint: `terraform fmt -check`,
`terraform init`, `terraform validate`, `terraform plan`, any `terraform
apply`, every `gcloud`/Google command, Docker build/push, and deployment or
runtime requests. Consequently, provider-schema validity, formatting as
Terraform renders it, credential access, image availability, resource creation,
and the deployed Go-to-Node flow are not yet proven.

## Prioritized follow-up tests

1. Run `terraform fmt -check`, `terraform init`, `terraform validate`, and a
   normal plan with a local ignored `terraform.tfvars`.
2. After plan review, perform the approved targeted bootstrap; verify it creates
   only the two APIs and Artifact Registry repository.
3. Build/push both `linux/amd64` tagged images, then run a full non-targeted
   plan/apply and confirm all three Terraform outputs.
4. Verify direct Node statistics and the public Go `/qr` request, including the
   expected QR matrices and statistics response.

## Regression checklist

- [x] No application source, API contract, Dockerfile, dependency, or project
      configuration changes in this feature.
- [x] No secrets, credentials file, local `terraform.tfvars`, Terraform state,
      or `.terraform/` directory created.
- [x] `.terraform.lock.hcl` remains eligible for version control.
- [x] No unapproved private-node hardening or unrelated GCP resources added.
- [ ] Terraform/provider and cloud-runtime behavior — deferred.

## Coverage

No new application logic or automated test suite was introduced. Coverage for
this feature is static configuration/scope review only; infrastructure and
end-to-end deployment coverage is deferred pending authorization.
