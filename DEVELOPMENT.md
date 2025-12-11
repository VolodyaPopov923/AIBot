# Implementation & Development Guide

## Project Setup

### Prerequisites
- Go 1.21+
- OpenAI API key
- macOS, Linux, or Windows

### First-Time Setup

```bash
# 1. Navigate to project
cd /Users/vladimirpopov/GolandProjects/AIBot

# 2. Install dependencies
make install

# 3. Setup environment
cp .env.example .env
# Edit .env with your OpenAI API key

# 4. Build project
make build

# 5. Run
make run
```

## Code Structure Overview

### Entry Point: `cmd/agent/main.go`

```go
// Initialization
browserMgr  // Playwright browser automation
aiClient    // OpenAI API client
agentInstance // Main agent loop

// CLI Loop
for {
  read user input
  if "task": run ExecuteTask
  if "go": navigate to URL
}
```

**Commands Available**:
- `task https://example.com "Find login button"`
- `go https://example.com`
- `exit`

### Browser Layer: `internal/browser/manager.go`

**Public API**:
```go
// Create manager
mgr, err := NewManager(ctx)

// Navigate to URL
err := mgr.Navigate(ctx, "https://example.com")

// Get current page state
content, err := mgr.GetPageContent(ctx)
// Returns: PageContent with title, URL, elements, text

// Interact with elements
err := mgr.Click(ctx, "button:nth-of-type(1)")
err := mgr.Fill(ctx, "input[name='email']", "user@example.com")

// Cleanup
mgr.Close(ctx)
```

**Key Algorithm - Dynamic Selector Generation**:
```go
// For each interactive element, generate selector
// Priority: ID > Name > nth-of-type position

// Example: <button id="login">Login</button>
// Selector: #login

// Example: <input name="email">
// Selector: input[name="email"]

// Example: <button>Unknown</button>  (no id/name)
// Selector: button:nth-of-type(1)
```

### AI Integration: `internal/ai/client.go`

**Public API**:
```go
// Create AI client
aiClient := NewClient(apiKey)

// Analyze page and get decision
analysis, err := aiClient.GetAnalysis(ctx, pageContent, task)
// Returns: Analysis string from GPT-4

// Make decision (structured response)
decision, err := aiClient.MakeDecision(ctx, systemPrompt, userInput)
// Returns: DecisionResponse with action and reasoning
```

**Prompt Example**:
```
System: You are an intelligent web browser automation agent...

User: Current task: Find login button
      Current page: Title: "Example Login"
                   URL: https://example.com
                   Elements:
                   1. [button] Sign In (selector: button#login)
                   2. [link] Forgot Password (selector: a[href="/reset"])
      
      What should we do next?

Response: I see a login button. Let's click on it.
```

### Agent Core: `internal/agent/agent.go`

**Main Loop Algorithm**:
```go
func (a *Agent) ExecuteTask(ctx context.Context, task string, url string) error {
  // 1. Navigate to initial URL
  a.browserMgr.Navigate(ctx, url)
  
  // 2. Main loop
  for iteration := 0; iteration < 20; iteration++ {
    // 2a. Get current page
    pageContent := a.browserMgr.GetPageContent(ctx)
    
    // 2b. Analyze and decide
    decision := a.analyzeAndDecide(ctx, pageContent)
    
    // 2c. Check if complete
    if decision.IsComplete { return nil }
    
    // 2d. Execute action
    a.executeAction(ctx, decision)
    
    // 2e. Wait for page
    time.Sleep(1 * time.Second)
  }
  
  return errors.New("max iterations exceeded")
}
```

**Decision Execution**:
```go
switch decision.Action {
case "navigate":
  browserMgr.Navigate(ctx, decision.URL)
  
case "click":
  browserMgr.Click(ctx, decision.Selector)
  
case "fill":
  browserMgr.Fill(ctx, decision.Selector, decision.Text)
  
case "complete":
  return nil
}
```

### Context Management: `internal/context/manager.go`

**Token Tracking**:
```go
// Create manager (8K token limit, 20 message history)
contextMgr := NewContextManager(8000, 20)

// Check if we can add tokens
if !contextMgr.TokenCounter().CanAddTokens(500) {
  // Remove old messages
}

// Track usage
contextMgr.TokenCounter().Add(
  promtTokens,
  completionTokens,
)
```

**Token Estimation**:
```
Rough: 4 characters ≈ 1 token
Page content: ~2KB = 500 tokens
Prompt: ~500 chars = 125 tokens
Response: ~500 chars = 125 tokens
Total: ~750 tokens per iteration
```

### Security Layer: `internal/security/security.go`

**Usage**:
```go
secMgr := NewValidator()

// Check if destructive
if secMgr.IsDestructive("delete account") {
  // Ask for confirmation
  approved, _ := secMgr.RequestConfirmation(
    DestructiveAction{
      Type: "delete",
      Description: "Delete email account",
      Severity: "high",
    },
  )
  
  if !approved {
    return errors.New("action denied")
  }
}
```

**Destructive Keywords**:
```
Delete: delete, remove, destroy, clear, wipe
Payment: payment, purchase, checkout, pay, charge
Account: logout, sign out, disable, close
```

## Common Tasks & Examples

### Example 1: Search on GitHub

```
Input: task https://github.com "Search for Go AI projects"

Flow:
1. Navigate to github.com
2. Find search button → Click
3. Find search input → Fill with "Go AI"
4. Submit search
5. Wait for results
6. Task complete
```

### Example 2: Add Item to Cart

```
Input: task https://amazon.com "Find a Go book and add to cart"

Flow:
1. Navigate to amazon.com
2. Find search box → Fill with "Go book"
3. Search → Click
4. Wait for results
5. Click first result
6. Find "Add to Cart" button → Click
7. Task complete
```

### Example 3: Fill Contact Form

```
Input: task https://example.com "Submit contact form"

Flow:
1. Navigate to website
2. Find name field → Fill with name
3. Find email field → Fill with email
4. Find message field → Fill with message
5. Find submit button → Click
6. Task complete
```

## Debugging

### Enable Verbose Output
In `main.go`, the agent is created with `verbose=true`:
```go
agentInstance := agent.NewAgent(browserMgr, aiClient, true)
```

This prints:
- Current iteration number
- Page URL and element count
- AI reasoning
- Actions taken

### Check OpenAI Integration
```go
// Test AI client
aiClient := NewClient(os.Getenv("OPENAI_API_KEY"))
analysis, err := aiClient.GetAnalysis(
  ctx,
  "Test page content",
  "Test task",
)
if err != nil {
  log.Fatal("OpenAI failed:", err)
}
```

### Test Browser Connection
```go
// Test browser manager
mgr, err := NewManager(ctx)
if err != nil {
  log.Fatal("Browser failed:", err)
}
defer mgr.Close(ctx)

err = mgr.Navigate(ctx, "https://example.com")
if err != nil {
  log.Fatal("Navigation failed:", err)
}
```

## Running Tests

```bash
# Run all tests
make test

# Or manually
go test -v ./...

# Run specific package
go test -v ./internal/browser
go test -v ./internal/ai
go test -v ./internal/context
go test -v ./internal/security
go test -v ./pkg/utils
```

## Performance Optimization

### Reduce Token Usage
```go
// Current: Full page description
// Output: ~500 tokens

// Optimize: Only critical elements
// Output: ~200 tokens

// Implement in GetPageContent():
func (m *Manager) GetPageContent(ctx context.Context) PageContent {
  // Only return visible elements
  // Truncate long text
  // Remove redundant info
}
```

### Faster Page Loads
```go
// Current: Wait for "networkidle"
page.Goto(url, WaitUntil("networkidle"))

// Faster: Wait for "domcontentloaded"
page.Goto(url, WaitUntil("domcontentloaded"))
```

### Parallel Processing (Future)
```go
// Pre-fetch next likely page while current processes
// Run multiple agents on different tabs
// Cache page analysis results
```

## Extending the Agent

### Add Custom Tools
```go
// In internal/browser/manager.go
func (m *Manager) CustomAction(ctx context.Context, arg string) error {
  // Implement custom action
}
```

### Implement Sub-Agents
```go
// Example: Specialized agent for e-commerce
type EcommerceAgent struct {
  *Agent
  cart []Product
}

func (e *EcommerceAgent) SearchProduct(ctx context.Context, query string) {
  // Specialized search logic
}
```

### Add Advanced Analysis
```go
// OCR for images
image := screenshot.CaptureElement(selector)
text := ocr.Extract(image)

// Visual element detection
elements := vision.DetectButtons(screenshot)

// ML-based click prediction
likelihood := ml.PredictClickable(element)
```

## Troubleshooting

### Issue: "Browser won't launch"
```bash
# Solution: Reinstall Playwright
go run github.com/playwright-community/playwright-go/cmd/playwright@latest install chromium
```

### Issue: "OpenAI API 401 error"
```bash
# Check API key is set
echo $OPENAI_API_KEY

# Verify it's valid in .env
cat .env | grep OPENAI_API_KEY

# Test with curl
curl https://api.openai.com/v1/models \
  -H "Authorization: Bearer $OPENAI_API_KEY"
```

### Issue: "Element not found"
- Browser might not have waited for JavaScript
- Try adding explicit wait
- Check selector generation logic

### Issue: "Token limit exceeded"
- Task is too complex
- Break into smaller subtasks
- Reduce page content verbosity

## Next Steps

1. **Test with Real Websites**
   - Try example.com (simple)
   - Try github.com (moderate)
   - Try amazon.com (complex)

2. **Implement Advanced Features**
   - Screenshot-based reasoning
   - Sub-agent architecture
   - Better context compression

3. **Optimize for Production**
   - Add comprehensive logging
   - Implement caching
   - Add performance metrics

4. **Enhance Security**
   - Add domain whitelist
   - Implement action budgets
   - Add IP restrictions

## Resources

- [Playwright Go Documentation](https://playwright.dev/go/)
- [OpenAI API Documentation](https://platform.openai.com/docs)
- [Go Best Practices](https://golang.org/doc/effective_go)
