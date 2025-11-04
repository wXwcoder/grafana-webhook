# Grafana Webhook for WeChat and Feishu

A webhook receiver server that forwards Grafana alerts and GitLab events to WeChat Work (企业微信) and Feishu (飞书) bots.

## Features

- Receives webhook events from Grafana and GitLab
- Formats and forwards alerts to messaging platforms:
  - WeChat Work (企业微信) robots
  - Feishu (飞书) bots
- Supports multiple event types:
  - Grafana alerts
  - GitLab push events
  - GitLab merge request events
- Configurable webhook endpoint

## Prerequisites

- Go 1.18 or higher
- Access to WeChat Work or Feishu bot webhook URLs

## Installation

1. Clone the repository:
   ```bash
   git clone <repository-url>
   cd webhook
   ```

2. Build the binary:
   ```bash
   go build -o webhook
   ```
   
   Or use the provided build scripts:
   - On Windows: `make.bat`
   - On Linux/macOS: `chmod +x start.sh && ./start.sh`

## Usage

Run the webhook server with the following parameters:

```bash
./webhook [port] [platform] [robot_url]
```

Where:
- `port`: The port number to listen on (e.g., 8081)
- `platform`: Either "weixin" for WeChat Work or "feishu" for Feishu
- `robot_url`: The full webhook URL for your bot

Example:
```bash
./webhook 8081 "feishu" "https://open.feishu.cn/open-apis/bot/v2/hook/your-robot-key"
```

### Configuration in GitLab

1. Go to your GitLab project settings
2. Navigate to Webhooks
3. Set the URL to: `http://[your-server-ip]:[port]/webhook`
4. Select the events you want to receive

### Configuration in Grafana

1. Go to Grafana Alerting settings
2. Add a new webhook contact point
3. Set the URL to: `http://[your-server-ip]:[port]/webhook`

## Deployment

The webhook server can be deployed in different ways:

### Direct Execution

```bash
./webhook 8081 "feishu" "https://open.feishu.cn/open-apis/bot/v2/hook/your-robot-key"
```

### Using the Start Script

On Linux/macOS systems, you can use the provided start script:

```bash
chmod +x start.sh
./start.sh
```

Note: Modify the parameters in `start.sh` according to your needs.

## Building for Different Platforms

### Cross-compilation for Linux (from Windows)

Use the provided `make.bat` script to build for Linux:

```cmd
make.bat
```

## How It Works

1. The webhook server listens for incoming HTTP POST requests on the configured port and path (`/webhook`)
2. When an event is received, it determines the source (Grafana or GitLab)
3. The event is formatted according to the target platform (WeChat Work or Feishu)
4. The formatted message is sent to the configured bot URL

## Supported Events

### Grafana Alerts

Receives and formats Grafana alert notifications with detailed information including:
- Alert status (firing/resolved)
- Alert title and message
- Labels and annotations
- Timestamps

### GitLab Events

Supports various GitLab events:
- Push events
- Merge request events

## Version Information

To check the version of the webhook server:

```bash
./webhook --version
```

## Help

To display help information:

```bash
./webhook --help
```

## Dependencies

- [github.com/gin-gonic/gin](https://github.com/gin-gonic/gin) - HTTP web framework
- [github.com/go-playground/webhooks/v6](https://github.com/go-playground/webhooks) - Webhook parser

View all dependencies in [go.mod](go.mod).

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details if it exists in your repository.
