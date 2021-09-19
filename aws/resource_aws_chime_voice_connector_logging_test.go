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

func TestAccAWSChimeVoiceConnectorLogging_basic(t *testing.T) {
	name := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_chime_voice_connector_logging.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, chime.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSChimeVoiceConnectorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSChimeVoiceConnectorLoggingConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSChimeVoiceConnectorLoggingExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "enable_sip_logs", "true"),
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

func TestAccAWSChimeVoiceConnectorLogging_disappears(t *testing.T) {
	name := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_chime_voice_connector_logging.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, chime.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSChimeVoiceConnectorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSChimeVoiceConnectorLoggingConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSChimeVoiceConnectorLoggingExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsChimeVoiceConnectorLogging(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSChimeVoiceConnectorLogging_update(t *testing.T) {
	name := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_chime_voice_connector_logging.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, chime.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSChimeVoiceConnectorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSChimeVoiceConnectorLoggingConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSChimeVoiceConnectorLoggingExists(resourceName),
				),
			},
			{
				Config: testAccAWSChimeVoiceConnectorLoggingUpdated(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSChimeVoiceConnectorLoggingExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "enable_sip_logs", "false"),
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

func testAccAWSChimeVoiceConnectorLoggingConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_chime_voice_connector" "chime" {
  name               = "vc-%[1]s"
  require_encryption = true
}

resource "aws_chime_voice_connector_logging" "test" {
  voice_connector_id = aws_chime_voice_connector.chime.id
  enable_sip_logs    = true
}
`, name)
}

func testAccAWSChimeVoiceConnectorLoggingUpdated(name string) string {
	return fmt.Sprintf(`
resource "aws_chime_voice_connector" "chime" {
  name               = "vc-%[1]s"
  require_encryption = true
}

resource "aws_chime_voice_connector_logging" "test" {
  voice_connector_id = aws_chime_voice_connector.chime.id
  enable_sip_logs    = false
}
`, name)
}

func testAccCheckAWSChimeVoiceConnectorLoggingExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no Chime Voice Connector logging ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).chimeconn
		input := &chime.GetVoiceConnectorLoggingConfigurationInput{
			VoiceConnectorId: aws.String(rs.Primary.ID),
		}

		resp, err := conn.GetVoiceConnectorLoggingConfiguration(input)
		if err != nil {
			return err
		}

		if resp == nil || resp.LoggingConfiguration == nil {
			return fmt.Errorf("no Chime Voice Connector logging configureation (%s) found", rs.Primary.ID)
		}

		return nil
	}
}
