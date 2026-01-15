package mcp

import "slices"

// Allowed MCP tools for the RCA agent.
var allowedTools = []string{
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

// FilterTools returns only the tools that are allowed.
func FilterTools(tools []*Tool) []*Tool {
	filtered := make([]*Tool, 0, len(allowedTools))
	for _, t := range tools {
		if slices.Contains(allowedTools, t.tool.Name) {
			filtered = append(filtered, t)
		}
	}
	return filtered
}
