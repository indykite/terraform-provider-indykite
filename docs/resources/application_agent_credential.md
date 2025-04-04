---
# generated by https://github.com/hashicorp/terraform-plugin-docs with custom templates
page_title: "indykite_application_agent_credential Resource - IndyKite"
subcategory: ""
description: |-

---

# indykite_application_agent_credential (Resource)



## Example Usage

```terraform
resource "indykite_application_agent_credential" "with_public" {
  app_agent_id = indykite_application_agent.agent.id
  display_name = "Key with custom private-public key pair"
  expire_time  = "2026-12-31T12:34:56-01:00"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `app_agent_id` (String) identifier of Application Agent

### Optional

- `display_name` (String)
- `expire_time` (String) Optional date-time when credentials are going to expire
- `public_key_jwk` (String, Deprecated) Provide your onw Public key in JWK format, otherwise new pair is generated
- `public_key_pem` (String, Deprecated) Provide your onw Public key in PEM format, otherwise new pair is generated
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `agent_config` (String, Sensitive)
- `app_space_id` (String) identifier of Application Space
- `application_id` (String) identifier of Application
- `create_time` (String) Timestamp when the Resource was created. Assigned by the server. A timestamp in RFC3339 UTC "Zulu" format, accurate to nanoseconds. Example: "2014-10-02T15:01:23.045123456Z".
- `customer_id` (String) identifier of Customer
- `id` (String) The ID of this resource.
- `kid` (String)

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String)
- `default` (String)
- `delete` (String)
- `read` (String)

## Import

Import is supported using the following syntax:
```shell
terraform import indykite_application_agent_credential.id gid:AAABBBCCC_000111222333
```
