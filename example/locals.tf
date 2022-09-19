// Copyright (c) 2022 IndyKite
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

locals {
  oauth2_grant_types = {
    authorization_code = "authorization_code"
    implicit           = "implicit"
    password           = "password"
    client_credentials = "client_credentials"
    refresh_token      = "refresh_token"
  }

  oauth2_response_types = {
    token = "token"
    code  = "code"
    token = "id_token"
  }

  oauth2_token_endpoint_auth_methods = {
    client_secret_basic = "client_secret_basic"
    client_secret_post  = "client_secret_post"
    private_key_jwt     = "private_key_jwt"
    none                = "none"
  }

  oauth2_client_subject_types = {
    public   = "public"
    pairwise = "pairwise"
  }

  supported_auth_signing_algs = {
    RS256  = "RS256"
    RS384  = "RS384"
    RS512  = "RS512"
    PS256  = "PS256"
    PS384  = "PS384"
    PS512  = "PS512"
    ES256  = "ES256"
    ES384  = "ES384"
    ES512  = "ES512"
    PS256K = "ES256K"
    HS256  = "HS256"
    HS384  = "HS384"
    HS512  = "HS512"
    EdDSA  = "EdDSA"
  }
}
