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

package indykite_test

import (
	"context"

	"github.com/golang/mock/gomock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/indykite/indykite-sdk-go/config"
	configm "github.com/indykite/indykite-sdk-go/test/config/v1beta1"

	"github.com/indykite/terraform-provider-indykite/indykite"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Provider Bookmarks", func() {
	var (
		mockCtrl         *gomock.Controller
		mockConfigClient *configm.MockConfigManagementAPIClient
		clientContext    *indykite.ClientContext
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(TerraformGomockT(GinkgoT()))
		mockConfigClient = configm.NewMockConfigManagementAPIClient(mockCtrl)

		p := indykite.Provider()
		cfgFunc := p.ConfigureContextFunc
		p.ConfigureContextFunc =
			func(ctx context.Context, data *schema.ResourceData) (interface{}, diag.Diagnostics) {
				client, _ := config.NewTestClient(ctx, mockConfigClient)
				ctx = indykite.WithClient(ctx, client)
				return cfgFunc(ctx, data)
			}

		meta, _ := p.ConfigureContextFunc(context.Background(), nil)
		clientContext = meta.(*indykite.ClientContext)
	})

	It("Verify adding and retrieving bookmarks", func() {
		clientContext.AddBookmarks("", "")

		bms := clientContext.GetBookmarks()
		Expect(bms).To(HaveCap(0))
		Expect(bms).To(BeEmpty())

		By("Adding 5 bookmarks")
		clientContext.AddBookmarks("a")
		clientContext.AddBookmarks("b")
		clientContext.AddBookmarks("c", "d", "", "e")

		By("Read all 5 bookmarks - order doesn't matter")
		bms = clientContext.GetBookmarks()
		Expect(bms).To(HaveCap(5))
		Expect(bms).To(ConsistOf("a", "b", "c", "d", "e"))

		By("Add 5 more bookmarks")
		clientContext.AddBookmarks("v", "w")
		clientContext.AddBookmarks("x")
		clientContext.AddBookmarks("y")
		clientContext.AddBookmarks("z")

		By("Read all 10 bookmarks - all should be here and order doesn't matter")
		bms = clientContext.GetBookmarks()
		Expect(bms).To(HaveCap(10))
		Expect(bms).To(ConsistOf("a", "b", "c", "d", "e", "v", "w", "x", "y", "z"))

		By("Add 3 more bookmarks")
		clientContext.AddBookmarks("1", "2", "3")

		By("Read all 10 bookmarks - oldest should be gone and order doesn't matter")
		bms = clientContext.GetBookmarks()
		Expect(bms).To(HaveCap(10))
		Expect(bms).To(ConsistOf("1", "2", "3", "d", "e", "v", "w", "x", "y", "z"))
	})
})
