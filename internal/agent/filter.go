package agent

import (
	"slices"

	"charm.land/fantasy"

	"rca.agent/test/internal/mcp"
)

// allowedMCPTools is the list of MCP tools allowed for the agent.
var allowedMCPTools = []string{
	// Observability
	"get_traces",
	"get_component_logs",
	"get_project_logs",
	"get_component_resource_metrics",
	// OpenChoreo
	"list_environments",
	"list_organizations",
	"list_projects",
	"list_components",
}

// filterMCPTools returns only the allowed tools as AgentTools.
func filterMCPTools(tools []*mcp.Tool) []fantasy.AgentTool {
	filtered := make([]fantasy.AgentTool, 0, len(allowedMCPTools))
	for _, t := range tools {
		if slices.Contains(allowedMCPTools, t.Name()) {
			filtered = append(filtered, t)
		}
	}
	return filtered
}
