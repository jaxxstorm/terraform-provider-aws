package aws

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/connect"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	tfconnect "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/connect"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/connect/waiter"
)

func resourceAwsConnectInstance() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAwsConnectInstanceCreate,
		ReadContext:   resourceAwsConnectInstanceRead,
		UpdateContext: resourceAwsConnectInstanceUpdate,
		DeleteContext: resourceAwsConnectInstanceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(waiter.ConnectInstanceCreatedTimeout),
			Delete: schema.DefaultTimeout(waiter.ConnectInstanceDeletedTimeout),
		},
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"auto_resolve_best_voices_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true, //verified default result from ListInstanceAttributes()
			},
			"contact_flow_logs_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false, //verified default result from ListInstanceAttributes()
			},
			"contact_lens_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true, //verified default result from ListInstanceAttributes()
			},
			"created_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"directory_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(12, 12),
				AtLeastOneOf: []string{"directory_id", "instance_alias"},
			},
			"early_media_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true, //verified default result from ListInstanceAttributes()
			},
			"identity_management_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(connect.DirectoryType_Values(), false),
			},
			"inbound_calls_enabled": {
				Type:     schema.TypeBool,
				Required: true,
			},
			"instance_alias": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				AtLeastOneOf: []string{"directory_id", "instance_alias"},
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 64),
					validation.StringMatch(regexp.MustCompile(`^([\da-zA-Z]+)([\da-zA-Z-]+)$`), "must contain only alphanumeric hyphen and underscore characters"),
					validation.StringDoesNotMatch(regexp.MustCompile(`^(d-).+$`), "can not start with d-"),
				),
			},
			"outbound_calls_enabled": {
				Type:     schema.TypeBool,
				Required: true,
			},
			"service_role": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			// Pre-release feature requiring allow-list from AWS. Removing all functionality until feature is GA
			// "use_custom_tts_voices_enabled": {
			// 	Type:     schema.TypeBool,
			// 	Optional: true,
			// 	Default:  false, //verified default result from ListInstanceAttributes()
			// },
		},
	}
}

func resourceAwsConnectInstanceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).connectconn

	input := &connect.CreateInstanceInput{
		ClientToken:            aws.String(resource.UniqueId()),
		IdentityManagementType: aws.String(d.Get("identity_management_type").(string)),
		InboundCallsEnabled:    aws.Bool(d.Get("inbound_calls_enabled").(bool)),
		OutboundCallsEnabled:   aws.Bool(d.Get("outbound_calls_enabled").(bool)),
	}

	if v, ok := d.GetOk("directory_id"); ok {
		input.DirectoryId = aws.String(v.(string))
	}
	if v, ok := d.GetOk("instance_alias"); ok {
		input.InstanceAlias = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating Connect Instance %s", input)
	output, err := conn.CreateInstanceWithContext(ctx, input)

	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating Connect Instance (%s): %w", d.Id(), err))
	}

	d.SetId(aws.StringValue(output.Id))

	if _, err := waiter.InstanceCreated(ctx, conn, d.Id()); err != nil {
		return diag.FromErr(fmt.Errorf("error waiting for Connect instance creation (%s): %w", d.Id(), err))
	}

	for att := range tfconnect.InstanceAttributeMapping() {
		rKey := tfconnect.InstanceAttributeMapping()[att]

		if v, ok := d.GetOk(rKey); ok {
			err := resourceAwsConnectInstanceUpdateAttribute(ctx, conn, d.Id(), att, strconv.FormatBool(v.(bool)))
			//Pre-release attribute, user/account/instance now allow-listed
			if err != nil && tfawserr.ErrCodeEquals(err, tfconnect.ErrCodeAccessDeniedException) || tfawserr.ErrMessageContains(err, tfconnect.ErrCodeAccessDeniedException, "not authorized to update") {
				log.Printf("[WARN] error setting Connect instance (%s) attribute (%s): %s", d.Id(), att, err)
			} else if err != nil {
				return diag.FromErr(fmt.Errorf("error setting Connect instance (%s) attribute (%s): %w", d.Id(), att, err))
			}
		}
	}

	return resourceAwsConnectInstanceRead(ctx, d, meta)
}

func resourceAwsConnectInstanceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).connectconn

	for att := range tfconnect.InstanceAttributeMapping() {
		rKey := tfconnect.InstanceAttributeMapping()[att]
		if d.HasChange(rKey) {
			_, n := d.GetChange(rKey)
			err := resourceAwsConnectInstanceUpdateAttribute(ctx, conn, d.Id(), att, strconv.FormatBool(n.(bool)))
			//Pre-release attribute, user/account/instance now allow-listed
			if err != nil && tfawserr.ErrCodeEquals(err, tfconnect.ErrCodeAccessDeniedException) || tfawserr.ErrMessageContains(err, tfconnect.ErrCodeAccessDeniedException, "not authorized to update") {
				log.Printf("[WARN] error setting Connect instance (%s) attribute (%s): %s", d.Id(), att, err)
			} else if err != nil {
				return diag.FromErr(fmt.Errorf("error setting Connect instance (%s) attribute (%s): %s", d.Id(), att, err))
			}
		}
	}

	return nil
}
func resourceAwsConnectInstanceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).connectconn

	input := connect.DescribeInstanceInput{
		InstanceId: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Reading Connect Instance %s", d.Id())
	output, err := conn.DescribeInstanceWithContext(ctx, &input)

	if !d.IsNewResource() && isAWSErr(err, connect.ErrCodeResourceNotFoundException, "") {
		log.Printf("[WARN] Connect Instance (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading Connect Instance (%s): %s", d.Id(), err))
	}

	instance := output.Instance

	d.SetId(aws.StringValue(instance.Id))
	d.Set("arn", instance.Arn)
	d.Set("created_time", instance.CreatedTime.Format(time.RFC3339))
	d.Set("identity_management_type", instance.IdentityManagementType)
	d.Set("inbound_calls_enabled", instance.InboundCallsEnabled)
	d.Set("instance_alias", instance.InstanceAlias)
	d.Set("outbound_calls_enabled", instance.OutboundCallsEnabled)
	d.Set("service_role", instance.ServiceRole)
	d.Set("status", instance.InstanceStatus)

	for att := range tfconnect.InstanceAttributeMapping() {
		value, err := resourceAwsConnectInstanceReadAttribute(ctx, conn, d.Id(), att)
		if err != nil {
			return diag.FromErr(fmt.Errorf("error reading Connect instance (%s) attribute (%s): %s", d.Id(), att, err))
		}
		d.Set(tfconnect.InstanceAttributeMapping()[att], value)
	}

	return nil
}

func resourceAwsConnectInstanceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).connectconn

	input := &connect.DeleteInstanceInput{
		InstanceId: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting Connect Instance %s", d.Id())

	_, err := conn.DeleteInstance(input)

	if isAWSErr(err, connect.ErrCodeResourceNotFoundException, "") {
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error deleting Connect Instance (%s): %s", d.Id(), err))
	}

	if _, err := waiter.InstanceDeleted(ctx, conn, d.Id()); err != nil {
		return diag.FromErr(fmt.Errorf("error waiting for Connect Instance deletion (%s): %s", d.Id(), err))
	}
	return nil
}

func resourceAwsConnectInstanceUpdateAttribute(ctx context.Context, conn *connect.Connect, instanceID string, attributeType string, value string) error {
	input := &connect.UpdateInstanceAttributeInput{
		InstanceId:    aws.String(instanceID),
		AttributeType: aws.String(attributeType),
		Value:         aws.String(value),
	}

	_, err := conn.UpdateInstanceAttributeWithContext(ctx, input)
	if err != nil {
		return err
	}
	return nil
}

func resourceAwsConnectInstanceReadAttribute(ctx context.Context, conn *connect.Connect, instanceID string, attributeType string) (bool, error) {
	input := &connect.DescribeInstanceAttributeInput{
		InstanceId:    aws.String(instanceID),
		AttributeType: aws.String(attributeType),
	}

	output, err := conn.DescribeInstanceAttributeWithContext(ctx, input)
	if err != nil {
		return false, err
	}
	result, parseerr := strconv.ParseBool(*output.Attribute.Value)
	return result, parseerr
}
