module github.com/paultyng/terraform-provider-twitter

go 1.12

replace github.com/paultyng/go-twitter => ../go-twitter

require (
	github.com/dghubble/oauth1 v0.6.0
	github.com/hashicorp/go-cty v1.4.1-0.20200414143053-d3edf31b6320
	github.com/hashicorp/terraform-plugin-sdk/v2 v2.0.0-rc.1.0.20200507004823-7d4190ae1797
	// pinning to c1e5228532ead77d14ec895fc4eced13f480c855 on this fork of github.com/dghubble/go-twitter
	// they have mute/block implemented, its proposed as PR: https://github.com/dghubble/go-twitter/pull/149
	github.com/paultyng/go-twitter v0.0.0-20200206005214-c1e5228532ea
	golang.org/x/net v0.0.0-20200506145744-7e3656a0809f // indirect
)
