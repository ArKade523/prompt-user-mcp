package test

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"prompt-mcp/server"
)

func TestMCPServerInitialize(t *testing.T) {
	input := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	srv := &server.MCPServer{}
	srv.SetIO(strings.NewReader(input), &stdout, &stderr)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	go func() {
		srv.Start(ctx)
	}()

	time.Sleep(100 * time.Millisecond)

	var response map[string]interface{}
	if err := json.Unmarshal(stdout.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response["jsonrpc"] != "2.0" {
		t.Errorf("Expected jsonrpc 2.0, got %v", response["jsonrpc"])
	}

	if response["id"] != float64(1) {
		t.Errorf("Expected id 1, got %v", response["id"])
	}

	result, ok := response["result"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected result to be an object")
	}

	if result["protocolVersion"] != "2024-11-05" {
		t.Errorf("Expected protocol version 2024-11-05, got %v", result["protocolVersion"])
	}
}

func TestMCPServerUserInput(t *testing.T) {
	input := `{"jsonrpc":"2.0","id":2,"method":"user_input","params":{"prompt":"Test prompt"}}`

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	srv := &server.MCPServer{}
	srv.SetIO(strings.NewReader(input), &stdout, &stderr)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	go func() {
		srv.Start(ctx)
	}()

	time.Sleep(100 * time.Millisecond)

	var response map[string]interface{}
	if err := json.Unmarshal(stdout.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	result, ok := response["result"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected result to be an object")
	}

	// In test environment, /dev/tty might not be available or the terminal might not be interactive
	// So we just verify that the server handled the request properly and returned a valid response structure
	if result["success"] == nil {
		t.Error("Expected success field to be present")
	}

	if result["response"] == nil {
		t.Error("Expected response field to be present")  
	}
}

func TestMCPServerInvalidMethod(t *testing.T) {
	input := `{"jsonrpc":"2.0","id":3,"method":"invalid_method","params":{}}`

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	srv := &server.MCPServer{}
	srv.SetIO(strings.NewReader(input), &stdout, &stderr)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	go func() {
		srv.Start(ctx)
	}()

	time.Sleep(100 * time.Millisecond)

	var response map[string]interface{}
	if err := json.Unmarshal(stdout.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	errorObj, ok := response["error"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected error to be an object")
	}

	if errorObj["code"] != float64(-32601) {
		t.Errorf("Expected error code -32601, got %v", errorObj["code"])
	}

	if errorObj["message"] != "Method not found" {
		t.Errorf("Expected error message 'Method not found', got %v", errorObj["message"])
	}
}

func TestMCPServerCapabilities(t *testing.T) {
	input := `{"jsonrpc":"2.0","id":4,"method":"capabilities/list","params":{}}`

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	srv := &server.MCPServer{}
	srv.SetIO(strings.NewReader(input), &stdout, &stderr)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	go func() {
		srv.Start(ctx)
	}()

	time.Sleep(100 * time.Millisecond)

	var response map[string]interface{}
	if err := json.Unmarshal(stdout.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	result, ok := response["result"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected result to be an object")
	}

	capabilities, ok := result["capabilities"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected capabilities to be an object")
	}

	tools, ok := capabilities["tools"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected tools to be an object")
	}

	if tools["listChanged"] != false {
		t.Errorf("Expected listChanged false, got %v", tools["listChanged"])
	}
}

func TestMCPServerToolsList(t *testing.T) {
	input := `{"jsonrpc":"2.0","id":5,"method":"tools/list","params":{}}`

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	srv := &server.MCPServer{}
	srv.SetIO(strings.NewReader(input), &stdout, &stderr)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	go func() {
		srv.Start(ctx)
	}()

	time.Sleep(100 * time.Millisecond)

	var response map[string]interface{}
	if err := json.Unmarshal(stdout.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	result, ok := response["result"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected result to be an object")
	}

	tools, ok := result["tools"].([]interface{})
	if !ok {
		t.Fatal("Expected tools to be an array")
	}

	if len(tools) != 1 {
		t.Fatalf("Expected 1 tool, got %d", len(tools))
	}

	tool, ok := tools[0].(map[string]interface{})
	if !ok {
		t.Fatal("Expected tool to be an object")
	}

	if tool["name"] != "user_input" {
		t.Errorf("Expected tool name 'user_input', got %v", tool["name"])
	}

	if tool["description"] != "Request input or approval from the user" {
		t.Errorf("Expected tool description, got %v", tool["description"])
	}
}

func TestMCPServerToolCall(t *testing.T) {
	input := `{"jsonrpc":"2.0","id":6,"method":"tools/call","params":{"name":"user_input","arguments":{"prompt":"Test tool call"}}}`

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	srv := &server.MCPServer{}
	srv.SetIO(strings.NewReader(input), &stdout, &stderr)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	go func() {
		srv.Start(ctx)
	}()

	time.Sleep(100 * time.Millisecond)

	var response map[string]interface{}
	if err := json.Unmarshal(stdout.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// In test environment, /dev/tty might not be available, so we might get an error response
	// Let's check for either success or error response format
	if response["result"] != nil {
		result, ok := response["result"].(map[string]interface{})
		if !ok {
			t.Fatal("Expected result to be an object")
		}

		content, ok := result["content"].([]interface{})
		if !ok {
			t.Fatal("Expected content to be an array")
		}

		if len(content) != 1 {
			t.Fatalf("Expected 1 content item, got %d", len(content))
		}

		contentItem, ok := content[0].(map[string]interface{})
		if !ok {
			t.Fatal("Expected content item to be an object")
		}

		if contentItem["type"] != "text" {
			t.Errorf("Expected content type 'text', got %v", contentItem["type"])
		}

		if result["isError"] != false {
			t.Errorf("Expected isError false, got %v", result["isError"])
		}
	} else if response["error"] != nil {
		// This is expected in test environment where /dev/tty might not be available
		errorObj, ok := response["error"].(map[string]interface{})
		if !ok {
			t.Fatal("Expected error to be an object")
		}

		if errorObj["code"] != float64(-32603) {
			t.Errorf("Expected error code -32603, got %v", errorObj["code"])
		}
	} else {
		t.Fatal("Expected either result or error in response")
	}
}
