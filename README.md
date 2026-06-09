# meshbrow CLI

The official command-line tool for [Meshbrow](https://meshbrow.dev) — Managed Browser Fleet for AI Agents.

## Installation

### Homebrew (macOS/Linux)

```bash
brew tap meshbrow-dev/tap
brew install meshbrow
```

### Shell Script

```bash
curl -sSL https://get.meshbrow.dev | sh
```

### Go Install

```bash
go install github.com/meshbrow-dev/meshbrow-cli@latest
```

### From Source

```bash
git clone https://github.com/meshbrow-dev/meshbrow-cli.git
cd meshbrow-cli
make install
```

## Quick Start

```bash
# Authenticate
meshbrow auth login --key mb_live_your_key_here

# Launch a stealth browser session
meshbrow sessions create --stealth max --proxy-country US

# List active sessions
meshbrow sessions list

# Take a screenshot
meshbrow screenshot sess_abc123 -f page.png

# Destroy a session
meshbrow sessions destroy sess_abc123
```

## Commands

| Command | Description |
|---------|-------------|
| `meshbrow auth login` | Save API key |
| `meshbrow auth status` | Check auth status |
| `meshbrow auth logout` | Remove stored key |
| `meshbrow sessions create` | Launch a browser session |
| `meshbrow sessions list` | List active sessions |
| `meshbrow sessions get` | Get session details |
| `meshbrow sessions destroy` | Destroy a session |
| `meshbrow profiles create` | Create a browser profile |
| `meshbrow profiles list` | List profiles |
| `meshbrow profiles delete` | Delete a profile |
| `meshbrow fleet create` | Create a multi-session fleet |
| `meshbrow fleet status` | Check fleet status |
| `meshbrow fleet destroy` | Destroy a fleet |
| `meshbrow screenshot` | Capture a screenshot |
| `meshbrow navigate` | Navigate to a URL |
| `meshbrow exec` | Execute JavaScript |
| `meshbrow cookies export` | Export session cookies |
| `meshbrow cookies import` | Import cookies |
| `meshbrow status` | System status and usage |
| `meshbrow version` | Print version info |

## Configuration

Config is stored in `~/.meshbrow.yaml`:

```yaml
api_key: mb_live_your_key_here
api_url: https://api.meshbrow.dev
```

Or use environment variables:

```bash
export MESHBROW_API_KEY=mb_live_your_key_here
export MESHBROW_API_URL=https://api.meshbrow.dev
```

## Documentation

Full CLI reference: [docs.meshbrow.dev/guides/cli](https://docs.meshbrow.dev/guides/cli)

## License

MIT
