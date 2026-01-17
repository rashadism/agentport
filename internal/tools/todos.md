Create and manage a structured task list for tracking progress on complex, multi-step tasks.

## When to Use

- Complex multi-step tasks requiring 3+ distinct steps
- Non-trivial tasks requiring careful planning
- User provides multiple tasks (numbered or comma-separated)
- After receiving new instructions to capture requirements
- When starting work on a task (mark as in_progress BEFORE beginning)
- After completing a task (mark as completed)

## Task Breakdown

Break down the user's request into subtasks based on what can be accomplished with your available tools. Each subtask should correspond to a tool call or a logical grouping of related tool calls. Consider the data dependencies between tools when ordering subtasks.

## When NOT to Use

- Single, straightforward task
- Task completable in less than 3 trivial steps
- Purely conversational or informational request

## Task States

- **pending**: Task not yet started
- **in_progress**: Currently working on
- **completed**: Task finished successfully

## Task Management

- Update task status in real-time as you work
- Mark tasks complete IMMEDIATELY after finishing
- Only have ONE task as in_progress at a time
- Complete current tasks before starting new ones
- Remove tasks that are no longer relevant

## Completion Requirements

ONLY mark a task as completed when FULLY accomplished.

Never mark completed if:
- There are unresolved issues or errors
- Work is partial or incomplete
- You encountered blockers

If blocked: keep task as in_progress and create a new task describing what needs resolution.
