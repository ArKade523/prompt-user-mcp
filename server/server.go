package server

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

type MCPServer struct {
	stdin  io.Reader
	stdout io.Writer
	stderr io.Writer
}

type MCPRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

type MCPResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *MCPError   `json:"error,omitempty"`
}

type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type UserInputRequest struct {
	Prompt  string `json:"prompt"`
	Timeout int    `json:"timeout,omitempty"`
}

type UserInputResult struct {
	Response string `json:"response"`
	Success  bool   `json:"success"`
}

func NewMCPServer() *MCPServer {
	return &MCPServer{
		stdin:  os.Stdin,
		stdout: os.Stdout,
		stderr: os.Stderr,
	}
}

func (s *MCPServer) SetIO(stdin io.Reader, stdout io.Writer, stderr io.Writer) {
	s.stdin = stdin
	s.stdout = stdout
	s.stderr = stderr
}

func (s *MCPServer) Start(ctx context.Context) error {
	scanner := bufio.NewScanner(s.stdin)

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var req MCPRequest
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			s.sendError(req.ID, -32700, "Parse error")
			continue
		}

		switch req.Method {
		case "initialize":
			s.handleInitialize(req)
		case "notifications/initialized":
			// No response needed for this notification
			continue
		case "capabilities/list":
			s.handleCapabilities(req)
		case "tools/list":
			s.handleToolsList(req)
		case "tools/call":
			s.handleToolCall(req, scanner)
		case "user_input":
			s.handleUserInput(req, scanner)
		default:
			s.sendError(req.ID, -32601, "Method not found")
		}
	}

	return scanner.Err()
}

func (s *MCPServer) handleInitialize(req MCPRequest) {
	result := map[string]interface{}{
		"protocolVersion": "2024-11-05",
		"capabilities": map[string]interface{}{
			"tools": map[string]interface{}{
				"listChanged": false,
			},
		},
		"serverInfo": map[string]interface{}{
			"name":    "prompt-mcp",
			"version": "1.0.0",
		},
	}

	s.sendResponse(req.ID, result)
}

func (s *MCPServer) handleCapabilities(req MCPRequest) {
	result := map[string]interface{}{
		"capabilities": map[string]interface{}{
			"tools": map[string]interface{}{
				"listChanged": false,
			},
		},
	}

	s.sendResponse(req.ID, result)
}

func (s *MCPServer) handleToolsList(req MCPRequest) {
	tools := []map[string]interface{}{
		{
			"name":        "user_input",
			"description": "Request input or approval from the user",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"prompt": map[string]interface{}{
						"type":        "string",
						"description": "The prompt to show to the user",
					},
					"timeout": map[string]interface{}{
						"type":        "integer",
						"description": "Optional timeout in seconds",
					},
				},
				"required": []string{"prompt"},
			},
		},
	}

	result := map[string]interface{}{
		"tools": tools,
	}

	s.sendResponse(req.ID, result)
}

func (s *MCPServer) handleToolCall(req MCPRequest, scanner *bufio.Scanner) {
	paramsBytes, err := json.Marshal(req.Params)
	if err != nil {
		s.sendError(req.ID, -32602, "Invalid params")
		return
	}

	var toolCall struct {
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments"`
	}
	if err := json.Unmarshal(paramsBytes, &toolCall); err != nil {
		s.sendError(req.ID, -32602, "Invalid params")
		return
	}

	switch toolCall.Name {
	case "user_input":
		s.handleUserInputTool(req, toolCall.Arguments, scanner)
	default:
		s.sendError(req.ID, -32601, "Unknown tool")
	}
}

func (s *MCPServer) handleUserInputTool(req MCPRequest, args map[string]interface{}, scanner *bufio.Scanner) {
	prompt, ok := args["prompt"].(string)
	if !ok {
		s.sendError(req.ID, -32602, "Missing or invalid prompt parameter")
		return
	}

	// Get user input from the controlling terminal, not from MCP stdin
	response, err := s.getUserInputFromTTY(prompt)
	if err != nil {
		s.sendError(req.ID, -32603, fmt.Sprintf("Failed to get user input: %v", err))
		return
	}

	result := map[string]interface{}{
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": response,
			},
		},
		"isError": false,
	}

	s.sendResponse(req.ID, result)
}

func (s *MCPServer) getUserInputFromTTY(prompt string) (string, error) {
	// Open the controlling terminal directly
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return "", fmt.Errorf("failed to open /dev/tty: %w", err)
	}
	defer tty.Close()

	// Write prompt to the terminal
	fmt.Fprintf(tty, "%s\n", prompt)
	fmt.Fprintf(tty, "Response: ")

	// Read response from the terminal
	scanner := bufio.NewScanner(tty)
	if scanner.Scan() {
		return strings.TrimSpace(scanner.Text()), nil
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("failed to read from terminal: %w", err)
	}

	return "", nil
}

func (s *MCPServer) handleUserInput(req MCPRequest, scanner *bufio.Scanner) {
	paramsBytes, err := json.Marshal(req.Params)
	if err != nil {
		s.sendError(req.ID, -32602, "Invalid params")
		return
	}

	var userReq UserInputRequest
	if err := json.Unmarshal(paramsBytes, &userReq); err != nil {
		s.sendError(req.ID, -32602, "Invalid params")
		return
	}

	// Get user input from the controlling terminal, not from MCP stdin
	response, err := s.getUserInputFromTTY(userReq.Prompt)
	if err != nil {
		result := UserInputResult{
			Response: "",
			Success:  false,
		}
		s.sendResponse(req.ID, result)
		return
	}

	result := UserInputResult{
		Response: response,
		Success:  true,
	}

	s.sendResponse(req.ID, result)
}

func (s *MCPServer) sendResponse(id interface{}, result interface{}) {
	resp := MCPResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}

	data, _ := json.Marshal(resp)
	fmt.Fprintf(s.stdout, "%s\n", data)
}

func (s *MCPServer) sendError(id interface{}, code int, message string) {
	resp := MCPResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &MCPError{
			Code:    code,
			Message: message,
		},
	}

	data, _ := json.Marshal(resp)
	fmt.Fprintf(s.stdout, "%s\n", data)
}
