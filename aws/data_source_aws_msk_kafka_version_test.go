package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/kafka"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSMskKafkaVersionDataSource_basic(t *testing.T) {
	dataSourceName := "data.aws_msk_kafka_version.test"
	version := "2.4.1.1"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccAWSMskKafkaVersionPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, kafka.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSMskKafkaVersionDataSourceBasicConfig(version),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "version", version),
					resource.TestCheckResourceAttrSet(dataSourceName, "status"),
				),
			},
		},
	})
}

func TestAccAWSMskKafkaVersionDataSource_preferred(t *testing.T) {
	dataSourceName := "data.aws_msk_kafka_version.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccAWSMskKafkaVersionPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, kafka.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSMskKafkaVersionDataSourcePreferredConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "version", "2.4.1.1"),
					resource.TestCheckResourceAttrSet(dataSourceName, "status"),
				),
			},
		},
	})
}

func testAccAWSMskKafkaVersionPreCheck(t *testing.T) {
	conn := testAccProvider.Meta().(*AWSClient).kafkaconn

	input := &kafka.ListKafkaVersionsInput{}

	_, err := conn.ListKafkaVersions(input)

	if testAccPreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccAWSMskKafkaVersionDataSourceBasicConfig(version string) string {
	return fmt.Sprintf(`
data "aws_msk_kafka_version" "test" {
  version = %[1]q
}
`, version)
}

func testAccAWSMskKafkaVersionDataSourcePreferredConfig() string {
	return `
data "aws_msk_kafka_version" "test" {
  preferred_versions = ["2.4.1.1", "2.4.1", "2.2.1"]
}
`
}
