package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/transfer"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	tftransfer "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/transfer"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/transfer/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func resourceAwsTransferAccess() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsTransferAccessCreate,
		Read:   resourceAwsTransferAccessRead,
		Update: resourceAwsTransferAccessUpdate,
		Delete: resourceAwsTransferAccessDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"external_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 256),
			},

			"home_directory": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 1024),
			},

			"home_directory_mappings": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 50,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"entry": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(0, 1024),
						},
						"target": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(0, 1024),
						},
					},
				},
			},

			"home_directory_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      transfer.HomeDirectoryTypePath,
				ValidateFunc: validation.StringInSlice(transfer.HomeDirectoryType_Values(), false),
			},

			"policy": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateFunc:     validateIAMPolicyJson,
				DiffSuppressFunc: suppressEquivalentAwsPolicyDiffs,
			},

			"posix_profile": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"gid": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"uid": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"secondary_gids": {
							Type:     schema.TypeSet,
							Elem:     &schema.Schema{Type: schema.TypeInt},
							Optional: true,
						},
					},
				},
			},

			"role": {
				Type: schema.TypeString,
				// Although Role is required in the API it is not currently returned on Read.
				// Required:     true,
				Optional:     true,
				ValidateFunc: validateArn,
			},

			"server_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateTransferServerID,
			},
		},
	}
}

func resourceAwsTransferAccessCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).transferconn

	externalID := d.Get("external_id").(string)
	serverID := d.Get("server_id").(string)
	id := tftransfer.AccessCreateResourceID(serverID, externalID)
	input := &transfer.CreateAccessInput{
		ExternalId: aws.String(externalID),
		ServerId:   aws.String(serverID),
	}

	if v, ok := d.GetOk("home_directory"); ok {
		input.HomeDirectory = aws.String(v.(string))
	}

	if v, ok := d.GetOk("home_directory_mappings"); ok {
		input.HomeDirectoryMappings = expandAwsTransferHomeDirectoryMappings(v.([]interface{}))
	}

	if v, ok := d.GetOk("home_directory_type"); ok {
		input.HomeDirectoryType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("policy"); ok {
		input.Policy = aws.String(v.(string))
	}

	if v, ok := d.GetOk("posix_profile"); ok {
		input.PosixProfile = expandTransferUserPosixUser(v.([]interface{}))
	}

	if v, ok := d.GetOk("role"); ok {
		input.Role = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating Transfer Access: %s", input)
	_, err := conn.CreateAccess(input)

	if err != nil {
		return fmt.Errorf("error creating Transfer Access (%s): %w", id, err)
	}

	d.SetId(id)

	return resourceAwsTransferAccessRead(d, meta)
}

func resourceAwsTransferAccessRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).transferconn

	serverID, externalID, err := tftransfer.AccessParseResourceID(d.Id())

	if err != nil {
		return fmt.Errorf("error parsing Transfer Access ID: %w", err)
	}

	access, err := finder.AccessByServerIDAndExternalID(conn, serverID, externalID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Transfer Access (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Transfer Access (%s): %w", d.Id(), err)
	}

	d.Set("external_id", access.ExternalId)
	d.Set("home_directory", access.HomeDirectory)
	if err := d.Set("home_directory_mappings", flattenAwsTransferHomeDirectoryMappings(access.HomeDirectoryMappings)); err != nil {
		return fmt.Errorf("error setting home_directory_mappings: %w", err)
	}
	d.Set("home_directory_type", access.HomeDirectoryType)
	d.Set("policy", access.Policy)
	if err := d.Set("posix_profile", flattenTransferUserPosixUser(access.PosixProfile)); err != nil {
		return fmt.Errorf("error setting posix_profile: %w", err)
	}
	// Role is currently not returned via the API.
	// d.Set("role", access.Role)
	d.Set("server_id", serverID)

	return nil
}

func resourceAwsTransferAccessUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).transferconn

	serverID, externalID, err := tftransfer.AccessParseResourceID(d.Id())

	if err != nil {
		return fmt.Errorf("error parsing Transfer Access ID: %w", err)
	}

	input := &transfer.UpdateAccessInput{
		ExternalId: aws.String(externalID),
		ServerId:   aws.String(serverID),
	}

	if d.HasChange("home_directory") {
		input.HomeDirectory = aws.String(d.Get("home_directory").(string))
	}

	if d.HasChange("home_directory_mappings") {
		input.HomeDirectoryMappings = expandAwsTransferHomeDirectoryMappings(d.Get("home_directory_mappings").([]interface{}))
	}

	if d.HasChange("home_directory_type") {
		input.HomeDirectoryType = aws.String(d.Get("home_directory_type").(string))
	}

	if d.HasChange("policy") {
		input.Policy = aws.String(d.Get("policy").(string))
	}

	if d.HasChange("posix_profile") {
		input.PosixProfile = expandTransferUserPosixUser(d.Get("posix_profile").([]interface{}))
	}

	if d.HasChange("role") {
		input.Role = aws.String(d.Get("role").(string))
	}

	log.Printf("[DEBUG] Updating Transfer Access: %s", input)
	_, err = conn.UpdateAccess(input)

	if err != nil {
		return fmt.Errorf("error updating Transfer Access (%s): %w", d.Id(), err)
	}

	return resourceAwsTransferAccessRead(d, meta)
}

func resourceAwsTransferAccessDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).transferconn

	serverID, externalID, err := tftransfer.AccessParseResourceID(d.Id())

	if err != nil {
		return fmt.Errorf("error parsing Transfer Access ID: %w", err)
	}

	log.Printf("[DEBUG] Deleting Transfer Access: %s", d.Id())
	_, err = conn.DeleteAccess(&transfer.DeleteAccessInput{
		ExternalId: aws.String(externalID),
		ServerId:   aws.String(serverID),
	})

	if tfawserr.ErrCodeEquals(err, transfer.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Transfer Access (%s): %w", d.Id(), err)
	}

	return nil
}
