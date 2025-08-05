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
	input := `{"jsonrpc":"2.0","id":2,"method":"user_input","params":{"prompt":"Test prompt"}}
test response`
	
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
	
	stderrOutput := stderr.String()
	if !strings.Contains(stderrOutput, "Test prompt") {
		t.Errorf("Expected stderr to contain prompt, got: %s", stderrOutput)
	}
	
	var response map[string]interface{}
	if err := json.Unmarshal(stdout.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
	
	result, ok := response["result"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected result to be an object")
	}
	
	if result["response"] != "test response" {
		t.Errorf("Expected response 'test response', got %v", result["response"])
	}
	
	if result["success"] != true {
		t.Errorf("Expected success true, got %v", result["success"])
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