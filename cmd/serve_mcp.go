package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/B4Dmonkey/bit-pro/bitdir"
	"github.com/B4Dmonkey/bit-pro/task"
	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"
)

const (
	serveMCPCmdUse   = "mcp"
	taskReadTool     = "task_read"
	taskListTool     = "task_list"
	taskCreateTool   = "task_create"
	taskUpdateTool   = "task_update"
	taskMoveTool     = "task_move"
	feedbackAddTool  = "feedback_add"
	taskCompleteTool = "task_complete"
	taskDeleteTool   = "task_delete"

	statusProperty = "status"
)

const taskReadDescription = `Read one task by ID, returning its fields and its body together.

A track is a top-level task — one whole scope — and its ID has no dot, as in BIT-7. A bar is a
child of a track — one plan step — and its ID is dotted, as in BIT-7.3. The result carries body
alongside id, title, status, approved, phase, phase_label and parent, so reading a task's prose and
reading its summary are the same call rather than two.`

const taskListDescription = `List tasks as structured fields, in the order bp prints them.

A track is a top-level task — one whole scope — and its ID has no dot, as in BIT-7. A bar is a
child of a track — one plan step — and its ID is dotted, as in BIT-7.3. Set parent to a track ID
to list that track's direct bars in the track's own order; omit it to list every task.`

const taskCreateDescription = `Create a task and return its minted ID.

A track is a top-level task — one whole scope — and its ID has no dot, as in BIT-7. A bar is a
child of a track — one plan step — and its ID is dotted, as in BIT-7.3. Omit parent to mint a
track; set it to a track ID to mint a dotted bar under that track and append it to the track's
order. after names a sibling bar and places the new bar directly after it in that order, and is
only meaningful together with parent; omit it to append. phase and phase_label tag the verse the
bar serves. The returned id is the new task's ID, and its status starts at todo.`

const taskUpdateDescription = `Update a task's fields and report whether approval survived.

Every field is optional: one that is omitted is left unchanged, and one that is sent is written.
The returned approved reflects whether the edit revoked approval. Sending any of
title, body, phase or phase_label revokes it, so that a replan cannot quietly alter what someone
already blessed — revocation fires on the field being sent, not on its value differing, so a body
rewrite of an approved task comes back with approved false even if the text is unchanged.
Writing status todo revokes approval too, because a task pulled back for rework has to be
re-reviewed before it runs again, while a forward move to doing or done keeps approval, being the
act of doing work that was already approved.

Status does not cascade to the parent: setting a bar's status leaves its track untouched, so a
caller that wants the track to reflect its bars sets the track's status in a separate call. The
status enum admits only todo, doing and done, so that rollup can fail only by being left undone.`

const taskMoveDescription = `Resequence a bar within its track.

bar names the bar to move, and exactly one of before or after names the sibling it moves
relative to — a sibling being another bar under the same track. A bar's ID is stable identity, so
moving it keeps every existing reference to it — a commit message, a feedback note, a plan
citation — valid.`

const taskCompleteDescription = `File a signed-off track and its bars under .bit/completed/.

A track is a top-level task — one whole scope — and its ID has no dot, as in BIT-7. Completing one
relocates the track and every bar under it out of the active list, so a finished cycle stops
showing up in task_list. It refuses a track that still has an unfinished bar and there is no
override — set every bar's status to done first. The ID stays reserved rather than being freed, so
older commit messages and feedback notes that reference it remain valid.`

const taskDeleteDescription = `Remove a task from the active list by relocating it to .bit/archive/tasks/.

The task file is relocated rather than destroyed, so it stays recoverable on disk, and its ID stays
reserved rather than being freed — a commit message or feedback note that cites it remains valid,
and the ID is never re-minted onto a different task. Deleting a bar — a child of a track, whose ID
is dotted, as in BIT-7.3 — also removes it from its track's order, which is what makes dropping one
mid-plan safe. Setting force deletes a track that still has unfinished bars; without it such a
track is refused.`

const feedbackAddDescription = `Record a feedback note against a track and return its path.

A note keys to a track — a top-level task, whose ID has no dot, as in BIT-7 — and cites the bar it
happened at in its own prose, because replanning renumbers bars and would orphan a note keyed to
one. The write is create-only: each note lands in a new file, so adding one can never damage a
note already recorded. A completed or archived track is accepted as readily as an active one.`

type taskReadInput struct {
	ID string `json:"id"`
}

type taskListInput struct {
	Parent string `json:"parent,omitempty"`
}

type taskSummary struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	Status     string `json:"status"`
	Approved   bool   `json:"approved"`
	Phase      int    `json:"phase"`
	PhaseLabel string `json:"phase_label"`
	Parent     string `json:"parent"`
}

type taskListOutput struct {
	Tasks []taskSummary `json:"tasks"`
}

type taskCreateInput struct {
	Title      string `json:"title"`
	Body       string `json:"body,omitempty"`
	Parent     string `json:"parent,omitempty"`
	After      string `json:"after,omitempty"`
	Phase      int    `json:"phase,omitempty"`
	PhaseLabel string `json:"phase_label,omitempty"`
}

type taskCreateOutput struct {
	ID string `json:"id"`
}

type taskUpdateInput struct {
	ID         string  `json:"id"`
	Title      *string `json:"title,omitempty"`
	Body       *string `json:"body,omitempty"`
	Status     *string `json:"status,omitempty"`
	Phase      *int    `json:"phase,omitempty"`
	PhaseLabel *string `json:"phase_label,omitempty"`
}

type taskCompleteInput struct {
	ID string `json:"id"`
}

type taskDeleteInput struct {
	ID    string `json:"id"`
	Force bool   `json:"force,omitempty"`
}

type feedbackAddInput struct {
	Track string `json:"track"`
	Body  string `json:"body"`
}

type feedbackAddOutput struct {
	Path string `json:"path"`
}

type taskMoveInput struct {
	Bar    string `json:"bar"`
	Before string `json:"before,omitempty"`
	After  string `json:"after,omitempty"`
}

// emptyOutput is the result of a tool whose success carries no fields. The SDK
// derives an object schema for it and marshals it to {}.
type emptyOutput struct{}

type taskUpdateOutput struct {
	ID       string `json:"id"`
	Approved bool   `json:"approved"`
}

type taskReadOutput struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	Status     string `json:"status"`
	Approved   bool   `json:"approved"`
	Phase      int    `json:"phase"`
	PhaseLabel string `json:"phase_label"`
	Parent     string `json:"parent"`
	Body       string `json:"body"`
}

func newServeMCPCmd() *cobra.Command {
	return &cobra.Command{
		Use:         serveMCPCmdUse,
		Short:       "Run the MCP server in the foreground",
		Args:        cobra.NoArgs,
		Annotations: map[string]string{quietAnnotation: quietEnabled},
		RunE: func(cmd *cobra.Command, args []string) error {
			root := os.Getenv("CLAUDE_PROJECT_DIR")

			return runMCPServer(cmd.Context(), root, &mcp.StdioTransport{})
		},
	}
}

func runMCPServer(ctx context.Context, root string, transport mcp.Transport) error {
	s := mcp.NewServer(&mcp.Implementation{Name: "bp", Version: "1"}, nil)
	mcp.AddTool(s, &mcp.Tool{Name: taskReadTool, Description: taskReadDescription}, taskReadHandler(root))
	mcp.AddTool(s, &mcp.Tool{Name: taskListTool, Description: taskListDescription}, taskListHandler(root))
	mcp.AddTool(s, &mcp.Tool{Name: taskCreateTool, Description: taskCreateDescription}, taskCreateHandler(root))

	updateSchema, err := taskUpdateSchema()
	if err != nil {
		return fmt.Errorf("building %s input schema: %w", taskUpdateTool, err)
	}

	mcp.AddTool(s, &mcp.Tool{
		Name:        taskUpdateTool,
		Description: taskUpdateDescription,
		InputSchema: updateSchema,
	}, taskUpdateHandler(root))

	mcp.AddTool(s, &mcp.Tool{Name: taskMoveTool, Description: taskMoveDescription}, taskMoveHandler(root))
	mcp.AddTool(s, &mcp.Tool{Name: feedbackAddTool, Description: feedbackAddDescription}, feedbackAddHandler(root))
	mcp.AddTool(s, &mcp.Tool{
		Name:        taskCompleteTool,
		Description: taskCompleteDescription,
	}, taskCompleteHandler(root))
	mcp.AddTool(s, &mcp.Tool{
		Name:        taskDeleteTool,
		Description: taskDeleteDescription,
	}, taskDeleteHandler(root))

	return s.Run(ctx, transport)
}

func taskUpdateSchema() (*jsonschema.Schema, error) {
	schema, err := jsonschema.For[taskUpdateInput](nil)
	if err != nil {
		return nil, err
	}

	schema.Properties[statusProperty].Enum = []any{task.StatusTodo, task.StatusDoing, task.StatusDone}

	return schema, nil
}

func taskReadHandler(root string) mcp.ToolHandlerFor[taskReadInput, taskReadOutput] {
	return func(
		_ context.Context,
		_ *mcp.CallToolRequest,
		in taskReadInput,
	) (*mcp.CallToolResult, taskReadOutput, error) {
		store := task.New(bitdir.ForRoot(root))

		t, err := store.Load(in.ID)
		if err != nil {
			return nil, taskReadOutput{}, fmt.Errorf("loading task %s: %w", in.ID, err)
		}

		return nil, taskReadOutput{
			ID:         t.ID,
			Title:      t.Title,
			Status:     t.Status,
			Approved:   t.Approved,
			Phase:      t.Phase,
			PhaseLabel: t.PhaseLabel,
			Parent:     parentOf(t.ID),
			Body:       t.Body,
		}, nil
	}
}

func taskListHandler(root string) mcp.ToolHandlerFor[taskListInput, taskListOutput] {
	return func(
		_ context.Context,
		_ *mcp.CallToolRequest,
		in taskListInput,
	) (*mcp.CallToolResult, taskListOutput, error) {
		store := task.New(bitdir.ForRoot(root))

		var (
			tasks []*task.Task
			err   error
		)

		if in.Parent == "" {
			tasks, err = store.List()
		} else {
			tasks, err = store.Children(in.Parent)
		}

		if err != nil {
			return nil, taskListOutput{}, fmt.Errorf("listing tasks: %w", err)
		}

		out := taskListOutput{Tasks: make([]taskSummary, 0, len(tasks))}
		for _, t := range tasks {
			out.Tasks = append(out.Tasks, taskSummary{
				ID:         t.ID,
				Title:      t.Title,
				Status:     t.Status,
				Approved:   t.Approved,
				Phase:      t.Phase,
				PhaseLabel: t.PhaseLabel,
				Parent:     parentOf(t.ID),
			})
		}

		return nil, out, nil
	}
}

func taskCreateHandler(root string) mcp.ToolHandlerFor[taskCreateInput, taskCreateOutput] {
	return func(
		_ context.Context,
		_ *mcp.CallToolRequest,
		in taskCreateInput,
	) (*mcp.CallToolResult, taskCreateOutput, error) {
		store := task.New(bitdir.ForRoot(root))

		t, err := store.Create(task.CreateParams{
			Title:      in.Title,
			Body:       in.Body,
			Parent:     in.Parent,
			After:      in.After,
			Phase:      in.Phase,
			PhaseLabel: in.PhaseLabel,
		})
		if err != nil {
			return nil, taskCreateOutput{}, fmt.Errorf("creating task: %w", err)
		}

		return nil, taskCreateOutput{ID: t.ID}, nil
	}
}

func taskUpdateHandler(root string) mcp.ToolHandlerFor[taskUpdateInput, taskUpdateOutput] {
	return func(
		_ context.Context,
		_ *mcp.CallToolRequest,
		in taskUpdateInput,
	) (*mcp.CallToolResult, taskUpdateOutput, error) {
		store := task.New(bitdir.ForRoot(root))

		t, err := store.Update(in.ID, task.Patch{
			Title:      in.Title,
			Body:       in.Body,
			Status:     in.Status,
			Phase:      in.Phase,
			PhaseLabel: in.PhaseLabel,
		})
		if err != nil {
			return nil, taskUpdateOutput{}, fmt.Errorf("updating task %s: %w", in.ID, err)
		}

		return nil, taskUpdateOutput{ID: t.ID, Approved: t.Approved}, nil
	}
}

func taskMoveHandler(root string) mcp.ToolHandlerFor[taskMoveInput, emptyOutput] {
	return func(
		_ context.Context,
		_ *mcp.CallToolRequest,
		in taskMoveInput,
	) (*mcp.CallToolResult, emptyOutput, error) {
		store := task.New(bitdir.ForRoot(root))

		if err := store.Move(in.Bar, in.Before, in.After); err != nil {
			return nil, emptyOutput{}, fmt.Errorf("moving bar %s: %w", in.Bar, err)
		}

		return nil, emptyOutput{}, nil
	}
}

func taskCompleteHandler(root string) mcp.ToolHandlerFor[taskCompleteInput, emptyOutput] {
	return func(
		_ context.Context,
		_ *mcp.CallToolRequest,
		in taskCompleteInput,
	) (*mcp.CallToolResult, emptyOutput, error) {
		store := task.New(bitdir.ForRoot(root))

		if err := store.Complete(in.ID); err != nil {
			return nil, emptyOutput{}, fmt.Errorf("completing task %s: %w", in.ID, err)
		}

		return nil, emptyOutput{}, nil
	}
}

func taskDeleteHandler(root string) mcp.ToolHandlerFor[taskDeleteInput, emptyOutput] {
	return func(
		_ context.Context,
		_ *mcp.CallToolRequest,
		in taskDeleteInput,
	) (*mcp.CallToolResult, emptyOutput, error) {
		store := task.New(bitdir.ForRoot(root))

		if err := store.Relocate(in.ID, in.Force); err != nil {
			return nil, emptyOutput{}, fmt.Errorf("deleting task %s: %w", in.ID, err)
		}

		return nil, emptyOutput{}, nil
	}
}

func feedbackAddHandler(root string) mcp.ToolHandlerFor[feedbackAddInput, feedbackAddOutput] {
	return func(
		_ context.Context,
		_ *mcp.CallToolRequest,
		in feedbackAddInput,
	) (*mcp.CallToolResult, feedbackAddOutput, error) {
		store := task.New(bitdir.ForRoot(root))

		path, err := store.AddNote(in.Track, in.Body)
		if err != nil {
			return nil, feedbackAddOutput{}, fmt.Errorf("adding note for %s: %w", in.Track, err)
		}

		return nil, feedbackAddOutput{Path: path}, nil
	}
}

func parentOf(id string) string {
	i := strings.LastIndex(id, ".")
	if i == -1 {
		return ""
	}

	return id[:i]
}
