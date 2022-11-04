// Code generated by internal/generate/tags/main.go; DO NOT EDIT.
package s3legacy

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

// []*SERVICE.Tag handling

// Tags returns s3 service tags.
func Tags(tags tftags.KeyValueTags) []*s3.Tag {
	result := make([]*s3.Tag, 0, len(tags))

	for k, v := range tags.Map() {
		tag := &s3.Tag{
			Key:   aws.String(k),
			Value: aws.String(v),
		}

		result = append(result, tag)
	}

	return result
}

// KeyValueTags creates tftags.KeyValueTags from s3 service tags.
func KeyValueTags(tags []*s3.Tag) tftags.KeyValueTags {
	m := make(map[string]*string, len(tags))

	for _, tag := range tags {
		m[aws.StringValue(tag.Key)] = tag.Value
	}

	return tftags.New(m)
}
