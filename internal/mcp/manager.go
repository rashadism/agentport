package mcp

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"charm.land/fantasy"
	gomcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

const defaultTimeout = 30 * time.Second

// Config represents configuration for an MCP server
type Config struct {
	Name          string
	URL           string
	Headers       map[string]string
	TLSSkipVerify bool
}

// Manager manages multiple MCP client connections
type Manager struct {
	mu       sync.RWMutex
	sessions map[string]*gomcp.ClientSession
	configs  map[string]Config
}

// NewManager creates a new MCP manager
func NewManager() *Manager {
	return &Manager{
		sessions: make(map[string]*gomcp.ClientSession),
		configs:  make(map[string]Config),
	}
}

// Initialize connects to all configured MCP servers concurrently
func (m *Manager) Initialize(ctx context.Context, configs []Config) {
	var wg sync.WaitGroup

	for _, cfg := range configs {
		m.configs[cfg.Name] = cfg

		wg.Add(1)
		go func(cfg Config) {
			defer wg.Done()
			m.connect(ctx, cfg)
		}(cfg)
	}

	wg.Wait()
}

func (m *Manager) connect(ctx context.Context, cfg Config) {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	session, err := m.createSession(ctx, cfg)
	if err != nil {
		slog.Error("MCP connection failed", "server", cfg.Name, "url", cfg.URL, "error", err)
		return
	}

	// List tools to verify connection
	tools, err := session.ListTools(ctx, &gomcp.ListToolsParams{})
	if err != nil {
		session.Close()
		slog.Error("MCP failed to list tools", "server", cfg.Name, "error", err)
		return
	}

	m.mu.Lock()
	m.sessions[cfg.Name] = session
	m.mu.Unlock()

	slog.Info("MCP connected", "server", cfg.Name, "tools", len(tools.Tools))
}

func (m *Manager) createSession(ctx context.Context, cfg Config) (*gomcp.ClientSession, error) {
	httpClient := &http.Client{
		Transport: &headerRoundTripper{
			headers:       cfg.Headers,
			tlsSkipVerify: cfg.TLSSkipVerify,
		},
	}

	transport := &gomcp.StreamableClientTransport{
		Endpoint:             cfg.URL,
		HTTPClient:           httpClient,
		DisableStandaloneSSE: true,
	}

	client := gomcp.NewClient(
		&gomcp.Implementation{
			Name:    "fantasydemo",
			Version: "1.0.0",
		},
		nil,
	)

	return client.Connect(ctx, transport, nil)
}

// GetSession returns a session, reconnecting if necessary
func (m *Manager) GetSession(ctx context.Context, name string) (*gomcp.ClientSession, error) {
	m.mu.RLock()
	session, ok := m.sessions[name]
	cfg, hasCfg := m.configs[name]
	m.mu.RUnlock()

	if !hasCfg {
		return nil, fmt.Errorf("mcp '%s' not configured", name)
	}

	// Check if session is alive
	if ok {
		pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		if err := session.Ping(pingCtx, nil); err == nil {
			return session, nil
		}
	}

	// Reconnect
	slog.Debug("MCP reconnecting", "server", name)

	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	newSession, err := m.createSession(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("reconnection failed: %w", err)
	}

	m.mu.Lock()
	m.sessions[name] = newSession
	m.mu.Unlock()

	slog.Info("MCP reconnected", "server", name)
	return newSession, nil
}

// GetAllTools returns Fantasy AgentTools for all connected MCP servers
func (m *Manager) GetAllTools(ctx context.Context) []fantasy.AgentTool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var allTools []fantasy.AgentTool

	for name, session := range m.sessions {
		tools, err := session.ListTools(ctx, &gomcp.ListToolsParams{})
		if err != nil {
			slog.Warn("MCP failed to list tools", "server", name, "error", err)
			continue
		}

		for _, tool := range tools.Tools {
			allTools = append(allTools, &Tool{
				manager:    m,
				serverName: name,
				tool:       tool,
			})
		}
	}

	return allTools
}

// Close closes all MCP client sessions
func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for name, session := range m.sessions {
		if err := session.Close(); err != nil &&
			!errors.Is(err, io.EOF) &&
			!errors.Is(err, context.Canceled) {
			slog.Warn("MCP close error", "server", name, "error", err)
		}
	}

	return nil
}
