resource "indykite_consent" "basic-user-data" {
  location    = "appSpace-gid"
  name        = "location-name-sharing"
  description = "This consent will allow third parties to access the location and name of the user"

  purpose          = "To send you your order you need to share your location and name with the delivery service"
  application_id   = "application-gid"
  validity_period  = 96400
  revoke_after_use = false
  data_points = [
    jsonencode(
      {
        "query": "->[:BELONGS]-(c:CAR)-[:MADEBY]->(o:MANUFACTURER)",
        "returns": [
          {
            "variable": "c",
            "properties": [
              "Model"
            ]
          },
          {
            "variable": "o",
            "properties": [
              "Name"
            ]
          }
        ]
      }
    )
  ]
}
