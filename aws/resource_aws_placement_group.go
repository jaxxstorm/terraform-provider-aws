package aws

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	tfec2 "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ec2"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ec2/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ec2/waiter"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func resourceAwsPlacementGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsPlacementGroupCreate,
		Read:   resourceAwsPlacementGroupRead,
		Update: resourceAwsPlacementGroupUpdate,
		Delete: resourceAwsPlacementGroupDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"partition_count": {
				Type:     schema.TypeInt,
				ForceNew: true,
				Optional: true,
				// https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/placement-groups.html#placement-groups-limitations-partition.
				ValidateFunc: validation.IntBetween(0, 7),
			},
			"placement_group_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"strategy": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(ec2.PlacementStrategy_Values(), false),
			},
			"tags":     tagsSchema(),
			"tags_all": tagsSchemaTrulyComputed(),
		},

		CustomizeDiff: customdiff.All(
			resourceAwsPlacementGroupCustomizeDiff,
			SetTagsDiff,
		),
	}
}

func resourceAwsPlacementGroupCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(keyvaluetags.New(d.Get("tags").(map[string]interface{})))

	name := d.Get("name").(string)
	input := &ec2.CreatePlacementGroupInput{
		GroupName:         aws.String(name),
		Strategy:          aws.String(d.Get("strategy").(string)),
		TagSpecifications: ec2TagSpecificationsFromKeyValueTags(tags, ec2.ResourceTypePlacementGroup),
	}

	if v, ok := d.GetOk("partition_count"); ok {
		input.PartitionCount = aws.Int64(int64(v.(int)))
	}

	log.Printf("[DEBUG] Creating EC2 Placement Group: %s", input)
	_, err := conn.CreatePlacementGroup(input)

	if err != nil {
		return fmt.Errorf("error creating EC2 Placement Group (%s): %w", name, err)
	}

	d.SetId(name)

	_, err = waiter.PlacementGroupCreated(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error waiting for EC2 Placement Group (%s) create: %w", d.Id(), err)
	}

	return resourceAwsPlacementGroupRead(d, meta)
}

func resourceAwsPlacementGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	pg, err := finder.PlacementGroupByName(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Placement Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading EC2 Placement Group (%s): %w", d.Id(), err)
	}

	d.Set("name", pg.GroupName)
	d.Set("partition_count", pg.PartitionCount)
	d.Set("placement_group_id", pg.GroupId)
	d.Set("strategy", pg.Strategy)

	tags := keyvaluetags.Ec2KeyValueTags(pg.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	arn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Service:   ec2.ServiceName,
		Region:    meta.(*AWSClient).region,
		AccountID: meta.(*AWSClient).accountid,
		Resource:  fmt.Sprintf("placement-group/%s", d.Id()),
	}.String()

	d.Set("arn", arn)

	return nil
}

func resourceAwsPlacementGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := keyvaluetags.Ec2UpdateTags(conn, d.Get("placement_group_id").(string), o, n); err != nil {
			return fmt.Errorf("error updating EC2 Placement Group (%s) tags: %w", d.Id(), err)
		}
	}

	return resourceAwsPlacementGroupRead(d, meta)
}

func resourceAwsPlacementGroupDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	log.Printf("[DEBUG] Deleting EC2 Placement Group: %s", d.Id())
	_, err := conn.DeletePlacementGroup(&ec2.DeletePlacementGroupInput{
		GroupName: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, tfec2.ErrCodeInvalidPlacementGroupUnknown) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting EC2 Placement Group (%s): %w", d.Id(), err)
	}

	_, err = waiter.PlacementGroupDeleted(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error waiting for EC2 Placement Group (%s) delete: %w", d.Id(), err)
	}

	return nil
}

func resourceAwsPlacementGroupCustomizeDiff(_ context.Context, diff *schema.ResourceDiff, v interface{}) error {
	if diff.Id() == "" {
		if partitionCount, strategy := diff.Get("partition_count").(int), diff.Get("strategy").(string); partitionCount > 0 && strategy != ec2.PlacementGroupStrategyPartition {
			return fmt.Errorf("partition_count must not be set when strategy = %q", strategy)
		}
	}

	return nil
}
