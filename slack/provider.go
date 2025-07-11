package slack

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/slack-go/slack"
)

// Provider returns a *schema.Provider
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"token": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("SLACK_TOKEN", nil),
				Description: "The Slack token",
			},
			"retry_timeout": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     DefaultRetryTimeoutSeconds,
				Description: "The timeout in seconds for retry operations when rate limited by Slack. Defaults to 60 seconds.",
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"slack_conversation": resourceSlackConversation(),
			"slack_usergroup":    resourceSlackUserGroup(),
		},

		DataSourcesMap: map[string]*schema.Resource{
			"slack_conversation": dataSourceConversation(),
			"slack_user":         dataSourceUser(),
			"slack_usergroup":    dataSourceUserGroup(),
		},

		ConfigureContextFunc: providerConfigure,
	}
}

// ProviderConfig holds the provider configuration
type ProviderConfig struct {
	Client      ClientInterface
	RetryConfig *RetryConfig
}

func providerConfigure(_ context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics

	token, ok := d.GetOk("token")
	if !ok {
		return nil, diag.Errorf("could not create slack client. Please provide a token.")
	}

	retryTimeout := d.Get("retry_timeout").(int)
	retryConfig := &RetryConfig{
		Timeout: time.Duration(retryTimeout) * time.Second,
	}

	slackClient := slack.New(token.(string))
	wrappedClient := NewClientWrapper(slackClient)

	config := &ProviderConfig{
		Client:      wrappedClient,
		RetryConfig: retryConfig,
	}

	return config, diags
}

func schemaSetToSlice(set *schema.Set) []string {
	if set == nil {
		return []string{}
	}
	s := make([]string, len(set.List()))
	for i, v := range set.List() {
		s[i] = v.(string)
	}
	return s
}

func remove(s []string, r string) []string {
	result := make([]string, 0, len(s))
	for _, v := range s {
		if v != r {
			result = append(result, v)
		}
	}
	return result
}
