package slack

import (
	"context"

	"github.com/slack-go/slack"
)

// ClientWrapper wraps the real slack.Client to implement ClientInterface
type ClientWrapper struct {
	client *slack.Client
}

// NewClientWrapper creates a new wrapper around a slack.Client
func NewClientWrapper(client *slack.Client) ClientInterface {
	return &ClientWrapper{client: client}
}

// User operations
func (w *ClientWrapper) GetUserByEmailContext(ctx context.Context, email string) (*slack.User, error) {
	return w.client.GetUserByEmailContext(ctx, email)
}

func (w *ClientWrapper) GetUsersContext(ctx context.Context) ([]slack.User, error) {
	return w.client.GetUsersContext(ctx)
}

// Conversation operations
func (w *ClientWrapper) CreateConversationContext(ctx context.Context, params slack.CreateConversationParams) (*slack.Channel, error) {
	return w.client.CreateConversationContext(ctx, params)
}

func (w *ClientWrapper) GetConversationInfoContext(ctx context.Context, input *slack.GetConversationInfoInput) (*slack.Channel, error) {
	return w.client.GetConversationInfoContext(ctx, input)
}

func (w *ClientWrapper) GetConversationsContext(ctx context.Context, params *slack.GetConversationsParameters) ([]slack.Channel, string, error) {
	return w.client.GetConversationsContext(ctx, params)
}

func (w *ClientWrapper) GetUsersInConversationContext(ctx context.Context, params *slack.GetUsersInConversationParameters) ([]string, string, error) {
	return w.client.GetUsersInConversationContext(ctx, params)
}

func (w *ClientWrapper) JoinConversationContext(ctx context.Context, channelID string) (*slack.Channel, string, []string, error) {
	return w.client.JoinConversationContext(ctx, channelID)
}

func (w *ClientWrapper) InviteUsersToConversationContext(ctx context.Context, channelID string, users ...string) (*slack.Channel, error) {
	return w.client.InviteUsersToConversationContext(ctx, channelID, users...)
}

func (w *ClientWrapper) KickUserFromConversationContext(ctx context.Context, channelID, user string) error {
	return w.client.KickUserFromConversationContext(ctx, channelID, user)
}

func (w *ClientWrapper) SetTopicOfConversationContext(ctx context.Context, channelID, topic string) (*slack.Channel, error) {
	return w.client.SetTopicOfConversationContext(ctx, channelID, topic)
}

func (w *ClientWrapper) SetPurposeOfConversationContext(ctx context.Context, channelID, purpose string) (*slack.Channel, error) {
	return w.client.SetPurposeOfConversationContext(ctx, channelID, purpose)
}

func (w *ClientWrapper) RenameConversationContext(ctx context.Context, channelID, name string) (*slack.Channel, error) {
	return w.client.RenameConversationContext(ctx, channelID, name)
}

func (w *ClientWrapper) ArchiveConversationContext(ctx context.Context, channelID string) error {
	return w.client.ArchiveConversationContext(ctx, channelID)
}

func (w *ClientWrapper) UnArchiveConversationContext(ctx context.Context, channelID string) error {
	return w.client.UnArchiveConversationContext(ctx, channelID)
}

// User group operations
func (w *ClientWrapper) CreateUserGroupContext(ctx context.Context, userGroup slack.UserGroup, options ...slack.CreateUserGroupOption) (slack.UserGroup, error) {
	return w.client.CreateUserGroupContext(ctx, userGroup, options...)
}

func (w *ClientWrapper) GetUserGroupsContext(ctx context.Context, options ...slack.GetUserGroupsOption) ([]slack.UserGroup, error) {
	return w.client.GetUserGroupsContext(ctx, options...)
}

func (w *ClientWrapper) UpdateUserGroupContext(ctx context.Context, userGroupID string, options ...slack.UpdateUserGroupsOption) (slack.UserGroup, error) {
	return w.client.UpdateUserGroupContext(ctx, userGroupID, options...)
}

func (w *ClientWrapper) UpdateUserGroupMembersContext(ctx context.Context, userGroupID, users string) (slack.UserGroup, error) {
	return w.client.UpdateUserGroupMembersContext(ctx, userGroupID, users)
}

func (w *ClientWrapper) DisableUserGroupContext(ctx context.Context, userGroup string, options ...slack.DisableUserGroupOption) (slack.UserGroup, error) {
	return w.client.DisableUserGroupContext(ctx, userGroup, options...)
}

func (w *ClientWrapper) EnableUserGroupContext(ctx context.Context, userGroup string, options ...slack.EnableUserGroupOption) (slack.UserGroup, error) {
	return w.client.EnableUserGroupContext(ctx, userGroup, options...)
}

// Auth operations
func (w *ClientWrapper) AuthTest() (*slack.AuthTestResponse, error) {
	return w.client.AuthTest()
}
