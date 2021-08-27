package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/fsx"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/fsx/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/fsx/waiter"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func resourceAwsFsxBackup() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsFsxBackupCreate,
		Read:   resourceAwsFsxBackupRead,
		Update: resourceAwsFsxBackupUpdate,
		Delete: resourceAwsFsxBackupDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"file_system_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"kms_key_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags":     tagsSchemaComputed(),
			"tags_all": tagsSchemaComputed(),
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		CustomizeDiff: customdiff.Sequence(
			SetTagsDiff,
		),
	}
}

func resourceAwsFsxBackupCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).fsxconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(keyvaluetags.New(d.Get("tags").(map[string]interface{})))

	input := &fsx.CreateBackupInput{
		ClientRequestToken: aws.String(resource.UniqueId()),
		FileSystemId:       aws.String(d.Get("file_system_id").(string)),
	}

	if len(tags) > 0 {
		input.Tags = tags.IgnoreAws().FsxTags()
	}

	result, err := conn.CreateBackup(input)
	if err != nil {
		return fmt.Errorf("error creating FSx Backup: %w", err)
	}

	d.SetId(aws.StringValue(result.Backup.BackupId))

	log.Println("[DEBUG] Waiting for FSx backup to become available")
	if _, err := waiter.BackupAvailable(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for FSx Backup (%s) to be available: %w", d.Id(), err)
	}

	return resourceAwsFsxBackupRead(d, meta)
}

func resourceAwsFsxBackupUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).fsxconn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := keyvaluetags.FsxUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating FSx Backup (%s) tags: %w", d.Get("arn").(string), err)
		}
	}

	return resourceAwsFsxBackupRead(d, meta)
}

func resourceAwsFsxBackupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).fsxconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	backup, err := finder.BackupByID(conn, d.Id())
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] FSx Backup (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading FSx Backup (%s): %w", d.Id(), err)
	}

	d.Set("arn", backup.ResourceARN)
	d.Set("type", backup.Type)

	fs := backup.FileSystem
	d.Set("file_system_id", fs.FileSystemId)

	d.Set("kms_key_id", backup.KmsKeyId)

	d.Set("owner_id", backup.OwnerId)

	tags := keyvaluetags.FsxKeyValueTags(backup.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceAwsFsxBackupDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).fsxconn

	request := &fsx.DeleteBackupInput{
		BackupId: aws.String(d.Id()),
	}

	log.Printf("[INFO] Deleting FSx Backup: %s", d.Id())
	_, err := conn.DeleteBackup(request)

	if err != nil {
		if tfawserr.ErrCodeEquals(err, fsx.ErrCodeBackupNotFound) {
			return nil
		}
		return fmt.Errorf("error deleting FSx Backup (%s): %w", d.Id(), err)
	}

	log.Println("[DEBUG] Waiting for backup to delete")
	if _, err := waiter.BackupDeleted(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for FSx Backup (%s) to deleted: %w", d.Id(), err)
	}

	return nil
}
