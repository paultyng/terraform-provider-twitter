package provider

import (
	"context"
	"net/http"

	"github.com/dghubble/oauth1"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/kivikakk/go-twitter/twitter"
)

func New() *schema.Provider {
	p := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"consumer_api_key": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("TWITTER_CONSUMER_KEY", ""),
			},
			"consumer_api_secret": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("TWITTER_CONSUMER_SECRET", ""),
			},
			"access_token": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("TWITTER_ACCESS_TOKEN", ""),
			},
			"access_token_secret": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("TWITTER_ACCESS_TOKEN_SECRET", ""),
			},
		},
		DataSourcesMap: map[string]*schema.Resource{
			//"scaffolding_data_source": dataSourceScaffolding(),
		},
		ResourcesMap: map[string]*schema.Resource{
			"twitter_list": resourceList(),
			// "twitter_list_member"
			// "twitter_block"
			// "twitter_mute"
		},
	}

	p.ConfigureContextFunc = configure(p)

	return p
}

type client struct {
	*twitter.Client
}

func configure(p *schema.Provider) schema.ConfigureContextFunc {
	return func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		consumerKey := d.Get("consumer_api_key").(string)
		consumerSecret := d.Get("consumer_api_secret").(string)
		accessToken := d.Get("access_token").(string)
		accessSecret := d.Get("access_token_secret").(string)

		oac := oauth1.NewConfig(consumerKey, consumerSecret)
		oat := oauth1.NewToken(accessToken, accessSecret)
		hc := oac.Client(ctx, oat)

		// add logging transport for debugging
		transport := hc.Transport
		if transport == nil {
			transport = http.DefaultTransport
		}

		api := twitter.NewClient(hc)
		c := &client{
			Client: api,
		}
		return c, nil
	}
}
