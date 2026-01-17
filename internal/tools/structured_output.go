package tools

import (
	"context"
	"reflect"

	"charm.land/fantasy"
	"charm.land/fantasy/schema"
)

const StructuredOutputToolName = "structured_output"

const structuredOutputDescription = `Submit your final structured response. Call this tool when you are ready to respond to the user.`

// NewStructuredOutputTool creates a structured_output tool with a dynamic schema.
// The schema is generated from the provided output type.
func NewStructuredOutputTool(outputSchema any) fantasy.AgentTool {
	s := schema.Generate(reflect.TypeOf(outputSchema))
	return &structuredOutputTool{schema: s}
}

// structuredOutputTool implements fantasy.AgentTool with a dynamic schema.
type structuredOutputTool struct {
	schema          fantasy.Schema
	providerOptions fantasy.ProviderOptions
}

func (t *structuredOutputTool) Info() fantasy.ToolInfo {
	required := t.schema.Required
	if required == nil {
		required = []string{}
	}
	return fantasy.ToolInfo{
		Name:        StructuredOutputToolName,
		Description: structuredOutputDescription,
		Parameters:  schema.ToParameters(t.schema),
		Required:    required,
	}
}

func (t *structuredOutputTool) Run(ctx context.Context, params fantasy.ToolCall) (fantasy.ToolResponse, error) {
	return fantasy.NewTextResponse("Response submitted"), nil
}

func (t *structuredOutputTool) ProviderOptions() fantasy.ProviderOptions {
	return t.providerOptions
}

func (t *structuredOutputTool) SetProviderOptions(opts fantasy.ProviderOptions) {
	t.providerOptions = opts
}
