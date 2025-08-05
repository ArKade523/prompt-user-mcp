# User input MCP server

This is an MCP server that is intended to increase interaction between a user and an agentic LLM tool by giving the LLM the option to ask for feedback without stopping its train of thought.

It has two primary use cases:

1.  Asking the user to test a feature and provide feedback.
    For example:
    - Asking the user to observe and test a UI. 
    - Asking the user to test a tool with a sensitive API key.
    - Asking the user to build and flash an embedded program and report behavior.

2. Waiting for approval from the user before continuing implementation.
    For example:
    - The user could require manual approval of each feature before continuing implementation. This is important for keeping development on track, instead of wasting time and resources on a slightly wrong application.
    - The user could ask for an additional feature set before continuing with a prior plan.
    - The user could ask the agent to use a different API before getting further entrenched in a specific ecosystem.
    - The user could ask the agent to switch to a different set of libraries, framework, or language.

## Installation

### Download Pre-built Binaries

Download the latest release for your platform from the [GitHub Releases](../../releases) page:

- **Windows**: `prompt-mcp-windows-amd64.exe` or `prompt-mcp-windows-arm64.exe`
- **Linux**: `prompt-mcp-linux-amd64` or `prompt-mcp-linux-arm64`  
- **macOS**: `prompt-mcp-darwin-amd64` or `prompt-mcp-darwin-arm64`
    - Given macOS's security constraints, you will have to 
        a. Make the executable runnable (`chmod +x <executable name>`)
        b. Allow the binary to run through settings > privacy > security
    - Basically, it might be a little easier to just build and run it yourself. up to you tho

Run the binary:

```bash
# Linux/macOS
./prompt-mcp-* serve

# Windows (PowerShell)
.\prompt-mcp-*.exe serve
```

### Build from Source

Requirements: Go 1.21+

```bash
git clone <repository-url>
cd prompt-mcp
make all
./bin/prompt-mcp serve
```

## Usage

The server supports two input methods:

### TTY Method (Terminal)
```bash
echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"user_input","arguments":{"prompt":"Enter your name:","method":"tty"}}}' | ./prompt-mcp serve
```

### Web Method (Browser)
```bash
echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"user_input","arguments":{"prompt":"Enter your name:","method":"web"}}}' | ./prompt-mcp serve
```

The web method automatically opens your browser to a simple input form and works well with Claude Code and other environments where stdin/stdout are redirected.

