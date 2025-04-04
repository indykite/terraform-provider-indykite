---
# generated by https://github.com/hashicorp/terraform-plugin-docs with custom templates
page_title: "indykite_knowledge_query Resource - IndyKite"
subcategory: ""
description: |-

---

# indykite_knowledge_query (Resource)





<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `location` (String) identifier of Location, where to create resource
- `name` (String) Unique client assigned immutable identifier. Can not be updated without creating a new resource.
- `policy_id` (String) ID of the Authorization Policy that is used to authorize the query.
- `query` (String) Configuration of Knowledge Query in JSON format, the same one exported by The Hub.
- `status` (String) Status of the Knowledge Query. Possible values are: active, draft, inactive.

### Optional

- `description` (String) Your own description of resource. Must be less than or equal to 256 UTF-8 bytes.
- `display_name` (String) The display name for the instance. Can be updated without creating a new resource.
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `app_space_id` (String) identifier of Application Space
- `create_time` (String) Timestamp when the Resource was created. Assigned by the server. A timestamp in RFC3339 UTC "Zulu" format, accurate to nanoseconds. Example: "2014-10-02T15:01:23.045123456Z".
- `customer_id` (String) identifier of Customer
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
