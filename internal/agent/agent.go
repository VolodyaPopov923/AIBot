package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/VolodyaPopov923/AIBot/internal/ai"
	"github.com/VolodyaPopov923/AIBot/internal/browser"
	ctxmgr "github.com/VolodyaPopov923/AIBot/internal/context"
	"github.com/VolodyaPopov923/AIBot/internal/security"
)

type Agent struct {
	browserMgr    *browser.Manager
	aiClient      *ai.Client
	contextMgr    *ctxmgr.ContextManager
	securityMgr   *security.Validator
	currentTask   string
	maxIterations int
	verbose       bool
}

func NewAgent(browserMgr *browser.Manager, aiClient *ai.Client, verbose bool) *Agent {
	return &Agent{
		browserMgr:    browserMgr,
		aiClient:      aiClient,
		contextMgr:    ctxmgr.NewContextManager(8000, 20),
		securityMgr:   security.NewValidator(),
		maxIterations: 20,
		verbose:       verbose,
	}
}

func (a *Agent) ExecuteTask(ctx context.Context, task string, initialURL string) error {
	a.currentTask = task

	a.contextMgr.ClearContext()
	a.contextMgr.ResetTokenCounter()

	if a.verbose {
		log.Printf("Starting task: %s\n", task)
		log.Printf("Initial URL: %s\n", initialURL)
	}

	if initialURL != "" && initialURL != "about:blank" {
		if err := a.browserMgr.Navigate(ctx, initialURL); err != nil {
			return fmt.Errorf("failed to navigate to initial URL: %w", err)
		}
		if err := a.browserMgr.WaitForNavigation(ctx); err != nil {
			log.Printf("Warning: navigation wait failed: %v\n", err)
		}
	}

	pageContent, err := a.browserMgr.GetPageContent(ctx)
	if err != nil {
		return fmt.Errorf("failed to get page content for planning: %w", err)
	}
	pageDesc := buildPageDescription(pageContent, a.browserMgr.ListOpenPages())

	steps, err := a.aiClient.PlanTask(ctx, task, pageDesc)
	if err != nil {
		if a.verbose {
			log.Printf("Planning failed, falling back to iterative mode: %v\n", err)
		}
		for iteration := 0; iteration < a.maxIterations; iteration++ {
			if a.verbose {
				log.Printf("\n=== Iteration %d ===\n", iteration+1)
			}

			pageContent, err := a.browserMgr.GetPageContent(ctx)
			if err != nil {
				return fmt.Errorf("failed to get page content: %w", err)
			}
			if isBlockedPage(pageContent) {
				log.Printf("CAPTCHA detected on %s. Waiting for you to solve it...\n", pageContent.URL)
				if err := a.waitForCaptchaSolution(ctx); err != nil {
					return fmt.Errorf("CAPTCHA wait failed: %w", err)
				}
				log.Printf("CAPTCHA solved, continuing task...\n")
				continue
			}

			decision, err := a.analyzeAndDecide(ctx, pageContent)
			if err != nil {
				return fmt.Errorf("decision making failed: %w", err)
			}
			if a.verbose {
				log.Printf("Decision: %s\n", decision.Reasoning)
			}
			if decision.IsComplete {
				if a.verbose {
					log.Printf("Task completed successfully\n")
				}
				return nil
			}
			if err := a.executeAction(ctx, decision); err != nil {
				if a.verbose {
					log.Printf("Action failed, attempting recovery: %v\n", err)
				}
				continue
			}
			time.Sleep(1 * time.Second)
		}
		return fmt.Errorf("max iterations (%d) reached without completing task: %s", a.maxIterations, a.currentTask)
	}

	if a.verbose {
		log.Printf("Plan generated with %d steps. Executing each step once.\n", len(steps))
	}

	for idx, step := range steps {
		if a.verbose {
			log.Printf("\n--- Executing plan step %d/%d: %s\n", idx+1, len(steps), step)
		}

		pc, err := a.browserMgr.GetPageContent(ctx)
		if err != nil {
			return fmt.Errorf("failed to get page content: %w", err)
		}
		if isBlockedPage(pc) {
			log.Printf("CAPTCHA detected on %s. Waiting for you to solve it...\n", pc.URL)
			if err := a.waitForCaptchaSolution(ctx); err != nil {
				return fmt.Errorf("CAPTCHA wait failed: %w", err)
			}
			log.Printf("CAPTCHA solved, continuing plan...\n")
		}

		systemPrompt := `You are an intelligent web automation agent. Provide a single concise action to accomplish the given step on the current page.
Valid actions: navigate, click, fill, focus, type, press, wait, switch_tab, complete, error.
Use "focus" before typing if needed, "type" for freeform text entry (text field provided in the decision), and "press" for keyboard keys like Enter.
Use "switch_tab" when you must operate on a different browser tab (specify tab index or part of the title/URL).`
		userInput := fmt.Sprintf("Task: %s\nPlan step: %s\nCurrent page:\n%s\n\nReturn a single JSON decision as before.", a.currentTask, step, buildPageDescription(pc, a.browserMgr.ListOpenPages()))

		a.contextMgr.AddMessage("system", systemPrompt)
		a.contextMgr.AddMessage("user", userInput)

		decision, err := a.aiClient.MakeDecision(ctx, systemPrompt, userInput)
		if err != nil {
			return fmt.Errorf("MakeDecision failed for step %d: %w", idx+1, err)
		}

		if a.verbose {
			log.Printf("Decision for step %d: %v\n", idx+1, decision.Reasoning)
		}

		if err := a.executeAction(ctx, decision); err != nil {
			if a.verbose {
				log.Printf("Execution of step %d failed: %v\n", idx+1, err)
			}
			continue
		}

		_ = a.browserMgr.WaitForNavigation(ctx)
		time.Sleep(1 * time.Second)
	}

	if a.verbose {
		log.Printf("Plan completed (all steps attempted).\n")
	}
	return nil
}

func (a *Agent) waitForCaptchaSolution(ctx context.Context) error {
	const checkInterval = 2 * time.Second
	const timeout = 5 * time.Minute
	deadline := time.Now().Add(timeout)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if time.Now().After(deadline) {
			return fmt.Errorf("CAPTCHA wait timeout")
		}

		time.Sleep(checkInterval)

		pageContent, err := a.browserMgr.GetPageContent(ctx)
		if err != nil {
			log.Printf("Checking page: %v\n", err)
			continue
		}

		if !isBlockedPage(pageContent) {
			log.Printf("CAPTCHA solved! Now at: %s\n", pageContent.URL)
			return nil
		}

		log.Printf("Waiting for CAPTCHA...\n")
	}
}

func (a *Agent) analyzeAndDecide(ctx context.Context, pageContent browser.PageContent) (ai.DecisionResponse, error) {
	pageDescription := buildPageDescription(pageContent, a.browserMgr.ListOpenPages())

	systemPrompt := `You are an intelligent web automation agent. Your task is to complete user requests by interacting with web pages.
You can:
- Click on buttons and links (action "click")
- Fill or type into form fields (actions "fill" or "type"; provide text to enter)
- Focus an element before typing if necessary (action "focus")
- Navigate to URLs (action "navigate")
- Switch between open tabs (action "switch_tab"; specify tab index or a fragment of the tab title/URL)
- Press keyboard keys (action "press"; set text to the key name, e.g. "Enter")
- Read page content
- Wait for page load or manual intervention (action "wait")

IMPORTANT INSTRUCTIONS:
- If you encounter a CAPTCHA or security challenge, use the "wait" action to give the user time to solve it manually. Do NOT use "error".
- After waiting, try to navigate again or continue the task.
- Be systematic, logical, and report when the task is complete.
- If no progress can be made after several retries on the same page, only then use "error" action.`

	userInput := fmt.Sprintf(`Current task: %s

Current page state:
%s

Based on the page content, what should be the next action? Respond with a clear decision.
Return a JSON object with:
- action: the action to take (navigate, click, fill, focus, type, press, switch_tab, wait, complete, error)
- selector: CSS selector for the element (if clicking or filling)
- text: text to fill (if filling a form)
- url: URL to navigate to (if navigating)
- reasoning: explanation of your decision
- is_complete: whether the task is complete
- needs_confirm: whether this action needs user confirmation
`, a.currentTask, pageDescription)

	a.contextMgr.AddMessage("system", systemPrompt)
	a.contextMgr.AddMessage("user", userInput)

	needed := ctxmgr.EstimateTokens(systemPrompt) + ctxmgr.EstimateTokens(userInput) + 400
	for !a.contextMgr.TokenCounter().CanAddTokens(needed) {
		a.contextMgr.RemoveOldest(1)
	}

	decision, err := a.aiClient.MakeDecision(ctx, systemPrompt, userInput)
	if err != nil {
		log.Printf("AI MakeDecision error: %v", err)
		return ai.DecisionResponse{Action: "error", Reasoning: err.Error(), IsComplete: false}, nil
	}

	if decision.Reasoning != "" {
		a.contextMgr.AddMessage("assistant", decision.Reasoning)
	} else {
		raw, _ := json.Marshal(decision)
		a.contextMgr.AddMessage("assistant", string(raw))
	}

	promptTokens := ctxmgr.EstimateTokens(systemPrompt) + ctxmgr.EstimateTokens(userInput)
	completionTokens := ctxmgr.EstimateTokens(decision.Reasoning)
	if err := a.contextMgr.TokenCounter().Add(promptTokens, completionTokens); err != nil {
		if a.verbose {
			log.Printf("Token limit exceeded after add: %v. Pruning history...\n", err)
		}
		a.contextMgr.RemoveOldest(1)
		_ = a.contextMgr.TokenCounter().Add(promptTokens, completionTokens)
	}

	if a.verbose {
		log.Printf("AI Decision: %+v\n", decision)
	}

	return decision, nil
}

func (a *Agent) executeAction(ctx context.Context, decision ai.DecisionResponse) error {
	if decision.NeedsConfirm {
		destructiveAction := security.DestructiveAction{
			Type:        decision.Action,
			Description: decision.Reasoning,
			Severity:    "high",
		}

		approved, err := a.securityMgr.RequestConfirmation(destructiveAction)
		if err != nil {
			return fmt.Errorf("confirmation check failed: %w", err)
		}

		security.LogAction(decision.Action, decision.Reasoning, approved)
		if !approved {
			return fmt.Errorf("action denied by user")
		}
	}

	action := strings.ToLower(decision.Action)

	switch action {
	case "navigate":
		if decision.URL != "" {
			if err := a.browserMgr.Navigate(ctx, decision.URL); err != nil {
				if strings.Contains(err.Error(), "page closed") {
					if a.verbose {
						log.Printf("Navigate: %v (will retry)\n", err)
					}
					return nil
				}
				return err
			}
			_ = a.browserMgr.WaitForNavigation(ctx)
		}
	case "click":
		if decision.Selector != "" {
			if err := a.browserMgr.Click(ctx, decision.Selector); err != nil {
				return err
			}
			_ = a.browserMgr.WaitForNavigation(ctx)
		}
	case "fill", "input":
		if decision.Selector != "" && decision.Text != "" {
			if err := a.browserMgr.Fill(ctx, decision.Selector, decision.Text); err != nil {
				return err
			}
		}
	case "focus":
		if decision.Selector != "" {
			if err := a.browserMgr.Focus(ctx, decision.Selector); err != nil {
				return err
			}
		}
	case "type":
		if decision.Selector != "" && decision.Text != "" {
			if err := a.browserMgr.TypeText(ctx, decision.Selector, decision.Text); err != nil {
				return err
			}
		}
	case "press", "keypress", "key":
		if decision.Text != "" {
			if err := a.browserMgr.PressKey(ctx, decision.Text); err != nil {
				return err
			}
		}
	case "switch_tab", "switch":
		target := decision.Text
		if target == "" {
			target = decision.URL
		}
		if err := a.browserMgr.SwitchToPage(ctx, target); err != nil {
			return err
		}
	case "wait":
		time.Sleep(2 * time.Second)
	case "complete":
		return nil
	case "error":
		time.Sleep(1 * time.Second)
	default:
		return fmt.Errorf("unknown action: %s", decision.Action)
	}

	return nil
}

func buildPageDescription(pageContent browser.PageContent, tabs []browser.TabInfo) string {
	desc := fmt.Sprintf(`Title: %s
URL: %s

Interactive Elements:
`, pageContent.Title, pageContent.URL)

	for i, elem := range pageContent.Elements {
		desc += fmt.Sprintf("%d. [%s] %s (selector: %s)\n", i+1, elem.Type, elem.Text, elem.Selector)
	}

	if len(tabs) > 0 {
		desc += "\nOpen Tabs:\n"
		for _, tab := range tabs {
			state := " "
			if tab.Active {
				state = "*"
			}
			desc += fmt.Sprintf("[%s] %d. %s (%s)\n", state, tab.Index, tab.Title, tab.URL)
		}
	}

	return desc
}

func isBlockedPage(pageContent browser.PageContent) bool {
	url := strings.ToLower(pageContent.URL)
	title := strings.ToLower(pageContent.Title)

	blockedPatterns := []string{
		"captcha",
		"showcaptcha",
		"challenge",
		"security check",
		"verify",
		"bot-check",
		"robot",
		"access denied",
		"403",
		"blocked",
	}

	for _, pattern := range blockedPatterns {
		if strings.Contains(url, pattern) || strings.Contains(title, pattern) {
			return true
		}
	}

	return false
}
