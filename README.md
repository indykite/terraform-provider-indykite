# Terraform Provider for IndyKite

[![Test](https://github.com/indykite/terraform-provider-indykite/actions/workflows/go-test.yaml/badge.svg)](https://github.com/indykite/terraform-provider-indykite/actions/workflows/go-test.yaml)&nbsp;
[![codecov](https://codecov.io/gh/indykite/terraform-provider-indykite/graph/badge.svg?token=bPeHVRUJaZ)](https://codecov.io/gh/indykite/terraform-provider-indykite)&nbsp;
[![registry](https://img.shields.io/github/v/release/indykite/terraform-provider-indykite)](https://registry.terraform.io/providers/indykite/indykite/latest)

The Terraform IndyKite provider is a plugin for Terraform that allows for the full
lifecycle management of IndyKite resources.
This provider is maintained internally by the IndyKite Provider team.

Please note: We take Terraform's security and our users' trust very seriously.
If you believe you have found a security issue in the IndyKite Terraform Provider,
please responsibly disclose by contacting us at security@indykite.com.

## Quick Starts

- [Provider documentation](https://registry.terraform.io/providers/indykite/indykite/latest/docs)

```hcl
terraform {
  required_providers {
    indykite = {
      source = "indykite/indykite"
      version = ">= 0.2.1"
    }
  }
}

provider "indykite" {
  # Configuration options
}
```

## Install

### Terraform

Be sure you have the correct Terraform version (0.13.0+), you can choose the binary here:

- https://releases.hashicorp.com/terraform/
