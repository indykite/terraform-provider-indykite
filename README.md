# Terraform Provider for IndyKite

[![Test](https://github.com/indykite/terraform-provider-indykite/actions/workflows/go-test.yaml/badge.svg)](https://github.com/indykite/terraform-provider-indykite/actions/workflows/go-test.yaml)&nbsp;
[![codecov](https://codecov.io/gh/indykite/terraform-provider-indykite/graph/badge.svg?token=bPeHVRUJaZ)](https://codecov.io/gh/indykite/terraform-provider-indykite)&nbsp;
[![registry](https://img.shields.io/github/v/release/indykite/terraform-provider-indykite)](https://registry.terraform.io/providers/indykite/indykite/latest)

The Terraform IndyKite provider is a plugin for Terraform that allows for the full
lifecycle management of IndyKite resources.
This provider is maintained internally by the IndyKite Provider team.

Please note: We take Terraform's security and our users' trust very seriously.
If you believe you have found a security issue in the IndyKite Terraform Provider,
please responsibly disclose by contacting us at <security@indykite.com>.

## Quick Starts

- [Provider documentation](https://registry.terraform.io/providers/indykite/indykite/latest/docs)

The provider need to be set:

```hcl
terraform {
  required_providers {
    indykite = {
      source = "indykite/indykite"
      version = "~> 0.27"
    }
  }
}

provider "indykite" {
  # Configuration options
}
```

And configured with one of the following environment variables:

- `INDYKITE_SERVICE_ACCOUNT_CREDENTIALS_FILE` with path to service account credentials file generated from our console.

- `INDYKITE_SERVICE_ACCOUNT_CREDENTIALS` with content of service account credentials file generated from our console.

## Install

### Terraform

Be sure you have the correct Terraform version (0.12.0+), you can choose the binary here

- <https://releases.hashicorp.com/terraform/>

## Example

You can find Terraform examples in our [Provider documentation](https://registry.terraform.io/providers/indykite/indykite/latest/docs).
A complete script example is available [test.tf](tests/provider/test.tf).

## Provider development

### GitHub workflows

`tfplugindocs` GitHub workflow automatically re-generates the provider documentation once commit is pushed to `master`.
It requires a PAT with the following permissions:

- Read access to metadata.
- Read and Write access to administration.
- Read and Write access to code.
