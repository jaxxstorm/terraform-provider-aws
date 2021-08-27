package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAwsDxConnection_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_dx_connection.test"
	datasourceName := "data.aws_dx_connection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { testAccPreCheck(t) },
		ErrorCheck: testAccErrorCheck(t, directconnect.EndpointsID),
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsDxConnectionConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "aws_device", resourceName, "aws_device"),
					resource.TestCheckResourceAttrPair(datasourceName, "bandwidth", resourceName, "bandwidth"),
					resource.TestCheckResourceAttrPair(datasourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(datasourceName, "location", resourceName, "location"),
					resource.TestCheckResourceAttrPair(datasourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(datasourceName, "owner_account_id", resourceName, "owner_account_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "provider_name", resourceName, "provider_name"),
				),
			},
		},
	})
}

func testAccDataSourceAwsDxConnectionConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_dx_locations" "test" {}

resource "aws_dx_connection" "test" {
  name      = %[1]q
  bandwidth = "1Gbps"
  location  = tolist(data.aws_dx_locations.test.location_codes)[0]
}

data "aws_dx_connection" "test" {
  name = aws_dx_connection.test.name
}
`, rName)
}
