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

	"github.com/indykite/indykite-sdk-go/config"
	configpb "github.com/indykite/indykite-sdk-go/gen/indykite/config/v1beta1"
	"github.com/indykite/indykite-sdk-go/grpc"
	apicfg "github.com/indykite/indykite-sdk-go/grpc/config"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var (
	client   *config.Client
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
		client, err = config.NewClient(context.Background(),
			grpc.WithCredentialsLoader(apicfg.DefaultEnvironmentLoader),
			grpc.WithServiceAccount(),
		)
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
		resp, err := client.ReadApplicationSpace(context.Background(),
			&configpb.ReadApplicationSpaceRequest{
				Identifier: &configpb.ReadApplicationSpaceRequest_Id{Id: myResult["appspace"]},
			},
		)
		Expect(err).To(Succeed())
		Expect(resp).NotTo(BeNil())

		Expect(resp).To(PointTo(MatchFields(IgnoreExtras, Fields{
			"AppSpace": PointTo(MatchFields(IgnoreExtras, Fields{
				"Id": Equal(myResult["appspace"]),
			})),
		})))
	})

	It("ReadTenant", func() {
		resp, err := client.ReadTenant(context.Background(),
			&configpb.ReadTenantRequest{
				Identifier: &configpb.ReadTenantRequest_Id{Id: myResult["tenant"]},
			},
		)
		Expect(err).To(Succeed())
		Expect(resp).NotTo(BeNil())

		Expect(resp).To(PointTo(MatchFields(IgnoreExtras, Fields{
			"Tenant": PointTo(MatchFields(IgnoreExtras, Fields{
				"Id": Equal(myResult["tenant"]),
			})),
		})))
	})

	It("ReadApplication", func() {
		resp, err := client.ReadApplication(context.Background(),
			&configpb.ReadApplicationRequest{
				Identifier: &configpb.ReadApplicationRequest_Id{Id: myResult["application"]},
			},
		)
		Expect(err).To(Succeed())
		Expect(resp).NotTo(BeNil())

		Expect(resp).To(PointTo(MatchFields(IgnoreExtras, Fields{
			"Application": PointTo(MatchFields(IgnoreExtras, Fields{
				"Id": Equal(myResult["application"]),
			})),
		})))
	})

	It("ReadAgent", func() {
		resp, err := client.ReadApplicationAgent(context.Background(),
			&configpb.ReadApplicationAgentRequest{
				Identifier: &configpb.ReadApplicationAgentRequest_Id{Id: myResult["agent"]},
			},
		)
		Expect(err).To(Succeed())
		Expect(resp).NotTo(BeNil())

		Expect(resp).To(PointTo(MatchFields(IgnoreExtras, Fields{
			"ApplicationAgent": PointTo(MatchFields(IgnoreExtras, Fields{
				"Id": Equal(myResult["agent"]),
			})),
		})))
	})

	It("ReadCredential", func() {
		resp, err := client.ReadApplicationAgentCredential(context.Background(),
			&configpb.ReadApplicationAgentCredentialRequest{
				Id: myResult["with_public"],
			},
		)
		Expect(err).To(Succeed())
		Expect(resp).NotTo(BeNil())

		Expect(resp).To(PointTo(MatchFields(IgnoreExtras, Fields{
			"ApplicationAgentCredential": PointTo(MatchFields(IgnoreExtras, Fields{
				"Id": Equal(myResult["with_public"]),
			})),
		})))
	})

	It("ReadPolicy", func() {
		configNodeRequest, err := config.NewRead(myResult["policy_drive_car"])
		Expect(err).To(Succeed())
		resp, err := client.ReadConfigNode(context.Background(), configNodeRequest)
		Expect(err).To(Succeed())
		Expect(resp).NotTo(BeNil())

		Expect(resp).To(PointTo(MatchFields(IgnoreExtras, Fields{
			"ConfigNode": PointTo(MatchFields(IgnoreExtras, Fields{
				"Id":     Equal(myResult["policy_drive_car"]),
				"Config": Not(BeNil()),
			})),
		})))
	})

	It("ReadEmail", func() {
		configNodeRequest, err := config.NewRead(myResult["email_conf"])
		Expect(err).To(Succeed())
		resp, err := client.ReadConfigNode(context.Background(), configNodeRequest)
		Expect(err).To(Succeed())
		Expect(resp).NotTo(BeNil())

		Expect(resp).To(PointTo(MatchFields(IgnoreExtras, Fields{
			"ConfigNode": PointTo(MatchFields(IgnoreExtras, Fields{
				"Id":     Equal(myResult["email_conf"]),
				"Config": Not(BeNil()),
			})),
		})))
	})

	It("ReadOAuth2Provider", func() {
		resp, err := client.ReadOAuth2Provider(context.Background(),
			&configpb.ReadOAuth2ProviderRequest{
				Id: myResult["oauth2_provider"],
			},
		)
		Expect(err).To(Succeed())
		Expect(resp).NotTo(BeNil())
		Expect(resp.Oauth2Provider.Id).To(Equal(myResult["oauth2_provider"]))
	})

	It("ReadOAuth2Application", func() {
		resp, err := client.ReadOAuth2Application(context.Background(),
			&configpb.ReadOAuth2ApplicationRequest{
				Id: myResult["oauth2_app"],
			},
		)
		Expect(err).To(Succeed())
		Expect(resp).NotTo(BeNil())
		Expect(resp.Oauth2Application.Id).To(Equal(myResult["oauth2_app"]))
	})

	It("ReadOAuth2Client", func() {
		configNodeRequest, err := config.NewRead(myResult["oauth2_client"])
		Expect(err).To(Succeed())
		resp, err := client.ReadConfigNode(context.Background(), configNodeRequest)
		Expect(err).To(Succeed())
		Expect(resp).NotTo(BeNil())

		Expect(resp).To(PointTo(MatchFields(IgnoreExtras, Fields{
			"ConfigNode": PointTo(MatchFields(IgnoreExtras, Fields{
				"Id":     Equal(myResult["oauth2_client"]),
				"Config": Not(BeNil()),
			})),
		})))
	})

	It("ReadCustomerConfig", func() {
		resp, err := client.ReadCustomerConfig(context.Background(),
			&configpb.ReadCustomerConfigRequest{
				Id: myResult["customer"],
			},
		)
		Expect(err).To(Succeed())
		Expect(resp).NotTo(BeNil())
		Expect(resp.Id).To(Equal(myResult["customer"]))
	})

	It("ReadAppSpaceConfig", func() {
		resp, err := client.ReadApplicationSpaceConfig(context.Background(),
			&configpb.ReadApplicationSpaceConfigRequest{
				Id: myResult["appspace"],
			},
		)
		Expect(err).To(Succeed())
		Expect(resp).NotTo(BeNil())
		Expect(resp.Id).To(Equal(myResult["appspace"]))
	})

	It("ReadTenantConfig", func() {
		resp, err := client.ReadTenantConfig(context.Background(),
			&configpb.ReadTenantConfigRequest{
				Id: myResult["tenant"],
			},
		)
		Expect(err).To(Succeed())
		Expect(resp).NotTo(BeNil())
		Expect(resp.Id).To(Equal(myResult["tenant"]))
	})
})
