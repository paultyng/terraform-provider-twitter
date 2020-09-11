resource "twitter_list" "hashicorp" {
  name = "HashiCorp Employees"
  mode = "public"

  description = "List of people publicly identifying as HashiCorp employees."

  members = [
    "ptyng",
  ]
}