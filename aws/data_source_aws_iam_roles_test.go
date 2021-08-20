package aws

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSIAMRolesDataSource_basic(t *testing.T) {
	dataSourceName := "data.aws_iam_roles.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { testAccPreCheck(t) },
		ErrorCheck: testAccErrorCheck(t, iam.EndpointsID),
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIAMRolesConfigDataSource_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(dataSourceName, "names.#", regexp.MustCompile("[^0].*$")),
				),
			},
		},
	})
}

func TestAccAWSIAMRolesDataSource_nameRegex(t *testing.T) {
	rCount := strconv.Itoa(acctest.RandIntRange(1, 4))
	rName := acctest.RandomWithPrefix("tf-acc-test")
	dataSourceName := "data.aws_iam_roles.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { testAccPreCheck(t) },
		ErrorCheck: testAccErrorCheck(t, iam.EndpointsID),
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIAMRolesConfigDataSource_nameRegex(rCount, rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "names.#", rCount),
					resource.TestCheckResourceAttr(dataSourceName, "arns.#", rCount),
				),
			},
		},
	})
}

func TestAccAWSIAMRolesDataSource_pathPrefix(t *testing.T) {
	rCount := strconv.Itoa(acctest.RandIntRange(1, 4))
	rName := acctest.RandomWithPrefix("tf-acc-test")
	rPathPrefix := acctest.RandomWithPrefix("tf-acc-path")
	dataSourceName := "data.aws_iam_roles.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { testAccPreCheck(t) },
		ErrorCheck: testAccErrorCheck(t, iam.EndpointsID),
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIAMRolesConfigDataSource_pathPrefix(rCount, rName, rPathPrefix),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "names.#", rCount),
					resource.TestCheckResourceAttr(dataSourceName, "arns.#", rCount),
				),
			},
		},
	})
}

func TestAccAWSIAMRolesDataSource_nonExistentPathPrefix(t *testing.T) {
	dataSourceName := "data.aws_iam_roles.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { testAccPreCheck(t) },
		ErrorCheck: testAccErrorCheck(t, iam.EndpointsID),
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIAMRolesConfigDataSource_nonExistentPathPrefix,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "arns.#", "0"),
					resource.TestCheckResourceAttr(dataSourceName, "names.#", "0"),
				),
			},
		},
	})
}

func TestAccAWSIAMRolesDataSource_nameRegexAndPathPrefix(t *testing.T) {
	rCount := strconv.Itoa(acctest.RandIntRange(1, 4))
	rName := acctest.RandomWithPrefix("tf-acc-test")
	rPathPrefix := acctest.RandomWithPrefix("tf-acc-path")
	dataSourceName := "data.aws_iam_roles.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { testAccPreCheck(t) },
		ErrorCheck: testAccErrorCheck(t, iam.EndpointsID),
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIAMRolesConfigDataSource_nameRegexAndPathPrefix(rCount, rName, rPathPrefix, "0"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "names.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "arns.#", "1"),
				),
			},
		},
	})
}

const testAccAWSIAMRolesConfigDataSource_basic = `
data "aws_iam_roles" "test" {}
`

func testAccAWSIAMRolesConfigDataSource_nameRegex(rCount, rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  count = %[1]s
  name  = "%[2]s-${count.index}-role"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.${data.aws_partition.current.dns_suffix}"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
  tags = {
    Seed = %[2]q
  }
}

data "aws_iam_roles" "test" {
  name_regex = "${aws_iam_role.test[0].tags["Seed"]}-.*-role"
}
`, rCount, rName)
}

func testAccAWSIAMRolesConfigDataSource_pathPrefix(rCount, rName, rPathPrefix string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  count = %[1]s
  name  = "%[2]s-${count.index}-role"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.${data.aws_partition.current.dns_suffix}"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF

  path = "/%[3]s/"
}

data "aws_iam_roles" "test" {
  path_prefix = aws_iam_role.test[0].path
}
`, rCount, rName, rPathPrefix)
}

const testAccAWSIAMRolesConfigDataSource_nonExistentPathPrefix = `
data "aws_iam_roles" "test" {
  path_prefix = "/dne/path"
}
`

func testAccAWSIAMRolesConfigDataSource_nameRegexAndPathPrefix(rCount, rName, rPathPrefix, rIndex string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  count = %[1]s
  name  = "%[2]s-${count.index}-role"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.${data.aws_partition.current.dns_suffix}"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF

  path = "/%[3]s/"
  tags = {
    Seed = %[2]q
  }
}

data "aws_iam_roles" "test" {
  name_regex  = "${aws_iam_role.test[0].tags["Seed"]}-%[4]s-role"
  path_prefix = aws_iam_role.test[0].path
}
`, rCount, rName, rPathPrefix, rIndex)
}
