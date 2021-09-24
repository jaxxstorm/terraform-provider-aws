package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSCloudwatchLogGroupsDataSource_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "data.aws_cloudwatch_log_groups.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { testAccPreCheck(t) },
		ErrorCheck: testAccErrorCheck(t, cloudwatchlogs.EndpointsID),
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAWSCloudwatchLogGroupsDataSourceConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "arns.#", "2"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "arns.*", "aws_cloudwatch_log_group.test1", "arn"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "arns.*", "aws_cloudwatch_log_group.test2", "arn"),
					resource.TestCheckResourceAttr(resourceName, "log_group_names.#", "2"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "log_group_names.*", "aws_cloudwatch_log_group.test1", "name"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "log_group_names.*", "aws_cloudwatch_log_group.test2", "name"),
				),
			},
		},
	})
}

func testAccCheckAWSCloudwatchLogGroupsDataSourceConfig(rName string) string {
	return fmt.Sprintf(`
resource aws_cloudwatch_log_group "test1" {
  name = "%[1]s/1"
}

resource aws_cloudwatch_log_group "test2" {
  name = "%[1]s/2"
}

data aws_cloudwatch_log_groups "test" {
  log_group_name_prefix = %[1]q

  depends_on = [aws_cloudwatch_log_group.test1,aws_cloudwatch_log_group.test2]
}
`, rName)
}
