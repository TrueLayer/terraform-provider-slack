package slack

import (
	"context"
	"fmt"

	"errors"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/slack-go/slack"
)

func dataSourceConversation() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceSlackConversationRead,

		Schema: map[string]*schema.Schema{
			"channel_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"is_private": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"topic": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"purpose": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"creator": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"is_archived": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"is_shared": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"is_ext_shared": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"is_org_shared": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"is_general": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func dataSourceSlackConversationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*slack.Client)
	channelID := d.Get("channel_id").(string)
	channelName := d.Get("name").(string)
	isPrivate := d.Get("is_private").(bool)

	var (
		channel *slack.Channel
		users   []string
		err     error
	)

	err = retry.RetryContext(ctx, slackRetryTimeout, func() *retry.RetryError {
		var rlerr *slack.RateLimitedError
		if channelID != "" {
			channel, err = client.GetConversationInfoContext(ctx, &slack.GetConversationInfoInput{
				ChannelID: channelID,
			})
			if errors.As(err, &rlerr) {
				time.Sleep(rlerr.RetryAfter)
				return retry.RetryableError(err)
			}
			if err != nil {
				return retry.NonRetryableError(fmt.Errorf("couldn't get conversation info for %s: %w", channelID, err))
			}
		} else if channelName != "" {
			channel, err = findExistingChannel(ctx, client, channelName, isPrivate)
			if errors.As(err, &rlerr) {
				time.Sleep(rlerr.RetryAfter)
				return retry.RetryableError(err)
			}
			if err != nil {
				return retry.NonRetryableError(fmt.Errorf("couldn't get conversation info for %s: %w", channelName, err))
			}
		} else {
			return retry.NonRetryableError(fmt.Errorf("channel_id or name must be set"))
		}
		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	err = retry.RetryContext(ctx, slackRetryTimeout, func() *retry.RetryError {
		var rlerr *slack.RateLimitedError
		users, _, err = client.GetUsersInConversationContext(ctx, &slack.GetUsersInConversationParameters{
			ChannelID: channel.ID,
		})
		if errors.As(err, &rlerr) {
			time.Sleep(rlerr.RetryAfter)
			return retry.RetryableError(err)
		}
		if err != nil {
			return retry.NonRetryableError(fmt.Errorf("couldn't get users in conversation for %s: %w", channel.ID, err))
		}
		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return updateChannelData(d, channel, users)
}
