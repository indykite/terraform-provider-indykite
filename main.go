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

// Package main implements terraform provider main.
package main

import (
	"flag"
	"os"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"

	"github.com/indykite/terraform-provider-indykite/indykite"
)

func main() {
	var debugMode bool
	// https://www.terraform.io/docs/extend/debugging.html#enabling-debugging-in-a-provider
	flag.BoolVar(&debugMode, "debug", false,
		"set to true to run the provider with support for debuggers like delve")
	flag.Parse()
	plugin.Serve(&plugin.ServeOpts{
		GRPCProviderFunc: func() tfprotov5.ProviderServer {
			return schema.NewGRPCProviderServer(indykite.Provider())
		},
		NoLogOutputOverride: acceptanceTesting(),
		Debug:               debugMode,
	})
}

func acceptanceTesting() bool {
	_, ok := os.LookupEnv("TF_TEST_ENV_ACCEPTANCE_TESTING")
	return ok
}
