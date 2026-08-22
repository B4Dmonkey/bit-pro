package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/B4Dmonkey/bit-pro/task"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"
)

const (
	serveMCPCmdUse = "mcp"
	taskReadTool   = "task_read"
)

type taskReadInput struct {
	ID string `json:"id"`
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

	return s.Run(ctx, transport)
}

func taskReadHandler(root string) mcp.ToolHandlerFor[taskReadInput, taskReadOutput] {
	return func(
		_ context.Context,
		_ *mcp.CallToolRequest,
		in taskReadInput,
	) (*mcp.CallToolResult, taskReadOutput, error) {
		store := task.New(filepath.Join(root, ".bit"))

		t, err := store.Load(in.ID)
		if err != nil {
			return nil, taskReadOutput{}, fmt.Errorf("loading task %s: %w", in.ID, err)
		}

		parent := ""
		if i := strings.LastIndex(t.ID, "."); i != -1 {
			parent = t.ID[:i]
		}

		return nil, taskReadOutput{
			ID:         t.ID,
			Title:      t.Title,
			Status:     t.Status,
			Approved:   t.Approved,
			Phase:      t.Phase,
			PhaseLabel: t.PhaseLabel,
			Parent:     parent,
			Body:       t.Body,
		}, nil
	}
}
