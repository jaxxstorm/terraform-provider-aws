package aws

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	tfec2 "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ec2"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ec2/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func resourceAwsFlowLog() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsLogFlowCreate,
		Read:   resourceAwsLogFlowRead,
		Update: resourceAwsLogFlowUpdate,
		Delete: resourceAwsLogFlowDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"destination_options": {
				Type:             schema.TypeList,
				Optional:         true,
				ForceNew:         true,
				MaxItems:         1,
				DiffSuppressFunc: suppressMissingOptionalConfigurationBlock,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"file_format": {
							Type:         schema.TypeString,
							ValidateFunc: validation.StringInSlice(ec2.DestinationFileFormat_Values(), false),
							Optional:     true,
							Default:      ec2.DestinationFileFormatPlainText,
						},
						"hive_compatible_partitions": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"per_hour_partition": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
					},
				},
			},
			"eni_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"eni_id", "subnet_id", "vpc_id"},
			},
			"iam_role_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validateArn,
			},
			"log_destination": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ValidateFunc:  validateArn,
				ConflictsWith: []string{"log_group_name"},
			},
			"log_destination_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      ec2.LogDestinationTypeCloudWatchLogs,
				ValidateFunc: validation.StringInSlice(ec2.LogDestinationType_Values(), false),
			},
			"log_format": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"log_group_name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"log_destination"},
				Deprecated:    "use 'log_destination' argument instead",
			},
			"max_aggregation_interval": {
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				Default:      600,
				ValidateFunc: validation.IntInSlice([]int{60, 600}),
			},
			"subnet_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"eni_id", "subnet_id", "vpc_id"},
			},
			"tags":     tagsSchema(),
			"tags_all": tagsSchemaTrulyComputed(),
			"traffic_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(ec2.TrafficType_Values(), false),
			},
			"vpc_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"eni_id", "subnet_id", "vpc_id"},
			},
		},

		CustomizeDiff: SetTagsDiff,
	}
}

func resourceAwsLogFlowCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(keyvaluetags.New(d.Get("tags").(map[string]interface{})))

	var resourceID string
	var resourceType string
	for _, v := range []struct {
		ID   string
		Type string
	}{
		{
			ID:   d.Get("vpc_id").(string),
			Type: ec2.FlowLogsResourceTypeVpc,
		},
		{
			ID:   d.Get("subnet_id").(string),
			Type: ec2.FlowLogsResourceTypeSubnet,
		},
		{
			ID:   d.Get("eni_id").(string),
			Type: ec2.FlowLogsResourceTypeNetworkInterface,
		},
	} {
		if v.ID != "" {
			resourceID = v.ID
			resourceType = v.Type
			break
		}
	}

	input := &ec2.CreateFlowLogsInput{
		LogDestinationType: aws.String(d.Get("log_destination_type").(string)),
		ResourceIds:        aws.StringSlice([]string{resourceID}),
		ResourceType:       aws.String(resourceType),
		TrafficType:        aws.String(d.Get("traffic_type").(string)),
	}

	if v, ok := d.GetOk("destination_options"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.DestinationOptions = expandEc2DestinationOptionsRequest(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("iam_role_arn"); ok {
		input.DeliverLogsPermissionArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("log_destination"); ok {
		input.LogDestination = aws.String(strings.TrimSuffix(v.(string), ":*"))
	}

	if v, ok := d.GetOk("log_format"); ok {
		input.LogFormat = aws.String(v.(string))
	}

	if v, ok := d.GetOk("log_group_name"); ok {
		input.LogGroupName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("max_aggregation_interval"); ok {
		input.MaxAggregationInterval = aws.Int64(int64(v.(int)))
	}

	if len(tags) > 0 {
		input.TagSpecifications = ec2TagSpecificationsFromKeyValueTags(tags, ec2.ResourceTypeVpcFlowLog)
	}

	log.Printf("[DEBUG] Creating Flow Log: %s", input)
	output, err := conn.CreateFlowLogs(input)

	if err == nil && output != nil {
		err = tfec2.UnsuccessfulItemsError(output.Unsuccessful)
	}

	if err != nil {
		return fmt.Errorf("error creating Flow Log (%s): %w", resourceID, err)
	}

	d.SetId(aws.StringValue(output.FlowLogIds[0]))

	return resourceAwsLogFlowRead(d, meta)
}

func resourceAwsLogFlowRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	fl, err := finder.FlowLogByID(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Flow Log %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Flow Log (%s): %w", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Service:   ec2.ServiceName,
		Region:    meta.(*AWSClient).region,
		AccountID: meta.(*AWSClient).accountid,
		Resource:  fmt.Sprintf("vpc-flow-log/%s", d.Id()),
	}.String()
	d.Set("arn", arn)
	if fl.DestinationOptions != nil {
		if err := d.Set("destination_options", []interface{}{flattenEc2DestinationOptionsResponse(fl.DestinationOptions)}); err != nil {
			return fmt.Errorf("error setting destination_options: %w", err)
		}
	} else {
		d.Set("destination_options", nil)
	}
	d.Set("iam_role_arn", fl.DeliverLogsPermissionArn)
	d.Set("log_destination", fl.LogDestination)
	d.Set("log_destination_type", fl.LogDestinationType)
	d.Set("log_format", fl.LogFormat)
	d.Set("log_group_name", fl.LogGroupName)
	d.Set("max_aggregation_interval", fl.MaxAggregationInterval)
	d.Set("traffic_type", fl.TrafficType)

	switch resourceID := aws.StringValue(fl.ResourceId); {
	case strings.HasPrefix(resourceID, "vpc-"):
		d.Set("vpc_id", resourceID)
	case strings.HasPrefix(resourceID, "subnet-"):
		d.Set("subnet_id", resourceID)
	case strings.HasPrefix(resourceID, "eni-"):
		d.Set("eni_id", resourceID)
	}

	tags := keyvaluetags.Ec2KeyValueTags(fl.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceAwsLogFlowUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := keyvaluetags.Ec2UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating Flow Log (%s) tags: %w", d.Id(), err)
		}
	}

	return resourceAwsLogFlowRead(d, meta)
}

func resourceAwsLogFlowDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	log.Printf("[INFO] Deleting Flow Log: %s", d.Id())
	output, err := conn.DeleteFlowLogs(&ec2.DeleteFlowLogsInput{
		FlowLogIds: aws.StringSlice([]string{d.Id()}),
	})

	if err == nil && output != nil {
		err = tfec2.UnsuccessfulItemsError(output.Unsuccessful)
	}

	if tfawserr.ErrCodeEquals(err, tfec2.ErrCodeInvalidFlowLogIdNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Flow Log (%s): %w", d.Id(), err)
	}

	return nil
}

func expandEc2DestinationOptionsRequest(tfMap map[string]interface{}) *ec2.DestinationOptionsRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &ec2.DestinationOptionsRequest{}

	if v, ok := tfMap["file_format"].(string); ok && v != "" {
		apiObject.FileFormat = aws.String(v)
	}

	if v, ok := tfMap["hive_compatible_partitions"].(bool); ok {
		apiObject.HiveCompatiblePartitions = aws.Bool(v)
	}

	if v, ok := tfMap["per_hour_partition"].(bool); ok {
		apiObject.PerHourPartition = aws.Bool(v)
	}

	return apiObject
}

func flattenEc2DestinationOptionsResponse(apiObject *ec2.DestinationOptionsResponse) map[string]interface{} {
	tfMap := map[string]interface{}{}

	if v := apiObject.FileFormat; v != nil {
		tfMap["file_format"] = aws.StringValue(v)
	}

	if v := apiObject.HiveCompatiblePartitions; v != nil {
		tfMap["hive_compatible_partitions"] = aws.BoolValue(v)
	}

	if v := apiObject.PerHourPartition; v != nil {
		tfMap["per_hour_partition"] = aws.BoolValue(v)
	}

	return tfMap
}
