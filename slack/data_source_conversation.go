package slack

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
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
	config := m.(*ProviderConfig)
	client := config.Client
	channelID := d.Get("channel_id").(string)
	channelName := d.Get("name").(string)
	isPrivate := d.Get("is_private").(bool)

	var (
		channel *slack.Channel
		users   []string
		err     error
	)

	if channelID != "" {
		channel, err = WithRetryWithResult(ctx, config.RetryConfig, func() (*slack.Channel, error) {
			return client.GetConversationInfoContext(ctx, &slack.GetConversationInfoInput{
				ChannelID: channelID,
			})
		})
		if err != nil {
			return diag.Errorf("couldn't get conversation info for %s: %s", channelID, err)
		}
	} else if channelName != "" {
		channel, err = WithRetryWithResult(ctx, config.RetryConfig, func() (*slack.Channel, error) {
			return findExistingChannel(ctx, client, channelName, isPrivate)
		})
		if err != nil {
			return diag.Errorf("couldn't get conversation info for %s: %s", channelName, err)
		}
	} else {
		return diag.Errorf("channel_id or name must be set")
	}

	err = WithRetry(ctx, config.RetryConfig, func() error {
		var retryErr error
		users, _, retryErr = client.GetUsersInConversationContext(ctx, &slack.GetUsersInConversationParameters{
			ChannelID: channel.ID,
		})
		return retryErr
	})
	if err != nil {
		return diag.Errorf("couldn't get users in conversation for %s: %s", channel.ID, err)
	}

	return updateChannelData(d, channel, users)
}
