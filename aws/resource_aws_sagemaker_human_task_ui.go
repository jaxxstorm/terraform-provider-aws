package aws

import (
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/sagemaker/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func resourceAwsSagemakerHumanTaskUi() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSagemakerHumanTaskUiCreate,
		Read:   resourceAwsSagemakerHumanTaskUiRead,
		Update: resourceAwsSagemakerHumanTaskUiUpdate,
		Delete: resourceAwsSagemakerHumanTaskUiDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ui_template": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"content": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringLenBetween(1, 128000),
						},
						"content_sha256": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"url": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"human_task_ui_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 63),
					validation.StringMatch(regexp.MustCompile(`^[a-z0-9](-*[a-z0-9])*$`), "Valid characters are a-z, A-Z, 0-9, and - (hyphen)."),
				),
			},
			"tags":     tagsSchema(),
			"tags_all": tagsSchemaComputed(),
		},
		CustomizeDiff: SetTagsDiff,
	}
}

func resourceAwsSagemakerHumanTaskUiCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sagemakerconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(keyvaluetags.New(d.Get("tags").(map[string]interface{})))

	name := d.Get("human_task_ui_name").(string)
	input := &sagemaker.CreateHumanTaskUiInput{
		HumanTaskUiName: aws.String(name),
		UiTemplate:      expandSagemakerHumanTaskUiUiTemplate(d.Get("ui_template").([]interface{})),
	}

	if len(tags) > 0 {
		input.Tags = tags.IgnoreAws().SagemakerTags()
	}

	log.Printf("[DEBUG] Creating SageMaker HumanTaskUi: %s", input)
	_, err := conn.CreateHumanTaskUi(input)

	if err != nil {
		return fmt.Errorf("error creating SageMaker HumanTaskUi (%s): %w", name, err)
	}

	d.SetId(name)

	return resourceAwsSagemakerHumanTaskUiRead(d, meta)
}

func resourceAwsSagemakerHumanTaskUiRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sagemakerconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	humanTaskUi, err := finder.HumanTaskUiByName(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SageMaker HumanTaskUi (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading SageMaker HumanTaskUi (%s): %w", d.Id(), err)
	}

	arn := aws.StringValue(humanTaskUi.HumanTaskUiArn)
	d.Set("arn", arn)
	d.Set("human_task_ui_name", humanTaskUi.HumanTaskUiName)

	if err := d.Set("ui_template", flattenSagemakerHumanTaskUiUiTemplate(humanTaskUi.UiTemplate, d.Get("ui_template.0.content").(string))); err != nil {
		return fmt.Errorf("error setting ui_template: %w", err)
	}

	tags, err := keyvaluetags.SagemakerListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for SageMaker HumanTaskUi (%s): %w", d.Id(), err)
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

func resourceAwsSagemakerHumanTaskUiUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sagemakerconn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := keyvaluetags.SagemakerUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating SageMaker HumanTaskUi (%s) tags: %w", d.Id(), err)
		}
	}

	return resourceAwsSagemakerHumanTaskUiRead(d, meta)
}

func resourceAwsSagemakerHumanTaskUiDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sagemakerconn

	log.Printf("[DEBUG] Deleting SageMaker HumanTaskUi: %s", d.Id())
	_, err := conn.DeleteHumanTaskUi(&sagemaker.DeleteHumanTaskUiInput{
		HumanTaskUiName: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, sagemaker.ErrCodeResourceNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting SageMaker HumanTaskUi (%s): %w", d.Id(), err)
	}

	return nil
}

func expandSagemakerHumanTaskUiUiTemplate(l []interface{}) *sagemaker.UiTemplate {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &sagemaker.UiTemplate{
		Content: aws.String(m["content"].(string)),
	}

	return config
}

func flattenSagemakerHumanTaskUiUiTemplate(config *sagemaker.UiTemplateInfo, content string) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"content_sha256": aws.StringValue(config.ContentSha256),
		"url":            aws.StringValue(config.Url),
		"content":        content,
	}

	return []map[string]interface{}{m}
}
