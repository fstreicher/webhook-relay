# Webhook to Alertzy Relay

A lightweight, standalone Go application that receives generic webhooks and forwards them to [Alertzy](https://alertzy.app).

## Features

- **Ultra-lightweight** — ~5MB memory footprint
- **Single binary** — no dependencies, easy deployment
- **Generic webhook support** — forwards any JSON payload
- **Optional authentication** — bearer token or query parameter
- **Configurable** — title, group, port via flags or environment variables

## Installation

### Using Pre-built Binaries

Download the latest release from the [Releases](../../releases) page. Binaries are available for:
- Linux x64 and ARM64
- macOS x64 and ARM64
- Windows x64

Extract and run:
```bash
chmod +x webhook-alertzy-relay-linux-*
./webhook-alertzy-relay-linux-* -key YOUR_ALERTZY_KEY -port 8080
```

### Building from Source

Requires Go 1.26+:
```bash
go build -o webhook-alertzy-relay main.go
```

## Deployment

### systemd Service

Create `/etc/systemd/system/webhook-alertzy-relay.service`:

```ini
[Unit]
Description=Webhook to Alertzy Relay
After=network.target

[Service]
Type=simple
User=www-data
ExecStart=/opt/webhook-alertzy-relay/webhook-alertzy-relay -key YOUR_ALERTZY_KEY -port 8080
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

Enable and start:
```bash
sudo systemctl daemon-reload
sudo systemctl enable webhook-alertzy-relay
sudo systemctl start webhook-alertzy-relay
```

### Reverse Proxy (Nginx)

```nginx
server {
    listen 80;
    server_name webhook.your-domain.com;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

### Reverse Proxy (Caddy in Docker)

```caddyfile
webhook.your-domain.com {
    reverse_proxy webhook-relay:8080
}
```

Docker Compose example:
```yaml
version: '3.8'
services:
  webhook-relay:
    image: webhook-alertzy-relay:latest
    environment:
      ALERTZY_KEY: your-api-key

  caddy:
    image: caddy:latest
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./Caddyfile:/etc/caddy/Caddyfile
    depends_on:
      - webhook-relay
```

Caddy automatically handles HTTPS with Let's Encrypt.

## Usage

### Basic Webhook

```bash
curl -X POST http://localhost:8080/webhook \
  -H "Content-Type: application/json" \
  -d '{"title": "Alert", "message": "Something happened"}'
```

### With Authentication (Bearer Token)

```bash
curl -X POST http://localhost:8080/webhook \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title": "Alert", "message": "Something happened"}'
```

### With Query Token

```bash
curl -X POST "http://localhost:8080/webhook?token=YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title": "Alert", "message": "Something happened"}'
```

## Configuration

### Flags

```
-port int          Port to listen on (default 8080)
-key string        Alertzy account key (or ALERTZY_KEY env var)
-group string      Alertzy group name (default "webhooks")
-title string      Default alert title (default "Webhook Alert")
-token string      Optional auth token (or WEBHOOK_TOKEN env var)
```

### Environment Variables

```bash
export ALERTZY_KEY="your_alertzy_key"
export WEBHOOK_TOKEN="token1,token2,token3"  # Comma-separated for multiple tokens
export PORT=8080
```

## Webhook Payload Format

The relay looks for these fields in your JSON payload:

```json
{
  "title": "Custom Title",
  "message": "Your message here",
  "any_other_field": "will be ignored"
}
```

If `title` or `message` are missing, the relay will:
- Use default title from config
- Use formatted full payload as message

## Health Check

```bash
curl http://localhost:8080/health
# {"status":"ok"}
```

## Logs

View systemd logs:
```bash
sudo journalctl -u webhook-alertzy-relay -f
```

## Performance

- **Memory**: ~5MB idle
- **CPU**: Minimal (I/O bound)
- **Concurrent requests**: Handles thousands per minute
- **Binary size**: ~6-8MB

## License

MIT
