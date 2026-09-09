resource "google_cloud_run_v2_service" "node_api" {
  name                 = var.node_service_name
  location             = var.region
  ingress              = "INGRESS_TRAFFIC_ALL"
  invoker_iam_disabled = true
  deletion_protection  = false

  template {
    scaling {
      min_instance_count = 0
      max_instance_count = 3
    }

    containers {
      image = var.node_image

      env {
        name  = "JWT_SECRET"
        value = var.jwt_secret
      }

      env {
        name  = "JWT_ISSUER"
        value = var.jwt_issuer
      }

      env {
        name  = "JWT_AUDIENCE"
        value = var.jwt_audience
      }

      resources {
        limits = {
          cpu    = "1"
          memory = "512Mi"
        }
        cpu_idle = true
      }
    }
  }

  depends_on = [
    google_project_service.required["artifactregistry.googleapis.com"],
    google_project_service.required["run.googleapis.com"],
    google_artifact_registry_repository.containers,
  ]
}

resource "google_cloud_run_v2_service" "go_api" {
  name                 = var.go_service_name
  location             = var.region
  ingress              = "INGRESS_TRAFFIC_ALL"
  invoker_iam_disabled = true
  deletion_protection  = false

  template {
    scaling {
      min_instance_count = 0
      max_instance_count = 3
    }

    containers {
      image = var.go_image

      env {
        name  = "NODE_API_URL"
        value = google_cloud_run_v2_service.node_api.uri
      }

      env {
        name  = "NODE_API_JWT_SECRET"
        value = var.jwt_secret
      }

      env {
        name  = "NODE_API_JWT_ISSUER"
        value = var.jwt_issuer
      }

      env {
        name  = "NODE_API_JWT_AUDIENCE"
        value = var.jwt_audience
      }

      env {
        name  = "CORS_ALLOWED_ORIGINS"
        value = var.cors_allowed_origins
      }

      resources {
        limits = {
          cpu    = "1"
          memory = "512Mi"
        }
        cpu_idle = true
      }
    }
  }

  depends_on = [
    google_project_service.required["artifactregistry.googleapis.com"],
    google_project_service.required["run.googleapis.com"],
    google_artifact_registry_repository.containers,
  ]
}
