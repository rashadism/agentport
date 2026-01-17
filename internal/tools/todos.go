package tools

import (
	"context"
	_ "embed"
	"fmt"
	"strings"

	"charm.land/fantasy"
)

//go:embed todos.md
var todosDescription string

const TodosToolName = "todos"

type TodosParams struct {
	Todos []Todo `json:"todos" description:"The updated todo list"`
}

type Todo struct {
	Content string `json:"content" description:"What needs to be done"`
	Status  string `json:"status" description:"Task status: pending, in_progress, or completed"`
}

func NewTodosTool() fantasy.AgentTool {
	return fantasy.NewAgentTool(
		TodosToolName,
		todosDescription,
		func(ctx context.Context, params TodosParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			// Validate statuses
			for _, todo := range params.Todos {
				switch todo.Status {
				case "pending", "in_progress", "completed":
				default:
					return fantasy.ToolResponse{}, fmt.Errorf("invalid status %q for todo %q", todo.Status, todo.Content)
				}
			}

			// Count by status
			var pending, inProgress, completed int
			for _, todo := range params.Todos {
				switch todo.Status {
				case "pending":
					pending++
				case "in_progress":
					inProgress++
				case "completed":
					completed++
				}
			}

			// Build response with current todo state
			var sb strings.Builder
			sb.WriteString("Todo list updated.\n\n")
			sb.WriteString("Current todos:\n")
			for _, todo := range params.Todos {
				var icon string
				switch todo.Status {
				case "pending":
					icon = "[ ]"
				case "in_progress":
					icon = "[~]"
				case "completed":
					icon = "[x]"
				}
				sb.WriteString(fmt.Sprintf("  %s %s\n", icon, todo.Content))
			}
			sb.WriteString(fmt.Sprintf("\nStatus: %d pending, %d in progress, %d completed",
				pending, inProgress, completed))

			return fantasy.NewTextResponse(sb.String()), nil
		})
}
