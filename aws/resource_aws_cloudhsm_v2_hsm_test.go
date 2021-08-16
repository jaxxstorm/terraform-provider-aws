package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudhsmv2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/cloudhsmv2/finder"
)

func init() {
	resource.AddTestSweepers("aws_cloudhsm_v2_hsm", &resource.Sweeper{
		Name: "aws_cloudhsm_v2_hsm",
		F:    testSweepCloudhsmv2Hsms,
	})
}

func testSweepCloudhsmv2Hsms(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).cloudhsmv2conn
	input := &cloudhsmv2.DescribeClustersInput{}
	sweepResources := make([]*testSweepResource, 0)

	err = conn.DescribeClustersPages(input, func(page *cloudhsmv2.DescribeClustersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, cluster := range page.Clusters {
			if cluster == nil {
				continue
			}

			for _, hsm := range cluster.Hsms {
				r := resourceAwsCloudHsmV2Hsm()
				d := r.Data(nil)
				d.SetId(aws.StringValue(hsm.HsmId))
				d.Set("cluster_id", cluster.ClusterId)
				sweepResources = append(sweepResources, NewTestSweepResource(r, d, client))
			}
		}

		return !lastPage
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping CloudHSMv2 HSM sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing CloudHSMv2 HSMs (%s): %w", region, err)
	}

	err = testSweepResourceOrchestrator(sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping CloudHSMv2 HSMs (%s): %w", region, err)
	}

	return nil
}

func testAccAWSCloudHsmV2Hsm_basic(t *testing.T) {
	resourceName := "aws_cloudhsm_v2_hsm.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, cloudhsmv2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudHsmV2HsmDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudHsmV2HsmConfigSubnetId(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCloudHsmV2HsmExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "availability_zone", "aws_subnet.test.0", "availability_zone"),
					resource.TestCheckResourceAttrPair(resourceName, "cluster_id", "aws_cloudhsm_v2_cluster.test", "id"),
					resource.TestMatchResourceAttr(resourceName, "hsm_eni_id", regexp.MustCompile(`^eni-.+`)),
					resource.TestMatchResourceAttr(resourceName, "hsm_id", regexp.MustCompile(`^hsm-.+`)),
					resource.TestCheckResourceAttr(resourceName, "hsm_state", cloudhsmv2.HsmStateActive),
					resource.TestCheckResourceAttrSet(resourceName, "ip_address"),
					resource.TestCheckResourceAttrPair(resourceName, "subnet_id", "aws_subnet.test.0", "id"),
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

func testAccAWSCloudHsmV2Hsm_disappears(t *testing.T) {
	resourceName := "aws_cloudhsm_v2_hsm.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, cloudhsmv2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudHsmV2ClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudHsmV2HsmConfigSubnetId(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCloudHsmV2ClusterExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsCloudHsmV2Hsm(), resourceName),
					// Verify Delete error handling
					testAccCheckResourceDisappears(testAccProvider, resourceAwsCloudHsmV2Hsm(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccAWSCloudHsmV2Hsm_disappears_Cluster(t *testing.T) {
	clusterResourceName := "aws_cloudhsm_v2_cluster.test"
	resourceName := "aws_cloudhsm_v2_hsm.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, cloudhsmv2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudHsmV2ClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudHsmV2HsmConfigSubnetId(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCloudHsmV2ClusterExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsCloudHsmV2Hsm(), resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsCloudHsmV2Cluster(), clusterResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccAWSCloudHsmV2Hsm_AvailabilityZone(t *testing.T) {
	resourceName := "aws_cloudhsm_v2_hsm.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, cloudhsmv2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudHsmV2HsmDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudHsmV2HsmConfigAvailabilityZone(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCloudHsmV2HsmExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "availability_zone", "aws_subnet.test.0", "availability_zone"),
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

func testAccAWSCloudHsmV2Hsm_IpAddress(t *testing.T) {
	resourceName := "aws_cloudhsm_v2_hsm.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, cloudhsmv2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudHsmV2HsmDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudHsmV2HsmConfigIpAddress(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCloudHsmV2HsmExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "ip_address", "10.0.0.5"),
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

func testAccAWSCloudHsmV2HsmConfigBase() string {
	return `
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "test" {
  count = 2

  availability_zone = element(data.aws_availability_zones.available.names, count.index)
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
  vpc_id            = aws_vpc.test.id
}

resource "aws_cloudhsm_v2_cluster" "test" {
  hsm_type   = "hsm1.medium"
  subnet_ids = aws_subnet.test[*].id
}
`
}

func testAccAWSCloudHsmV2HsmConfigAvailabilityZone() string {
	return composeConfig(
		testAccAWSCloudHsmV2HsmConfigBase(),
		`
resource "aws_cloudhsm_v2_hsm" "test" {
  availability_zone = aws_subnet.test[0].availability_zone
  cluster_id        = aws_cloudhsm_v2_cluster.test.cluster_id
}
`)
}

func testAccAWSCloudHsmV2HsmConfigIpAddress() string {
	return composeConfig(
		testAccAWSCloudHsmV2HsmConfigBase(),
		`
resource "aws_cloudhsm_v2_hsm" "test" {
  cluster_id = aws_cloudhsm_v2_cluster.test.cluster_id
  ip_address = cidrhost(aws_subnet.test[0].cidr_block, 5)
  subnet_id  = aws_subnet.test[0].id
}
`)
}

func testAccAWSCloudHsmV2HsmConfigSubnetId() string {
	return composeConfig(
		testAccAWSCloudHsmV2HsmConfigBase(),
		`
resource "aws_cloudhsm_v2_hsm" "test" {
  cluster_id = aws_cloudhsm_v2_cluster.test.cluster_id
  subnet_id  = aws_subnet.test[0].id
}
`)
}

func testAccCheckAWSCloudHsmV2HsmDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).cloudhsmv2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudhsm_v2_hsm" {
			continue
		}

		hsm, err := finder.Hsm(conn, rs.Primary.ID, rs.Primary.Attributes["hsm_eni_id"])

		if err != nil {
			return err
		}

		if hsm != nil && aws.StringValue(hsm.State) != "DELETED" {
			return fmt.Errorf("HSM still exists:\n%s", hsm)
		}
	}

	return nil
}

func testAccCheckAWSCloudHsmV2HsmExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).cloudhsmv2conn

		it, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		_, err := finder.Hsm(conn, it.Primary.ID, it.Primary.Attributes["hsm_eni_id"])
		if err != nil {
			return fmt.Errorf("CloudHSM cluster not found: %s", err)
		}

		return nil
	}
}
