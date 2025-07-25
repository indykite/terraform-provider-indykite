---
# generated by https://github.com/hashicorp/terraform-plugin-docs with custom templates
page_title: "indykite_application_space Resource - IndyKite"
subcategory: ""
description: |-
  It is workspace or environment for your applications.
---

# indykite_application_space (Resource)

It is workspace or environment for your applications.

## Example Usage

```terraform
resource "indykite_application_space" "appspace" {
  customer_id    = "CustomerGID"
  name           = "AppSpaceName"
  display_name   = "Terraform appspace"
  description    = "Application space for terraform configuration"
  region         = "us-east1"
  ikg_size       = "4GB"
  replica_region = "us-west1"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `customer_id` (String) Identifier of Customer
- `name` (String) Unique client assigned immutable identifier. Can not be updated without creating a new resource.
- `region` (String) Region where the application space is located.
		Valid values are: europe-west1, us-east1.

### Optional

- `deletion_protection` (Boolean) Whether or not to allow Terraform to destroy the instance. Unless this field is set to false in Terraform state, a terraform destroy or terraform apply that would delete the instance will fail.
- `description` (String) Your own description of the resource. Must be less than or equal to 256 UTF-8 bytes.
- `display_name` (String) The display name for the instance. Can be updated without creating a new resource.
- `ikg_size` (String) IKG size that will be allocated, which corresponds also to number of CPU nodes (default 2GB).
		Valid values are: 2GB (1 CPU), 4GB (1 CPU), 8GB (2 CPUs), 16GB (3 CPUs), 32GB (6 CPUs), 64GB (12 CPUs),
		128GB (24 CPUs), 192GB (36 CPUs), 256GB (48 CPUs), 384GB (82 CPUs), and 512GB (96 CPUs).
- `replica_region` (String) Replica region specifies where the replica IKG is created.
		Replica must be a different region than the master, but also on the same geographical continent.
		Valid values are: europe-west1, us-east1, us-west1.
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `create_time` (String) Timestamp when the Resource was created. Assigned by the server. A timestamp in RFC3339 UTC "Zulu" format, accurate to nanoseconds. Example: "2014-10-02T15:01:23.045123456Z".
- `id` (String) The ID of this resource.
- `update_time` (String) Timestamp when the Resource was last updated. Assigned by the server. A timestamp in RFC3339 UTC "Zulu" format, accurate to nanoseconds. Example: "2014-10-02T15:01:23.045123456Z".

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String)
- `default` (String)
- `delete` (String)
- `read` (String)
- `update` (String)
