# Webhook Relay

A lightweight, standalone Go application that receives generic webhooks and forwards them to configured services (for example [Alertzy](https://alertzy.app) and Pushover).

## Features

- **Ultra-lightweight** — ~5MB memory footprint
- **Single binary** — no dependencies, easy deployment
- **Generic webhook support** — forwards service-specific payloads
- **Pluggable services** — currently includes Alertzy and Pushover
- **Optional authentication** — bearer token
- **Service-agnostic config** — per-request service config in JSON body

## Installation

### Using Pre-built Binaries

Download the latest release from the [Releases](../../releases) page. Binaries are available for:
- Linux x64 and ARM64
- macOS x64 and ARM64
- Windows x64

Extract and run:
```bash
chmod +x webhook-relay-linux-*
./webhook-relay-linux-* -port 8080 -token YOUR_TOKEN
```

### Building from Source

Requires Go 1.26+:
```bash
go build -o webhook-relay ./cmd/webhook-relay
```

### Running in Development

```bash
go run ./cmd/webhook-relay -port 8080 -token YOUR_TOKEN
```

For live reload on file changes, use [air](https://github.com/air-verse/air):
```bash
go install github.com/air-verse/air@latest
air --build.cmd "go build -o ./tmp/main ./cmd/webhook-relay" --build.bin "./tmp/main -- -port 8080 -token YOUR_TOKEN"
```

## Deployment

### systemd Service

Create `/etc/systemd/system/webhook-relay.service`:

```ini
[Unit]
Description=Webhook Relay
After=network.target

[Service]
Type=simple
User=www-data
ExecStart=/opt/webhook-relay/webhook-relay -port 8080 -token YOUR_TOKEN
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

Enable and start:
```bash
sudo systemctl daemon-reload
sudo systemctl enable webhook-relay
sudo systemctl start webhook-relay
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

The relay runs bare metal while Caddy runs in Docker. Use `host.docker.internal` (Docker Desktop) or the host's IP address:

```caddyfile
webhook.your-domain.com {
    reverse_proxy host.docker.internal:8080
}
```

For Linux hosts, replace `host.docker.internal` with the host's IP address in the Caddyfile.

## Usage

### List Available Services

```bash
curl http://localhost:8080/services
# {"services":["alertzy","pushover"]}
```

### Send To Alertzy

```bash
curl -X POST http://localhost:8080/alertzy \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "config": {
      "accountKey": "TARGET_ALERTZY_ACCOUNT_KEY",
      "group": "ops"
    },
    "payload": {
      "title": "Deploy failed",
      "message": "api deployment failed on prod",
      "buildId": "12345"
    }
  }'
```

### Send To Pushover

```bash
curl -X POST http://localhost:8080/pushover \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "config": {
      "token": "PUSHOVER_APP_TOKEN",
      "user": "PUSHOVER_USER_KEY",
      "device": "iphone",
      "sound": "pushover"
    },
    "payload": {
      "title": "Deploy finished",
      "message": "Production deploy completed successfully"
    }
  }'
```

## Configuration

### Flags

```
-port int          Port to listen on (default 8080)
-token string      Optional auth token (or WEBHOOK_TOKEN env var)
```

### Environment Variables

```bash
export WEBHOOK_TOKEN="token1,token2,token3"  # Comma-separated for multiple tokens
```

## Webhook Payload Format

Each request to `POST /{service}` must use the envelope format:

```json
{
  "config": {
    "service_specific_key": "value"
  },
  "payload": {
    "title": "Optional title",
    "message": "Optional message",
    "anythingElse": "forwarded data"
  }
}
```

Rules:
- `config` contains service-specific credentials/settings.
- `payload` contains the event data to forward.
- If `payload.title` is missing, the relay uses `Webhook Alert`.
- If `payload.message` is missing, the relay uses formatted `payload` JSON as message.

### Service Config Requirements

- Alertzy (`POST /alertzy`): `config.accountKey`, `config.group`
- Pushover (`POST /pushover`): `config.token`, `config.user`

## Health Check

```bash
curl http://localhost:8080/health
# {"status":"ok"}
```

## Logs

View systemd logs:
```bash
sudo journalctl -u webhook-relay -f
```

## Performance

- **Memory**: ~5MB idle
- **CPU**: Minimal (I/O bound)
- **Concurrent requests**: Handles thousands per minute
- **Binary size**: ~6-8MB

## License

MIT
