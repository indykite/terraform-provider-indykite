---
# generated by https://github.com/hashicorp/terraform-plugin-docs with custom templates
page_title: "indykite_tenant_configuration Resource - IndyKite"
subcategory: ""
description: |-
  Tenant configuration resource manage defaults and other configuration for given Tenant.
  Most likely it will contain references to other configurations created under Tenant,
  so it must be managed as a separated resource to avoid circular dependencies.
  But be careful, that only 1 configuration per Tenant can be created.
  This resource cannot be imported, because it is tight to Tenant. It will be created automatically
  when Tenant is created, and deleted when Tenant is deleted.
---

# indykite_tenant_configuration (Resource)

Tenant configuration resource manage defaults and other configuration for given Tenant.

Most likely it will contain references to other configurations created under Tenant,
so it must be managed as a separated resource to avoid circular dependencies.
But be careful, that only 1 configuration per Tenant can be created.

This resource cannot be imported, because it is tight to Tenant. It will be created automatically
when Tenant is created, and deleted when Tenant is deleted.

## Example Usage

{{tffile "/home/runner/work/terraform-provider-indykite/terraform-provider-indykite/examples/resources/indykite_tenant_configuration/resource.tf"}}

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `tenant_id` (String) identifier of Tenant

### Optional

- `default_auth_flow_id` (String) ID of default Authentication flow
- `default_email_service_id` (String) ID of default Email notification provider
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))
- `username_policy` (Block List, Max: 1) (see [below for nested schema](#nestedblock--username_policy))

### Read-Only

- `id` (String) The ID of this resource.

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `default` (String)
- `read` (String)
- `update` (String)


<a id="nestedblock--username_policy"></a>
### Nested Schema for `username_policy`

Optional:

- `allowed_email_domains` (List of String) Allowed email domains to register. Can be shared among tenants.
- `allowed_username_formats` (List of String) Which username format is allowed. Valid values are email, mobile and username
- `exclusive_email_domains` (List of String) Allowed email domains to register. Can be shared among tenants.
- `valid_email` (Boolean) If email must be valid with MX record
- `verify_email` (Boolean) If email must be verified by a link sent to the owner
- `verify_email_grace_period` (String)
