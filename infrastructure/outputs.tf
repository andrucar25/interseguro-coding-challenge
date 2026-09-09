output "artifact_registry_repository_url" {
  description = "Base URL of the managed regional Artifact Registry Docker repository."
  value       = "${google_artifact_registry_repository.containers.location}-docker.pkg.dev/${var.project_id}/${google_artifact_registry_repository.containers.repository_id}"
}

output "node_service_url" {
  description = "Public URL of the Node statistics Cloud Run service."
  value       = google_cloud_run_v2_service.node_api.uri
}

output "go_service_url" {
  description = "Public URL of the Go QR Cloud Run service."
  value       = google_cloud_run_v2_service.go_api.uri
}
