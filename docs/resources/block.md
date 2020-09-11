---
page_title: "twitter_block Resource - terraform-provider-twitter"
subcategory: ""
description: |-
  
---

# Resource `twitter_block`



## Example Usage

```terraform
resource "twitter_block" "blocks" {
  screen_name = "ptyng"
}
```

## Schema

### Required

- **screen_name** (String, Required) The screen name of the potentially blocked user.

### Optional

- **id** (String, Optional) The ID of this resource.


