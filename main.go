package main

import (
	_ "caddy-domeneshop/provider"

	caddycmd "github.com/caddyserver/caddy/v2/cmd"
	_ "github.com/caddyserver/caddy/v2/modules/standard"
	_ "github.com/mholt/caddy-l4/modules/l4tls"
)

func main() {
	caddycmd.Main()
}
