package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/chime"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSChimeVoiceConnectorTermination_basic(t *testing.T) {
	name := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_chime_voice_connector_termination.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, chime.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSChimeVoiceConnectorTerminationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSChimeVoiceConnectorTerminationConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSChimeVoiceConnectorTerminationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "cps_limit", "1"),
					resource.TestCheckResourceAttr(resourceName, "calling_regions.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "cidr_allow_list.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "disabled", "false"),
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

func TestAccAWSChimeVoiceConnectorTermination_disappears(t *testing.T) {
	name := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_chime_voice_connector_termination.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, chime.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSChimeVoiceConnectorTerminationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSChimeVoiceConnectorTerminationConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSChimeVoiceConnectorTerminationExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsChimeVoiceConnectorTermination(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSChimeVoiceConnectorTermination_update(t *testing.T) {
	name := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_chime_voice_connector_termination.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, chime.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSChimeVoiceConnectorTerminationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSChimeVoiceConnectorTerminationConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSChimeVoiceConnectorTerminationExists(resourceName),
				),
			},
			{
				Config: testAccAWSChimeVoiceConnectorTerminationUpdated(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSChimeVoiceConnectorTerminationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "cps_limit", "1"),
					resource.TestCheckResourceAttr(resourceName, "calling_regions.#", "3"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cidr_allow_list.*", "100.35.78.97/32"),
					resource.TestCheckResourceAttr(resourceName, "disabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_phone_number", ""),
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

func testAccAWSChimeVoiceConnectorTerminationConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_chime_voice_connector" "chime" {
  name               = "vc-%[1]s"
  require_encryption = true
}

resource "aws_chime_voice_connector_termination" "test" {
  voice_connector_id = aws_chime_voice_connector.chime.id

  calling_regions = ["US", "RU"]
  cidr_allow_list = ["50.35.78.97/32"]
}
`, name)
}

func testAccAWSChimeVoiceConnectorTerminationUpdated(name string) string {
	return fmt.Sprintf(`
resource "aws_chime_voice_connector" "chime" {
  name               = "vc-%[1]s"
  require_encryption = true
}

resource "aws_chime_voice_connector_termination" "test" {
  voice_connector_id = aws_chime_voice_connector.chime.id
  disabled           = false
  calling_regions    = ["US", "RU", "CA"]
  cidr_allow_list    = ["100.35.78.97/32"]
}
`, name)
}

func testAccCheckAWSChimeVoiceConnectorTerminationExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no Chime voice connector group ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).chimeconn
		input := &chime.GetVoiceConnectorTerminationInput{
			VoiceConnectorId: aws.String(rs.Primary.ID),
		}

		resp, err := conn.GetVoiceConnectorTermination(input)
		if err != nil {
			return err
		}

		if resp == nil || resp.Termination == nil {
			return fmt.Errorf("Chime Voice Connector Termintation (%s) not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckAWSChimeVoiceConnectorTerminationDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_chime_voice_connector_termination" {
			continue
		}
		conn := testAccProvider.Meta().(*AWSClient).chimeconn
		input := &chime.GetVoiceConnectorTerminationInput{
			VoiceConnectorId: aws.String(rs.Primary.ID),
		}
		resp, err := conn.GetVoiceConnectorTermination(input)

		if isAWSErr(err, chime.ErrCodeNotFoundException, "") {
			continue
		}

		if err != nil {
			return err
		}

		if resp != nil && resp.Termination != nil {
			return fmt.Errorf("error Chime Voice Connector Termination still exists")
		}
	}

	return nil
}
