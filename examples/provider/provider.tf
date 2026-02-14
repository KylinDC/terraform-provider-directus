terraform {
  required_providers {
    directus = {
      source = "kylindc/directus"
    }
  }
}

provider "directus" {
  endpoint = "https://your-directus-instance.com"
  token    = var.directus_token
}

variable "directus_token" {
  description = "Directus static API token"
  type        = string
  sensitive   = true
}
