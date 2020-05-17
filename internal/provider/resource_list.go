package provider

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/paultyng/go-twitter/twitter"
)

var (
	validateName     schema.SchemaValidateFunc
	validateStringID schema.SchemaValidateFunc
)

func init() {
	nameRegexp := regexp.MustCompile("[a-zA-Z][a-zA-Z0-9-_]{0,25}")
	validateName = validation.StringMatch(nameRegexp, `must start with a letter and can consist only of 25 or fewer letters, numbers, "-", or "_" characters`)

	// using IDs as strings due to built-in ID attribute already being a string...
	idRegexp := regexp.MustCompile("\\d+")
	validateStringID = validation.StringMatch(idRegexp, "must be an integeger")
}

func resourceList() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceListCreate,
		ReadContext:   resourceListRead,
		UpdateContext: resourceListUpdate,
		DeleteContext: resourceListDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Description:  `The name for the list. A list's name must start with a letter and can consist only of 25 or fewer letters, numbers, "-", or "_" characters.`,
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateName,
			},
			"mode": {
				Description:  "Whether your list is public or private. Values can be `public` or `private`.",
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "public",
				ValidateFunc: validation.StringInSlice([]string{"public", "private"}, false),
			},
			"description": {
				Description: "The description to give the list.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"members": {
				Description: "The screen names of the user for whom to return results.",
				Type:        schema.TypeSet,
				Optional:    true,
				// Temporarily restricted until paging is implemented
				MaxItems: 100,
				Elem:     &schema.Schema{Type: schema.TypeString},
				// TODO: diff suppress for case insensitivity
			},

			"slug": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"uri": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceListSyncMembers(c *client, listID int64, from []string, to []string) error {
	fromMap := map[string]bool{}
	for _, m := range from {
		fromMap[m] = true
	}

	create := []string{}
	toMap := map[string]bool{}
	for _, m := range to {
		toMap[m] = true
		if !fromMap[m] {
			create = append(create, m)
		}
	}

	destroy := []string{}
	for _, m := range from {
		if !toMap[m] {
			destroy = append(destroy, m)
		}
	}

	log.Printf("[TRACE] syncing members, creating %d, destroying %d", len(create), len(destroy))

	if len(create) > 100 || len(destroy) > 100 {
		return fmt.Errorf("can only process 100 create (%d) or destroy (%d) member operations in a single apply", len(create), len(destroy))
	}

	if len(destroy) > 0 {
		_, err := c.Lists.MembersDestroyAll(&twitter.ListsMembersDestroyAllParams{
			ListID:     listID,
			ScreenName: strings.Join(destroy, ","),
		})
		if err != nil {
			return err
		}
	}

	if len(create) > 0 {
		_, err := c.Lists.MembersCreateAll(&twitter.ListsMembersCreateAllParams{
			ListID:     listID,
			ScreenName: strings.Join(create, ","),
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func resourceListCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*client)

	name := d.Get("name").(string)
	mode := d.Get("mode").(string)
	desc := d.Get("description").(string)
	membersSet := d.Get("members").(*schema.Set)
	members, err := setToStringSlice(membersSet)
	if err != nil {
		diags := diag.FromErr(err)
		diags[0].AttributePath = cty.GetAttrPath("members")
		return diags
	}

	list, _, err := c.Lists.Create(name, &twitter.ListsCreateParams{
		Mode:        mode,
		Description: desc,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	err = resourceListSyncMembers(c, list.ID, nil, members)
	if err != nil {
		diags := diag.FromErr(err)
		diags[0].AttributePath = cty.GetAttrPath("members")
		return diags
	}

	d.SetId(list.IDStr)
	return resourceListSetData(d, list, members)
}

func resourceListReadMembers(ctx context.Context, c *client, id int64) ([]string, error) {
	// sigh, pointer booleans
	f := false
	t := true

	// TODO: paging
	memberUsers, _, err := c.Lists.Members(&twitter.ListsMembersParams{
		ListID:          id,
		IncludeEntities: &f,
		SkipStatus:      &t,
		Count:           100,
	})
	if err != nil {
		return nil, err
	}

	members := []string{}
	for _, mu := range memberUsers.Users {
		members = append(members, mu.ScreenName)
	}
	log.Printf("[TRACE] members read: %v", members)
	return members, nil
}

func resourceListRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*client)

	sid := d.Id()
	id, err := strconv.ParseInt(sid, 10, 64)
	if err != nil {
		return diag.FromErr(err)
	}

	list, _, err := c.Lists.Show(&twitter.ListsShowParams{
		ListID: id,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	members, err := resourceListReadMembers(ctx, c, id)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceListSetData(d, list, members)
}

func resourceListSetData(d *schema.ResourceData, list *twitter.List, members []string) diag.Diagnostics {
	d.Set("name", list.Name)
	d.Set("mode", list.Mode)
	d.Set("description", list.Description)
	d.Set("slug", list.Slug)
	d.Set("uri", list.URI)
	d.Set("members", stringSliceToSet(members))

	return nil
}

func resourceListUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*client)

	sid := d.Id()
	id, err := strconv.ParseInt(sid, 10, 64)
	if err != nil {
		return diag.FromErr(err)
	}

	name := d.Get("name").(string)
	mode := d.Get("mode").(string)
	desc := d.Get("description").(string)
	membersSet := d.Get("members").(*schema.Set)
	toMembers, err := setToStringSlice(membersSet)
	if err != nil {
		diags := diag.FromErr(err)
		diags[0].AttributePath = cty.GetAttrPath("members")
		return diags
	}

	_, err = c.Lists.Update(&twitter.ListsUpdateParams{
		ListID:      id,
		Name:        name,
		Mode:        mode,
		Description: desc,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	list, _, err := c.Lists.Show(&twitter.ListsShowParams{
		ListID: id,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	fromMembers, err := resourceListReadMembers(ctx, c, id)
	if err != nil {
		return diag.FromErr(err)
	}

	err = resourceListSyncMembers(c, list.ID, fromMembers, toMembers)
	if err != nil {
		diags := diag.FromErr(err)
		diags[0].AttributePath = cty.GetAttrPath("members")
		return diags
	}

	return resourceListSetData(d, list, toMembers)
}

func resourceListDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*client)

	sid := d.Id()
	id, err := strconv.ParseInt(sid, 10, 64)
	if err != nil {
		return diag.FromErr(err)
	}

	_, _, err = c.Lists.Destroy(&twitter.ListsDestroyParams{
		ListID: id,
	})
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}
