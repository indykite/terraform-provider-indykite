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
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"sync"

	"github.com/golang/mock/gomock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/indykite/indykite-sdk-go/config"
	configpb "github.com/indykite/indykite-sdk-go/gen/indykite/config/v1beta1"
	objects "github.com/indykite/indykite-sdk-go/gen/indykite/objects/v1beta1"
	configm "github.com/indykite/indykite-sdk-go/test/config/v1beta1"
	"github.com/onsi/gomega/types"
	"github.com/pborman/uuid"
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
		mockCtrl         *gomock.Controller
		mockConfigClient *configm.MockConfigManagementAPIClient
		provider         *schema.Provider
		mockedBookmark   string

		// gid:/customer/1/appSpace/1/tenant/1/mail/1
		mailConfIDForTenant = "gid:L2N1c3RvbWVyLzEvYXBwU3BhY2UvMS90ZW5hbnQvMS9tYWlsLzE"
		// gid:/customer/1/mail/1
		mailConfIDForCustomer = "gid:L2N1c3RvbWVyLzEvbWFpbC8x"
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(TerraformGomockT(GinkgoT()))
		mockConfigClient = configm.NewMockConfigManagementAPIClient(mockCtrl)

		// Bookmark must be longer than 40 chars - have just 1 added before the first write to test all cases
		mockedBookmark = "for-email" + uuid.NewRandom().String()
		bmOnce := &sync.Once{}

		provider = indykite.Provider()
		cfgFunc := provider.ConfigureContextFunc
		provider.ConfigureContextFunc =
			func(ctx context.Context, data *schema.ResourceData) (interface{}, diag.Diagnostics) {
				client, _ := config.NewTestClient(ctx, mockConfigClient)
				ctx = indykite.WithClient(ctx, client)
				i, d := cfgFunc(ctx, data)
				// ConfigureContextFunc is called repeatedly, add initial bookmark just once
				bmOnce.Do(func() {
					i.(*indykite.ClientContext).AddBookmarks(mockedBookmark)
				})
				return i, d
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
						InvitationMessage:      &configpb.EmailDefinition{Email: fullTemplateMsg},
						ResetPasswordMessage:   &configpb.EmailDefinition{Email: fullTemplateMsg},
						VerificationMessage:    &configpb.EmailDefinition{Email: fullTemplateMsg},
						OneTimePasswordMessage: &configpb.EmailDefinition{Email: fullTemplateMsg},
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

		createBM := "created-email" + uuid.NewRandom().String()
		createBM2 := "created-email-2" + uuid.NewRandom().String()
		updateBM := "updated-email" + uuid.NewRandom().String()
		deleteBM := "deleted-email" + uuid.NewRandom().String()
		deleteBM2 := "deleted-email-2" + uuid.NewRandom().String()

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
						"InvitationMessage":      PointTo(MatchFields(IgnoreExtras, Fields{"Email": templateMatch})),
						"ResetPasswordMessage":   PointTo(MatchFields(IgnoreExtras, Fields{"Email": templateMatch})),
						"VerificationMessage":    PointTo(MatchFields(IgnoreExtras, Fields{"Email": templateMatch})),
						"OneTimePasswordMessage": PointTo(MatchFields(IgnoreExtras, Fields{"Email": templateMatch})),
					}))},
				)),
				"Bookmarks": ConsistOf(mockedBookmark),
			})))).
			Return(&configpb.CreateConfigNodeResponse{
				Id:         fullEmailConfigResp.ConfigNode.Id,
				CreateTime: timestamppb.Now(),
				UpdateTime: timestamppb.Now(),
				Bookmark:   createBM,
			}, nil)

		mockConfigClient.EXPECT().
			CreateConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
				"Name":      Equal(withTenantLocEmailConfigResp.ConfigNode.Name),
				"Location":  Equal(tenantID),
				"Bookmarks": ConsistOf(mockedBookmark, createBM, updateBM, deleteBM),
			})))).
			Return(&configpb.CreateConfigNodeResponse{
				Id:         withTenantLocEmailConfigResp.ConfigNode.Id,
				CreateTime: timestamppb.Now(),
				UpdateTime: timestamppb.Now(),
				Bookmark:   createBM2,
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
					})),
				})),
				"Bookmarks": ConsistOf(mockedBookmark, createBM),
			})))).
			Return(&configpb.UpdateConfigNodeResponse{
				Id:       fullEmailConfigResp.ConfigNode.Id,
				Bookmark: updateBM,
			}, nil)

		// Read in given order
		gomock.InOrder(
			mockConfigClient.EXPECT().
				ReadConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
					"Id":        Equal(fullEmailConfigResp.ConfigNode.Id),
					"Bookmarks": ConsistOf(mockedBookmark, createBM),
				})))).
				Times(4).
				Return(fullEmailConfigResp, nil),

			mockConfigClient.EXPECT().
				ReadConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
					"Id":        Equal(fullEmailConfigResp.ConfigNode.Id),
					"Bookmarks": ConsistOf(mockedBookmark, createBM, updateBM),
				})))).
				Times(3).
				Return(minimalEmailConfigResp, nil),

			mockConfigClient.EXPECT().
				ReadConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
					"Id":        Equal(withTenantLocEmailConfigResp.ConfigNode.Id),
					"Bookmarks": ConsistOf(mockedBookmark, createBM, updateBM, deleteBM, createBM2),
				})))).
				Times(2).
				Return(withTenantLocEmailConfigResp, nil),
		)

		// Delete
		mockConfigClient.EXPECT().
			DeleteConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
				"Id":        Equal(fullEmailConfigResp.ConfigNode.Id),
				"Bookmarks": ConsistOf(mockedBookmark, createBM, updateBM),
			})))).
			Return(&configpb.DeleteConfigNodeResponse{Bookmark: deleteBM}, nil)

		mockConfigClient.EXPECT().
			DeleteConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
				"Id":        Equal(withTenantLocEmailConfigResp.ConfigNode.Id),
				"Bookmarks": ConsistOf(mockedBookmark, createBM, updateBM, deleteBM, createBM2),
			})))).
			Return(&configpb.DeleteConfigNodeResponse{Bookmark: deleteBM2}, nil)

		resource.Test(GinkgoT(), resource.TestCase{
			Providers: map[string]*schema.Provider{
				"indykite": provider,
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
					//nolint:lll
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
					//nolint:lll
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
						testEmailNotificationResourceDataExists(resourceName, fullEmailConfigResp, nil),
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
						testEmailNotificationResourceDataExists(resourceName, minimalEmailConfigResp, Keys{
							// Those extra fields are sometimes required. Not sure why, but probably on first
							// apply it is present, when previously it wasn't empty. Next apply does not contain those.
							"invitation_message.0.template.0.categories.#":       Equal("0"),
							"invitation_message.0.template.0.custom_arguments.%": Equal("0"),
							"invitation_message.0.template.0.headers.%":          Equal("0"),
						}),
					),
				},
				{
					// Checking ForceNew on location change
					// Read(minimalEmailConfigResp), Delete, Create and Read(withTenantLocEmailConfigResp)
					Config: getMinimalEmailNotificationConfig(true),
					Check: resource.ComposeTestCheckFunc(
						testEmailNotificationResourceDataExists(resourceName, withTenantLocEmailConfigResp, nil),
					),
				},
			},
		})
	})
})

func testEmailNotificationResourceDataExists(
	n string,
	data *configpb.ReadConfigNodeResponse,
	extraKeys Keys,
) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}
		if rs.Primary.ID != data.ConfigNode.Id {
			return errors.New("ID does not match")
		}

		mailConf := data.ConfigNode.GetEmailServiceConfig()
		keys := Keys{
			"id": Equal(data.ConfigNode.Id),
			"%":  Not(BeEmpty()), // This is Terraform helper

			"location":     Not(BeEmpty()), // Response does not return this
			"customer_id":  Equal(data.ConfigNode.CustomerId),
			"app_space_id": Equal(data.ConfigNode.AppSpaceId),
			"tenant_id":    Equal(data.ConfigNode.TenantId),
			"name":         Equal(data.ConfigNode.Name),
			"display_name": Equal(data.ConfigNode.DisplayName),
			"description":  Equal(data.ConfigNode.GetDescription().GetValue()),
			"create_time":  Not(BeEmpty()),
			"update_time":  Not(BeEmpty()),
		}

		addEmailSchemaToKeys(keys, mailConf.DefaultFromAddress, "default_from_address")
		addMailMessageDataToKeys(keys, mailConf.InvitationMessage, "invitation_message")
		addMailMessageDataToKeys(keys, mailConf.ResetPasswordMessage, "reset_password_message")
		addMailMessageDataToKeys(keys, mailConf.VerificationMessage, "email_verification_message")
		addMailMessageDataToKeys(keys, mailConf.OneTimePasswordMessage, "one_time_password_message")

		testAmazonProviderData(keys, mailConf.GetAmazon())
		testSendgridProviderData(keys, mailConf.GetSendgrid())

		for k, v := range extraKeys {
			keys[k] = v
		}

		return convertOmegaMatcherToError(MatchAllKeys(keys), rs.Primary.Attributes)
	}
}

func addEmailSchemaToKeys(keys Keys, data *configpb.Email, key string) {
	if data == nil {
		addEmailArraySchemaToKeys(keys, nil, key)
	} else {
		addEmailArraySchemaToKeys(keys, []*configpb.Email{data}, key)
	}
}

func addEmailArraySchemaToKeys(keys Keys, data []*configpb.Email, key string) {
	keys[key+".#"] = Equal(strconv.Itoa(len(data)))
	for i, v := range data {
		currKey := fmt.Sprintf("%s.%d.", key, i)
		keys[currKey+"%"] = Not(BeEmpty())
		keys[currKey+"address"] = Equal(v.Address)
		keys[currKey+"name"] = Equal(v.Name)
	}
}

func testAmazonProviderData(keys Keys, provider *configpb.AmazonSESProviderConfig) {
	if provider == nil {
		keys["amazon_ses.#"] = Equal("0")
		return
	}
	keys["amazon_ses.#"] = Equal("1")

	keys["amazon_ses.0.%"] = Not(BeEmpty())
	keys["amazon_ses.0.access_key_id"] = Equal(provider.AccessKeyId)
	keys["amazon_ses.0.secret_access_key"] = Equal(provider.SecretAccessKey)
	keys["amazon_ses.0.region"] = Equal(provider.Region)
	keys["amazon_ses.0.configuration_set_name"] = Equal(provider.ConfigurationSetName)
	keys["amazon_ses.0.feedback_forwarding_email_address"] = Equal(provider.FeedbackForwardingEmailAddress)

	addEmailSchemaToKeys(keys, provider.DefaultFromAddress, "amazon_ses.0.default_from_address")
	addStringArrayToKeys(keys, "amazon_ses.0.reply_to_addresses", provider.ReplyToAddresses)
}

func testSendgridProviderData(keys Keys, provider *configpb.SendGridProviderConfig) {
	if provider == nil {
		keys["sendgrid.#"] = Equal("0")
		return
	}

	keys["sendgrid.#"] = Equal("1")
	keys["sendgrid.0.%"] = Not(BeEmpty())
	keys["sendgrid.0.api_key"] = Equal(provider.ApiKey)
	keys["sendgrid.0.host"] = Equal(provider.Host.GetValue())
	keys["sendgrid.0.ip_pool_name"] = Equal(provider.IpPoolName.GetValue())
	keys["sendgrid.0.sandbox_mode"] = Equal(strconv.FormatBool(provider.SandboxMode))
}

func addMailMessageDataToKeys(keys Keys, data *configpb.EmailDefinition, key string) {
	if data == nil {
		keys[key+".#"] = Equal("0")
		return
	}
	keys[key+".#"] = Equal("1")

	tpl := data.GetTemplate()
	key += ".0."
	keys[key+"%"] = Not(BeEmpty()) // Terraform helper
	addEmailSchemaToKeys(keys, tpl.From, key+"from")
	addEmailSchemaToKeys(keys, tpl.ReplyTo, key+"reply_to")
	addEmailArraySchemaToKeys(keys, tpl.To, key+"to")
	addEmailArraySchemaToKeys(keys, tpl.Cc, key+"cc")
	addEmailArraySchemaToKeys(keys, tpl.Bcc, key+"bcc")
	keys[key+"subject"] = Equal(tpl.Subject)

	keys[key+"template.#"] = Not(BeEmpty()) // Terraform helper
	tplKey := key + "template.0."
	keys[tplKey+"%"] = Not(BeEmpty()) // Terraform helper

	keys[tplKey+"id"] = Equal(tpl.TemplateId)
	keys[tplKey+"version"] = Equal(tpl.TemplateVersion.GetValue())
	keys[tplKey+"event_payload"] = Equal(tpl.EventPayload.GetValue())
	keys[tplKey+"ses_arn"] = Equal(tpl.TemplateArn)

	addStringArrayToKeys(keys, tplKey+"categories", tpl.Categories)
	addStringMapMatcherToKeys(keys, tplKey+"headers", tpl.Headers)
	addStringMapMatcherToKeys(keys, tplKey+"custom_arguments", tpl.CustomArgs)
	keys[tplKey+"template_dynamic_values"] = Not(BeNil())
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

			one_time_password_message {
				%s
			}
		}
	`, messageDef, messageDef, messageDef, messageDef)
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
