package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/efs"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/efs/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func resourceAwsEfsFileSystemPolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsEfsFileSystemPolicyPut,
		Read:   resourceAwsEfsFileSystemPolicyRead,
		Update: resourceAwsEfsFileSystemPolicyPut,
		Delete: resourceAwsEfsFileSystemPolicyDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"bypass_policy_lockout_safety_check": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"file_system_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"policy": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: suppressEquivalentJsonDiffs,
			},
		},
	}
}

func resourceAwsEfsFileSystemPolicyPut(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).efsconn

	fsID := d.Get("file_system_id").(string)
	input := &efs.PutFileSystemPolicyInput{
		BypassPolicyLockoutSafetyCheck: aws.Bool(d.Get("bypass_policy_lockout_safety_check").(bool)),
		FileSystemId:                   aws.String(fsID),
		Policy:                         aws.String(d.Get("policy").(string)),
	}

	log.Printf("[DEBUG] Putting EFS File System Policy: %s", input)
	_, err := conn.PutFileSystemPolicy(input)

	if err != nil {
		return fmt.Errorf("error putting EFS File System Policy (%s): %w", fsID, err)
	}

	d.SetId(fsID)

	return resourceAwsEfsFileSystemPolicyRead(d, meta)
}

func resourceAwsEfsFileSystemPolicyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).efsconn

	output, err := finder.FileSystemPolicyByID(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EFS File System Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading EFS File System Policy (%s): %w", d.Id(), err)
	}

	d.Set("file_system_id", output.FileSystemId)
	d.Set("policy", output.Policy)

	return nil
}

func resourceAwsEfsFileSystemPolicyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).efsconn

	log.Printf("[DEBUG] Deleting EFS File System Policy: %s", d.Id())
	_, err := conn.DeleteFileSystemPolicy(&efs.DeleteFileSystemPolicyInput{
		FileSystemId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, efs.ErrCodeFileSystemNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting EFS File System Policy (%s): %w", d.Id(), err)
	}

	return nil
}
