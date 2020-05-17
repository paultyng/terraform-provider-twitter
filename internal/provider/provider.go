package provider

import (
	"context"
	"net/http"
	"time"

	"github.com/dghubble/oauth1"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/paultyng/go-batcher"
	"github.com/paultyng/go-twitter/twitter"
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
			"twitter_block": resourceBlock(),
			"twitter_mute":  resourceMute(),
			// "twitter_mute_word":
			// "twitter_saved_search"
		},
	}

	p.ConfigureContextFunc = configure(p)

	return p
}

type client struct {
	*twitter.Client

	blockBatcher *batcher.Batcher
	muteBatcher  *batcher.Batcher
}

func configure(p *schema.Provider) schema.ConfigureContextFunc {
	return func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		consumerKey := d.Get("consumer_api_key").(string)
		consumerSecret := d.Get("consumer_api_secret").(string)
		accessToken := d.Get("access_token").(string)
		accessSecret := d.Get("access_token_secret").(string)

		oac := oauth1.NewConfig(consumerKey, consumerSecret)
		oat := oauth1.NewToken(accessToken, accessSecret)
		// this is a little hacky, but didn't see a different way this was
		// exposed to combine this with retry
		// only a Transport is set on the client, so pull that out
		oauthTransport := oac.Client(ctx, oat).Transport.(*oauth1.Transport)

		// add exponential backoff for rate limiting
		rc := retryablehttp.NewClient()
		rc.RetryWaitMin = 4 * time.Second
		rc.RetryWaitMax = 8 * time.Minute
		rc.RetryMax = 10
		rc.CheckRetry = func(ctx context.Context, resp *http.Response, err error) (bool, error) {
			// do not retry on context.Canceled or context.DeadlineExceeded
			if ctx.Err() != nil {
				return false, ctx.Err()
			}

			if resp.StatusCode == 429 {
				return true, nil
			}

			return false, nil
		}
		hc := rc.StandardClient()

		oauthTransport.Base = hc.Transport
		hc.Transport = oauthTransport

		// TODO: add logging transport?

		api := twitter.NewClient(hc)
		c := &client{
			Client: api,

			blockBatcher: batcher.New(2*time.Second, fetchBlockBatch(api)),
			muteBatcher:  batcher.New(2*time.Second, fetchMuteBatch(api)),
		}
		return c, nil
	}
}
