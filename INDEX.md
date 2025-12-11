# AIBot - Complete Project Index

## 📋 Quick Navigation

### 🚀 Getting Started
1. **[QUICKSTART.md](QUICKSTART.md)** - 5-minute setup guide
2. **[README.md](README.md)** - Full overview and features
3. **[Makefile](Makefile)** - Build and run commands

### 📚 Documentation
1. **[PROJECT_SUMMARY.md](PROJECT_SUMMARY.md)** - Complete project overview
2. **[ARCHITECTURE.md](ARCHITECTURE.md)** - Design deep-dive and algorithms
3. **[DEVELOPMENT.md](DEVELOPMENT.md)** - Development guide and examples
4. **[INDEX.md](INDEX.md)** - This file

### 💻 Source Code

#### Entry Point
- **[cmd/agent/main.go](cmd/agent/main.go)** - CLI interface and main loop

#### Core Components
- **[internal/browser/manager.go](internal/browser/manager.go)** - Playwright automation
- **[internal/ai/client.go](internal/ai/client.go)** - OpenAI integration
- **[internal/agent/agent.go](internal/agent/agent.go)** - Main autonomous loop
- **[internal/context/manager.go](internal/context/manager.go)** - Token management
- **[internal/security/security.go](internal/security/security.go)** - Security layer

#### Utilities
- **[pkg/utils/strings.go](pkg/utils/strings.go)** - String helpers
- **[config/config.go](config/config.go)** - Configuration loading

#### Tests
- **[internal/browser/manager_test.go](internal/browser/manager_test.go)**
- **[internal/ai/client_test.go](internal/ai/client_test.go)** *(placeholder)*
- **[internal/context/manager_test.go](internal/context/manager_test.go)**
- **[internal/security/security_test.go](internal/security/security_test.go)**
- **[pkg/utils/strings_test.go](pkg/utils/strings_test.go)**

### ⚙️ Configuration
- **[go.mod](go.mod)** - Go dependencies
- **[.env.example](.env.example)** - Environment template
- **[.gitignore](.gitignore)** - Git exclusions

---

## 📖 Reading Guide

### For First-Time Users
```
1. Start here: QUICKSTART.md
2. Run: make install && make run
3. Try a task: task https://example.com "Click something"
4. Read: README.md for overview
```

### For Developers
```
1. Read: PROJECT_SUMMARY.md (structure)
2. Read: ARCHITECTURE.md (design)
3. Read: DEVELOPMENT.md (implementation)
4. Explore: internal/ source code
5. Run tests: make test
```

### For Architecture Review
```
1. Read: ARCHITECTURE.md
2. Review: internal/agent/agent.go (main loop)
3. Review: internal/browser/manager.go (element detection)
4. Review: internal/ai/client.go (AI integration)
5. Review: internal/security/security.go (safety layer)
```

---

## 🗂️ Directory Structure

```
AIBot/
├── cmd/agent/                 # Executable entry point
│   └── main.go               # CLI interface (130 lines)
│
├── internal/                  # Private packages
│   ├── agent/                # Autonomous loop
│   │   └── agent.go          # Main orchestration (200 lines)
│   ├── ai/                   # AI integration
│   │   └── client.go         # OpenAI wrapper (80 lines)
│   ├── browser/              # Browser automation
│   │   ├── manager.go        # Playwright control (250 lines)
│   │   └── manager_test.go   # Browser tests (30 lines)
│   ├── context/              # Token management
│   │   ├── manager.go        # Context manager (100 lines)
│   │   └── manager_test.go   # Context tests (40 lines)
│   └── security/             # Security layer
│       ├── security.go       # Action validation (70 lines)
│       └── security_test.go  # Security tests (30 lines)
│
├── pkg/utils/                # Public utilities
│   ├── strings.go            # String helpers (50 lines)
│   └── strings_test.go       # Utility tests (50 lines)
│
├── config/                   # Configuration
│   └── config.go            # Config loading (20 lines)
│
├── Documentation
│   ├── README.md            # Full documentation
│   ├── QUICKSTART.md        # Quick start guide
│   ├── ARCHITECTURE.md      # Design documentation
│   ├── DEVELOPMENT.md       # Development guide
│   ├── PROJECT_SUMMARY.md   # Project overview
│   └── INDEX.md             # This file
│
└── Configuration
    ├── go.mod               # Go dependencies
    ├── .env.example         # Environment template
    ├── .gitignore          # Git exclusions
    └── Makefile            # Build commands
```

---

## 🎯 Key Features Overview

| Feature | Location | Status |
|---------|----------|--------|
| Browser Automation | `internal/browser/` | ✅ Complete |
| Dynamic Element Detection | `internal/browser/manager.go` | ✅ Complete |
| OpenAI Integration | `internal/ai/client.go` | ✅ Complete |
| Autonomous Loop | `internal/agent/agent.go` | ✅ Complete |
| Token Management | `internal/context/manager.go` | ✅ Complete |
| Security Layer | `internal/security/security.go` | ✅ Complete |
| Error Recovery | `internal/agent/agent.go` | ✅ Complete |
| CLI Interface | `cmd/agent/main.go` | ✅ Complete |

---

## 🔍 Key Components Explained

### 1. Browser Manager (`internal/browser/manager.go`)
**What it does**: Controls Chromium browser via Playwright

**Key Methods**:
- `NewManager()` - Initialize browser
- `Navigate()` - Go to URL
- `GetPageContent()` - Extract page structure
- `Click()` / `Fill()` - Interact with elements
- `Close()` - Clean up

**Key Algorithm**: Dynamically generates CSS selectors without hardcoding

### 2. AI Client (`internal/ai/client.go`)
**What it does**: Communicates with OpenAI GPT-4

**Key Methods**:
- `NewClient()` - Initialize OpenAI client
- `MakeDecision()` - Get AI decision on next action
- `GetAnalysis()` - Ask AI to analyze page

**Model**: GPT-4 Turbo (2024)

### 3. Agent Core (`internal/agent/agent.go`)
**What it does**: Main autonomous decision loop

**Algorithm**:
1. Navigate to URL
2. Extract page content
3. Ask AI: "What should we do?"
4. Execute action
5. Repeat (max 20 iterations)

### 4. Context Manager (`internal/context/manager.go`)
**What it does**: Manages token budget and conversation history

**Features**:
- 8K token limit per task
- 20-message history window
- Automatic cleanup of old messages
- Token usage tracking

### 5. Security Layer (`internal/security/security.go`)
**What it does**: Prevents accidental destructive actions

**Process**:
1. Detect destructive keyword
2. Show confirmation prompt
3. Log decision
4. Allow/deny action

---

## 📊 Code Statistics

```
Total Lines of Code:     ~1,200 (without tests/docs)
Go Source Files:         12
Test Files:              4
Documentation Files:     5
Configuration Files:     3

Main Package Sizes:
- internal/browser:      ~280 lines (code + tests)
- internal/agent:        ~200 lines
- internal/ai:           ~80 lines
- internal/context:      ~140 lines (code + tests)
- internal/security:     ~100 lines (code + tests)
- pkg/utils:             ~100 lines (code + tests)
- cmd/agent:             ~130 lines
- config:                ~20 lines
```

---

## 🚀 Quick Commands

```bash
# Setup
make install        # Install dependencies
make build         # Build project
make run           # Build and run
make test          # Run all tests
make clean         # Clean artifacts

# Direct execution
go run ./cmd/agent

# Testing specific packages
go test -v ./internal/browser
go test -v ./internal/ai
go test -v ./internal/agent
go test -v ./internal/context
go test -v ./internal/security
go test -v ./pkg/utils
```

---

## 🎓 Learning Path

### Beginner: Understanding the Project
1. Read: QUICKSTART.md (5 minutes)
2. Read: README.md (10 minutes)
3. Run: `make install && make run` (5 minutes)
4. Try: Simple task on example.com (5 minutes)

### Intermediate: Understanding Architecture
1. Read: PROJECT_SUMMARY.md (10 minutes)
2. Read: ARCHITECTURE.md (20 minutes)
3. Review: `internal/agent/agent.go` (10 minutes)
4. Review: `internal/browser/manager.go` (10 minutes)

### Advanced: Implementation Details
1. Read: DEVELOPMENT.md (15 minutes)
2. Review: All source files (30 minutes)
3. Review: Test files (15 minutes)
4. Run: `make test` (5 minutes)
5. Modify: Source code for custom features

### Expert: Extending the Framework
1. Implement sub-agents
2. Add screenshot-based reasoning
3. Implement advanced context compression
4. Create specialized workflows

---

## 🔗 File Cross-References

### Main Loop Flow
```
cmd/agent/main.go
  ↓
internal/agent/agent.go (ExecuteTask)
  ├─→ internal/browser/manager.go (Navigate)
  ├─→ internal/browser/manager.go (GetPageContent)
  ├─→ internal/ai/client.go (GetAnalysis)
  ├─→ internal/security/security.go (Validate)
  └─→ internal/browser/manager.go (Click/Fill)
```

### Configuration Flow
```
cmd/agent/main.go
  ├─→ config/config.go (LoadConfig)
  ├─→ internal/ai/client.go (NewClient)
  └─→ internal/browser/manager.go (NewManager)
```

### Data Flow
```
cmd/agent/main.go (User Input)
  ↓
internal/agent/agent.go (Execute)
  ├─→ internal/browser/manager.go (GetPageContent)
  │   └─→ pkg/utils/strings.go (CleanText)
  ├─→ internal/context/manager.go (Track Tokens)
  ├─→ internal/ai/client.go (Decide)
  ├─→ internal/security/security.go (Validate)
  └─→ internal/browser/manager.go (Execute Action)
```

---

## 📝 Documentation Quality Matrix

| Document | Purpose | Audience | Reading Time |
|----------|---------|----------|--------------|
| README.md | Full overview | Everyone | 10 min |
| QUICKSTART.md | Get started fast | New users | 5 min |
| PROJECT_SUMMARY.md | Project structure | Developers | 10 min |
| ARCHITECTURE.md | Design deep-dive | Architects | 20 min |
| DEVELOPMENT.md | Implementation | Developers | 15 min |
| This file (INDEX.md) | Navigation | Everyone | 5 min |

---

## ✅ Quality Checklist

- ✅ Clean code architecture
- ✅ Comprehensive documentation
- ✅ Unit tests for core components
- ✅ No hardcoded values
- ✅ Proper error handling
- ✅ Security validation
- ✅ Token management
- ✅ Interactive CLI
- ✅ Professional Go practices
- ✅ Easy to extend

---

## 🎯 What's Next?

### Immediate
1. Run `make install` to setup
2. Create `.env` with your OpenAI key
3. Run `make run` to start
4. Try: `task https://example.com "Find and click anything"`

### Short-term
1. Test with real websites
2. Monitor token usage
3. Try more complex tasks
4. Read ARCHITECTURE.md

### Medium-term
1. Implement custom features
2. Add screenshot analysis
3. Create sub-agents
4. Enhance context management

### Long-term
1. Production deployment
2. API server mode
3. Multi-agent coordination
4. Advanced analytics

---

## 📞 Support Resources

### Documentation
- **Architecture questions**: See ARCHITECTURE.md
- **Implementation questions**: See DEVELOPMENT.md
- **Setup questions**: See QUICKSTART.md
- **General questions**: See README.md

### Code Examples
- **Browser automation**: internal/browser/manager.go
- **AI integration**: internal/ai/client.go
- **Agent loop**: internal/agent/agent.go
- **Testing**: *_test.go files

### Common Issues
- Browser issues: QUICKSTART.md Troubleshooting
- OpenAI issues: DEVELOPMENT.md Troubleshooting
- Design questions: ARCHITECTURE.md Design Principles

---

**Project Status**: ✅ Framework Complete & Ready for Use

**Latest Update**: December 9, 2025

**Maintained by**: Vladimir Popov (@VolodyaPopov923)
