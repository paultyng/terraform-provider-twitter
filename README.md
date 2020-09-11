[Published on the Terraform Registry](https://registry.terraform.io/providers/paultyng/twitter/latest)

# Terraform Provider Twitter

Maintain lists, blocks, and muted accounts in Terraform.

```terraform
resource "twitter_list" "hashicorp" {
  name = "HashiCorp Employees"
  mode = "public"

  description = "List of people publicly identifying as HashiCorp employees."

  members = [
    "ptyng",
  ]
}

resource "twitter_block" "blocks" {
  screen_name = "ptyng"
}
```

## Rate Limiting

Twitters [rate limiting](https://developer.twitter.com/en/docs/basics/rate-limiting) for mutes and blocks is 15 requests per 15 minutes. In addition to just following the prescribed exponential backoff, the provider attempts to mitigate the chattiness of normal Terraform interactions by batching reads to list operations, you can maximize this by bumping up the `parallelism` flag. This is a bit experimental.
