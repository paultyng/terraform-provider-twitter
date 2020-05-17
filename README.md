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
