package aws

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/service/apprunner"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	tfapprunner "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/apprunner"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/apprunner/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/apprunner/waiter"
)

func TestAccAwsAppRunnerCustomDomainAssociation_basic(t *testing.T) {
	domain := os.Getenv("APPRUNNER_CUSTOM_DOMAIN")
	if domain == "" {
		t.Skip("Environment variable APPRUNNER_CUSTOM_DOMAIN is not set")
	}

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_apprunner_custom_domain_association.test"
	serviceResourceName := "aws_apprunner_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAppRunner(t) },
		ErrorCheck:   testAccErrorCheck(t, apprunner.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppRunnerCustomDomainAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppRunnerCustomDomainAssociation_basic(rName, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppRunnerCustomDomainAssociationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "certificate_validation_records.#", "3"),
					resource.TestCheckResourceAttrSet(resourceName, "dns_target"),
					resource.TestCheckResourceAttr(resourceName, "domain_name", domain),
					resource.TestCheckResourceAttr(resourceName, "enable_www_subdomain", "true"),
					resource.TestCheckResourceAttr(resourceName, "status", waiter.CustomDomainAssociationStatusPendingCertificateDnsValidation),
					resource.TestCheckResourceAttrPair(resourceName, "service_arn", serviceResourceName, "arn"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"dns_target"},
			},
		},
	})
}

func TestAccAwsAppRunnerCustomDomainAssociation_disappears(t *testing.T) {
	domain := os.Getenv("APPRUNNER_CUSTOM_DOMAIN")
	if domain == "" {
		t.Skip("Environment variable APPRUNNER_CUSTOM_DOMAIN is not set")
	}

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_apprunner_custom_domain_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAppRunner(t) },
		ErrorCheck:   testAccErrorCheck(t, apprunner.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppRunnerCustomDomainAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppRunnerCustomDomainAssociation_basic(rName, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppRunnerCustomDomainAssociationExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsAppRunnerCustomDomainAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAwsAppRunnerCustomDomainAssociationDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_apprunner_connection" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).apprunnerconn

		domainName, serviceArn, err := tfapprunner.CustomDomainAssociationParseID(rs.Primary.ID)

		if err != nil {
			return err
		}

		customDomain, err := finder.CustomDomain(context.Background(), conn, domainName, serviceArn)

		if tfawserr.ErrCodeEquals(err, apprunner.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return err
		}

		if customDomain != nil {
			return fmt.Errorf("App Runner Custom Domain Association (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAwsAppRunnerCustomDomainAssociationExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No App Runner Custom Domain Association ID is set")
		}

		domainName, serviceArn, err := tfapprunner.CustomDomainAssociationParseID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := testAccProvider.Meta().(*AWSClient).apprunnerconn

		customDomain, err := finder.CustomDomain(context.Background(), conn, domainName, serviceArn)

		if err != nil {
			return err
		}

		if customDomain == nil {
			return fmt.Errorf("App Runner Custom Domain Association (%s) not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccAppRunnerCustomDomainAssociation_basic(rName, domain string) string {
	return fmt.Sprintf(`
resource "aws_apprunner_service" "test" {
  service_name = %[1]q

  source_configuration {
    auto_deployments_enabled = false
    image_repository {
      image_configuration {
        port = "80"
      }
      image_identifier      = "public.ecr.aws/nginx/nginx:latest"
      image_repository_type = "ECR_PUBLIC"
    }
  }
}

resource "aws_apprunner_custom_domain_association" "test" {
  domain_name = %[2]q
  service_arn = aws_apprunner_service.test.arn
}
`, rName, domain)
}
