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

package indykite_test

import (
	"context"
	"fmt"
	"regexp"
	"strconv"

	"github.com/golang/mock/gomock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/indykite/jarvis-sdk-go/config"
	configpb "github.com/indykite/jarvis-sdk-go/gen/indykite/config/v1beta1"
	objects "github.com/indykite/jarvis-sdk-go/gen/indykite/objects/v1beta1"
	configm "github.com/indykite/jarvis-sdk-go/test/config/v1beta1"
	"github.com/onsi/gomega/types"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/indykite/terraform-provider-indykite/indykite"
	"github.com/indykite/terraform-provider-indykite/indykite/test"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

func matchEmail(address, name string) types.GomegaMatcher {
	return PointTo(MatchFields(IgnoreExtras, Fields{
		"Address": Equal(address),
		"Name":    Equal(name),
	}))
}

var _ = Describe("Resource Email Notification", func() {
	const resourceName = "indykite_email_notification.wonka"
	var (
		mockCtrl                *gomock.Controller
		mockConfigClient        *configm.MockConfigManagementAPIClient
		indykiteProviderFactory func() (*schema.Provider, error)

		// gid:/customer/1/appSpace/1/tenant/1/mail/1
		mailConfIDForTenant = "gid:L2N1c3RvbWVyLzEvYXBwU3BhY2UvMS90ZW5hbnQvMS9tYWlsLzE"
		// gid:/customer/1/mail/1
		mailConfIDForCustomer = "gid:L2N1c3RvbWVyLzEvbWFpbC8x"
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(TerraformGomockT(GinkgoT()))
		mockConfigClient = configm.NewMockConfigManagementAPIClient(mockCtrl)

		indykiteProviderFactory = func() (*schema.Provider, error) {
			p := indykite.Provider()
			cfgFunc := p.ConfigureContextFunc
			p.ConfigureContextFunc =
				func(ctx context.Context, data *schema.ResourceData) (interface{}, diag.Diagnostics) {
					client, _ := config.NewTestClient(ctx, mockConfigClient)
					ctx = context.WithValue(ctx, indykite.ClientContext, client)
					return cfgFunc(ctx, data)
				}
			return p, nil
		}
	})

	It("Test all CRUD", func() {
		fullTemplateMsg := &configpb.EmailDefinition_Template{Template: &configpb.EmailTemplate{
			TemplateId:      "MTID-2",
			TemplateVersion: wrapperspb.String("v3"),
			From:            &configpb.Email{Address: "wonka@chocolate-factory.com", Name: "Willy Wonka"},
			ReplyTo:         &configpb.Email{Address: "oompa@chocolate-factory.com", Name: "Oompa Loompa"},
			To: []*configpb.Email{
				{Address: "customer@example.com", Name: "John Doe"},
				{Address: "customer-2@example.com", Name: "John Doe"},
			},
			Cc: []*configpb.Email{
				{Address: "customer-3@example.com", Name: "Will Doe"},
				{Address: "another@example.com", Name: "Jane Roe"},
			},
			Bcc: []*configpb.Email{
				{Address: "hidden@example.com", Name: "Secret Doe"},
				{Address: "customer-hidden@example.com", Name: "Secret Roe"},
			},
			Subject: "Subject of the message",
			Headers: map[string]string{
				"SomeHeader":          "a",
				"X-Mailgun-Variables": `{"user-id": "Mailgun accept JSON in headers as variables"}`,
			},
			CustomArgs:            map[string]string{"arg1": "val1", "arg2": "val2"},
			DynamicTemplateValues: map[string]*objects.Value{"a": objects.Bool(true), "b": objects.Float64(159)},
			Categories:            []string{"a", "b", "c"},
			EventPayload:          wrapperspb.String("abc_def"),
			TemplateArn:           "SES_ARN_number",
		}}
		fullEmailConfigResp := &configpb.ReadConfigNodeResponse{
			ConfigNode: &configpb.ConfigNode{
				CustomerId:  customerID,
				Id:          mailConfIDForCustomer,
				Name:        "wonka-email-service",
				DisplayName: "Wonka ChocoEmail Factory",
				Description: wrapperspb.String("Description of the best ChocoMail service by Wonka inc."),
				CreateTime:  timestamppb.Now(),
				UpdateTime:  timestamppb.Now(),
				Config: &configpb.ConfigNode_EmailServiceConfig{
					EmailServiceConfig: &configpb.EmailServiceConfig{
						DefaultFromAddress: &configpb.Email{
							Address: "default-wonka@chocolate-factory.com",
							Name:    "Willy Wonka default",
						},
						Provider: &configpb.EmailServiceConfig_Sendgrid{Sendgrid: &configpb.SendGridProviderConfig{
							ApiKey:      "sendgrid_access_key_id_for_wonka",
							SandboxMode: true,
							IpPoolName:  wrapperspb.String("oompa_pool"),
							Host:        wrapperspb.String("https://wonka.sengrid.com"),
						}},
						InvitationMessage:    &configpb.EmailDefinition{Email: fullTemplateMsg},
						ResetPasswordMessage: &configpb.EmailDefinition{Email: fullTemplateMsg},
						VerificationMessage:  &configpb.EmailDefinition{Email: fullTemplateMsg},
					},
				},
			},
		}

		minimalEmailConfigResp := &configpb.ReadConfigNodeResponse{
			ConfigNode: &configpb.ConfigNode{
				CustomerId:  customerID,
				Id:          fullEmailConfigResp.ConfigNode.Id,
				Name:        fullEmailConfigResp.ConfigNode.Name,
				DisplayName: fullEmailConfigResp.ConfigNode.DisplayName,
				CreateTime:  fullEmailConfigResp.ConfigNode.CreateTime,
				UpdateTime:  timestamppb.Now(),
				Config: &configpb.ConfigNode_EmailServiceConfig{
					EmailServiceConfig: &configpb.EmailServiceConfig{
						Provider: &configpb.EmailServiceConfig_Amazon{Amazon: &configpb.AmazonSESProviderConfig{
							AccessKeyId:     "amazon_wonka_factory_access_key",
							SecretAccessKey: "amazon_wonka_factory_secret_key",
							Region:          "eu-north-1",
						}},
						InvitationMessage: &configpb.EmailDefinition{Email: &configpb.EmailDefinition_Template{
							Template: &configpb.EmailTemplate{
								TemplateId: "MTID-2",
							},
						}},
					},
				},
			},
		}

		withTenantLocEmailConfigResp := &configpb.ReadConfigNodeResponse{
			ConfigNode: &configpb.ConfigNode{
				CustomerId:  customerID,
				AppSpaceId:  appSpaceID,
				TenantId:    tenantID,
				Id:          mailConfIDForTenant,
				Name:        fullEmailConfigResp.ConfigNode.Name,
				DisplayName: fullEmailConfigResp.ConfigNode.DisplayName,
				CreateTime:  timestamppb.Now(),
				UpdateTime:  timestamppb.Now(),
				Config: &configpb.ConfigNode_EmailServiceConfig{
					EmailServiceConfig: &configpb.EmailServiceConfig{
						Provider: &configpb.EmailServiceConfig_Amazon{Amazon: &configpb.AmazonSESProviderConfig{
							AccessKeyId:          "amazon_wonka_factory_access_key",
							SecretAccessKey:      "amazon_wonka_factory_secret_key",
							Region:               "eu-north-1",
							ConfigurationSetName: "just_a_ses_config_name",
							DefaultFromAddress:   &configpb.Email{Address: "ses@example.com"},
							ReplyToAddresses:     []string{"wonka@choco-factory.com", "oompa@choco-factory.com"},

							FeedbackForwardingEmailAddress: "feedback@example.com",
						}},
						InvitationMessage: &configpb.EmailDefinition{Email: &configpb.EmailDefinition_Template{
							Template: &configpb.EmailTemplate{
								TemplateId: "MTID-2",
							},
						}},
					},
				},
			},
		}

		// MOCKS
		// There are 3 test steps
		// 1. step call: Create + Read
		// 2. step call: Read, Update, Read
		// 3. step call is recreate: Read, Delete, Create and Read
		// after steps Delete is called

		// Create
		templateMatch := PointTo(MatchFields(IgnoreExtras, Fields{"Template": PointTo(MatchFields(IgnoreExtras, Fields{
			"TemplateId":      Equal("MTID-2"),
			"TemplateVersion": PointTo(MatchFields(IgnoreExtras, Fields{"Value": Equal("v3")})),
			"Subject":         Equal("Subject of the message"),
			"From":            matchEmail("wonka@chocolate-factory.com", "Willy Wonka"),
			"ReplyTo":         matchEmail("oompa@chocolate-factory.com", "Oompa Loompa"),
			"To": ContainElements(
				matchEmail("customer@example.com", "John Doe"),
				matchEmail("customer-2@example.com", "John Doe"),
			),
			"Cc": ContainElements(
				matchEmail("customer-3@example.com", "Will Doe"),
				matchEmail("another@example.com", "Jane Roe"),
			),
			"Bcc": ContainElements(
				matchEmail("hidden@example.com", "Secret Doe"),
				matchEmail("customer-hidden@example.com", "Secret Roe"),
			),
			"Headers": MatchAllKeys(Keys{
				"SomeHeader":          Equal("a"),
				"X-Mailgun-Variables": ContainSubstring(`"user-id"`),
			}),
			"CustomArgs": MatchAllKeys(Keys{
				"arg1": Equal("val1"),
				"arg2": ContainSubstring("val2"),
			}),
			"DynamicTemplateValues": MatchAllKeys(Keys{
				"a": PointTo(MatchFields(IgnoreExtras, Fields{"Value": PointTo(MatchFields(IgnoreExtras, Fields{
					"BoolValue": BeTrue(),
				}))})),
				"b": PointTo(MatchFields(IgnoreExtras, Fields{"Value": PointTo(MatchFields(IgnoreExtras, Fields{
					"DoubleValue": BeEquivalentTo(159),
				}))})),
			}),
			"Categories":   ContainElements("a", "b", "c"),
			"EventPayload": PointTo(MatchFields(IgnoreExtras, Fields{"Value": Equal("abc_def")})),
			"TemplateArn":  Equal("SES_ARN_number"),
		}))}))
		mockConfigClient.EXPECT().
			CreateConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
				"Name": Equal(fullEmailConfigResp.ConfigNode.Name),
				"DisplayName": PointTo(MatchFields(IgnoreExtras, Fields{
					"Value": Equal(fullEmailConfigResp.ConfigNode.DisplayName),
				})),
				"Description": PointTo(MatchFields(IgnoreExtras, Fields{
					"Value": Equal(fullEmailConfigResp.ConfigNode.Description.Value),
				})),
				"Location": Equal(customerID),
				"Config": PointTo(MatchFields(IgnoreExtras, Fields{
					"EmailServiceConfig": PointTo(MatchFields(IgnoreExtras, Fields{
						"DefaultFromAddress": matchEmail("default-wonka@chocolate-factory.com", "Willy Wonka default"),
						"Provider": PointTo(MatchFields(IgnoreExtras, Fields{
							"Sendgrid": PointTo(MatchFields(IgnoreExtras, Fields{
								"ApiKey":      Equal("sendgrid_access_key_id_for_wonka"),
								"SandboxMode": BeTrue(),
								"IpPoolName":  PointTo(MatchFields(IgnoreExtras, Fields{"Value": Equal("oompa_pool")})),
								"Host": PointTo(MatchFields(IgnoreExtras, Fields{
									"Value": Equal("https://wonka.sengrid.com"),
								})),
							})),
						})),
						"InvitationMessage":    PointTo(MatchFields(IgnoreExtras, Fields{"Email": templateMatch})),
						"ResetPasswordMessage": PointTo(MatchFields(IgnoreExtras, Fields{"Email": templateMatch})),
						"VerificationMessage":  PointTo(MatchFields(IgnoreExtras, Fields{"Email": templateMatch})),
					}))},
				)),
			})))).
			Return(&configpb.CreateConfigNodeResponse{
				Id:         fullEmailConfigResp.ConfigNode.Id,
				CreateTime: timestamppb.Now(),
				UpdateTime: timestamppb.Now(),
			}, nil)

		mockConfigClient.EXPECT().
			CreateConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
				"Name":     Equal(withTenantLocEmailConfigResp.ConfigNode.Name),
				"Location": Equal(tenantID),
			})))).
			Return(&configpb.CreateConfigNodeResponse{
				Id:         withTenantLocEmailConfigResp.ConfigNode.Id,
				CreateTime: timestamppb.Now(),
				UpdateTime: timestamppb.Now(),
			}, nil)

		// update
		mockConfigClient.EXPECT().
			UpdateConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
				"Id":          Equal(fullEmailConfigResp.ConfigNode.Id),
				"DisplayName": BeNil(),
				"Description": PointTo(MatchFields(IgnoreExtras, Fields{"Value": BeEmpty()})),
				"Config": PointTo(MatchFields(IgnoreExtras, Fields{
					"EmailServiceConfig": PointTo(MatchFields(IgnoreExtras, Fields{
						"DefaultFromAddress": BeNil(),
						"Provider": PointTo(MatchFields(IgnoreExtras, Fields{
							"Amazon": PointTo(MatchFields(IgnoreExtras, Fields{
								"AccessKeyId":     Equal("amazon_wonka_factory_access_key"),
								"SecretAccessKey": Equal("amazon_wonka_factory_secret_key"),
								"Region":          Equal("eu-north-1"),
							})),
						})),
						"InvitationMessage": PointTo(MatchFields(IgnoreExtras, Fields{
							"Email": PointTo(MatchFields(IgnoreExtras, Fields{
								"Template": PointTo(MatchFields(IgnoreExtras, Fields{
									"TemplateId":            Equal("MTID-2"),
									"TemplateVersion":       BeNil(),
									"Subject":               BeEmpty(),
									"From":                  BeNil(),
									"ReplyTo":               BeNil(),
									"To":                    BeNil(),
									"Cc":                    BeNil(),
									"Bcc":                   BeNil(),
									"Headers":               BeNil(),
									"CustomArgs":            BeNil(),
									"DynamicTemplateValues": BeNil(),
									"Categories":            BeNil(),
									"EventPayload":          BeNil(),
									"TemplateArn":           BeEmpty(),
								}))}))})),
						"ResetPasswordMessage": BeNil(),
						"VerificationMessage":  BeNil(),
					}))},
				)),
			})))).
			Return(&configpb.UpdateConfigNodeResponse{Id: fullEmailConfigResp.ConfigNode.Id}, nil)

		// Read in given order
		gomock.InOrder(
			mockConfigClient.EXPECT().
				ReadConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
					"Id": Equal(fullEmailConfigResp.ConfigNode.Id),
				})))).
				Times(4).
				Return(fullEmailConfigResp, nil),

			mockConfigClient.EXPECT().
				ReadConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
					"Id": Equal(fullEmailConfigResp.ConfigNode.Id),
				})))).
				Times(3).
				Return(minimalEmailConfigResp, nil),

			mockConfigClient.EXPECT().
				ReadConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
					"Id": Equal(withTenantLocEmailConfigResp.ConfigNode.Id),
				})))).
				Times(2).
				Return(withTenantLocEmailConfigResp, nil),
		)

		// Delete
		mockConfigClient.EXPECT().
			DeleteConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
				"Id": Equal(fullEmailConfigResp.ConfigNode.Id),
			})))).
			Return(&configpb.DeleteConfigNodeResponse{}, nil)

		mockConfigClient.EXPECT().
			DeleteConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
				"Id": Equal(withTenantLocEmailConfigResp.ConfigNode.Id),
			})))).
			Return(&configpb.DeleteConfigNodeResponse{}, nil)

		resource.Test(GinkgoT(), resource.TestCase{
			ProviderFactories: map[string]func() (*schema.Provider, error){
				"indykite": indykiteProviderFactory,
			},
			Steps: []resource.TestStep{
				// Error cases should be always first, easier to avoid missing mocks or incomplete plan
				{
					Config: `resource "indykite_email_notification" "wonka" {
						name = "wonka-email-service"
						sendgrid{
							api_key = "sendgrid_access_key_id_for_wonka"
						}
					}
					`,
					ExpectError: regexp.MustCompile(`"location" is required, but no definition was found`),
				},
				{
					Config: `resource "indykite_email_notification" "wonka" {
  						location = "` + customerID + `"
  						name = "wonka-email-service"
					}
					`,
					ExpectError: regexp.MustCompile("one of `amazon_ses,sendgrid` must be specified"),
				},
				{
					Config: `resource "indykite_email_notification" "wonka" {
						location = "` + customerID + `"
						customer_id = "69857924-098e-4f11-8800-62b10bb188ea"
						name = "wonka-email-service"
						sendgrid{
							api_key = "sendgrid_access_key_id_for_wonka"
						}
					}
					`,
					ExpectError: regexp.MustCompile(
						`Can't configure a value for "customer_id": its value will be decided`),
				},
				{
					Config: `resource "indykite_email_notification" "wonka" {
  						location = "` + customerID + `"
  						name = "wonka-email-service"
						sendgrid{
							api_key = "sendgrid_access_key_id_for_wonka"
						}
						default_from_address {
							address = "cc"
						}
					}
					`,
					ExpectError: regexp.MustCompile("Value is not valid email address"),
				},
				{
					Config: `resource "indykite_email_notification" "wonka" {
  						location = "` + customerID + `"
  						name = "wonka-email-service"
						amazon_ses {
						}
					}
					`,
					// nolint:lll
					ExpectError: regexp.MustCompile(`((?s)("access_key_id" is required.*)|("secret_access_key" is required.*)|("region" is required.*)){3}`),
				},
				{
					Config: `resource "indykite_email_notification" "wonka" {
  						location = "` + customerID + `"
  						name = "wonka-email-service"
						sendgrid {
						}
					}
					`,
					ExpectError: regexp.MustCompile(`"api_key" is required`),
				},
				{
					Config: `resource "indykite_email_notification" "wonka" {
  						location = "` + customerID + `"
  						name = "wonka-email-service"
						sendgrid {
							api_key = "sendgrid_access_key_id_for_wonka"
						}
						invitation_message {
							template {
								id = "T1ID"
							}
							template {
								id = "T2ID"
							}
						}
					}
					`,
					// There is update of Terraform which produce 'Too many template blocks error'
					// or 'Attribute supports 1 item maximum, but config has 2 declared'
					// but older version still produce 'List longer than MaxItems'
					// nolint:lll
					ExpectError: regexp.MustCompile(`Too many template blocks|List longer than MaxItems|Attribute supports 1 item maximum`),
				},
				{
					Config: `resource "indykite_email_notification" "wonka" {
  						location = "` + customerID + `"
  						name = "wonka-email-service"
						sendgrid {
							api_key = "sendgrid_access_key_id_for_wonka"
						}
						invitation_message {
							template {
								id = "T1ID"
								template_dynamic_values = "\"value\": 123"
							}
						}
					}
					`,
					ExpectError: regexp.MustCompile(`invalid JSON: invalid character ':' after top-level value`),
				},
				{
					Config: `resource "indykite_email_notification" "wonka" {
  						location = "` + customerID + `"
  						name = "wonka-email-service"
						sendgrid {
							api_key = "sendgrid_access_key_id_for_wonka"
						}
						invitation_message {
							template {
								id = "T1ID"
								template_dynamic_values = "[\"value\", 123]"
							}
						}
					}
					`,
					ExpectError: regexp.MustCompile(`cannot build dynamic template values from JSON`),
				},
				{
					// Checking Create and Read (fullEmailConfigResp)
					Config: getFullEmailNotificationConfig(),
					Check: resource.ComposeTestCheckFunc(
						testEmailNotificationResourceDataExists(resourceName, fullEmailConfigResp),
					),
				},
				{
					// Performs 1 read (fullEmailConfigResp)
					ResourceName:  resourceName,
					ImportState:   true,
					ImportStateId: fullEmailConfigResp.ConfigNode.Id,
				},
				{
					// Checking Read(fullEmailConfigResp), Update and Read(minimalEmailConfigResp)
					Config: getMinimalEmailNotificationConfig(false),
					Check: resource.ComposeTestCheckFunc(
						testEmailNotificationResourceDataExists(resourceName, minimalEmailConfigResp),
					),
				},
				{
					// Checking ForceNew on location change
					// Read(minimalEmailConfigResp), Delete, Create and Read(withTenantLocEmailConfigResp)
					Config: getMinimalEmailNotificationConfig(true),
					Check: resource.ComposeTestCheckFunc(
						testEmailNotificationResourceDataExists(resourceName, withTenantLocEmailConfigResp),
					),
				},
			},
		})
	})
})

func testEmailNotificationResourceDataExists(n string, data *configpb.ReadConfigNodeResponse) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}
		if rs.Primary.ID != data.ConfigNode.Id {
			return fmt.Errorf("ID does not match")
		}
		attrs := rs.Primary.Attributes
		if v, has := attrs["name"]; !has || v != data.ConfigNode.Name {
			return fmt.Errorf("invalid name: %s", v)
		}
		if v, has := attrs["display_name"]; !has || v != data.ConfigNode.DisplayName {
			return fmt.Errorf("invalid display name: %s", v)
		}
		if v, has := attrs["description"]; !has || v != data.ConfigNode.Description.GetValue() {
			return fmt.Errorf("invalid description: %s", v)
		}

		if v, has := attrs["customer_id"]; !has || v != data.ConfigNode.CustomerId {
			return fmt.Errorf("invalid customer_id: %s", v)
		}
		if v, has := attrs["app_space_id"]; !has || v != data.ConfigNode.AppSpaceId {
			return fmt.Errorf("invalid app_space_id: %s", v)
		}
		if v, has := attrs["tenant_id"]; !has || v != data.ConfigNode.TenantId {
			return fmt.Errorf("invalid tenant_id: %s", v)
		}

		var err error
		mailConf := data.ConfigNode.GetEmailServiceConfig()
		if err = testEmailSchemaData(attrs, mailConf.DefaultFromAddress, "default_from_address"); err != nil {
			return err
		}
		switch p := mailConf.Provider.(type) {
		case *configpb.EmailServiceConfig_Amazon:
			err = testAmazonProviderData(attrs, p)
		case *configpb.EmailServiceConfig_Sendgrid:
			err = testSendgridProviderData(attrs, p)
		}
		if err != nil {
			return err
		}

		return testMailMessageData(attrs, mailConf.ResetPasswordMessage, "reset_password_message")
	}
}

func testEmailSchemaData(attrs map[string]string, data *configpb.Email, key string) error {
	if data == nil {
		return testEmailArraySchemaData(attrs, nil, key)
	}
	return testEmailArraySchemaData(attrs, []*configpb.Email{data}, key)
}

func testEmailArraySchemaData(attrs map[string]string, data []*configpb.Email, key string) error {
	cnt, _ := strconv.Atoi(attrs[key+".#"])
	if cnt != len(data) {
		return fmt.Errorf("under key %s got %d email schemas, expected %d", key, len(data), cnt)
	}
	for i := 0; i < cnt; i++ {
		curKey := fmt.Sprintf("%s.%d.%s", key, i, "address")
		if v, has := attrs[curKey]; !has || v != data[i].Address {
			return fmt.Errorf("invalid %s: %s", curKey, v)
		}
		curKey = fmt.Sprintf("%s.%d.%s", key, i, "name")
		if v, has := attrs[curKey]; !has || v != data[i].Name {
			return fmt.Errorf("invalid %s: %s", curKey, v)
		}
	}
	return nil
}

func testAmazonProviderData(attrs map[string]string, provider *configpb.EmailServiceConfig_Amazon) error {
	if v, has := attrs["amazon_ses.0.access_key_id"]; !has || v != provider.Amazon.AccessKeyId {
		return fmt.Errorf("invalid amazon_ses.0.access_key_id: %s", v)
	}
	if v, has := attrs["amazon_ses.0.secret_access_key"]; !has || v != provider.Amazon.SecretAccessKey {
		return fmt.Errorf("invalid amazon_ses.0.secret_access_key: %s", v)
	}
	if v, has := attrs["amazon_ses.0.region"]; !has || v != provider.Amazon.Region {
		return fmt.Errorf("invalid amazon_ses.0.region: %s", v)
	}

	if v, has := attrs["amazon_ses.0.configuration_set_name"]; !has || v != provider.Amazon.ConfigurationSetName {
		return fmt.Errorf("invalid amazon_ses.0.configuration_set_name: %s", v)
	}
	v, has := attrs["amazon_ses.0.feedback_forwarding_email_address"]
	if !has || v != provider.Amazon.FeedbackForwardingEmailAddress {
		return fmt.Errorf("invalid amazon_ses.0.feedback_forwarding_email_address: %s", v)
	}
	err := testEmailSchemaData(attrs, provider.Amazon.DefaultFromAddress, "amazon_ses.0.default_from_address")
	if err != nil {
		return err
	}
	err = testStringArraySchemaData(attrs, "amazon_ses.0.reply_to_addresses", provider.Amazon.ReplyToAddresses)
	if err != nil {
		return err
	}

	return nil
}

func testSendgridProviderData(attrs map[string]string, provider *configpb.EmailServiceConfig_Sendgrid) error {
	if v, has := attrs["sendgrid.0.api_key"]; !has || v != provider.Sendgrid.ApiKey {
		return fmt.Errorf("invalid sendgrid.0.api_key: %s", v)
	}
	if v, has := attrs["sendgrid.0.host"]; !has || v != provider.Sendgrid.Host.GetValue() {
		return fmt.Errorf("invalid sendgrid.0.host: %s", v)
	}
	if v, has := attrs["sendgrid.0.ip_pool_name"]; !has || v != provider.Sendgrid.IpPoolName.GetValue() {
		return fmt.Errorf("invalid sendgrid.0.ip_pool_name: %s", v)
	}
	if v, has := attrs["sendgrid.0.sandbox_mode"]; !has || v != strconv.FormatBool(provider.Sendgrid.SandboxMode) {
		return fmt.Errorf("invalid sendgrid.0.sandbox_mode: %s", v)
	}
	return nil
}

func testMailMessageData(attrs map[string]string, data *configpb.EmailDefinition, key string) error {
	if data == nil {
		return nil
	}
	tpl := data.GetTemplate()
	key += ".0"
	if err := testEmailSchemaData(attrs, tpl.From, key+".from"); err != nil {
		return err
	}
	if err := testEmailSchemaData(attrs, tpl.ReplyTo, key+".reply_to"); err != nil {
		return err
	}
	if err := testEmailArraySchemaData(attrs, tpl.To, key+".to"); err != nil {
		return err
	}
	if err := testEmailArraySchemaData(attrs, tpl.Cc, key+".cc"); err != nil {
		return err
	}
	if err := testEmailArraySchemaData(attrs, tpl.Bcc, key+".bcc"); err != nil {
		return err
	}
	if v, has := attrs[key+".subject"]; !has || v != tpl.Subject {
		return fmt.Errorf("invalid %s.subject: %s", key, v)
	}

	tplKey := key + ".template.0"
	if v, has := attrs[tplKey+".id"]; !has || v != tpl.TemplateId {
		return fmt.Errorf("invalid %s.id: %s", tplKey, v)
	}
	if v, has := attrs[tplKey+".version"]; !has || v != tpl.TemplateVersion.GetValue() {
		return fmt.Errorf("invalid %s.version: %s", tplKey, v)
	}
	if v, has := attrs[tplKey+".event_payload"]; !has || v != tpl.EventPayload.GetValue() {
		return fmt.Errorf("invalid %s.event_payload: %s", tplKey, v)
	}
	if v, has := attrs[tplKey+".ses_arn"]; !has || v != tpl.TemplateArn {
		return fmt.Errorf("invalid %s.ses_arn: %s", tplKey, v)
	}
	err := testStringArraySchemaData(attrs, tplKey+".categories", tpl.Categories)
	if err != nil {
		return err
	}

	err = testStringMapSchemaData(attrs, tplKey+".headers", tpl.Headers)
	if err != nil {
		return err
	}
	err = testStringMapSchemaData(attrs, tplKey+".custom_arguments", tpl.CustomArgs)
	if err != nil {
		return err
	}
	return nil
}

func getFullEmailNotificationConfig() string {
	messageDef := `
		from {
			address = "wonka@chocolate-factory.com"
			name = "Willy Wonka"
		}
		reply_to {
			address = "oompa@chocolate-factory.com"
			name = "Oompa Loompa"
		}

		to {
			address = "customer@example.com"
			name = "John Doe"
		}
		to {
			address = "customer-2@example.com"
			name = "John Doe"
		}

		cc {
			address = "customer-3@example.com"
			name = "Will Doe"
		}
		cc {
			address = "another@example.com"
			name = "Jane Roe"
		}
		bcc {
			address = "hidden@example.com"
			name = "Secret Doe"
		}
		bcc {
			address = "customer-hidden@example.com"
			name = "Secret Roe"
		}
		subject = "Subject of the message"

		template {
			id = "MTID-2"
			version = "v3"

			headers = {
				SomeHeader = "a"
				"X-Mailgun-Variables" = "{\"user-id\": \"Mailgun accept JSON in headers as variables\"}"
			}

			custom_arguments = {
				arg1 = "val1"
				arg2 = "val2"
			}
			template_dynamic_values = "{ \"a\": true, \"b\": 159 }"
			categories = ["a", "b", "c"]
			event_payload = "abc_def"
			ses_arn = "SES_ARN_number"
		}`

	return fmt.Sprintf(`
		resource "indykite_email_notification" "wonka" {
			location = "`+customerID+`"
			name = "wonka-email-service"
			display_name = "Wonka ChocoEmail Factory"
			description = "Description of the best ChocoMail service by Wonka inc."

			default_from_address {
				address = "default-wonka@chocolate-factory.com"
				name = "Willy Wonka default"
			}

			sendgrid {
				api_key = "sendgrid_access_key_id_for_wonka"
				sandbox_mode = true
				ip_pool_name = "oompa_pool"
				host = "https://wonka.sengrid.com"
			}

			invitation_message {
				%s
			}

			reset_password_message {
				%s
			}

			email_verification_message {
				%s
			}
		}
	`, messageDef, messageDef, messageDef)
}

func getMinimalEmailNotificationConfig(forTenantAndFullSES bool) string {
	locationLine := `location = "` + customerID + `"`
	sesExtraLines := ""
	if forTenantAndFullSES {
		locationLine = `location = "` + tenantID + `"`
		sesExtraLines = `configuration_set_name =  "just_a_ses_config_name"
						default_from_address {
							address = "ses@example.com"
						}
						feedback_forwarding_email_address = "feedback@example.com"
						reply_to_addresses = ["wonka@choco-factory.com", "oompa@choco-factory.com"]`
	}
	return fmt.Sprintf(`
		resource "indykite_email_notification" "wonka" {
			%s
			name = "wonka-email-service"
			display_name = "Wonka ChocoEmail Factory"

			amazon_ses {
				access_key_id = "amazon_wonka_factory_access_key"
				secret_access_key = "amazon_wonka_factory_secret_key"
				region = "eu-north-1"
				%s
			}

			invitation_message {
				template {
					id = "MTID-2"
				}
			}
		}
	`, locationLine, sesExtraLines)
}
