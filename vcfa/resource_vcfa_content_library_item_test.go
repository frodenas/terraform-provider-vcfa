//go:build tm || contentlibrary || ALL || functional

package vcfa

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcfaContentLibraryItem(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	vCenterHcl, vCenterHclRef := getVCenterHcl(t)
	nsxManagerHcl, nsxManagerHclRef := getNsxManagerHcl(t)
	regionHcl, regionHclRef := getRegionHcl(t, vCenterHclRef, nsxManagerHclRef)
	contentLibraryHcl, contentLibraryHclRef := getContentLibraryHcl(t, regionHclRef)

	var params = StringMap{
		"Name":              t.Name(),
		"ContentLibraryRef": fmt.Sprintf("%s.id", contentLibraryHclRef),
		"OvaPath":           "../test-resources/test_vapp_template.ova",
		"Tags":              "tm contentlibrary",
	}
	testParamsNotEmpty(t, params)

	preRequisites := vCenterHcl + nsxManagerHcl + regionHcl + contentLibraryHcl

	configText1 := templateFill(preRequisites+testAccVcfaContentLibraryItemStep1, params)
	params["FuncName"] = t.Name() + "-step2"
	params["Name"] = t.Name() + "Updated"
	configText2 := templateFill(preRequisites+testAccVcfaContentLibraryItemStep1, params)
	params["FuncName"] = t.Name() + "-step3"
	configText3 := templateFill(preRequisites+testAccVcfaContentLibraryItemStep3, params)

	debugPrintf("#[DEBUG] CONFIGURATION step1: %s\n", configText1)
	debugPrintf("#[DEBUG] CONFIGURATION step2: %s\n", configText2)
	debugPrintf("#[DEBUG] CONFIGURATION step3: %s\n", configText3)
	if vcfaShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resourceName := "vcfa_content_library_item.cli"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", t.Name()),
					resource.TestCheckResourceAttr(resourceName, "description", t.Name()),
					resource.TestCheckResourceAttrPair(resourceName, "content_library_id", contentLibraryHclRef, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "creation_date"),
					resource.TestCheckResourceAttr(resourceName, "is_subscribed", "false"),
					resource.TestCheckResourceAttr(resourceName, "is_published", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "image_identifier"),
					resource.TestCheckResourceAttr(resourceName, "item_type", "TEMPLATE"),
					resource.TestMatchResourceAttr(resourceName, "owner_org_id", regexp.MustCompile("urn:vcloud:org:")),
					resource.TestCheckResourceAttr(resourceName, "status", "READY"),
					resource.TestCheckResourceAttr(resourceName, "last_successful_sync", ""),
					resource.TestCheckResourceAttr(resourceName, "version", "1"),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", t.Name()+"Updated"),
					resource.TestCheckResourceAttr(resourceName, "description", t.Name()+"Updated"),
					resource.TestCheckResourceAttrPair(resourceName, "content_library_id", contentLibraryHclRef, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "creation_date"),
					resource.TestCheckResourceAttr(resourceName, "is_subscribed", "false"),
					resource.TestCheckResourceAttr(resourceName, "is_published", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "image_identifier"),
					resource.TestCheckResourceAttr(resourceName, "item_type", "TEMPLATE"),
					resource.TestMatchResourceAttr(resourceName, "owner_org_id", regexp.MustCompile("urn:vcloud:org:")),
					resource.TestCheckResourceAttr(resourceName, "status", "READY"),
					resource.TestCheckResourceAttr(resourceName, "last_successful_sync", ""),
					resource.TestCheckResourceAttr(resourceName, "version", "1"),
				),
			},
			{
				Config: configText3,
				Check: resource.ComposeAggregateTestCheckFunc(
					// file_path and upload_piece_size cannot be obtained during reads, that's why it does not appear in data source schema
					resourceFieldsEqual(resourceName, "data.vcfa_content_library_item.cli_ds", []string{"file_path", "upload_piece_size", "%"}),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateId:           fmt.Sprintf("%s%s%s", testConfig.Tm.ContentLibrary, ImportSeparator, params["Name"].(string)),
				ImportStateVerifyIgnore: []string{"file_path", "upload_piece_size"}, // file_path and upload_piece_size cannot be obtained during imports, that's why it's Optional
			},
		},
	})

	postTestChecks(t)
}

const testAccVcfaContentLibraryItemStep1 = `
resource "vcfa_content_library_item" "cli" {
  name               = "{{.Name}}"
  description        = "{{.Name}}"
  content_library_id = {{.ContentLibraryRef}}
  file_path          = "{{.OvaPath}}"
}
`

const testAccVcfaContentLibraryItemStep3 = testAccVcfaContentLibraryItemStep1 + `
data "vcfa_content_library_item" "cli_ds" {
  name               = vcfa_content_library_item.cli.name
  content_library_id = vcfa_content_library_item.cli.content_library_id
}
`
