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
				"user_input": map[string]interface{}{
					"description": "Request input or approval from the user",
				},
			},
		},
		"serverInfo": map[string]interface{}{
			"name":    "prompt-mcp",
			"version": "1.0.0",
		},
	}
	
	s.sendResponse(req.ID, result)
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
	
	fmt.Fprintf(s.stderr, "%s\n", userReq.Prompt)
	fmt.Fprintf(s.stderr, "Response: ")
	
	var response string
	if scanner.Scan() {
		response = strings.TrimSpace(scanner.Text())
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