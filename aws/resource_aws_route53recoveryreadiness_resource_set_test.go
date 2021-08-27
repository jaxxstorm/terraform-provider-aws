package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/route53recoveryreadiness"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAwsRoute53RecoveryReadinessResourceSet_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	cwArn := arn.ARN{
		AccountID: "123456789012",
		Partition: endpoints.AwsPartitionID,
		Region:    endpoints.EuWest1RegionID,
		Resource:  "alarm:zzzzzzzzz",
		Service:   "cloudwatch",
	}.String()
	resourceName := "aws_route53recoveryreadiness_resource_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccPreCheckAwsRoute53RecoveryReadiness(t) },
		ErrorCheck:        testAccErrorCheck(t, route53recoveryreadiness.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsRoute53RecoveryReadinessResourceSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsRoute53RecoveryReadinessResourceSetConfig(rName, cwArn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsRoute53RecoveryReadinessResourceSetExists(resourceName),
					testAccMatchResourceAttrGlobalARN(resourceName, "arn", "route53-recovery-readiness", regexp.MustCompile(`resource-set.+`)),
					resource.TestCheckResourceAttr(resourceName, "resources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func TestAccAwsRoute53RecoveryReadinessResourceSet_disappears(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	cwArn := arn.ARN{
		AccountID: "123456789012",
		Partition: endpoints.AwsPartitionID,
		Region:    endpoints.EuWest1RegionID,
		Resource:  "alarm:zzzzzzzzz",
		Service:   "cloudwatch",
	}.String()
	resourceName := "aws_route53recoveryreadiness_resource_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccPreCheckAwsRoute53RecoveryReadiness(t) },
		ErrorCheck:        testAccErrorCheck(t, route53recoveryreadiness.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsRoute53RecoveryReadinessResourceSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsRoute53RecoveryReadinessResourceSetConfig(rName, cwArn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsRoute53RecoveryReadinessResourceSetExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsRoute53RecoveryReadinessResourceSet(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAwsRoute53RecoveryReadinessResourceSet_tags(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_route53recoveryreadiness_resource_set.test"
	cwArn := arn.ARN{
		AccountID: "123456789012",
		Partition: endpoints.AwsPartitionID,
		Region:    endpoints.EuWest1RegionID,
		Resource:  "alarm:zzzzzzzzz",
		Service:   "cloudwatch",
	}.String()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccPreCheckAwsRoute53RecoveryReadiness(t) },
		ErrorCheck:        testAccErrorCheck(t, route53recoveryreadiness.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsRoute53RecoveryReadinessResourceSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsRoute53RecoveryReadinessResourceSetConfig_Tags1(rName, cwArn, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsRoute53RecoveryReadinessResourceSetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAwsRoute53RecoveryReadinessResourceSetConfig_Tags2(rName, cwArn, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsRoute53RecoveryReadinessResourceSetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAwsRoute53RecoveryReadinessResourceSetConfig_Tags1(rName, cwArn, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsRoute53RecoveryReadinessResourceSetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAwsRoute53RecoveryReadinessResourceSet_readinessScope(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_route53recoveryreadiness_resource_set.test"
	cwArn := arn.ARN{
		AccountID: "123456789012",
		Partition: endpoints.AwsPartitionID,
		Region:    endpoints.EuWest1RegionID,
		Resource:  "alarm:zzzzzzzzz",
		Service:   "cloudwatch",
	}.String()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccPreCheckAwsRoute53RecoveryReadiness(t) },
		ErrorCheck:        testAccErrorCheck(t, route53recoveryreadiness.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsRoute53RecoveryReadinessResourceSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsRoute53RecoveryReadinessResourceSetConfig_ReadinessScopes(rName, cwArn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsRoute53RecoveryReadinessResourceSetExists(resourceName),
					testAccMatchResourceAttrGlobalARN(resourceName, "arn", "route53-recovery-readiness", regexp.MustCompile(`resource-set.+`)),
					resource.TestCheckResourceAttr(resourceName, "resources.0.readiness_scopes.#", "1"),
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

func TestAccAwsRoute53RecoveryReadinessResourceSet_basicDNSTargetResource(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_route53recoveryreadiness_resource_set.test"
	domainName := "myTestDomain.test"
	hzArn := arn.ARN{
		AccountID: "123456789012",
		Partition: endpoints.AwsPartitionID,
		Region:    endpoints.EuWest1RegionID,
		Resource:  "hostedzone/zzzzzzzzz",
		Service:   "route53",
	}.String()
	recordType := "A"
	recordSetId := "12345"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckAwsRoute53RecoveryReadinessResourceSet(t)
		},
		ErrorCheck:        testAccErrorCheck(t, route53recoveryreadiness.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsRoute53RecoveryReadinessResourceSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsRoute53RecoveryReadinessResourceSetBasicDnsTargetResourceConfig(rName, domainName, hzArn, recordType, recordSetId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsRoute53RecoveryReadinessResourceSetExists(resourceName),
					testAccMatchResourceAttrGlobalARN(resourceName, "arn", "route53-recovery-readiness", regexp.MustCompile(`resource-set.+`)),
					resource.TestCheckResourceAttr(resourceName, "resources.0.dns_target_resource.0.domain_name", domainName),
					resource.TestCheckResourceAttrSet(resourceName, "resources.0.dns_target_resource.0.hosted_zone_arn"),
					resource.TestCheckResourceAttr(resourceName, "resources.0.dns_target_resource.0.record_type", recordType),
					resource.TestCheckResourceAttr(resourceName, "resources.0.dns_target_resource.0.record_set_id", recordSetId),
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

func TestAccAwsRoute53RecoveryReadinessResourceSet_dnsTargetResourceNLBTarget(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_route53recoveryreadiness_resource_set.test"
	hzArn := arn.ARN{
		AccountID: "123456789012",
		Partition: endpoints.AwsPartitionID,
		Region:    endpoints.EuWest1RegionID,
		Resource:  "hostedzone/zzzzzzzzz",
		Service:   "route53",
	}.String()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckAwsRoute53RecoveryReadiness(t)
		},
		ErrorCheck:        testAccErrorCheck(t, route53recoveryreadiness.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsRoute53RecoveryReadinessResourceSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsRoute53RecoveryReadinessResourceSetDnsTargetResourceNlbTargetConfig(rName, hzArn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsRoute53RecoveryReadinessResourceSetExists(resourceName),
					testAccMatchResourceAttrGlobalARN(resourceName, "arn", "route53-recovery-readiness", regexp.MustCompile(`resource-set.+`)),
					resource.TestCheckResourceAttrSet(resourceName, "resources.0.dns_target_resource.0.target_resource.0.nlb_resource.0.arn"),
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

func TestAccAwsRoute53RecoveryReadinessResourceSet_dnsTargetResourceR53Target(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_route53recoveryreadiness_resource_set.test"
	hzArn := arn.ARN{
		AccountID: "123456789012",
		Partition: endpoints.AwsPartitionID,
		Region:    endpoints.EuWest1RegionID,
		Resource:  "hostedzone/zzzzzzzzz",
		Service:   "route53",
	}.String()
	domainName := "my.target.domain"
	recordSetId := "987654321"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckAwsRoute53RecoveryReadiness(t)
		},
		ErrorCheck:        testAccErrorCheck(t, route53recoveryreadiness.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsRoute53RecoveryReadinessResourceSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsRoute53RecoveryReadinessResourceSetDnsTargetResourceR53TargetConfig(rName, hzArn, domainName, recordSetId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsRoute53RecoveryReadinessResourceSetExists(resourceName),
					testAccMatchResourceAttrGlobalARN(resourceName, "arn", "route53-recovery-readiness", regexp.MustCompile(`resource-set.+`)),
					resource.TestCheckResourceAttr(resourceName, "resources.0.dns_target_resource.0.target_resource.0.r53_resource.0.domain_name", domainName),
					resource.TestCheckResourceAttr(resourceName, "resources.0.dns_target_resource.0.target_resource.0.r53_resource.0.record_set_id", recordSetId),
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

func TestAccAwsRoute53RecoveryReadinessResourceSet_timeout(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	cwArn := arn.ARN{
		AccountID: "123456789012",
		Partition: endpoints.AwsPartitionID,
		Region:    endpoints.EuWest1RegionID,
		Resource:  "alarm:zzzzzzzzz",
		Service:   "cloudwatch",
	}.String()
	resourceName := "aws_route53recoveryreadiness_resource_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccPreCheckAwsRoute53RecoveryReadiness(t) },
		ErrorCheck:        testAccErrorCheck(t, route53recoveryreadiness.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsRoute53RecoveryReadinessResourceSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsRoute53RecoveryReadinessResourceSetConfig_Timeout(rName, cwArn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsRoute53RecoveryReadinessResourceSetExists(resourceName),
					testAccMatchResourceAttrGlobalARN(resourceName, "arn", "route53-recovery-readiness", regexp.MustCompile(`resource-set.+`)),
					resource.TestCheckResourceAttr(resourceName, "resources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func testAccCheckAwsRoute53RecoveryReadinessResourceSetDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).route53recoveryreadinessconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_route53recoveryreadiness_resource_set" {
			continue
		}

		input := &route53recoveryreadiness.GetResourceSetInput{
			ResourceSetName: aws.String(rs.Primary.ID),
		}

		_, err := conn.GetResourceSet(input)
		if err == nil {
			return fmt.Errorf("Route53RecoveryReadiness Resource Set (%s) not deleted", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAwsRoute53RecoveryReadinessResourceSetExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := testAccProvider.Meta().(*AWSClient).route53recoveryreadinessconn

		input := &route53recoveryreadiness.GetResourceSetInput{
			ResourceSetName: aws.String(rs.Primary.ID),
		}

		_, err := conn.GetResourceSet(input)

		return err
	}
}

func testAccPreCheckAwsRoute53RecoveryReadinessResourceSet(t *testing.T) {
	conn := testAccProvider.Meta().(*AWSClient).route53recoveryreadinessconn

	input := &route53recoveryreadiness.ListResourceSetsInput{}

	_, err := conn.ListResourceSets(input)

	if testAccPreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccAwsRoute53RecoveryReadinessResourceSetConfig_NLB(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_lb" "test" {
  name = %[1]q

  subnets = [
    aws_subnet.test1.id,
    aws_subnet.test2.id,
  ]

  load_balancer_type         = "network"
  internal                   = true
  idle_timeout               = 60
  enable_deletion_protection = false
}

data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_subnet" "test1" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.0.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]
}

resource "aws_subnet" "test2" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.0.2.0/24"
  availability_zone = data.aws_availability_zones.available.names[1]
}

data "aws_caller_identity" "current" {}
`, rName)
}

func testAccAwsRoute53RecoveryReadinessResourceSetConfig(rName, cwArn string) string {
	return fmt.Sprintf(`
resource "aws_route53recoveryreadiness_resource_set" "test" {
  resource_set_name = %[1]q
  resource_set_type = "AWS::CloudWatch::Alarm"

  resources {
    resource_arn = %[2]q
  }
}
`, rName, cwArn)
}

func testAccAwsRoute53RecoveryReadinessResourceSetConfig_Tags1(rName, cwArn, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_route53recoveryreadiness_resource_set" "test" {
  resource_set_name = %[1]q
  resource_set_type = "AWS::CloudWatch::Alarm"

  resources {
    resource_arn = %[2]q
  }

  tags = {
    %[3]q = %[4]q
  }
}
`, rName, cwArn, tagKey1, tagValue1)
}

func testAccAwsRoute53RecoveryReadinessResourceSetConfig_Tags2(rName, cwArn, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_route53recoveryreadiness_resource_set" "test" {
  resource_set_name = %[1]q
  resource_set_type = "AWS::CloudWatch::Alarm"

  resources {
    resource_arn = %[2]q
  }

  tags = {
    %[3]q = %[4]q
    %[5]q = %[6]q
  }
}
`, rName, cwArn, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccAwsRoute53RecoveryReadinessResourceSetConfig_ReadinessScopes(rName, cwArn string) string {
	return fmt.Sprintf(`
resource "aws_route53recoveryreadiness_cell" "test" {
  cell_name = "resource_set_test_cell"
}

resource "aws_route53recoveryreadiness_resource_set" "test" {
  resource_set_name = %[1]q
  resource_set_type = "AWS::CloudWatch::Alarm"

  resources {
    resource_arn     = %[2]q
    readiness_scopes = [aws_route53recoveryreadiness_cell.test.arn]
  }
}
`, rName, cwArn)
}

func testAccAwsRoute53RecoveryReadinessResourceSetBasicDnsTargetResourceConfig(rName, domainName, hzArn, recordType, recordSetId string) string {
	return composeConfig(fmt.Sprintf(`
resource "aws_route53recoveryreadiness_resource_set" "test" {
  resource_set_name = %[1]q
  resource_set_type = "AWS::Route53RecoveryReadiness::DNSTargetResource"

  resources {
    dns_target_resource {
      domain_name     = %[2]q
      hosted_zone_arn = %[3]q
      record_type     = %[4]q
      record_set_id   = %[5]q
    }
  }
}
`, rName, domainName, hzArn, recordType, recordSetId))
}

func testAccAwsRoute53RecoveryReadinessResourceSetDnsTargetResourceNlbTargetConfig(rName, hzArn string) string {
	return composeConfig(testAccAwsRoute53RecoveryReadinessResourceSetConfig_NLB(rName), fmt.Sprintf(`
resource "aws_route53recoveryreadiness_resource_set" "test" {
  resource_set_name = %[1]q
  resource_set_type = "AWS::Route53RecoveryReadiness::DNSTargetResource"

  resources {
    dns_target_resource {
      domain_name     = "myTestDomain.test"
      hosted_zone_arn = %[2]q
      record_type     = "A"
      record_set_id   = "12345"

      target_resource {
        nlb_resource {
          arn = aws_lb.test.arn
        }
      }
    }
  }
}
`, rName, hzArn))
}

func testAccAwsRoute53RecoveryReadinessResourceSetDnsTargetResourceR53TargetConfig(rName, hzArn, domainName, recordSetId string) string {
	return fmt.Sprintf(`
resource "aws_route53recoveryreadiness_resource_set" "test" {
  resource_set_name = %[1]q
  resource_set_type = "AWS::Route53RecoveryReadiness::DNSTargetResource"

  resources {
    dns_target_resource {
      domain_name     = "myTestDomain.test"
      hosted_zone_arn = %[2]q
      record_type     = "A"
      record_set_id   = "12345"

      target_resource {
        r53_resource {
          domain_name   = %[3]q
          record_set_id = %[4]q
        }
      }
    }
  }
}
`, rName, hzArn, domainName, recordSetId)
}

func testAccAwsRoute53RecoveryReadinessResourceSetConfig_Timeout(rName, cwArn string) string {
	return fmt.Sprintf(`
resource "aws_route53recoveryreadiness_resource_set" "test" {
  resource_set_name = %[1]q
  resource_set_type = "AWS::CloudWatch::Alarm"

  resources {
    resource_arn = %[2]q
  }

  timeouts {
    delete = "10m"
  }
}
`, rName, cwArn)
}
