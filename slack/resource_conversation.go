package slack

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/slack-go/slack"
)

const (
	conversationActionOnDestroyNone    = "none"
	conversationActionOnDestroyArchive = "archive"

	conversationActionOnUpdatePermanentMembersNone = "none"
	conversationActionOnUpdatePermanentMembersKick = "kick"

	// 100 is default, slack docs recommend no more than 200, but 1000 is the max.
	// See also https://github.com/slack-go/slack/blob/master/users.go#L305
	cursorLimit = 200
)

var (
	conversationActionValidValues = []string{
		conversationActionOnDestroyNone,
		conversationActionOnDestroyArchive,
	}
	conversationActionOnUpdatePermanentMembersValidValues = []string{
		conversationActionOnUpdatePermanentMembersNone,
		conversationActionOnUpdatePermanentMembersKick,
	}

	validateConversationActionOnDestroyValue           = validation.StringInSlice(conversationActionValidValues, false)
	validateConversationActionOnUpdatePermanentMembers = validation.StringInSlice(conversationActionOnUpdatePermanentMembersValidValues, false)
)

func resourceSlackConversation() *schema.Resource {
	return &schema.Resource{
		ReadContext:   resourceSlackConversationRead,
		CreateContext: resourceSlackConversationCreate,
		UpdateContext: resourceSlackConversationUpdate,
		DeleteContext: resourceSlackConversationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"topic": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"purpose": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"permanent_members": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Set:      schema.HashString,
				Optional: true,
			},
			"created": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"creator": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"is_private": {
				Type:     schema.TypeBool,
				Required: true,
			},
			"is_archived": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
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
			"action_on_destroy": {
				Type:         schema.TypeString,
				Description:  "Either of none or archive",
				Optional:     true,
				Default:      "archive",
				ValidateFunc: validateConversationActionOnDestroyValue,
			},
			"action_on_update_permanent_members": {
				Type:         schema.TypeString,
				Description:  "Either of none or kick",
				Optional:     true,
				Default:      "kick",
				ValidateFunc: validateConversationActionOnUpdatePermanentMembers,
			},
			"adopt_existing_channel": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
		},
	}
}

func resourceSlackConversationCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	config := m.(*ProviderConfig)
	client := config.Client

	name := d.Get("name").(string)
	isPrivate := d.Get("is_private").(bool)

	channel, err := client.CreateConversationContext(ctx, slack.CreateConversationParams{
		ChannelName: name,
		IsPrivate:   isPrivate,
	})
	if err != nil && err.Error() == "name_taken" && d.Get("adopt_existing_channel").(bool) {
		channel, err = findExistingChannel(ctx, client, name, isPrivate)
		if err == nil && channel.IsArchived {
			// ensure unarchived first if adopting existing channel, else other calls below will fail
			if err := client.UnArchiveConversationContext(ctx, channel.ID); err != nil {
				if err.Error() != "not_archived" {
					return diag.Errorf("couldn't unarchive conversation %s: %s", channel.ID, err)
				}
			}
		}
	}
	if err != nil {
		return diag.Errorf("could not create conversation %s: %s", name, err)
	}

	err = updateChannelMembers(ctx, d, client, channel.ID)
	if err != nil {
		return diag.FromErr(err)
	}

	if topic, ok := d.GetOk("topic"); ok {
		if _, err := client.SetTopicOfConversationContext(ctx, channel.ID, topic.(string)); err != nil {
			return diag.Errorf("couldn't set conversation topic %s: %s", topic.(string), err)
		}
	}

	if purpose, ok := d.GetOk("purpose"); ok {
		if _, err := client.SetPurposeOfConversationContext(ctx, channel.ID, purpose.(string)); err != nil {
			return diag.Errorf("couldn't set conversation purpose %s: %s", purpose.(string), err)
		}
	}

	if isArchived, ok := d.GetOk("is_archived"); ok {
		if isArchived.(bool) {
			err := archiveConversationWithContext(ctx, client, channel.ID)
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	d.SetId(channel.ID)
	return resourceSlackConversationRead(ctx, d, m)
}

func findExistingChannel(ctx context.Context, client ClientInterface, name string, isPrivate bool) (*slack.Channel, error) {
	// find the existing channel. Sadly, there is no non-admin API to search by name,
	// so we must search through ALL the channels
	// Note: This function is called from within WithRetryWithResult, so rate limiting is handled by the wrapper
	tflog.Info(ctx, "Looking for channel %s", map[string]interface{}{"channel": name})
	paginationComplete := false
	cursor := ""       // initial empty cursor to begin at start of list
	var types []string // default value with empty list is "public_channel"
	if isPrivate {
		types = append(types, "private_channel")
	}
	for !paginationComplete {
		channels, nextCursor, err := client.GetConversationsContext(ctx, &slack.GetConversationsParameters{
			Cursor:          cursor,
			Limit:           cursorLimit,
			Types:           types,
			ExcludeArchived: true,
		})
		tflog.Debug(ctx, "new page of channels",
			map[string]interface{}{
				"numChannels": len(channels),
				"nextCursor":  nextCursor,
				"err":         err})
		if err != nil {
			return nil, fmt.Errorf("couldn't get conversation context: %s", err.Error())
		}

		// see if channel in current batch
		for _, c := range channels {
			tflog.Trace(ctx, "checking channel", map[string]interface{}{"channel": c.Name})
			if c.Name == name {
				tflog.Info(ctx, "found channel")
				return &c, nil
			}
		}
		// not found so far, move on to next cursor, if pagination incomplete
		paginationComplete = nextCursor == ""
		cursor = nextCursor
	}
	// looked through entire list, but didn't find matching name
	return nil, fmt.Errorf("could not find channel with name %s", name)
}

func updateChannelMembers(ctx context.Context, d *schema.ResourceData, client ClientInterface, channelID string) error {
	members := d.Get("permanent_members").(*schema.Set)

	userIDs := schemaSetToSlice(members)
	channel, err := client.GetConversationInfoContext(ctx, &slack.GetConversationInfoInput{
		ChannelID: channelID,
	})
	if err != nil {
		return fmt.Errorf("could not retrieve conversation info for ID %s: %w", channelID, err)
	}

	apiUserInfo, err := client.AuthTest()

	if err != nil {
		return fmt.Errorf("error authenticating with slack %w", err)
	}
	userIDs = remove(userIDs, apiUserInfo.UserID)
	userIDs = remove(userIDs, channel.Creator)

	channelUsers, _, err := client.GetUsersInConversationContext(ctx, &slack.GetUsersInConversationParameters{
		ChannelID: channel.ID,
	})

	if err != nil {
		return fmt.Errorf("could not retrieve conversation users for ID %s: %w", channelID, err)
	}

	// first, ensure the api user is in the channel, otherwise other member modifications below may fail
	if _, _, _, err := client.JoinConversationContext(ctx, channelID); err != nil {
		if err.Error() != "already_in_channel" && err.Error() != "method_not_supported_for_channel_type" {
			return fmt.Errorf("api user could not join conversation: %w", err)
		}
	}

	action := d.Get("action_on_update_permanent_members").(string)
	if action == conversationActionOnUpdatePermanentMembersKick {
		for _, currentMember := range channelUsers {
			if currentMember != channel.Creator && currentMember != apiUserInfo.UserID && !contains(userIDs, currentMember) {
				if err := client.KickUserFromConversationContext(ctx, channelID, currentMember); err != nil {
					return fmt.Errorf("couldn't kick user from conversation: %w", err)
				}
			}
		}
	}

	if len(userIDs) > 0 {
		if _, err := client.InviteUsersToConversationContext(ctx, channelID, userIDs...); err != nil {
			if err.Error() != "already_in_channel" {
				return fmt.Errorf("couldn't invite users to conversation: %w", err)
			}
		}
	}

	return nil
}

func resourceSlackConversationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	config := m.(*ProviderConfig)
	client := config.Client
	id := d.Id()
	var (
		diags   diag.Diagnostics
		channel *slack.Channel
		users   []string
		err     error
	)

	channel, err = WithRetryWithResult(ctx, config.RetryConfig, func() (*slack.Channel, error) {
		return client.GetConversationInfoContext(ctx, &slack.GetConversationInfoInput{
			ChannelID: id,
		})
	})
	if err != nil {
		if err.Error() == "channel_not_found" {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  fmt.Sprintf("channel with ID %s not found, removing from state", id),
			})
			d.SetId("")
			return diags
		}
		return diag.Errorf("couldn't get conversation info for %s: %s", id, err)
	}
	if d.Id() == "" {
		return diags
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

func resourceSlackConversationUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	config := m.(*ProviderConfig)
	client := config.Client

	id := d.Id()

	if d.HasChange("name") {
		if _, err := client.RenameConversationContext(ctx, id, d.Get("name").(string)); err != nil {
			return diag.Errorf("couldn't rename conversation: %s", err)
		}
	}

	if d.HasChange("topic") {
		topic := d.Get("topic")
		if _, err := client.SetTopicOfConversationContext(ctx, id, topic.(string)); err != nil {
			return diag.Errorf("couldn't set conversation topic %s: %s", topic.(string), err)
		}
	}

	if d.HasChange("purpose") {
		purpose := d.Get("purpose")
		if _, err := client.SetPurposeOfConversationContext(ctx, id, purpose.(string)); err != nil {
			return diag.Errorf("couldn't set conversation purpose %s: %s", purpose.(string), err)
		}
	}

	if d.HasChange("is_archived") {
		isArchived := d.Get("is_archived")
		if isArchived.(bool) {
			err := archiveConversationWithContext(ctx, client, id)
			if err != nil {
				return diag.FromErr(err)
			}
		} else {
			if err := client.UnArchiveConversationContext(ctx, id); err != nil {
				if err.Error() != "not_archived" {
					return diag.Errorf("couldn't unarchive conversation %s: %s", id, err)
				}
			}
		}
	}

	if d.HasChange("permanent_members") {
		err := updateChannelMembers(ctx, d, client, id)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceSlackConversationRead(ctx, d, m)
}

func resourceSlackConversationDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	config := m.(*ProviderConfig)
	client := config.Client

	id := d.Id()
	action := d.Get("action_on_destroy").(string)
	switch action {
	case conversationActionOnDestroyNone:
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  fmt.Sprintf("conversation %s (%s) won't be archived on destroy", id, d.Get("name")),
			Detail:   fmt.Sprintf("action_on_destroy is set to %s which does not archive the conversation ", conversationActionOnDestroyNone),
		})
	case conversationActionOnDestroyArchive:
		err := archiveConversationWithContext(ctx, client, id)
		if err != nil {
			if err.Error() == "channel_not_found" {
				return diags
			}
			return diag.FromErr(err)
		}
	default:
		return diag.Errorf("unknown action_on_destroy value. Valid values are %v", conversationActionValidValues)
	}

	return diags
}

func updateChannelData(d *schema.ResourceData, channel *slack.Channel, _ []string) diag.Diagnostics {
	if channel.ID == "" {
		return diag.Errorf("error setting id: returned channel does not have an id")
	}
	d.SetId(channel.ID)

	if d.Get("channel_id") != nil {
		if err := d.Set("channel_id", channel.ID); err != nil {
			return diag.Errorf("error setting channel_id: %s", err)
		}
	}

	if err := d.Set("name", channel.Name); err != nil {
		return diag.Errorf("error setting name: %s", err)
	}

	if err := d.Set("topic", channel.Topic.Value); err != nil {
		return diag.Errorf("error setting topic: %s", err)
	}

	if err := d.Set("purpose", channel.Purpose.Value); err != nil {
		return diag.Errorf("error setting purpose: %s", err)
	}

	if err := d.Set("is_archived", channel.IsArchived); err != nil {
		return diag.Errorf("error setting is_archived: %s", err)
	}

	if err := d.Set("is_shared", channel.IsShared); err != nil {
		return diag.Errorf("error setting is_shared: %s", err)
	}

	if err := d.Set("is_ext_shared", channel.IsExtShared); err != nil {
		return diag.Errorf("error setting is_ext_shared: %s", err)
	}

	if err := d.Set("is_org_shared", channel.IsOrgShared); err != nil {
		return diag.Errorf("error setting is_org_shared: %s", err)
	}

	if err := d.Set("created", channel.Created); err != nil {
		return diag.Errorf("error setting created: %s", err)
	}

	if err := d.Set("creator", channel.Creator); err != nil {
		return diag.Errorf("error setting creator: %s", err)
	}

	if err := d.Set("is_private", channel.IsPrivate); err != nil {
		return diag.Errorf("error setting is_private: %s", err)
	}

	if err := d.Set("is_general", channel.IsGeneral); err != nil {
		return diag.Errorf("error setting is_general: %s", err)
	}

	return nil
}

func archiveConversationWithContext(ctx context.Context, client ClientInterface, id string) error {
	if err := client.ArchiveConversationContext(ctx, id); err != nil {
		if err.Error() != "already_archived" {
			return fmt.Errorf("couldn't archive conversation %s: %s", id, err)
		}
	}
	return nil
}

func contains(s []string, e string) bool {
	var found bool
	for _, x := range s {
		if x == e {
			return true
		}
	}
	return found
}
