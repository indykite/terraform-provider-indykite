// Copyright (c) 2023 IndyKite
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

//go:build integration

package terraform_test

import (
	"context"
	"encoding/json"
	"os"

	"github.com/indykite/terraform-provider-indykite/indykite"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	client   *indykite.RestClient
	myResult = make(map[string]string)
)

type Data struct {
	Resources []Resource `json:"resources"`
}

type Resource struct {
	Mode      string     `json:"mode"`
	Type      string     `json:"type"`
	Name      string     `json:"name"`
	Instances []Instance `json:"instances"`
}

type Instance struct {
	Attributes Attribute `json:"attributes"`
}

type Attribute struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

var _ = Describe("Terraform", func() {
	BeforeEach(func() {
		var err error
		client, err = indykite.NewRestClient(context.Background())
		Expect(err).To(Succeed())

		jsonData, err := os.ReadFile("../provider/terraform.tfstate")
		Expect(err).To(Succeed())
		Expect(jsonData).NotTo(BeNil())

		data := Data{}
		err = json.Unmarshal(jsonData, &data)
		Expect(err).To(Succeed())

		for _, getid := range data.Resources {
			myResult[getid.Name] = getid.Instances[0].Attributes.ID
		}
		Expect(myResult).NotTo(BeNil())
	})

	It("ReadAppSpace", func() {
		var resp indykite.ApplicationSpaceResponse
		err := client.Get(context.Background(), "/projects/"+myResult["appspace"], &resp)
		Expect(err).To(Succeed())
		Expect(resp.ID).To(Equal(myResult["appspace"]))
	})

	It("ReadApplication", func() {
		var resp indykite.ApplicationResponse
		err := client.Get(context.Background(), "/applications/"+myResult["application"], &resp)
		Expect(err).To(Succeed())
		Expect(resp.ID).To(Equal(myResult["application"]))
	})

	It("ReadAgent", func() {
		var resp indykite.ApplicationAgentResponse
		err := client.Get(context.Background(), "/application-agents/"+myResult["agent"], &resp)
		Expect(err).To(Succeed())
		Expect(resp.ID).To(Equal(myResult["agent"]))
	})

	It("ReadCredential", func() {
		var resp indykite.ApplicationAgentCredentialResponse
		err := client.Get(context.Background(), "/application-agent-credentials/"+myResult["with_public"], &resp)
		Expect(err).To(Succeed())
		Expect(resp.ID).To(Equal(myResult["with_public"]))
	})

	It("ReadPolicy", func() {
		var resp indykite.AuthorizationPolicyResponse
		err := client.Get(context.Background(), "/authorization-policies/"+myResult["policy_drive_car"], &resp)
		Expect(err).To(Succeed())
		Expect(resp.ID).To(Equal(myResult["policy_drive_car"]))
	})
})
