package slack

import (
	"context"

	"github.com/slack-go/slack"
)

// ClientInterface defines the interface for Slack client operations
// This allows for easier mocking in unit tests
//
//nolint:dupl // Interface and implementation have similar signatures by design
type ClientInterface interface {
	// User operations
	GetUserByEmailContext(ctx context.Context, email string) (*slack.User, error)
	GetUsersContext(ctx context.Context) ([]slack.User, error)

	// Conversation operations
	CreateConversationContext(ctx context.Context, params slack.CreateConversationParams) (*slack.Channel, error)
	GetConversationInfoContext(ctx context.Context, input *slack.GetConversationInfoInput) (*slack.Channel, error)
	GetConversationsContext(ctx context.Context, params *slack.GetConversationsParameters) ([]slack.Channel, string, error)
	GetUsersInConversationContext(ctx context.Context, params *slack.GetUsersInConversationParameters) ([]string, string, error)
	JoinConversationContext(ctx context.Context, channelID string) (*slack.Channel, string, []string, error)
	InviteUsersToConversationContext(ctx context.Context, channelID string, users ...string) (*slack.Channel, error)
	KickUserFromConversationContext(ctx context.Context, channelID, user string) error
	SetTopicOfConversationContext(ctx context.Context, channelID, topic string) (*slack.Channel, error)
	SetPurposeOfConversationContext(ctx context.Context, channelID, purpose string) (*slack.Channel, error)
	RenameConversationContext(ctx context.Context, channelID, name string) (*slack.Channel, error)
	ArchiveConversationContext(ctx context.Context, channelID string) error
	UnArchiveConversationContext(ctx context.Context, channelID string) error

	// User group operations
	CreateUserGroupContext(ctx context.Context, userGroup slack.UserGroup, options ...slack.CreateUserGroupOption) (slack.UserGroup, error)
	GetUserGroupsContext(ctx context.Context, options ...slack.GetUserGroupsOption) ([]slack.UserGroup, error)
	UpdateUserGroupContext(ctx context.Context, userGroupID string, options ...slack.UpdateUserGroupsOption) (slack.UserGroup, error)
	UpdateUserGroupMembersContext(ctx context.Context, userGroupID, users string) (slack.UserGroup, error)
	DisableUserGroupContext(ctx context.Context, userGroup string, options ...slack.DisableUserGroupOption) (slack.UserGroup, error)
	EnableUserGroupContext(ctx context.Context, userGroup string, options ...slack.EnableUserGroupOption) (slack.UserGroup, error)

	// Auth operations
	AuthTest() (*slack.AuthTestResponse, error)
}

// MockSlackClient is a mock implementation of ClientInterface for testing
//
//nolint:dupl // Interface and implementation have similar signatures by design
type MockSlackClient struct {
	// User mocks
	MockGetUserByEmail func(ctx context.Context, email string) (*slack.User, error)
	MockGetUsers       func(ctx context.Context) ([]slack.User, error)

	// Conversation mocks
	MockCreateConversation        func(ctx context.Context, params slack.CreateConversationParams) (*slack.Channel, error)
	MockGetConversationInfo       func(ctx context.Context, input *slack.GetConversationInfoInput) (*slack.Channel, error)
	MockGetConversations          func(ctx context.Context, params *slack.GetConversationsParameters) ([]slack.Channel, string, error)
	MockGetUsersInConversation    func(ctx context.Context, params *slack.GetUsersInConversationParameters) ([]string, string, error)
	MockJoinConversation          func(ctx context.Context, channelID string) (*slack.Channel, string, []string, error)
	MockInviteUsersToConversation func(ctx context.Context, channelID string, users ...string) (*slack.Channel, error)
	MockKickUserFromConversation  func(ctx context.Context, channelID, user string) error
	MockSetTopicOfConversation    func(ctx context.Context, channelID, topic string) (*slack.Channel, error)
	MockSetPurposeOfConversation  func(ctx context.Context, channelID, purpose string) (*slack.Channel, error)
	MockRenameConversation        func(ctx context.Context, channelID, name string) (*slack.Channel, error)
	MockArchiveConversation       func(ctx context.Context, channelID string) error
	MockUnArchiveConversation     func(ctx context.Context, channelID string) error

	// User group mocks
	MockCreateUserGroup        func(ctx context.Context, userGroup slack.UserGroup, options ...slack.CreateUserGroupOption) (slack.UserGroup, error)
	MockGetUserGroups          func(ctx context.Context, options ...slack.GetUserGroupsOption) ([]slack.UserGroup, error)
	MockUpdateUserGroup        func(ctx context.Context, userGroupID string, options ...slack.UpdateUserGroupsOption) (slack.UserGroup, error)
	MockUpdateUserGroupMembers func(ctx context.Context, userGroupID, users string) (slack.UserGroup, error)
	MockDisableUserGroup       func(ctx context.Context, userGroup string, options ...slack.DisableUserGroupOption) (slack.UserGroup, error)
	MockEnableUserGroup        func(ctx context.Context, userGroup string, options ...slack.EnableUserGroupOption) (slack.UserGroup, error)

	// Auth mocks
	MockAuthTest func() (*slack.AuthTestResponse, error)
}

// Ensure MockSlackClient implements ClientInterface
var _ ClientInterface = (*MockSlackClient)(nil)

// User operations
func (m *MockSlackClient) GetUserByEmailContext(ctx context.Context, email string) (*slack.User, error) {
	if m.MockGetUserByEmail != nil {
		return m.MockGetUserByEmail(ctx, email)
	}
	return nil, nil
}

func (m *MockSlackClient) GetUsersContext(ctx context.Context) ([]slack.User, error) {
	if m.MockGetUsers != nil {
		return m.MockGetUsers(ctx)
	}
	return nil, nil
}

// Conversation operations
func (m *MockSlackClient) CreateConversationContext(ctx context.Context, params slack.CreateConversationParams) (*slack.Channel, error) {
	if m.MockCreateConversation != nil {
		return m.MockCreateConversation(ctx, params)
	}
	return nil, nil
}

func (m *MockSlackClient) GetConversationInfoContext(ctx context.Context, input *slack.GetConversationInfoInput) (*slack.Channel, error) {
	if m.MockGetConversationInfo != nil {
		return m.MockGetConversationInfo(ctx, input)
	}
	return nil, nil
}

func (m *MockSlackClient) GetConversationsContext(ctx context.Context, params *slack.GetConversationsParameters) ([]slack.Channel, string, error) {
	if m.MockGetConversations != nil {
		return m.MockGetConversations(ctx, params)
	}
	return nil, "", nil
}

func (m *MockSlackClient) GetUsersInConversationContext(ctx context.Context, params *slack.GetUsersInConversationParameters) ([]string, string, error) {
	if m.MockGetUsersInConversation != nil {
		return m.MockGetUsersInConversation(ctx, params)
	}
	return nil, "", nil
}

func (m *MockSlackClient) JoinConversationContext(ctx context.Context, channelID string) (*slack.Channel, string, []string, error) {
	if m.MockJoinConversation != nil {
		return m.MockJoinConversation(ctx, channelID)
	}
	return nil, "", nil, nil
}

func (m *MockSlackClient) InviteUsersToConversationContext(ctx context.Context, channelID string, users ...string) (*slack.Channel, error) {
	if m.MockInviteUsersToConversation != nil {
		return m.MockInviteUsersToConversation(ctx, channelID, users...)
	}
	return nil, nil
}

func (m *MockSlackClient) KickUserFromConversationContext(ctx context.Context, channelID, user string) error {
	if m.MockKickUserFromConversation != nil {
		return m.MockKickUserFromConversation(ctx, channelID, user)
	}
	return nil
}

func (m *MockSlackClient) SetTopicOfConversationContext(ctx context.Context, channelID, topic string) (*slack.Channel, error) {
	if m.MockSetTopicOfConversation != nil {
		return m.MockSetTopicOfConversation(ctx, channelID, topic)
	}
	return nil, nil
}

func (m *MockSlackClient) SetPurposeOfConversationContext(ctx context.Context, channelID, purpose string) (*slack.Channel, error) {
	if m.MockSetPurposeOfConversation != nil {
		return m.MockSetPurposeOfConversation(ctx, channelID, purpose)
	}
	return nil, nil
}

func (m *MockSlackClient) RenameConversationContext(ctx context.Context, channelID, name string) (*slack.Channel, error) {
	if m.MockRenameConversation != nil {
		return m.MockRenameConversation(ctx, channelID, name)
	}
	return nil, nil
}

func (m *MockSlackClient) ArchiveConversationContext(ctx context.Context, channelID string) error {
	if m.MockArchiveConversation != nil {
		return m.MockArchiveConversation(ctx, channelID)
	}
	return nil
}

func (m *MockSlackClient) UnArchiveConversationContext(ctx context.Context, channelID string) error {
	if m.MockUnArchiveConversation != nil {
		return m.MockUnArchiveConversation(ctx, channelID)
	}
	return nil
}

// User group operations
func (m *MockSlackClient) CreateUserGroupContext(ctx context.Context, userGroup slack.UserGroup, options ...slack.CreateUserGroupOption) (slack.UserGroup, error) {
	if m.MockCreateUserGroup != nil {
		return m.MockCreateUserGroup(ctx, userGroup, options...)
	}
	return slack.UserGroup{}, nil
}

func (m *MockSlackClient) GetUserGroupsContext(ctx context.Context, options ...slack.GetUserGroupsOption) ([]slack.UserGroup, error) {
	if m.MockGetUserGroups != nil {
		return m.MockGetUserGroups(ctx, options...)
	}
	return nil, nil
}

func (m *MockSlackClient) UpdateUserGroupContext(ctx context.Context, userGroupID string, options ...slack.UpdateUserGroupsOption) (slack.UserGroup, error) {
	if m.MockUpdateUserGroup != nil {
		return m.MockUpdateUserGroup(ctx, userGroupID, options...)
	}
	return slack.UserGroup{}, nil
}

func (m *MockSlackClient) UpdateUserGroupMembersContext(ctx context.Context, userGroupID, users string) (slack.UserGroup, error) {
	if m.MockUpdateUserGroupMembers != nil {
		return m.MockUpdateUserGroupMembers(ctx, userGroupID, users)
	}
	return slack.UserGroup{}, nil
}

func (m *MockSlackClient) DisableUserGroupContext(ctx context.Context, userGroup string, options ...slack.DisableUserGroupOption) (slack.UserGroup, error) {
	if m.MockDisableUserGroup != nil {
		return m.MockDisableUserGroup(ctx, userGroup, options...)
	}
	return slack.UserGroup{}, nil
}

func (m *MockSlackClient) EnableUserGroupContext(ctx context.Context, userGroup string, options ...slack.EnableUserGroupOption) (slack.UserGroup, error) {
	if m.MockEnableUserGroup != nil {
		return m.MockEnableUserGroup(ctx, userGroup, options...)
	}
	return slack.UserGroup{}, nil
}

// Auth operations
func (m *MockSlackClient) AuthTest() (*slack.AuthTestResponse, error) {
	if m.MockAuthTest != nil {
		return m.MockAuthTest()
	}
	return nil, nil
}
