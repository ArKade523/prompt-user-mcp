You are building the project detailed in README.md. Follow these implementation instructions:

- Use Go for implementation.
    - Create a cli entry point in cli/.
    - Create separate entry points in other locations.
- Prioritize creating an intuitive CLI experience.
    - Give good error messages.
    - Create flags with a `--` prefix and a shorthand with a `-` prefix.
- Create regression tests for every feature as they are implemented. 
    - Store tests in test/.
- Use git rigorously.
    - Use feature branches as you work to keep the project organized.
    - Typecheck must pass before creating a commit.
    - All tests must pass before creating a commit.
    - When adding files to a commit, specify EVERY file you intend to add.
    - Commit messages MUST be 1 sentence long.
- Use a Makefile for building, cleaning, and testing.
    - Create scripts as you go.
    - For example, create and use `make clean`, `make test`, and `make all`.
- Take notes in CLAUDE.md about implementation details.
    - For example, take notes on:
        - The data structures you used.
        - The libraries you used.
        - The assumptions you made.
        - The features you implemented and the ones that you pushed off for later.

NOTES:

## MCP Server Implementation

### Data Structures
- `MCPServer`: Main server struct with stdin/stdout/stderr for MCP communication
- `MCPRequest`/`MCPResponse`: JSON-RPC message structures following MCP protocol 2024-11-05
- `UserInputRequest`/`UserInputResult`: Legacy user input structures (maintained for backwards compatibility)

### Libraries Used  
- Standard Go libraries: `encoding/json`, `bufio`, `context`, `os`, `fmt`, `io`, `strings`
- No external dependencies to keep the server lightweight

### Key Implementation Details

#### Terminal Access Solution
**Problem**: MCP protocol uses stdin/stdout for JSON-RPC communication, but user input tools also need to read from user. This created an infinite loop where MCP messages and user input conflicted.

**Solution**: Use `/dev/tty` to access the controlling terminal directly, bypassing the redirected stdin used for MCP communication. This allows:
- MCP JSON-RPC messages to flow through stdin/stdout 
- User input prompts and responses to go through the actual terminal (`/dev/tty`)
- No conflict between protocol communication and user interaction

#### MCP Protocol Compliance
Implemented full MCP 2024-11-05 specification support:
- `initialize` - Server capability negotiation
- `notifications/initialized` - Post-initialization notification handling
- `capabilities/list` - Server capability discovery 
- `tools/list` - Tool enumeration with JSON schema
- `tools/call` - Tool execution with proper error handling

#### User Input Tool
- **Name**: `user_input`
- **Purpose**: Allow LLM agents to request user input/approval without breaking their execution flow
- **Schema**: 
  - Required: `prompt` string parameter
  - Optional: `timeout` integer, `method` string (`"tty"` or `"web"`, defaults to `"tty"`)
- **Input Methods**:
  - `"tty"`: Direct terminal access via `/dev/tty` (works when run directly from terminal)
  - `"web"`: Opens browser tab with input form (works with Claude Code and other redirected environments)
- **Response**: Returns user's text response in MCP content format
- **Error Handling**: Graceful fallback if chosen input method fails

### Features Implemented
✅ Full MCP server protocol compliance
✅ JSON-RPC message handling  
✅ Dual input methods: TTY (terminal) and web (browser)
✅ Terminal-based user input via `/dev/tty` (solves stdin conflict)
✅ Web-based user input with automatic browser opening
✅ Tool schema definitions with method parameter
✅ Comprehensive test coverage
✅ Backwards compatibility with legacy `user_input` method

### Assumptions Made
- `/dev/tty` is available on target Unix systems (macOS/Linux) for TTY method
- Web browser is available and accessible for web method
- Default browser can be opened programmatically on target platform
- Claude Code or other MCP clients handle JSON-RPC communication properly

### Future Enhancements (Deferred)
- Windows TTY support (would need different approach than `/dev/tty`)
- Enhanced timeout handling for user input requests
- Rich prompting with styled output in web interface
- Multiple input types (confirmation dialogs, choice menus, file uploads)
- Secure HTTPS option for web method
- Custom port configuration for web server

