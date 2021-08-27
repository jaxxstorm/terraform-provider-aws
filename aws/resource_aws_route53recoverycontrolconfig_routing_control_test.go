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

func testAccAWSRoute53RecoveryControlConfigRoutingControl_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_route53recoverycontrolconfig_routing_control.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(r53rcc.EndpointsID, t) },
		ErrorCheck:   testAccErrorCheck(t, r53rcc.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsRoute53RecoveryControlConfigRoutingControlDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsRoute53RecoveryControlConfigRoutingControlConfig_InDefaultControlPanel(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsRoute53RecoveryControlConfigRoutingControlExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "status", "DEPLOYED"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"cluster_arn", // not available in DescribeRoutingControlOutput
				},
			},
		},
	})
}

func testAccAWSRoute53RecoveryControlConfigRoutingControl_disappears(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_route53recoverycontrolconfig_routing_control.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(r53rcc.EndpointsID, t) },
		ErrorCheck:   testAccErrorCheck(t, r53rcc.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsRoute53RecoveryControlConfigRoutingControlDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsRoute53RecoveryControlConfigRoutingControlConfig_InDefaultControlPanel(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsRoute53RecoveryControlConfigRoutingControlExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsRoute53RecoveryControlConfigRoutingControl(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccAWSRoute53RecoveryControlConfigRoutingControl_nonDefaultControlPanel(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_route53recoverycontrolconfig_routing_control.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(r53rcc.EndpointsID, t) },
		ErrorCheck:   testAccErrorCheck(t, r53rcc.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsRoute53RecoveryControlConfigRoutingControlDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsRoute53RecoveryControlConfigRoutingControlConfig_InNonDefaultControlPanel(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsRoute53RecoveryControlConfigRoutingControlExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "status", "DEPLOYED"),
				),
			},
		},
	})
}

func testAccCheckAwsRoute53RecoveryControlConfigRoutingControlExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := testAccProvider.Meta().(*AWSClient).route53recoverycontrolconfigconn

		input := &r53rcc.DescribeRoutingControlInput{
			RoutingControlArn: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeRoutingControl(input)

		return err
	}
}

func testAccCheckAwsRoute53RecoveryControlConfigRoutingControlDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).route53recoverycontrolconfigconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_route53recoverycontrolconfig_routing_control" {
			continue
		}

		input := &r53rcc.DescribeRoutingControlInput{
			RoutingControlArn: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeRoutingControl(input)

		if err == nil {
			return fmt.Errorf("Route53RecoveryControlConfig Routing Control (%s) not deleted", rs.Primary.ID)
		}
	}

	return nil
}

func testAccAwsRoute53RecoveryControlConfigClusterBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_route53recoverycontrolconfig_cluster" "test" {
  name = %[1]q
}
`, rName)
}

func testAccAwsRoute53RecoveryControlConfigRoutingControlConfig_InDefaultControlPanel(rName string) string {
	return composeConfig(
		testAccAwsRoute53RecoveryControlConfigClusterBase(rName), fmt.Sprintf(`
resource "aws_route53recoverycontrolconfig_routing_control" "test" {
  name        = %[1]q
  cluster_arn = aws_route53recoverycontrolconfig_cluster.test.arn
}
`, rName))
}

func testAccAwsRoute53RecoveryControlConfigControlPanelBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_route53recoverycontrolconfig_control_panel" "test" {
  name        = %[1]q
  cluster_arn = aws_route53recoverycontrolconfig_cluster.test.arn
}
`, rName)
}

func testAccAwsRoute53RecoveryControlConfigRoutingControlConfig_InNonDefaultControlPanel(rName string) string {
	return composeConfig(
		testAccAwsRoute53RecoveryControlConfigClusterBase(rName),
		testAccAwsRoute53RecoveryControlConfigControlPanelBase(rName),
		fmt.Sprintf(`
resource "aws_route53recoverycontrolconfig_routing_control" "test" {
  name              = %[1]q
  cluster_arn       = aws_route53recoverycontrolconfig_cluster.test.arn
  control_panel_arn = aws_route53recoverycontrolconfig_control_panel.test.arn
}
`, rName))
}
