package provider

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/paultyng/go-twitter/twitter"
)

func allBlocks(c *twitter.Client) ([]twitter.User, error) {
	blockUsers := []twitter.User{}
	err := allPages(func(cursor int64) (int64, error) {
		blocks, _, err := c.Blocks.List(&twitter.BlockListParams{
			Cursor:              cursor,
			IncludeUserEntities: twitter.Bool(false),
			SkipStatus:          twitter.Bool(true),
		})
		if err != nil {
			return 0, err
		}
		blockUsers = append(blockUsers, blocks.Users...)
		return blocks.NextCursor, nil
	})
	if err != nil {
		return nil, err
	}
	return blockUsers, nil
}

type fetchBlockRequest struct {
	IDStr      string
	ScreenName string
}

func fetchBlockBatch(api *twitter.Client) func(reqs []interface{}) ([]interface{}, error) {
	return func(reqs []interface{}) ([]interface{}, error) {
		blocks, err := allBlocks(api)
		if err != nil {
			return nil, err
		}

		found := make([]interface{}, len(reqs))
		for i, req := range reqs {
			fbr := req.(fetchBlockRequest)
			for _, block := range blocks {
				switch {
				case fbr.IDStr != "":
					if fbr.IDStr == block.IDStr {
						found[i] = &block
						goto nextReq
					}
				case fbr.ScreenName != "":
					if fbr.ScreenName == block.ScreenName {
						found[i] = &block
						goto nextReq
					}
				}
			}
		nextReq:
		}
		return found, nil
	}
}

func resourceBlock() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceBlockCreate,
		ReadContext:   resourceBlockRead,
		DeleteContext: resourceBlockDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceBlockImport,
		},

		Schema: map[string]*schema.Schema{
			"screen_name": {
				Description:      `The screen name of the potentially blocked user.`,
				Type:             schema.TypeString,
				ForceNew:         true,
				Required:         true,
				DiffSuppressFunc: caseInsensitiveStringSuppressDiff,
			},
		},
	}
}

func resourceBlockImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	c := meta.(*client)
	// we import by screenname, not ID
	name := d.Id()
	resp, err := c.blockBatcher.Get(ctx, fetchBlockRequest{ScreenName: name})
	if err != nil {
		return nil, err
	}
	if resp == nil {
		// TODO: how to handle not found? err?
		return nil, nil
	}
	block := resp.(*twitter.User)
	d.SetId(block.IDStr)
	d.Set("screen_name", block.ScreenName)
	return []*schema.ResourceData{d}, nil
}

func resourceBlockCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*client)
	name := d.Get("screen_name").(string)

	block, _, err := c.Blocks.Create(&twitter.BlockCreateParams{
		ScreenName:      name,
		IncludeEntities: twitter.Bool(false),
		SkipStatus:      twitter.Bool(true),
	})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(block.IDStr)
	return nil
}

func resourceBlockRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*client)
	idstr := d.Id()
	resp, err := c.blockBatcher.Get(ctx, fetchBlockRequest{IDStr: idstr})
	if err != nil {
		return diag.FromErr(err)
	}
	if resp == nil {
		d.SetId("")
		return nil
	}
	block := resp.(*twitter.User)
	d.SetId(block.IDStr)
	d.Set("screen_name", block.ScreenName)
	return nil
}

func resourceBlockDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*client)
	idstr := d.Id()

	id, err := strconv.ParseInt(idstr, 10, 64)
	if err != nil {
		return diag.FromErr(err)
	}

	_, _, err = c.Blocks.Destroy(&twitter.BlockDestroyParams{
		UserID:          id,
		IncludeEntities: twitter.Bool(false),
		SkipStatus:      twitter.Bool(true),
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
