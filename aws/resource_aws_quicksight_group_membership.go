package aws

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/quicksight"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/quicksight/finder"
)

func resourceAwsQuickSightGroupMembership() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAwsQuickSightGroupMembershipCreate,
		ReadWithoutTimeout:   resourceAwsQuickSightGroupMembershipRead,
		DeleteWithoutTimeout: resourceAwsQuickSightGroupMembershipDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"aws_account_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"member_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"group_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"namespace": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "default",
				ValidateFunc: validation.StringInSlice([]string{
					"default",
				}, false),
			},
		},
	}
}

func resourceAwsQuickSightGroupMembershipCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).quicksightconn

	awsAccountID := meta.(*AWSClient).accountid
	namespace := d.Get("namespace").(string)
	groupName := d.Get("group_name").(string)
	memberName := d.Get("member_name").(string)

	if v, ok := d.GetOk("aws_account_id"); ok {
		awsAccountID = v.(string)
	}

	createOpts := &quicksight.CreateGroupMembershipInput{
		AwsAccountId: aws.String(awsAccountID),
		GroupName:    aws.String(groupName),
		MemberName:   aws.String(memberName),
		Namespace:    aws.String(namespace),
	}

	resp, err := conn.CreateGroupMembershipWithContext(ctx, createOpts)
	if err != nil {
		return diag.Errorf("error adding QuickSight user (%s) to group (%s): %s", memberName, groupName, err)
	}

	d.SetId(fmt.Sprintf("%s/%s/%s/%s", awsAccountID, namespace, groupName, aws.StringValue(resp.GroupMember.MemberName)))

	return resourceAwsQuickSightGroupMembershipRead(ctx, d, meta)
}

func resourceAwsQuickSightGroupMembershipRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).quicksightconn

	awsAccountID, namespace, groupName, userName, err := resourceAwsQuickSightGroupMembershipParseID(d.Id())
	if err != nil {
		return diag.Errorf("%s", err)
	}

	listInput := &quicksight.ListGroupMembershipsInput{
		AwsAccountId: aws.String(awsAccountID),
		Namespace:    aws.String(namespace),
		GroupName:    aws.String(groupName),
	}

	found, err := finder.GroupMembership(conn, listInput, userName)
	if err != nil {
		return diag.Errorf("Error listing QuickSight Group Memberships (%s): %s", d.Id(), err)
	}

	if !found {
		log.Printf("[WARN] QuickSight User-group membership %s is not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("aws_account_id", awsAccountID)
	d.Set("namespace", namespace)
	d.Set("member_name", userName)
	d.Set("group_name", groupName)

	return nil
}

func resourceAwsQuickSightGroupMembershipDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).quicksightconn

	awsAccountID, namespace, groupName, userName, err := resourceAwsQuickSightGroupMembershipParseID(d.Id())
	if err != nil {
		return diag.Errorf("%s", err)
	}

	deleteOpts := &quicksight.DeleteGroupMembershipInput{
		AwsAccountId: aws.String(awsAccountID),
		Namespace:    aws.String(namespace),
		MemberName:   aws.String(userName),
		GroupName:    aws.String(groupName),
	}

	if _, err := conn.DeleteGroupMembershipWithContext(ctx, deleteOpts); err != nil {
		if isAWSErr(err, quicksight.ErrCodeResourceNotFoundException, "") {
			return nil
		}
		return diag.Errorf("Error deleting QuickSight User-group membership %s: %s", d.Id(), err)
	}

	return nil
}

func resourceAwsQuickSightGroupMembershipParseID(id string) (string, string, string, string, error) {
	parts := strings.SplitN(id, "/", 4)
	if len(parts) < 4 || parts[0] == "" || parts[1] == "" || parts[2] == "" || parts[3] == "" {
		return "", "", "", "", fmt.Errorf("unexpected format of ID (%s), expected AWS_ACCOUNT_ID/NAMESPACE/GROUP_NAME/USER_NAME", id)
	}
	return parts[0], parts[1], parts[2], parts[3], nil
}
