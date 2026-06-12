package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func runMCPServer(state *AppState, logger *slog.Logger) {
	srv := server.NewMCPServer(
		"askd",
		"1.0.0",
	)

	tool := mcp.NewTool("ask_user_questions",
		mcp.WithDescription("Ask the user one or more questions and wait for answers. "+
			"Only one session can be pending at a time. "+
			"If a previous session is awaiting a response, the call is rejected. "+
			"allowCustom defaults to true — set it to false only when options are exhaustive."),
		mcp.WithInputSchema[AskQuestionsInput](),
	)

	srv.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := req.Params.Arguments.(map[string]interface{})
		if !ok {
			return mcp.NewToolResultError("invalid arguments"), nil
		}
		questionsRaw, ok := args["questions"]
		if !ok {
			return mcp.NewToolResultError("missing required parameter: questions"), nil
		}

		raw, err := json.Marshal(questionsRaw)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("processing questions: %v", err)), nil
		}

		var questions []Question
		if err := json.Unmarshal(raw, &questions); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("invalid questions format: %v", err)), nil
		}

		if len(questions) == 0 {
			return mcp.NewToolResultError("questions must not be empty"), nil
		}

		for i := range questions {
			if questions[i].AllowCustom == nil {
				trueVal := true
				questions[i].AllowCustom = &trueVal
			}
		}

		session := Session{
			ID:        uuid.New().String(),
			Questions: questions,
		}

		select {
		case state.sessionChan <- session:
		default:
			return mcp.NewToolResultError("previous session still pending"), nil
		}

		select {
		case answer := <-state.answerChan:
			result, err := json.Marshal(answer)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("processing answer: %v", err)), nil
			}
			return mcp.NewToolResultText(string(result)), nil
		case <-state.done:
			return mcp.NewToolResultError("application is shutting down"), nil
		case <-time.After(15 * time.Minute):
			select {
			case state.timeoutChan <- struct{}{}:
			default:
			}
			return mcp.NewToolResultError("timeout: user did not respond within 5 minutes"), nil
		}
	})

	if err := server.ServeStdio(srv); err != nil {
		logger.Error("MCP server error", "error", err)
	}

	close(state.done)
}
