# indykite provider integrates IndyKite platform with Terraform scripting.

terraform {
  required_version = ">= 1.9"

  required_providers {
    indykite = {
      source  = "indykite/indykite"
      version = ">=0.0.1"
    }
  }
}

# Provider for now does not support any parameters and all is set within service account credential file.
provider "indykite" {}
