# Webhook to Alertzy Relay

A lightweight, standalone Go application that receives generic webhooks and forwards them to [Alertzy](https://alertzy.app).

## Features

- **Ultra-lightweight** — ~5MB memory footprint
- **Single binary** — no dependencies, easy deployment
- **Generic webhook support** — forwards any JSON payload
- **Optional authentication** — bearer token or query parameter
- **Configurable** — title, group, port via flags or environment variables

## Building

### Prerequisites
- Go 1.21+ installed

### Compile for Linux ARM

First, determine your VPS ARM architecture:
```bash
# On your VPS
uname -m
```

Then build accordingly:
```bash
# For ARM64 (aarch64) - most common modern ARM VPS
GOOS=linux GOARCH=arm64 go build -o webhook-alertzy-relay main.go

# For ARMv7 (armv7l)
GOOS=linux GOARCH=arm GOARM=7 go build -o webhook-alertzy-relay main.go

# For ARMv6 (armv6l) - older systems/Raspberry Pi
GOOS=linux GOARCH=arm GOARM=6 go build -o webhook-alertzy-relay main.go

# For current platform (if already on Linux ARM)
go build -o webhook-alertzy-relay main.go
```

## Deployment

### 1. Prepare on VPS

```bash
# Copy binary to VPS
scp webhook-alertzy-relay user@your-vps:/opt/webhook-alertzy-relay/

# Or download directly on VPS if you've uploaded it to a release
tar -xzf webhook-alertzy-relay-linux-amd64.tar.gz -C /opt/webhook-alertzy-relay/
```

### 2. Create systemd Service

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

### 3. Start Service

```bash
sudo systemctl daemon-reload
sudo systemctl enable webhook-alertzy-relay
sudo systemctl start webhook-alertzy-relay
sudo systemctl status webhook-alertzy-relay
```

### 4. Configure Reverse Proxy (Nginx)

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

Enable SSL with Let's Encrypt:
```bash
sudo certbot certonly --nginx -d webhook.your-domain.com
```

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
