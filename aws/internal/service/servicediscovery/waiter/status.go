package waiter

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicediscovery"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/servicediscovery/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

// OperationStatus fetches the Operation and its Status
func OperationStatus(conn *servicediscovery.ServiceDiscovery, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := finder.OperationByID(conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}
