---
subcategory: "Elemental MediaLive"
layout: "aws"
page_title: "AWS: aws_medialive_channel"
description: |-
  Resource for managing an AWS MediaLive Channel.
---

# Resource: aws_medialive_channel

Resource for managing an AWS MediaLive Channel.

## Example Usage

### Basic Usage

```terraform
resource "aws_medialive_channel" "example" {
  name          = "example-channel"
  channel_class = "STANDARD"
  role_arn      = aws_iam_role.example.arn

  input_specification {
    codec            = "AVC"
    input_resolution = "HD"
    maximum_bitrate  = "MAX_20_MBPS"
  }

  input_attachments {
    input_attachment_name = "example-input"
    input_id              = aws_medialive_input.example.id

  }

  destinations {
    id = "destination"

    settings {
      url = "s3://${aws_s3_bucket.main.id}/test1"
    }

    settings {
      url = "s3://${aws_s3_bucket.main2.id}/test2"
    }
  }

  encoder_settings {
    timecode_config {
      source = "EMBEDDED"
    }

    audio_descriptions {
      audio_selector_name = "example audio selector"
      name                = "audio-selector"
    }

    video_descriptions {
      name = "example-vidoe"
    }

    output_groups {
      output_group_settings {
        archive_group_settings {
          destination {
            destination_ref_id = "destination"
          }
        }
      }

      outputs {
        output_name             = "example-name"
        video_description_name  = "example-vidoe"
        audio_description_names = ["audio-selector"]
        output_settings {
          archive_output_settings {
            name_modifier = "_1"
            extension     = "m2ts"
            container_settings {
              m2ts_settings {
                audio_buffer_model = "ATSC"
                buffer_model       = "MULTIPLEX"
                rate_mode          = "CBR"
              }
            }
          }
        }
      }
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `channel_class` - (Required) Concise argument description.
* `destinations` - (Required) Destinations for channel. See [Destinations](#destinations) for more details.
* `encoder_settings` - (Required) Encoder settings. See [Encoder Settings](#encoder-settings) for more details.
* `input_specification` - (Required) Specification of network and file inputs for the channel.
* `name` - (Required) Name of the Channel.

The following arguments are optional:

* `cdi_input_specification` - (Optional) Specification of CDI inputs for this channel. See [CDI Input Specification](#cdi-input-specification) for more details.
* `input_attachments` - (Optional) Input attachments for the channel. See [Input Attachments](#input-attachments) for more details.
* `log_level` - (Optional) The log level to write to Cloudwatch logs.
* `maintenance` - (Optional) Maintenance settings for this channel. See [Maintenance](#maintenance) for more details.
* `role_arn` - (Optional) Concise argument description.
* `tags` - (Optional) A map of tags to assign to the channel. If configured with a provider `default_tags` configuration block present, tags with matching keys will overwrite those defined at the provider-level.
* `vpc` - (Optional) Settings for the VPC outputs.

### CDI Input Specification

* `resolution` - (Required) - Maximum CDI input resolution.

### Destinations

* `id` - (Required) User-specified id. Ths is used in an output group or an output.
* `media_package_settings` - (Optional) Destination settings for a MediaPackage output; one destination for both encoders. See [Media Package Settings](#media-package-settings) for more details.
* `multiplex_settings` - (Optional) Destination settings for a Multiplex output; one destination for both encoders. See [Multiplex Settings](#multiplex-settings) for more details.
* `settings` - (Optional) Destination settings for a standard output; one destination for each redundant encoder. See [Settings](#settings) for more details.

### Encoder Settings

* `audio_descriptions` - (Required) Audio descriptions for the channel. See [Audio Descriptions](#audio-descriptions) for more details.
* `output_groups` - (Required) Output groups for the channel. See [Output Groups](#output-groups) for more details.
* `timecode_config` - (Required) Contains settings used to acquire and adjust timecode information from inputs. See [Timecode Config](#timecode-config) for more details.
* `video_descriptions` - (Required) Video Descriptions. See [Video Descriptions](#video-descriptions) for more details.

### Input Attachments

* `input_attachment_name` - (Optional) User-specified name for the attachment.
* `input_id` - (Required) The ID of the input.
* `input_settings` - (Optional) Settings of an input. See [Input Settings](#input-settings) for more details

### Input Settings

* `audio_selectors` - (Optional) Used to select the audio stream to decode for inputs that have multiple. See [Audio Selectors](#audio-selectors) for more details.

### Audio Selectors

* `name` - (Required) The name of the audio selector.

### Maintenance

* `maintenance_day` - (Optional) The day of the week to use for maintenance.
* `maintenance_start_time` - (Optional) The hour maintenance will start.

### Media Package Settings

* `channel_id` - (Required) ID of the channel in MediaPackage that is the destination for this output group.

### Multiplex Settings

* `multiplex_id` - (Required) The ID of the Multiplex that the encoder is providing output to.
* `program_name` - (Optional) The program name of the Multiplex program that the encoder is providing output to.

### Settings

* `password_param` - (Optional) Key used to extract the password from EC2 Parameter store.
* `stream_name` - (Optional) Stream name RTMP destinations (URLs of type rtmp://)
* `url` - (Optional) A URL specifying a destination.
* `username` - (Optional) Username for destination.

### Audio Descriptions

* `audio_selector_name` - (Required) The name of the audio selector used as the source for this AudioDescription.
* `name` - (Required) The name of this audio description.
* `audio_normalization_settings` - (Optional) Advanced audio normalization settings. See [Audio Normalization Settings](#audio-normalization-settings) for more details.
* `audio_type` - (Optional) Applies only if audioTypeControl is useConfigured. The values for audioType are defined in ISO-IEC 13818-1.
* `audio_type_control` - (Optional) Determined how audio type is determined.
* `audio_watermark_settings` - (Optional) Settings to configure one or more solutions that insert audio watermarks in the audio encode. See [Audio Watermark Settings](#audio-watermark-settings) for more details.

### Audio Normalization Settings

* `algorithm` - (Optional) Audio normalization algorithm to use. itu17701 conforms to the CALM Act specification, itu17702 to the EBU R-128 specification.
* `algorithm_control` - (Optional) Algorithm control for the audio description.
* `target_lkfs` - (Optional) Target LKFS (loudness) to adjust volume to.

### Audio Watermark Settings

* `nielsen_watermark_settings` - (Optional) Settings to configure Nielsen Watermarks in the audio encode. See [Nielsen Watermark Settings](#nielsen-watermark-settings) for more details.

### Nielsen Watermark Settings

* `nielsen_cbet_settings` - (Optional) Used to insert watermarks of type Nielsen CBET. See [Nielsen CBET Settings](#nielsen-cbet-settings) for more details.
* `nielsen_distribution_type` - (Optional) Distribution types to assign to the watermarks. Options are `PROGRAM_CONTENT` and `FINAL_DISTRIBUTOR`.
* `nielsen_naes_ii_nw_settings` - (Optional) Used to insert watermarks of type Nielsen NAES, II (N2) and Nielsen NAES VI (NW). See [Nielsen NAES II NW Settings](#nielsen-naes-ii-nw-settings) for more details.

### Nielsen CBET Settings

* `cbet_check_digit` - (Required) CBET check digits to use for the watermark.
* `cbet_stepaside` - (Required) Determines the method of CBET insertion mode when prior encoding is detected on the same layer.
* `csid` - (Required) CBET source ID to use in the watermark.

### Nielsen NAES II NW Settings

* `check_digit` - (Required) Check digit string for the watermark.
* `sid` - (Required) The Nielsen Source ID to include in the watermark.

### Output Groups

* `output_group_settings` - (Required) Settings associated with the output group. See [Output Group Settings](#output-group-settings) for more details.
* `outputs` - (Required) List of outputs. See [Outputs](#outputs) for more details.
* `name` - (Optional) Custom output group name defined by the user.

### Output Group Settings

* `archive_group_settings` - (Optional) Archive group settings. See [Archive Group Settings](#archive-group-settings) for more details.
* `media_package_group_settings` - (Optional) Media package group settings. See [Media Package Group Settings](#media-package-group-settings) for more details.
* `multiplex_group_sttings` - (Optional) Multiplex group settings. Attribute can be passed as an empty block.
* `rtmp_group_settings` - (Optional) RTMP group settings. See [RTMP Group Settings](#rtmp-group-settings) for more details.
* `udp_group_sttings` - (Optional) UDP group settings. See [UDP Group Settings](#udp-group-settings) for more details.

### Outputs

* `output_settings` - (Required) Settings for output. See [Output Settings](#output-settings) for more details.
* `audio_description_names` - (Optional) The names of the audio descriptions used as audio sources for the output.
* `caption_description_names` - (Optional) The names of the caption descriptions used as audio sources for the output.
* `output_name` - (Required) The name used to identify an output.
* `video_description_name` - (Optional) The name of the video description used as audio sources for the output.

### Timecode Config

* `source` - (Optional) The source for the timecode that will be associated with the events outputs.
* `sync_threshold` - (Optional) Threshold in frames beyond which output timecode is resynchronized to the input timecode.

### Video Descriptions

* `height` - Output video height in pixels.
* `respond_to_afd` - (Optional) Indicate how to respond to the AFD values that might be in the input video.
* `scaling_behavior` - (Optional) Behavior on how to scale.
* `sharpness` - (Optional) Changes the strength of the anti-alias filter used for scaling.
* `width` - (Optional) Output video width in pixels.

### Archive Group Settings

* `destination` - (Required) A director and base filename where archive files should be written. See [Destination](#destination) for more details.
* `archive_cdn_settings` - (Optional) Parameters that control the interactions with the CDN. See [Archive CDN Settings](#archive-cdn-settings) for more details.
* `rollover_interval` - (Optional) Number of seconds to write to archive file before closing and starting a new one.

### Media Package Group Settings

* `destination` - (Required) A director and base filename where archive files should be written. See [Destination](#destination) for more details.

### RTMP Group Settings

* `ad_markers` - (Optional) The ad marker type for this output group.
* `authentication_scheme` - (Optional) Authentication scheme to use when connecting with CDN.
* `cache_full_behavior` - (Optional) Controls behavior when content cache fills up.
* `cache_length` - (Optional) Cache length in seconds, is used to calculate buffer size.
* `caption_data` - (Optional) Controls the types of data that passes to onCaptionInfo outputs.
* `input_loss_action` - (Optional) Controls the behavior of the RTMP group if input becomes unavailable.
* `restart_delay` - (Optional) Number of seconds to wait until a restart is initiated.

### UDP Group Settings

* `input_loss_action` - (Optional) Specifies behavior of last resort when input video os lost.
* `timed_metadata_id3_frame` - (Optional) Indicates ID3 frame that has the timecode.
* `timed_metadta_id3_perios`- (Optional) Timed metadata interval in seconds.

### Destination

* `destination_ref_id` - (Required) Reference ID for the destination.

### Archive CDN Settings

* `archive_s3_settings` - (Optional) Archive S3 Settings. See [Archive S3 Settings](#archive-s3-settings) for more details.

### Archive S3 Settings

* `canned_acl` - (Optional) Specify the canned ACL to apply to each S3 request.

### Output Settings

* `archive_output_settings` - (Optional) Archive output settings. See [Archive Output Settings](#archive-output-settings) for more details.
* `media_package_output_settings` - (Optional) Media package output settings. This can be set as an empty block.
* `multiplex_output_settings` - (Optional) Multiplex output settings. See [Multiplex Output Settings](#multiplex-output-settings) for more details.
* `rtmp_output_settings` - (Optional) RTMP output settings. See [RTMP Output Settings](#rtmp-output-settings) for more details.
* `udp_output_settings` - (Optional) UDP output settings. See [UDP Output Settings](#udp-output-settings) for more details

### Archive Output Settings

* `container_settings` - (Required) Settings specific to the container type of the file. See [Container Settings](#container-settings) for more details.
* `extension` - (Optional) Output file extension.
* `name_modifier` - (Optional) String concatenated to the end of the destination filename. Required for multiple outputs of the same type.

### Multiplex Output Settings

* `destination` - (Required) Destination is a multiplex. See [Destination](#destination) for more details.

### RTMP Output Settings

- `destination` - (Required) The RTMP endpoint excluding the stream name. See [Destination](#destination) for more details.
- `certificate_mode` - (Optional) Setting to allow self signed or verified RTMP certificates.
- `connection_retry_interval` - (Optional) Number of seconds to wait before retrying connection to the flash media server if the connection is lost.
- `num_retries` - (Optional) Number of retry attempts.

### Container Settings

* `m2ts_settings` - (Optional) M2ts Settings. See [M2ts Settings](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-medialive-channel-m2tssettings.html) for more details.
* `raw_settings`- (Optional) Raw Settings. This can be set as an empty block.

### UDP Output Settings

* `container_settings` - (Required) UDP container settings. See [Container Settings](#container-settings) for more details.
* `destination` - (Required) Destination address and port number for RTP or UDP packets. See [Destination](#destination) for more details.
* `buffer_msec` - (Optional) UDP output buffering in milliseconds.
* `fec_output_setting` - (Optional) Settings for enabling and adjusting Forward Error Correction on UDP outputs. See [FEC Output Settings](#fec-output-settings) for more details.

### FEC Output Settings

* `column_depth` - (Optional) The height of the FEC protection matrix.
* `include_fec` - (Optional) Enables column oly or column and row based FEC.
* `row_length` - (Optional) The width of the FEC protection matrix.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - ARN of the Channel.
* `channel_id` - ID of the Channel.

## Timeouts

Configuration options:

* `create` - (Default `15m`)
* `update` - (Default `15m`)
* `delete` - (Default `15m`)

## Import

MediaLive Channel can be imported using the `channel_id`, e.g.,

```
$ terraform import aws_medialive_channel.example 1234567
```
