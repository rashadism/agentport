# Fantasy Demo

An AI-powered analysis server using the Fantasy framework for agent orchestration with multi-provider LLM support and Model Context Protocol (MCP) integration.

## Features

- Multi-provider LLM support (OpenAI, Anthropic, Google)
- Model Context Protocol (MCP) for external tool integration
- OAuth2 authentication for secure MCP server access
- RESTful API for analysis requests

## Project Structure

```
fantasydemo/
├── cmd/
│   └── fantasydemo/
│       └── main.go              # Application entry point
├── internal/
│   ├── auth/                    # OAuth2 token management
│   ├── config/                  # Configuration loading
│   ├── handler/                 # HTTP request/response types
│   ├── mcp/                     # MCP client management
│   ├── provider/                # LLM provider abstraction
│   └── server/                  # HTTP server and handlers
├── .gitignore
├── Makefile
├── go.mod
└── README.md
```

## Prerequisites

- Go 1.25+
- LLM API key (OpenAI, Anthropic, or Google)

## Configuration

Set the following environment variables (or use `make run-local` with a `.env` file):

| Variable | Description | Required |
|----------|-------------|----------|
| `RCA_LLM_API_KEY` | API key for the LLM provider | Yes |
| `RCA_MODEL_NAME` | Model to use (e.g., `gpt-4o`, `claude-sonnet-4-20250514`) | No (default: `gpt-4o`) |
| `SERVER_PORT` | HTTP server port | No (default: `8080`) |
| `OBSERVER_MCP_URL` | Observer MCP server URL | No |
| `OPENCHOREO_MCP_URL` | OpenChoreo MCP server URL | No |

## Usage

### Build

```bash
make build
```

### Run

```bash
export RCA_LLM_API_KEY=your-api-key
make run
```

### API Endpoints

#### Health Check
```bash
curl http://localhost:8080/health
```

#### Analyze
```bash
curl -X POST http://localhost:8080/analyze \
  -H "Content-Type: application/json" \
  -d '{"prompt": "Your analysis request here"}'
```

## Development

```bash
# Run tests
make test

# Run linter
make vet

# Format code
make fmt
```
