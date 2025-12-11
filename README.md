# AI Browser Automation Agent

Autonomous AI agent that controls a web browser to complete complex multi-step tasks. Built in Go with OpenAI integration and Playwright browser automation.

## Features

- 🤖 **Autonomous AI Control**: Uses OpenAI GPT-4 to make intelligent decisions
- 🌐 **Visible Browser**: Non-headless Chromium browser for transparency
- 💾 **Session Persistence**: Maintains browser context and cookies
- 🔐 **Security Layer**: Asks for confirmation on destructive actions
- 📊 **Token Management**: Handles OpenAI token limits efficiently
- 🎯 **Dynamic Element Detection**: No hardcoded selectors - agent learns elements on the fly
- 🛡️ **Error Recovery**: Gracefully handles failed actions and adapts

## Architecture

```
AIBot/
├── cmd/
│   └── agent/              # Entry point
│       └── main.go         # CLI interface
├── internal/
│   ├── agent/              # Core agent loop
│   │   └── agent.go        # Main agent logic
│   ├── ai/                 # OpenAI integration
│   │   └── client.go       # AI client wrapper
│   ├── browser/            # Browser automation (Playwright)
│   │   └── manager.go      # Browser control
│   ├── context/            # Token & context management
│   │   └── manager.go      # Context and token tracking
│   └── security/           # Security & confirmation layer
│       └── security.go     # Destructive action validation
├── pkg/
│   └── utils/              # Shared utilities
│       └── strings.go      # String helpers
└── config/
    └── config.go           # Configuration management
```

## Tech Stack

- **Language**: Go 1.21+
- **Browser Automation**: Playwright Go (Chromium)
- **AI Model**: OpenAI GPT-4 Turbo
- **CLI**: Interactive terminal interface
- **Environment**: dotenv for configuration

## Quick Start

### Prerequisites

- Go 1.21 or higher
- OpenAI API key
- macOS, Linux, or Windows

### Installation

1. Navigate to project directory:
```bash
cd /Users/vladimirpopov/GolandProjects/AIBot
```

2. Install Go dependencies:
```bash
make install
```

Or manually:
```bash
go mod download
go mod tidy
go run github.com/playwright-community/playwright-go/cmd/playwright@latest install chromium
```

3. Configure environment:
```bash
cp .env.example .env
# Edit .env and add your OPENAI_API_KEY
```

4. Build and run:
```bash
make build
make run
```

Or directly:
```bash
go run ./cmd/agent
```

## Usage

### Interactive CLI

```
> task <URL> <description>  - Execute an autonomous task
> go <URL>                   - Navigate to a URL
> exit                       - Exit the program
```

### Example Tasks

```
> task https://example.com "Find and click the login button"
> task https://github.com "Search for repositories about machine learning"
> task https://amazon.com "Search for Go books and add one to cart"
> task https://mail.google.com "Check unread emails"
```

## Key Components

### 1. Browser Manager (`internal/browser/manager.go`)
**Responsibilities:**
- Non-headless Chromium browser control via Playwright
- Dynamic element extraction and analysis
- CSS selector generation without hardcoding
- Page content extraction with interactive element identification
- Navigation and form field management

**Key Methods:**
- `Navigate()` - Go to a URL
- `GetPageContent()` - Extract page structure and elements
- `Click()` / `Fill()` - Interact with elements
- `extractElements()` - Find buttons, links, inputs dynamically
- `getSelector()` - Generate CSS selectors on-the-fly

### 2. AI Client (`internal/ai/client.go`)
**Responsibilities:**
- OpenAI API integration
- Decision making through GPT-4 Turbo
- Page analysis and understanding

**Key Methods:**
- `MakeDecision()` - Ask AI what action to take
- `GetAnalysis()` - Request page content analysis

### 3. Agent (`internal/agent/agent.go`)
**Responsibilities:**
- Main autonomous loop (up to 20 iterations max)
- Task execution orchestration
- Decision analysis and action execution
- Error handling and recovery
- Security check enforcement

**Algorithm:**
1. Navigate to initial URL
2. For each iteration (max 20):
   - Extract current page content
   - Ask AI for next action
   - Check if task is complete
   - Execute action with security validation
   - Wait for page to settle

### 4. Context Manager (`internal/context/manager.go`)
**Responsibilities:**
- Token counting and limit enforcement
- Conversation history management
- Memory efficiency

**Features:**
- 8000 token limit per task
- 20 message history window
- Automatic old message pruning

### 5. Security Layer (`internal/security/security.go`)
**Responsibilities:**
- Identify destructive operations (delete, payment, logout, etc.)
- Request user confirmation before execution
- Log all security events

**Destructive Keywords:**
- `delete`, `remove`, `destroy`
- `payment`, `purchase`, `checkout`
- `logout`, `sign out`
- `clear`, `reset`, `disable`, `close account`

## Implementation Strategy

### Token Management
- **Estimation**: ~4 characters = 1 token (GPT-4 approximation)
- **Truncation**: Old messages removed when approaching limits
- **Optimization**: Concise page descriptions instead of full HTML

### Element Detection
```
Priority:
1. Use element ID (#id)
2. Use element name attribute (input[name="..."])
3. Generate position-based selector (nth-of-type)
No hardcoded selectors - purely dynamic discovery
```

### Page Content Extraction
Only sends to AI:
- Page title and URL
- Interactive elements (buttons, links, inputs)
- Element text and selectors
- Main body text (truncated if needed)

### Error Recovery
- Logs failed actions
- Continues to next iteration
- Agent re-assesses page state

## Project Capabilities

The agent will intelligently:
- ✅ Determine what elements are clickable
- ✅ Identify form fields and what to enter
- ✅ Navigate between pages based on content
- ✅ Handle dynamic pages and popups
- ✅ Avoid destructive actions without confirmation
- ✅ Manage token usage within limits
- ✅ Recover from failures

The agent will NOT:
- ❌ Use hardcoded selectors
- ❌ Follow preset scripts
- ❌ Get hints about page structure
- ❌ Know URLs in advance
- ❌ Know button labels in advance

## Makefile Commands

```bash
make install    # Install dependencies and Playwright browsers
make build      # Build the executable
make run        # Build and run
make test       # Run tests
make clean      # Remove build artifacts
make fmt        # Format code
make lint       # Run linter
```

## Environment Variables

```env
OPENAI_API_KEY    - Your OpenAI API key (required)
BROWSER_PATH      - Path to Chromium (auto-detected)
DEBUG             - Enable debug logging (true/false)
```

## Future Enhancements

- [ ] Sub-agent architecture for specialized workflows
- [ ] Screenshot-based reasoning for complex UIs
- [ ] OCR for image text detection
- [ ] Advanced context compression
- [ ] Multi-page workflow state management
- [ ] External API integration
- [ ] Performance metrics and analytics
- [ ] Webhook notifications for task completion

## Troubleshooting

### Browser won't launch
```bash
# Install Playwright browsers
go run github.com/playwright-community/playwright-go/cmd/playwright@latest install
```

### OpenAI API errors
- Verify `OPENAI_API_KEY` is set correctly
- Check OpenAI account has available credits
- Verify API key has sufficient permissions

### Token limit exceeded
- Tasks too complex for current context window
- Reduce initial page content size
- Break task into smaller subtasks

## License

MIT

## Author

Vladimir Popov (@VolodyaPopov923)
