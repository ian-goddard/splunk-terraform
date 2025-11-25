package appvalidation_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/splunk/terraform-provider-scp/internal/acctest"
)

const validationDataSourceTemplate = `
data "scp_app_validation" %[1]q {
	request_id = %[1]q
}
`

func TestAcc_ScpAppValidation_DataSource_basic(t *testing.T) {
	requestID := "example-request-id"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(validationDataSourceTemplate, requestID),
				Check: resource.TestCheckResourceAttr(
					fmt.Sprintf("data.scp_app_validation.%s", requestID), "request_id", requestID,
				),
			},
		},
	})
}
