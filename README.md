# Caddy.exe med Domeneshop DNS Challenge integrasjon og TCP/TLS Passthrough med Layer4 plugin.

Custom build av Caddy.exe med Domeneshop API integrasjon og TCP/TLS Passthrough med Layer4 plugin.

Denne tar med TCP/TLS Passthrough med Layer4 plugin (github.com/mholt/caddy-l4@latest) Den trenger du kanskje ikke og kan enkelt fjernes fra bygget ved å fjerne *_ "github.com/mholt/caddy-l4/modules/l4tls"* fra main.go og droppe *go get -u github.com/mholt/caddy-l4@latest* når du bygger.

## Bygge med oppdaterte avhengigheter
```bash
go get -u github.com/caddyserver/caddy/v2@latest
go get -u github.com/mholt/caddy-l4@latest
go mod tidy
go build -o .\build\caddy.exe
```
