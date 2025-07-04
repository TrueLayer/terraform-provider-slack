package slack

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/slack-go/slack"
)

func dataSourceUserGroup() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceUserGroupRead,

		Schema: map[string]*schema.Schema{
			"usergroup_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"name", "usergroup_id"},
			},
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"name", "usergroup_id"},
			},
			"channels": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Set:      schema.HashString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"handle": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"users": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Set:      schema.HashString,
				Computed: true,
			},
		},
	}
}

func dataSourceUserGroupRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var (
		group *slack.UserGroup
		err   error
	)

	err = retry.RetryContext(ctx, slackRetryTimeout, func() *retry.RetryError {
		var rlerr *slack.RateLimitedError
		if name, ok := d.GetOk("name"); ok {
			u, ferr := findUserGroupByName(ctx, name.(string), false, m)
			if errors.As(ferr, &rlerr) {
				time.Sleep(rlerr.RetryAfter)
				return retry.RetryableError(ferr)
			}
			if ferr != nil {
				return retry.NonRetryableError(ferr)
			}
			group = &u
		} else if id, ok := d.GetOk("usergroup_id"); ok {
			u, ferr := findUserGroupByID(ctx, id.(string), false, m)
			if errors.As(ferr, &rlerr) {
				time.Sleep(rlerr.RetryAfter)
				return retry.RetryableError(ferr)
			}
			if ferr != nil {
				return retry.NonRetryableError(ferr)
			}
			group = &u
		} else {
			return retry.NonRetryableError(fmt.Errorf("your query returned no results. Please change your search criteria and try again"))
		}
		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(group.ID)
	if err := d.Set("usergroup_id", group.ID); err != nil {
		return diag.FromErr(fmt.Errorf("error setting usergroup ID: %s", err))
	}
	return updateUserGroupData(d, *group)
}
