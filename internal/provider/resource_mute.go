package provider

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/paultyng/go-twitter/twitter"
)

func allMutes(c *twitter.Client) ([]twitter.User, error) {
	muteUsers := []twitter.User{}
	err := allPages(func(cursor int64) (int64, error) {
		mutes, _, err := c.Mutes.List(&twitter.MuteListParams{
			Cursor:              cursor,
			IncludeUserEntities: twitter.Bool(false),
			SkipStatus:          twitter.Bool(true),
		})
		if err != nil {
			return 0, err
		}
		muteUsers = append(muteUsers, mutes.Users...)
		return mutes.NextCursor, nil
	})
	if err != nil {
		return nil, err
	}
	return muteUsers, nil
}

type fetchMuteRequest struct {
	IDStr      string
	ScreenName string
}

func fetchMuteBatch(api *twitter.Client) func(reqs []interface{}) ([]interface{}, error) {
	return func(reqs []interface{}) ([]interface{}, error) {
		mutes, err := allMutes(api)
		if err != nil {
			return nil, err
		}

		found := make([]interface{}, len(reqs))
		for i, req := range reqs {
			fbr := req.(fetchMuteRequest)
			for _, mute := range mutes {
				switch {
				case fbr.IDStr != "":
					if fbr.IDStr == mute.IDStr {
						found[i] = &mute
						goto nextReq
					}
				case fbr.ScreenName != "":
					if fbr.ScreenName == mute.ScreenName {
						found[i] = &mute
						goto nextReq
					}
				}
			}
		nextReq:
		}
		return found, nil
	}
}

func resourceMute() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceMuteCreate,
		ReadContext:   resourceMuteRead,
		DeleteContext: resourceMuteDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceMuteImport,
		},

		Schema: map[string]*schema.Schema{
			"screen_name": {
				Description:      `The screen name of the potentially muteed user.`,
				Type:             schema.TypeString,
				ForceNew:         true,
				Required:         true,
				DiffSuppressFunc: caseInsensitiveStringSuppressDiff,
			},
		},
	}
}

func resourceMuteImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	c := meta.(*client)
	// we import by screenname, not ID
	name := d.Id()
	resp, err := c.muteBatcher.Get(ctx, fetchMuteRequest{ScreenName: name})
	if err != nil {
		return nil, err
	}
	if resp == nil {
		// TODO: how to handle not found? err?
		return nil, nil
	}
	mute := resp.(*twitter.User)
	d.SetId(mute.IDStr)
	d.Set("screen_name", mute.ScreenName)
	return []*schema.ResourceData{d}, nil
}

func resourceMuteCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*client)
	name := d.Get("screen_name").(string)

	mute, _, err := c.Mutes.Create(&twitter.MuteCreateParams{
		ScreenName: name,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(mute.IDStr)
	return nil
}

func resourceMuteRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*client)
	idstr := d.Id()
	resp, err := c.muteBatcher.Get(ctx, fetchMuteRequest{IDStr: idstr})
	if err != nil {
		return diag.FromErr(err)
	}
	if resp == nil {
		d.SetId("")
		return nil
	}
	mute := resp.(*twitter.User)
	d.SetId(mute.IDStr)
	d.Set("screen_name", mute.ScreenName)
	return nil
}

func resourceMuteDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*client)
	idstr := d.Id()

	id, err := strconv.ParseInt(idstr, 10, 64)
	if err != nil {
		return diag.FromErr(err)
	}

	_, _, err = c.Mutes.Destroy(&twitter.MuteDestroyParams{
		UserID: id,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
