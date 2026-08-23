package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/B4Dmonkey/bit-pro/bitdir"
	"github.com/B4Dmonkey/bit-pro/task"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"
)

const (
	serveMCPCmdUse = "mcp"
	taskReadTool   = "task_read"
	taskListTool   = "task_list"
)

const taskListDescription = `List tasks as structured fields, in the order bp prints them.

A track is a top-level task — one whole scope — and its ID has no dot, as in BIT-7. A bar is a
child of a track — one plan step — and its ID is dotted, as in BIT-7.3. Set parent to a track ID
to list that track's direct bars in the track's own order; omit it to list every task.`

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
		Use:   serveMCPCmdUse,
		Short: "Run the MCP server in the foreground",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			root := os.Getenv("CLAUDE_PROJECT_DIR")

			return runMCPServer(cmd.Context(), root, &mcp.StdioTransport{})
		},
	}
}

func runMCPServer(ctx context.Context, root string, transport mcp.Transport) error {
	s := mcp.NewServer(&mcp.Implementation{Name: "bp", Version: "1"}, nil)
	mcp.AddTool(s, &mcp.Tool{Name: taskReadTool, Description: "Read a task by ID"}, taskReadHandler(root))
	mcp.AddTool(s, &mcp.Tool{Name: taskListTool, Description: taskListDescription}, taskListHandler(root))

	return s.Run(ctx, transport)
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

func parentOf(id string) string {
	i := strings.LastIndex(id, ".")
	if i == -1 {
		return ""
	}

	return id[:i]
}
