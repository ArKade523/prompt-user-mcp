package server

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"
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

type WebInputHandler struct {
	prompt     string
	response   chan string
	serverDone chan struct{}
	mu         sync.Mutex
	server     *http.Server
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
					"method": map[string]interface{}{
						"type":        "string",
						"description": "Input method: 'tty' (terminal) or 'web' (browser)",
						"enum":        []string{"tty", "web"},
						"default":     "tty",
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

	// Get input method, default to TTY
	method := "tty"
	if methodArg, exists := args["method"]; exists {
		if methodStr, ok := methodArg.(string); ok {
			method = methodStr
		}
	}

	var response string
	var err error

	switch method {
	case "web":
		response, err = s.getUserInputFromWeb(prompt)
	case "tty":
		fallthrough
	default:
		response, err = s.getUserInputFromTTY(prompt)
	}

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

func (s *MCPServer) getUserInputFromWeb(prompt string) (string, error) {
	handler := &WebInputHandler{
		prompt:     prompt,
		response:   make(chan string, 1),
		serverDone: make(chan struct{}, 1),
	}

	// Find an available port
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return "", fmt.Errorf("failed to find available port: %w", err)
	}

	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close() // Close so we can use the port for HTTP server

	mux := http.NewServeMux()
	mux.HandleFunc("/", handler.handleRoot)
	mux.HandleFunc("/submit", handler.handleSubmit)

	handler.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	// Start server in background
	go func() {
		if err := handler.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "Web server error: %v\n", err)
		}
		handler.serverDone <- struct{}{}
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	url := fmt.Sprintf("http://localhost:%d", port)

	// Open browser
	if err := openBrowser(url); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open browser automatically. Please visit: %s\n", url)
	} else {
		fmt.Fprintf(os.Stderr, "Opening browser for input: %s\n", url)
	}

	// Wait for response or timeout
	select {
	case response := <-handler.response:
		handler.shutdown()
		return response, nil
	case <-time.After(5 * time.Minute): // 5 minute timeout
		handler.shutdown()
		return "", fmt.Errorf("timeout waiting for web input")
	}
}

func (h *WebInputHandler) handleRoot(w http.ResponseWriter, r *http.Request) {
	tmpl := `<!DOCTYPE html>
<html>
<head>
    <title>User Input Required</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 600px; margin: 50px auto; padding: 20px; }
        .prompt { background: #f5f5f5; padding: 15px; border-left: 4px solid #007cba; margin: 20px 0; }
        input[type="text"] { width: 100%; padding: 10px; font-size: 16px; border: 1px solid #ddd; }
        button { background: #007cba; color: white; padding: 10px 20px; border: none; font-size: 16px; cursor: pointer; }
        button:hover { background: #005a87; }
    </style>
</head>
<body>
    <h1>User Input Required</h1>
    <div class="prompt">{{.Prompt}}</div>
    <form action="/submit" method="post">
        <input type="text" name="response" placeholder="Enter your response..." autofocus required>
        <br><br>
        <button type="submit">Submit</button>
    </form>
    <script>
        document.querySelector('form').addEventListener('submit', function() {
            document.querySelector('button').textContent = 'Submitting...';
            document.querySelector('button').disabled = true;
        });
    </script>
</body>
</html>`

	t, err := template.New("input").Parse(tmpl)
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	data := struct{ Prompt string }{Prompt: h.prompt}
	if err := t.Execute(w, data); err != nil {
		http.Error(w, "Template execution error", http.StatusInternalServerError)
		return
	}
}

func (h *WebInputHandler) handleSubmit(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := r.FormValue("response")
	if response == "" {
		http.Error(w, "Response cannot be empty", http.StatusBadRequest)
		return
	}

	// Send response
	select {
	case h.response <- response:
		fmt.Fprintf(w, "<html><body><h1>Thank you!</h1><p>Your response has been submitted. You can close this tab.</p></body></html>")
	default:
		http.Error(w, "Response already submitted", http.StatusBadRequest)
	}
}

func (h *WebInputHandler) shutdown() {
	if h.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		h.server.Shutdown(ctx)
	}
}

func openBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)
	return exec.Command(cmd, args...).Start()
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
