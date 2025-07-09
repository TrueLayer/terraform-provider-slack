package slack

import (
	"context"
	"errors"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
)

func TestSearchByName(t *testing.T) {
	tests := []struct {
		name          string
		searchName    string
		mockUsers     []slack.User
		mockError     error
		expectedUser  *slack.User
		expectedError string
	}{
		{
			name:       "user found",
			searchName: "testuser",
			mockUsers: []slack.User{
				{ID: "U123", Name: "testuser", Profile: slack.UserProfile{Email: "test@example.com"}},
				{ID: "U456", Name: "otheruser", Profile: slack.UserProfile{Email: "other@example.com"}},
			},
			expectedUser: &slack.User{ID: "U123", Name: "testuser", Profile: slack.UserProfile{Email: "test@example.com"}},
		},
		{
			name:       "user not found",
			searchName: "nonexistent",
			mockUsers: []slack.User{
				{ID: "U123", Name: "testuser", Profile: slack.UserProfile{Email: "test@example.com"}},
			},
			expectedError: "your query returned no results",
		},
		{
			name:       "multiple users found",
			searchName: "testuser",
			mockUsers: []slack.User{
				{ID: "U123", Name: "testuser", Profile: slack.UserProfile{Email: "test1@example.com"}},
				{ID: "U456", Name: "testuser", Profile: slack.UserProfile{Email: "test2@example.com"}},
			},
			expectedError: "your query returned more than one result",
		},
		{
			name:          "API error",
			searchName:    "testuser",
			mockError:     errors.New("API error"),
			expectedError: "couldn't get workspace users",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockSlackClient{
				MockGetUsers: func(_ context.Context) ([]slack.User, error) {
					return tt.mockUsers, tt.mockError
				},
			}

			ctx := context.Background()
			result, err := searchByName(ctx, tt.searchName, mockClient)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedUser, result)
			}
		})
	}
}

func TestDataSourceUserRead_ByName(t *testing.T) {
	tests := []struct {
		name          string
		userName      string
		mockUser      *slack.User
		mockError     error
		expectedDiags diag.Diagnostics
		expectedID    string
	}{
		{
			name:     "user found by name",
			userName: "testuser",
			mockUser: &slack.User{
				ID:   "U123",
				Name: "testuser",
				Profile: slack.UserProfile{
					Email: "test@example.com",
				},
			},
			expectedID: "U123",
		},
		{
			name:      "user not found by name",
			userName:  "nonexistent",
			mockError: errors.New("user not found"),
			expectedDiags: diag.Diagnostics{
				diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "not found nonexistent: couldn't get workspace users: user not found",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockSlackClient{
				MockGetUsers: func(_ context.Context) ([]slack.User, error) {
					if tt.mockUser != nil {
						return []slack.User{*tt.mockUser}, tt.mockError
					}
					return nil, tt.mockError
				},
			}

			config := &ProviderConfig{
				Client:      mockClient,
				RetryConfig: DefaultRetryConfig(),
			}

			resourceData := schema.TestResourceDataRaw(t, dataSourceUser().Schema, map[string]interface{}{
				"name": tt.userName,
			})

			ctx := context.Background()
			diags := dataSourceUserRead(ctx, resourceData, config)

			if len(tt.expectedDiags) > 0 {
				assert.Equal(t, tt.expectedDiags, diags)
			} else {
				assert.Empty(t, diags)
				assert.Equal(t, tt.expectedID, resourceData.Id())
				assert.Equal(t, tt.userName, resourceData.Get("name"))
				assert.Equal(t, tt.mockUser.Profile.Email, resourceData.Get("email"))
			}
		})
	}
}

func TestDataSourceUserRead_ByEmail(t *testing.T) {
	tests := []struct {
		name          string
		userEmail     string
		mockUser      *slack.User
		mockError     error
		expectedDiags diag.Diagnostics
		expectedID    string
	}{
		{
			name:      "user found by email",
			userEmail: "test@example.com",
			mockUser: &slack.User{
				ID:   "U123",
				Name: "testuser",
				Profile: slack.UserProfile{
					Email: "test@example.com",
				},
			},
			expectedID: "U123",
		},
		{
			name:      "user not found by email",
			userEmail: "nonexistent@example.com",
			mockError: errors.New("users_not_found"),
			expectedDiags: diag.Diagnostics{
				diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "not found nonexistent@example.com: users_not_found",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockSlackClient{
				MockGetUserByEmail: func(_ context.Context, _ string) (*slack.User, error) {
					return tt.mockUser, tt.mockError
				},
			}

			config := &ProviderConfig{
				Client:      mockClient,
				RetryConfig: DefaultRetryConfig(),
			}

			resourceData := schema.TestResourceDataRaw(t, dataSourceUser().Schema, map[string]interface{}{
				"email": tt.userEmail,
			})

			ctx := context.Background()
			diags := dataSourceUserRead(ctx, resourceData, config)

			if len(tt.expectedDiags) > 0 {
				assert.Equal(t, tt.expectedDiags, diags)
			} else {
				assert.Empty(t, diags)
				assert.Equal(t, tt.expectedID, resourceData.Id())
				assert.Equal(t, tt.mockUser.Name, resourceData.Get("name"))
				assert.Equal(t, tt.userEmail, resourceData.Get("email"))
			}
		})
	}
}

func TestDataSourceUserRead_ValidationErrors(t *testing.T) {
	tests := []struct {
		name          string
		data          map[string]interface{}
		expectedError string
	}{
		{
			name:          "no fields set",
			data:          map[string]interface{}{},
			expectedError: "your query returned no results",
		},
		{
			name: "both name and email set",
			data: map[string]interface{}{
				"name":  "testuser",
				"email": "test@example.com",
			},
			expectedError: "your query returned no results",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockSlackClient{}
			config := &ProviderConfig{
				Client:      mockClient,
				RetryConfig: DefaultRetryConfig(),
			}

			resourceData := schema.TestResourceDataRaw(t, dataSourceUser().Schema, tt.data)

			ctx := context.Background()
			diags := dataSourceUserRead(ctx, resourceData, config)

			assert.NotEmpty(t, diags)
			assert.Contains(t, diags[0].Summary, tt.expectedError)
		})
	}
}
