variable "project_id" {
  description = "Google Cloud project ID that owns the deployment."
  type        = string

  validation {
    condition     = length(trimspace(var.project_id)) > 0
    error_message = "project_id must not be empty."
  }
}

variable "region" {
  description = "Google Cloud region for Artifact Registry and Cloud Run."
  type        = string

  validation {
    condition     = length(trimspace(var.region)) > 0
    error_message = "region must not be empty."
  }
}

variable "artifact_registry_repository" {
  description = "Name of the regional Docker Artifact Registry repository."
  type        = string

  validation {
    condition     = length(trimspace(var.artifact_registry_repository)) > 0
    error_message = "artifact_registry_repository must not be empty."
  }
}

variable "go_service_name" {
  description = "Cloud Run service name for the public Go QR API."
  type        = string

  validation {
    condition     = length(trimspace(var.go_service_name)) > 0
    error_message = "go_service_name must not be empty."
  }
}

variable "node_service_name" {
  description = "Cloud Run service name for the public Node statistics API."
  type        = string

  validation {
    condition     = length(trimspace(var.node_service_name)) > 0
    error_message = "node_service_name must not be empty."
  }
}

variable "go_image" {
  description = "Complete, explicitly tagged Artifact Registry image reference for the Go API."
  type        = string

  validation {
    condition     = can(regex("^[^[:space:]@]+:[^[:space:]@:]+$", var.go_image)) && !endswith(var.go_image, ":latest")
    error_message = "go_image must be a complete image reference with an explicit tag other than :latest."
  }
}

variable "node_image" {
  description = "Complete, explicitly tagged Artifact Registry image reference for the Node API."
  type        = string

  validation {
    condition     = can(regex("^[^[:space:]@]+:[^[:space:]@:]+$", var.node_image)) && !endswith(var.node_image, ":latest")
    error_message = "node_image must be a complete image reference with an explicit tag other than :latest."
  }
}

variable "cors_allowed_origins" {
  description = "Origins allowed to call the Go API from a browser."
  type        = string

  validation {
    condition     = length(trimspace(var.cors_allowed_origins)) > 0
    error_message = "cors_allowed_origins must not be empty."
  }
}
