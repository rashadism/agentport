package mcp

// ResponseTransformer transforms tool responses into structured formats for LLM consumption.
type ResponseTransformer interface {
	Transform(content map[string]any) (string, error)
}

// Registry maps tool names to their transformers
var transformers = map[string]ResponseTransformer{
	"get_component_logs":             &LogsTransformer{},
	"get_project_logs":               &ProjectLogsTransformer{},
	"get_component_resource_metrics": &MetricsTransformer{},
	"get_traces":                     &TracesTransformer{},
}

// GetTransformer returns the transformer for a tool, or nil if none exists.
func GetTransformer(toolName string) ResponseTransformer {
	return transformers[toolName]
}

// LogsTransformer formats component logs as markdown tables.
type LogsTransformer struct{}

func (t *LogsTransformer) Transform(content map[string]any) (string, error) {
	// TODO: Format logs as markdown table
	return "", nil
}

// ProjectLogsTransformer groups and formats logs by component.
type ProjectLogsTransformer struct{}

func (t *ProjectLogsTransformer) Transform(content map[string]any) (string, error) {
	// TODO: Group logs by component, format as markdown
	return "", nil
}

// MetricsTransformer calculates statistics and detects anomalies in metrics data.
type MetricsTransformer struct{}

func (t *MetricsTransformer) Transform(content map[string]any) (string, error) {
	// TODO: Calculate stats, detect anomalies, format as markdown
	return "", nil
}

// TracesTransformer builds hierarchical span trees from trace data.
type TracesTransformer struct{}

func (t *TracesTransformer) Transform(content map[string]any) (string, error) {
	// TODO: Build span tree, format as hierarchical markdown
	return "", nil
}
