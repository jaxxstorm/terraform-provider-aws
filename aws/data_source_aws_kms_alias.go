package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/kms/finder"
)

func dataSourceAwsKmsAlias() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsKmsAliasRead,
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateAwsKmsName,
			},
			"target_key_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"target_key_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsKmsAliasRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kmsconn

	target := d.Get("name").(string)

	alias, err := finder.AliasByName(conn, target)

	if err != nil {
		return fmt.Errorf("error reading KMS Alias (%s): %w", target, err)
	}

	d.SetId(aws.StringValue(alias.AliasArn))
	d.Set("arn", alias.AliasArn)

	// ListAliases can return an alias for an AWS service key (e.g.
	// alias/aws/rds) without a TargetKeyId if the alias has not yet been
	// used for the first time. In that situation, calling DescribeKey will
	// associate an actual key with the alias, and the next call to
	// ListAliases will have a TargetKeyId for the alias.
	//
	// For a simpler codepath, we always call DescribeKey with the alias
	// name to get the target key's ARN and Id direct from AWS.
	//
	// https://docs.aws.amazon.com/kms/latest/APIReference/API_ListAliases.html

	keyMetadata, err := finder.KeyByID(conn, target)

	if err != nil {
		return fmt.Errorf("error reading KMS Key (%s): %w", target, err)
	}

	d.Set("target_key_arn", keyMetadata.Arn)
	d.Set("target_key_id", keyMetadata.KeyId)

	return nil
}
