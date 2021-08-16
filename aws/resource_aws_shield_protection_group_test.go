package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/shield"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSShieldProtectionGroup_basic(t *testing.T) {
	resourceName := "aws_shield_protection_group.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPartitionHasServicePreCheck(shield.EndpointsID, t)
			testAccPreCheckAWSShield(t)
		},
		ErrorCheck:   testAccErrorCheck(t, shield.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSShieldProtectionGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccShieldProtectionGroupConfig_basic_all(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSShieldProtectionGroupExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "aggregation", shield.ProtectionGroupAggregationMax),
					resource.TestCheckNoResourceAttr(resourceName, "members"),
					resource.TestCheckResourceAttr(resourceName, "pattern", shield.ProtectionGroupPatternAll),
					resource.TestCheckNoResourceAttr(resourceName, "resource_type"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "0"),
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

func TestAccAWSShieldProtectionGroup_disappears(t *testing.T) {
	resourceName := "aws_shield_protection_group.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPartitionHasServicePreCheck(shield.EndpointsID, t)
			testAccPreCheckAWSShield(t)
		},
		ErrorCheck:   testAccErrorCheck(t, shield.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSShieldProtectionGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccShieldProtectionGroupConfig_basic_all(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSShieldProtectionGroupExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsShieldProtectionGroup(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSShieldProtectionGroup_aggregation(t *testing.T) {
	resourceName := "aws_shield_protection_group.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPartitionHasServicePreCheck(shield.EndpointsID, t)
			testAccPreCheckAWSShield(t)
		},
		ErrorCheck:   testAccErrorCheck(t, shield.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSShieldProtectionGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccShieldProtectionGroupConfig_aggregation(rName, shield.ProtectionGroupAggregationMean),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSShieldProtectionGroupExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "aggregation", shield.ProtectionGroupAggregationMean),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccShieldProtectionGroupConfig_aggregation(rName, shield.ProtectionGroupAggregationSum),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSShieldProtectionGroupExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "aggregation", shield.ProtectionGroupAggregationSum),
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

func TestAccAWSShieldProtectionGroup_members(t *testing.T) {
	resourceName := "aws_shield_protection_group.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPartitionHasServicePreCheck(shield.EndpointsID, t)
			testAccPreCheckAWSShield(t)
		},
		ErrorCheck:   testAccErrorCheck(t, shield.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSShieldProtectionGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccShieldProtectionGroupConfig_members(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSShieldProtectionGroupExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "pattern", shield.ProtectionGroupPatternArbitrary),
					resource.TestCheckResourceAttr(resourceName, "members.#", "1"),
					testAccMatchResourceAttrRegionalARN(resourceName, "members.0", "ec2", regexp.MustCompile(`eip-allocation/eipalloc-.+`)),
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

func TestAccAWSShieldProtectionGroup_protectionGroupId(t *testing.T) {
	resourceName := "aws_shield_protection_group.test"
	testID1 := acctest.RandomWithPrefix("tf-acc-test")
	testID2 := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPartitionHasServicePreCheck(shield.EndpointsID, t)
			testAccPreCheckAWSShield(t)
		},
		ErrorCheck:   testAccErrorCheck(t, shield.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSShieldProtectionGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccShieldProtectionGroupConfig_basic_all(testID1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSShieldProtectionGroupExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "protection_group_id", testID1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccShieldProtectionGroupConfig_basic_all(testID2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSShieldProtectionGroupExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "protection_group_id", testID2),
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

func TestAccAWSShieldProtectionGroup_resourceType(t *testing.T) {
	resourceName := "aws_shield_protection_group.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPartitionHasServicePreCheck(shield.EndpointsID, t)
			testAccPreCheckAWSShield(t)
		},
		ErrorCheck:   testAccErrorCheck(t, shield.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSShieldProtectionGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccShieldProtectionGroupConfig_resourceType(rName, shield.ProtectedResourceTypeElasticIpAllocation),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSShieldProtectionGroupExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "pattern", shield.ProtectionGroupPatternByResourceType),
					resource.TestCheckResourceAttr(resourceName, "resource_type", shield.ProtectedResourceTypeElasticIpAllocation),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccShieldProtectionGroupConfig_resourceType(rName, shield.ProtectedResourceTypeApplicationLoadBalancer),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSShieldProtectionGroupExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "pattern", shield.ProtectionGroupPatternByResourceType),
					resource.TestCheckResourceAttr(resourceName, "resource_type", shield.ProtectedResourceTypeApplicationLoadBalancer),
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

func testAccCheckAWSShieldProtectionGroupDestroy(s *terraform.State) error {
	shieldconn := testAccProvider.Meta().(*AWSClient).shieldconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_shield_protection_group" {
			continue
		}

		input := &shield.DescribeProtectionGroupInput{
			ProtectionGroupId: aws.String(rs.Primary.ID),
		}

		resp, err := shieldconn.DescribeProtectionGroup(input)

		if tfawserr.ErrCodeEquals(err, shield.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return err
		}

		if resp != nil && resp.ProtectionGroup != nil && aws.StringValue(resp.ProtectionGroup.ProtectionGroupId) == rs.Primary.ID {
			return fmt.Errorf("The Shield protection group with ID %v still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAWSShieldProtectionGroupExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := testAccProvider.Meta().(*AWSClient).shieldconn

		input := &shield.DescribeProtectionGroupInput{
			ProtectionGroupId: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeProtectionGroup(input)

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccShieldProtectionGroupConfig_basic_all(rName string) string {
	return fmt.Sprintf(`
resource "aws_shield_protection_group" "test" {
  protection_group_id = "%s"
  aggregation         = "MAX"
  pattern             = "ALL"
}
`, rName)
}

func testAccShieldProtectionGroupConfig_aggregation(rName string, aggregation string) string {
	return fmt.Sprintf(`
resource "aws_shield_protection_group" "test" {
  protection_group_id = "%[1]s"
  aggregation         = "%[2]s"
  pattern             = "ALL"
}
`, rName, aggregation)
}

func testAccShieldProtectionGroupConfig_resourceType(rName string, resType string) string {
	return fmt.Sprintf(`
resource "aws_shield_protection_group" "test" {
  protection_group_id = "%[1]s"
  aggregation         = "MAX"
  pattern             = "BY_RESOURCE_TYPE"
  resource_type       = "%[2]s"
}
`, rName, resType)
}

func testAccShieldProtectionGroupConfig_members(rName string) string {
	return composeConfig(testAccShieldProtectionElasticIPAddressConfig(rName), fmt.Sprintf(`
resource "aws_shield_protection_group" "test" {
  depends_on = [aws_shield_protection.acctest]

  protection_group_id = "%[1]s"
  aggregation         = "MAX"
  pattern             = "ARBITRARY"
  members             = ["arn:${data.aws_partition.current.partition}:ec2:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:eip-allocation/${aws_eip.acctest.id}"]
}
`, rName))
}
