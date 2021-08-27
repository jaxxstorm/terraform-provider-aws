package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	r53rcc "github.com/aws/aws-sdk-go/service/route53recoverycontrolconfig"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testAccAWSRoute53RecoveryControlConfigControlPanel_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_route53recoverycontrolconfig_control_panel.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(r53rcc.EndpointsID, t) },
		ErrorCheck:   testAccErrorCheck(t, r53rcc.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsRoute53RecoveryControlConfigControlPanelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsRoute53RecoveryControlConfigControlPanelConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsRoute53RecoveryControlConfigControlPanelExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "status", "DEPLOYED"),
					resource.TestCheckResourceAttr(resourceName, "default_control_panel", "false"),
					resource.TestCheckResourceAttr(resourceName, "routing_control_count", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccAWSRoute53RecoveryControlConfigControlPanel_disappears(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_route53recoverycontrolconfig_control_panel.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(r53rcc.EndpointsID, t) },
		ErrorCheck:   testAccErrorCheck(t, r53rcc.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsRoute53RecoveryControlConfigControlPanelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsRoute53RecoveryControlConfigControlPanelConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsRoute53RecoveryControlConfigControlPanelExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsRoute53RecoveryControlConfigControlPanel(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAwsRoute53RecoveryControlConfigControlPanelDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).route53recoverycontrolconfigconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_route53recoverycontrolconfig_control_panel" {
			continue
		}

		input := &r53rcc.DescribeControlPanelInput{
			ControlPanelArn: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeControlPanel(input)

		if err == nil {
			return fmt.Errorf("Route53RecoveryControlConfig Control Panel (%s) not deleted", rs.Primary.ID)
		}
	}

	return nil
}

func testAccAwsRoute53RecoveryControlConfigClusterSetUp(rName string) string {
	return fmt.Sprintf(`
resource "aws_route53recoverycontrolconfig_cluster" "test" {
  name = %[1]q
}
`, rName)
}

func testAccAwsRoute53RecoveryControlConfigControlPanelConfig(rName string) string {
	return composeConfig(testAccAwsRoute53RecoveryControlConfigClusterSetUp(rName), fmt.Sprintf(`
resource "aws_route53recoverycontrolconfig_control_panel" "test" {
  name        = %[1]q
  cluster_arn = aws_route53recoverycontrolconfig_cluster.test.arn
}
`, rName))
}

func testAccCheckAwsRoute53RecoveryControlConfigControlPanelExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := testAccProvider.Meta().(*AWSClient).route53recoverycontrolconfigconn

		input := &r53rcc.DescribeControlPanelInput{
			ControlPanelArn: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeControlPanel(input)

		return err
	}
}
