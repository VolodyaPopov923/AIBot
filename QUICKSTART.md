# Quick Start Guide

## ⚡ 5-Minute Setup

### Step 1: Prerequisites Check
```bash
# Check Go version (need 1.21+)
go version

# You should have: go version go1.21+ ...
```

### Step 2: Clone & Navigate
```bash
cd /Users/vladimirpopov/GolandProjects/AIBot
```

### Step 3: Install Dependencies
```bash
make install
```

This will:
- Download Go dependencies
- Install Playwright browsers (Chromium)
- Set up the environment

### Step 4: Configure OpenAI
```bash
cp .env.example .env

# Open .env and add your OpenAI API key
# Get one from: https://platform.openai.com/api-keys
```

Edit `.env`:
```env
OPENAI_API_KEY=sk-your-key-here
BROWSER_PATH=/Applications/Chromium.app/Contents/MacOS/Chromium  # macOS
DEBUG=false
```

### Step 5: Build & Run
```bash
# Build the project
make build

# Run the agent
make run
```

Or directly:
```bash
go run ./cmd/agent
```

## 🎯 First Task

Once the agent is running, try these commands:

### Simple Test
```
> go https://example.com
```
This just navigates - you should see Chromium open and go to example.com.

### Find an Element
```
> task https://example.com "Find the 'More information' link and click it"
```

### GitHub Search
```
> task https://github.com "Search for Go projects"
```

### Real-World Example
```
> task https://news.ycombinator.com "Find and click the newest story"
```

## 📁 Project Structure

```
AIBot/
├── cmd/agent/main.go           ← Entry point (interactive CLI)
├── internal/
│   ├── agent/agent.go          ← Main autonomous loop
│   ├── ai/client.go            ← OpenAI integration
│   ├── browser/manager.go      ← Playwright automation
│   ├── context/manager.go      ← Token management
│   └── security/security.go    ← Confirmation checks
├── pkg/utils/strings.go        ← Helper functions
├── go.mod                       ← Dependencies
├── .env.example                ← Config template
├── README.md                    ← Full documentation
├── ARCHITECTURE.md             ← Design deep-dive
└── DEVELOPMENT.md              ← Development guide
```

## 🔧 Common Commands

```bash
# Development
make build          # Build executable
make run            # Build and run
make test           # Run unit tests
make fmt            # Format code
make clean          # Clean build artifacts

# Direct commands
go run ./cmd/agent              # Run directly
go test -v ./...                # Run tests
go mod tidy                     # Clean dependencies
```

## 🐛 Troubleshooting

### "Playwright installation failed"
```bash
# Reinstall Playwright
go run github.com/playwright-community/playwright-go/cmd/playwright@latest install
```

### "OpenAI API key error"
```bash
# Verify the key exists and is valid
echo $OPENAI_API_KEY

# Make sure it starts with "sk-"
# Get new key at: https://platform.openai.com/api-keys
```

### "Browser won't open"
- Make sure you have Chrome or Chromium installed
- On macOS, it should auto-detect Chromium
- Check the browser is not already running

### "Token limit exceeded"
- The task was too complex
- Try breaking it into smaller tasks
- The agent will track token usage and inform you

## 📚 What to Read Next

1. **README.md** - Overview and features
2. **ARCHITECTURE.md** - How everything works
3. **DEVELOPMENT.md** - Implementation details

## 🚀 Key Features

✅ **Autonomous** - Solves tasks without step-by-step guidance
✅ **Transparent** - You see the browser and AI decisions
✅ **Secure** - Asks before deleting or paying
✅ **Efficient** - Manages OpenAI token usage
✅ **Smart** - Dynamically detects elements (no hardcoding)

## 🎓 How It Works

```
1. You: "Find and click the login button"
   ↓
2. Agent: Navigates to page, extracts interactive elements
   ↓
3. AI: Analyzes page, decides "click login button"
   ↓
4. Agent: Executes action, waits for page
   ↓
5. Repeat until task complete
```

## 📊 Example Session

```
> task https://example.com "Click the 'More information' link"

🚀 Initializing browser...
🤖 Initializing AI client...

=== Iteration 1 ===
Current URL: https://example.com
Found 5 interactive elements
Decision: I can see the page content. There's a link labeled 'More information'.

✅ Navigation successful!

=== Complete ===
```

## ⚠️ Important Notes

- The browser will be **visible** (not headless) - you'll see it working
- Each task costs OpenAI tokens (roughly $0.001-0.01 per task)
- Some websites might block automation (that's OK - task will fail gracefully)
- Tasks should be in English
- Agent has max 20 iterations per task

## 🤔 Common Questions

**Q: Can it handle complex websites?**
A: Yes! The agent adapts to any website structure dynamically.

**Q: Will it work without hardcoded selectors?**
A: Yes! That's the whole point - it learns the page on the fly.

**Q: Can I make it do risky actions?**
A: No - for deletion/payment, it asks for confirmation.

**Q: How much does it cost?**
A: Depends on task complexity. Rough: $0.001-0.01 per task.

**Q: Can it handle logins?**
A: Yes! Browser maintains cookies/sessions between tasks.

**Q: What if a task fails?**
A: It tries to recover. If impossible, returns error with explanation.

## 🎯 Next Steps

1. ✅ Run a simple task: `task https://example.com "Click something"`
2. ✅ Read ARCHITECTURE.md to understand how it works
3. ✅ Try complex tasks on different websites
4. ✅ Explore the code in `internal/` folders
5. ✅ Consider implementing your own features!

## 💡 Project Ideas

- [ ] Add screenshot-based reasoning
- [ ] Create sub-agents for specialized tasks
- [ ] Implement OCR for image text
- [ ] Add webhook notifications
- [ ] Create a web UI for task submission
- [ ] Build a database of completed tasks

---

**Happy automating! 🤖**

For more help: See README.md, ARCHITECTURE.md, or DEVELOPMENT.md
