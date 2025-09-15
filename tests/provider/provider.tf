terraform {
  required_version = ">= 1.9"

  required_providers {
    indykite = {
      source  = "indykite/indykite"
      version = ">=0.0.1" # Keep hardcoded version >=0.0.1 as provider is build locally with this version
    }
  }
}

provider "indykite" {
  # Configuration options
}
