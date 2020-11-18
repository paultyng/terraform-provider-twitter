---
page_title: "twitter_list Resource - terraform-provider-twitter"
subcategory: ""
description: |-
  
---

# Resource `twitter_list`



## Example Usage

```terraform
resource "twitter_list" "hashicorp" {
  name = "HashiCorp Employees"
  mode = "public"

  description = "List of people publicly identifying as HashiCorp employees."

  members = [
    "ptyng",
  ]
}
```

## Schema

### Required

- **name** (String, Required) The name for the list. A list's name must start with a letter and can consist only of 25 or fewer letters, numbers, "-", or "_" characters.

### Optional

- **description** (String, Optional) The description to give the list. At most 100 characters.
- **id** (String, Optional) The ID of this resource.
- **members** (Set of String, Optional) The screen names of the user for whom to return results.
- **mode** (String, Optional) Whether your list is public or private. Values can be `public` or `private`.

### Read-only

- **slug** (String, Read-only)
- **uri** (String, Read-only)


