package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceAwsEksClusters() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsEksClustersRead,

		Schema: map[string]*schema.Schema{
			"names": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceAwsEksClustersRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).eksconn

	var clusters []*string

	err := conn.ListClustersPages(&eks.ListClustersInput{}, func(page *eks.ListClustersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		clusters = append(clusters, page.Clusters...)

		return !lastPage
	})

	if err != nil {
		return fmt.Errorf("error listing EKS Clusters: %w", err)
	}

	d.SetId(meta.(*AWSClient).region)

	d.Set("names", aws.StringValueSlice(clusters))

	return nil
}
