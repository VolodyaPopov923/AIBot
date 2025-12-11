# Architecture & Design Document

## Overview

AIBot is an autonomous AI agent that controls a web browser to complete complex multi-step tasks. It combines intelligent decision-making with browser automation, security checks, and efficient token management.

## Design Principles

### 1. **No Hardcoded Knowledge**
- ❌ No preset scripts or workflows
- ❌ No hardcoded CSS selectors
- ❌ No knowledge of specific websites
- ✅ Agent learns page structure dynamically
- ✅ Agent determines actions based on current state

### 2. **Transparency**
- Visible browser (non-headless) for full visibility
- Detailed logging of agent decisions
- User can see exactly what's happening
- Clear confirmation prompts for sensitive actions

### 3. **Token Efficiency**
- OpenAI API tokens are expensive
- Never send full HTML to the AI
- Extract only relevant interactive elements
- Implement context windows and message history limits

### 4. **Security First**
- Identify and block destructive actions
- Request explicit user confirmation
- Log all security-relevant decisions
- Prevent accidental data deletion or unauthorized payments

### 5. **Resilience**
- Graceful error handling
- Automatic recovery from failed actions
- Re-assessment of page state after failures
- Timeouts and iteration limits

## Component Architecture

### Browser Layer (`internal/browser/manager.go`)

**Purpose**: Abstraction over Playwright for browser automation

**Key Responsibilities**:
1. Launch and manage Chromium browser
2. Extract page structure without hardcoding
3. Generate dynamic CSS selectors
4. Perform interactions (click, fill, navigate)

**Algorithm for Element Detection**:
```
For each element type (button, link, input):
  1. Query all elements of that type
  2. Extract visible text/placeholder
  3. Generate selector:
     - If element has ID: use #id
     - If element has name: use [name="..."]
     - Otherwise: use nth-of-type position
  4. Return element metadata
```

**Example Output**:
```json
{
  "elements": [
    {
      "type": "button",
      "text": "Sign In",
      "selector": "button:nth-of-type(1)",
      "index": 0
    },
    {
      "type": "input",
      "text": "Email",
      "selector": "input[name='email']",
      "index": 0
    }
  ]
}
```

### AI Integration Layer (`internal/ai/client.go`)

**Purpose**: Wrap OpenAI API for decision-making

**Key Methods**:
- `MakeDecision()` - Request next action based on page state
- `GetAnalysis()` - Analyze page content and task progress

**Prompt Engineering Strategy**:
1. **System Prompt**: Define agent role and capabilities
2. **Current State**: Page title, URL, available elements
3. **Task Context**: User's request and progress so far
4. **Query**: What should the agent do next?

**Response Format** (JSON):
```json
{
  "action": "click|navigate|fill|wait|complete",
  "selector": "CSS selector or URL",
  "text": "Text to fill",
  "reasoning": "Why this action",
  "is_complete": false,
  "needs_confirm": false
}
```

### Agent Core (`internal/agent/agent.go`)

**Purpose**: Orchestrate the autonomous loop

**Main Algorithm**:
```
1. Navigate to initial URL
2. Wait for page to load
3. For iteration 1 to MAX_ITERATIONS:
   a. Extract current page content
   b. Ask AI: "What should we do next?"
   c. Check if task is complete
   d. If destructive action: request confirmation
   e. Execute action (click, fill, navigate, etc.)
   f. Wait for page to settle
   g. Return to step 3
4. Return success or failure
```

**Decision Flow**:
```
┌─────────────────────┐
│  Get Page Content   │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│  Ask AI Next Action │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐     Yes    ┌──────────────┐
│  Task Complete?     ├───────────▶│  Return OK   │
└──────────┬──────────┘            └──────────────┘
           │
          No
           │
           ▼
┌─────────────────────┐
│  Destructive?       │──Yes──┐
└──────────┬──────────┘       │
           │                  │
          No                  ▼
           │          ┌──────────────────┐
           │          │  Ask Confirmation │
           │          └────────┬─────────┘
           │                   │
           │                  No
           │                   │
           │         ┌─────────▼────────┐
           │         │  Return Error    │
           │         └──────────────────┘
           │                   
           ▼
┌──────────────────────┐
│  Execute Action      │
└──────────┬───────────┘
           │
           ▼
┌──────────────────────┐
│  Wait for Page       │
└──────────┬───────────┘
           │
           ▼
    [Loop back to top]
```

### Context Management (`internal/context/manager.go`)

**Purpose**: Manage tokens and conversation history

**Token Calculation**:
- Rough estimation: 4 characters ≈ 1 token
- More accurate: Use tokenizers library for production
- Default limits: 8000 tokens per task

**History Management**:
```
Message Queue (max 20 messages):
[System] ──────────────────────────
[User] "Find login button" ─────────
[Assistant] "I see a button..." ──
[User] "Click it" ───────────────
[Assistant] "Clicked, now at..." ──
...

When limit exceeded:
Remove oldest user message, keep recent & system
```

**Token Tracking**:
```go
Before each API call:
  1. Estimate tokens for request
  2. Check: remaining_tokens >= needed_tokens
  3. If not: Remove oldest messages
  4. Send request
  5. Track actual tokens used
  6. Fail if total exceeds max
```

### Security Layer (`internal/security/security.go`)

**Purpose**: Prevent accidental destructive actions

**Destructive Action Keywords**:
```
Payment-related:
  - payment, purchase, checkout, pay, order, buy, charge

Deletion-related:
  - delete, remove, destroy, clear, wipe

Account-related:
  - logout, sign out, close account, disable

Data-related:
  - reset, clear, format
```

**Confirmation Flow**:
```
AI decides: execute_action(payment)
  │
  ▼
Is_destructive? → Yes
  │
  ▼
┌─────────────────────────────────┐
│ ⚠️  SECURITY CONFIRMATION       │
│ Action: payment                 │
│ Description: Process order      │
│ Severity: high                  │
│                                 │
│ Do you want to proceed? (y/n)   │
└─────────────────────────────────┘
  │
  ├─ Yes ──▶ [Security Log] Proceed
  │
  └─ No ───▶ [Security Log] Denied & Return Error
```

## Data Flow

### Task Execution Flow

```
User Input
  │
  ▼
┌──────────────────────┐
│ CLI Interface        │
│ task <URL> <desc>    │
└──────────┬───────────┘
           │
           ▼
┌──────────────────────┐
│ Agent.ExecuteTask()  │
│ - Navigate to URL    │
│ - Start main loop    │
└──────────┬───────────┘
           │
           ▼
    ┌──────────────┐
    │ Main Loop    │ (max 20 iterations)
    └──────┬───────┘
           │
           ▼
┌──────────────────────┐
│ Browser.GetContent() │
│ - Extract elements   │
│ - Build description  │
└──────────┬───────────┘
           │
           ▼
┌──────────────────────┐
│ AI.GetAnalysis()     │
│ - Call OpenAI API    │
│ - Get next action    │
└──────────┬───────────┘
           │
           ▼
┌──────────────────────┐
│ Security.Validate()  │
│ - Check if safe      │
│ - Request confirm    │
└──────────┬───────────┘
           │
           ▼
┌──────────────────────┐
│ Agent.ExecuteAction()│
│ - Click/Fill/Nav     │
│ - Wait for page      │
└──────────┬───────────┘
           │
      [Loop if not complete]
      [Return if complete or error]
```

## Token Usage Optimization

### Strategy 1: Selective Element Extraction
```
Full HTML: ~50KB = ~12,500 tokens ❌ Too much!
Extracted Elements: ~2KB = ~500 tokens ✅ Good

Only send:
- Title & URL
- Element type, text, selector
- Main text excerpt (truncated)
```

### Strategy 2: Message History Compression
```
Old messages (not needed):
  Remove after 5+ iterations

Keep:
  - Current page state
  - Recent decisions & results
  - System prompt
  - Last 2-3 actions
```

### Strategy 3: Iterative Refinement
```
First prompt: Full exploration (consume more tokens)
Next prompts: Focused queries (consume fewer tokens)

Example:
1. "What elements are clickable?" (200 tokens)
2. "Click login area" (100 tokens)
3. "Fill password field" (100 tokens)
```

## Error Handling Strategy

### Non-Fatal Errors
```
Error: Click failed
Reason: Element not visible
Action: Continue to next iteration
  - AI re-analyzes page
  - Might try different element
  - Or navigate away and retry
```

### Fatal Errors
```
Error: Token limit exceeded
Reason: Task too complex
Action: Return error and exit
```

### Timeout Handling
```
Set max iterations = 20
If not complete after 20 attempts:
  Return "Could not complete task"
  Suggest breaking into smaller tasks
```

## Security Considerations

### 1. Destructive Action Detection
- Pattern matching on action keywords
- Case-insensitive checking
- Extensible keyword list

### 2. User Confirmation
```
⚠️ SECURITY CONFIRMATION REQUIRED
Action Type: payment (high severity)
Description: Process checkout order
Target: https://amazon.com/checkout

Do you want to proceed? (yes/no):
```

### 3. Action Logging
```
[SECURITY LOG] APPROVED - Type: payment, Desc: "Checkout order"
[SECURITY LOG] DENIED - Type: delete_account, Desc: "Delete email account"
```

### 4. Future Enhancements
- Whitelist/blacklist of domains
- Time-based action restrictions
- Budget limits for payments
- IP-based access controls

## Testing Strategy

### Unit Tests
```
✅ Browser element extraction
✅ Token counting
✅ Destructive action detection
✅ URL normalization
```

### Integration Tests (Future)
```
- Real browser navigation
- OpenAI API calls (mocked)
- End-to-end task execution
```

### Test Sites (Future)
```
- scrapingbee.com (practice site)
- example.com (minimal site)
- Custom test server
```

## Performance Considerations

### Bottlenecks
1. **Network latency**: Page loading time
   - Solution: Reduce page load waits with smart timing

2. **OpenAI API latency**: 1-5 seconds per request
   - Solution: Parallel requests or caching

3. **Token costs**: $0.01 per 1K prompt tokens
   - Solution: Aggressive context management

### Optimization Opportunities
```
Caching:
  - Store page summaries
  - Reuse selectors for known patterns
  
Parallel Processing:
  - Pre-fetch next page
  - Parallel AI requests
  
Smart Waiting:
  - Detect page load vs. wait for stability
  - Adaptive timing based on page complexity
```

## Future Roadmap

### Phase 1: MVP (Current)
- ✅ Browser automation
- ✅ OpenAI integration
- ✅ Security layer
- ✅ Token management

### Phase 2: Advanced
- [ ] Sub-agent architecture (specialized agents)
- [ ] Screenshot-based reasoning
- [ ] OCR for text in images
- [ ] Multi-page workflows
- [ ] State persistence

### Phase 3: Production
- [ ] Advanced context compression
- [ ] Webhook notifications
- [ ] Analytics dashboard
- [ ] User authentication
- [ ] Enterprise security

## Deployment Considerations

### Local Development
```bash
go run ./cmd/agent
```

### Docker Container
```dockerfile
FROM golang:1.21
RUN playwright install chromium
COPY . /app
WORKDIR /app
RUN go build -o aibot ./cmd/agent
CMD ["./aibot"]
```

### Cloud Deployment
- AWS Lambda (headless only - need to handle)
- GCP Cloud Run
- Kubernetes cluster
- With persistent storage for browser sessions
