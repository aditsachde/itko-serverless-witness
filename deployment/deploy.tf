variable "project_id" {
  description = "The ID of the Google Cloud project."
  type        = string
}

variable "regions" {
  description = "Regions to deploy witnesses to."
  type        = list(string)
}

variable "tag" {
  description = "Docker image tag"
  type        = string
}

variable "domain" {
  description = "Base domain for domain mapping"
  type        = string
}

provider "google" {
  project = var.project_id
}

data "google_project" "project" {
}

resource "google_secret_manager_secret" "configuration" {
  secret_id = "configuration"

  replication {
    auto {}
  }
}

resource "google_secret_manager_secret_iam_member" "secret-access" {
  secret_id  = google_secret_manager_secret.configuration.id
  role       = "roles/secretmanager.secretAccessor"
  member     = "serviceAccount:${data.google_project.project.number}-compute@developer.gserviceaccount.com"
  depends_on = [google_secret_manager_secret.configuration]
}


resource "google_artifact_registry_repository" "repo" {
  location      = "us-central1"
  repository_id = "witness"
  format        = "DOCKER"

  cleanup_policy_dry_run = false
  cleanup_policies {
    id     = "keep-last-2"
    action = "KEEP"
    most_recent_versions {
      keep_count = 2
    }
  }
}

data "google_artifact_registry_docker_image" "image" {
  location      = google_artifact_registry_repository.repo.location
  repository_id = google_artifact_registry_repository.repo.repository_id
  image_name    = "witness:${var.tag}"
}

resource "google_cloud_run_v2_service" "witness" {
  for_each = toset(var.regions)

  name                = "witness-${each.key}"
  location            = each.key
  deletion_protection = false
  ingress             = "INGRESS_TRAFFIC_ALL"

  template {
    containers {
      image = data.google_artifact_registry_docker_image.image.self_link

      env {
        name  = "REGION"
        value = each.key
      }

      env {
        name  = "CONFIG_SECRET"
        value = "${google_secret_manager_secret.configuration.id}/versions/latest"
      }
    }
  }
}

resource "google_cloud_run_domain_mapping" "witness_domain" {
  for_each = toset(var.regions)

  name     = "${each.key}.${var.domain}"
  location = google_cloud_run_v2_service.witness[each.key].location
  metadata {
    namespace = data.google_project.project.project_id
  }
  spec {
    route_name = google_cloud_run_v2_service.witness[each.key].name
  }
}


resource "google_cloud_run_service_iam_binding" "public_access" {
  for_each = toset(var.regions)

  location = google_cloud_run_v2_service.witness[each.key].location
  service  = google_cloud_run_v2_service.witness[each.key].name
  role     = "roles/run.invoker"
  members = [
    "allUsers"
  ]
}

resource "google_pubsub_topic" "witness_refresher" {
  name = "witness-refresher"
}

resource "google_cloud_scheduler_job" "every_five_minutes" {
  name        = "witness-refresher"
  description = "Refresh witnesses every 5 minutes"
  schedule    = "*/5 * * * *" # Runs every 5 minutes
  time_zone   = "Etc/UTC"
  region      = "us-central1"

  pubsub_target {
    topic_name = google_pubsub_topic.witness_refresher.id
    data       = base64encode("Refresh witnesses")
  }
}

resource "google_pubsub_subscription" "subscription" {
  for_each = toset(var.regions)

  name  = "witness-refresher-subscription-${each.key}"
  topic = google_pubsub_topic.witness_refresher.name
  push_config {
    push_endpoint = "${google_cloud_run_v2_service.witness[each.key].uri}/witness/v0/logs/refresh"
    attributes = {
      x-goog-version = "v1"
    }
  }
  depends_on = [google_cloud_run_v2_service.witness]
}

output "urls" {
  value = { for r in var.regions : r => google_cloud_run_v2_service.witness[r].uri }
}

output "domain_mappings" {
  value = { for r in var.regions : r => google_cloud_run_domain_mapping.witness_domain[r].name }
}
