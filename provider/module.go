package provider

import (
	"context"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/libdns/libdns"
)

func init() {
	caddy.RegisterModule(DomeneshopProvider{})
}

type DomeneshopProvider struct {
	Provider  *Provider `json:"-"`
	APIToken  string    `json:"api_token,omitempty"`
	APISecret string    `json:"api_secret,omitempty"`
}

func (DomeneshopProvider) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "dns.providers.domeneshop",
		New: func() caddy.Module { return new(DomeneshopProvider) },
	}
}

func (d *DomeneshopProvider) Provision(ctx caddy.Context) error {
	d.Provider = &Provider{
		APIToken:  d.APIToken,
		APISecret: d.APISecret,
	}
	return nil
}

func (d *DomeneshopProvider) UnmarshalCaddyfile(dispenser *caddyfile.Dispenser) error {
	for dispenser.Next() {
		if dispenser.NextArg() {
			return dispenser.ArgErr()
		}
		for nesting := dispenser.Nesting(); dispenser.NextBlock(nesting); {
			switch dispenser.Val() {
			case "api_token":
				if d.APIToken != "" {
					return dispenser.Err("API token already set")
				}
				if dispenser.NextArg() {
					d.APIToken = dispenser.Val()
				}
				if dispenser.NextArg() {
					return dispenser.ArgErr()
				}
			case "api_secret":
				if d.APISecret != "" {
					return dispenser.Err("API secret already set")
				}
				if dispenser.NextArg() {
					d.APISecret = dispenser.Val()
				}
				if dispenser.NextArg() {
					return dispenser.ArgErr()
				}
			default:
				return dispenser.Errf("unrecognized subdirective '%s'", dispenser.Val())
			}
		}
	}
	if d.APIToken == "" {
		return dispenser.Err("missing API token")
	}
	if d.APISecret == "" {
		return dispenser.Err("missing API secret")
	}
	return nil
}

func (d *DomeneshopProvider) GetRecords(ctx context.Context, zone string) ([]libdns.Record, error) {
	return d.Provider.GetRecords(ctx, zone)
}

func (d *DomeneshopProvider) AppendRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	return d.Provider.AppendRecords(ctx, zone, records)
}

func (d *DomeneshopProvider) SetRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	return d.Provider.SetRecords(ctx, zone, records)
}

func (d *DomeneshopProvider) DeleteRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	return d.Provider.DeleteRecords(ctx, zone, records)
}

var (
	_ caddyfile.Unmarshaler = (*DomeneshopProvider)(nil)
	_ caddy.Provisioner     = (*DomeneshopProvider)(nil)
	_ libdns.RecordGetter   = (*DomeneshopProvider)(nil)
	_ libdns.RecordAppender = (*DomeneshopProvider)(nil)
	_ libdns.RecordSetter   = (*DomeneshopProvider)(nil)
	_ libdns.RecordDeleter  = (*DomeneshopProvider)(nil)
)
