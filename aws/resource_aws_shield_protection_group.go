package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/shield"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsShieldProtectionGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsShieldProtectionGroupCreate,
		Read:   resourceAwsShieldProtectionGroupRead,
		Update: resourceAwsShieldProtectionGroupUpdate,
		Delete: resourceAwsShieldProtectionGroupDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"aggregation": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(shield.ProtectionGroupAggregation_Values(), false),
			},
			"members": {
				Type:          schema.TypeList,
				Optional:      true,
				MinItems:      0,
				MaxItems:      10000,
				ConflictsWith: []string{"resource_type"},
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.All(validateArn,
						validation.StringLenBetween(1, 2048),
					),
				},
			},
			"pattern": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(shield.ProtectionGroupPattern_Values(), false),
			},
			"protection_group_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 36),
				ForceNew:     true,
			},
			"protection_group_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"resource_type": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"members"},
				ValidateFunc:  validation.StringInSlice(shield.ProtectedResourceType_Values(), false),
			},
			"tags":     tagsSchema(),
			"tags_all": tagsSchemaComputed(),
		},
		CustomizeDiff: SetTagsDiff,
	}
}

func resourceAwsShieldProtectionGroupCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).shieldconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(keyvaluetags.New(d.Get("tags").(map[string]interface{})))

	protectionGroupID := d.Get("protection_group_id").(string)
	input := &shield.CreateProtectionGroupInput{
		Aggregation:       aws.String(d.Get("aggregation").(string)),
		Pattern:           aws.String(d.Get("pattern").(string)),
		ProtectionGroupId: aws.String(protectionGroupID),
		Tags:              tags.IgnoreAws().ShieldTags(),
	}

	if v, ok := d.GetOk("members"); ok {
		input.Members = expandStringList(v.([]interface{}))
	}

	if v, ok := d.GetOk("resource_type"); ok {
		input.ResourceType = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating Shield Protection Group: %s", input)
	_, err := conn.CreateProtectionGroup(input)

	if err != nil {
		return fmt.Errorf("error creating Shield Protection Group (%s): %w", protectionGroupID, err)
	}

	d.SetId(protectionGroupID)

	return resourceAwsShieldProtectionGroupRead(d, meta)
}

func resourceAwsShieldProtectionGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).shieldconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	input := &shield.DescribeProtectionGroupInput{
		ProtectionGroupId: aws.String(d.Id()),
	}

	resp, err := conn.DescribeProtectionGroup(input)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, shield.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Shield Protection Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Shield Protection Group (%s): %w", d.Id(), err)
	}

	arn := aws.StringValue(resp.ProtectionGroup.ProtectionGroupArn)
	d.Set("protection_group_arn", arn)
	d.Set("aggregation", resp.ProtectionGroup.Aggregation)
	d.Set("protection_group_id", resp.ProtectionGroup.ProtectionGroupId)
	d.Set("pattern", resp.ProtectionGroup.Pattern)

	if resp.ProtectionGroup.Members != nil {
		d.Set("members", resp.ProtectionGroup.Members)
	}

	if resp.ProtectionGroup.ResourceType != nil {
		d.Set("resource_type", resp.ProtectionGroup.ResourceType)
	}

	tags, err := keyvaluetags.ShieldListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for Shield Protection Group (%s): %w", arn, err)
	}

	tags = tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceAwsShieldProtectionGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).shieldconn

	input := &shield.UpdateProtectionGroupInput{
		Aggregation:       aws.String(d.Get("aggregation").(string)),
		Pattern:           aws.String(d.Get("pattern").(string)),
		ProtectionGroupId: aws.String(d.Id()),
	}

	if v, ok := d.GetOk("members"); ok {
		input.Members = expandStringList(v.([]interface{}))
	}

	if v, ok := d.GetOk("resource_type"); ok {
		input.ResourceType = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Updating Shield Protection Group: %s", input)
	_, err := conn.UpdateProtectionGroup(input)

	if err != nil {
		return fmt.Errorf("error updating Shield Protection Group (%s): %w", d.Id(), err)
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := keyvaluetags.ShieldUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %w", err)
		}
	}

	return resourceAwsShieldProtectionGroupRead(d, meta)
}

func resourceAwsShieldProtectionGroupDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).shieldconn

	log.Printf("[DEBUG] Deletinh Shield Protection Group: %s", d.Id())
	_, err := conn.DeleteProtectionGroup(&shield.DeleteProtectionGroupInput{
		ProtectionGroupId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, shield.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Shield Protection Group (%s): %w", d.Id(), err)
	}

	return nil
}
