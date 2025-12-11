package browser

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/playwright-community/playwright-go"
)

// Manager handles browser automation with persistent sessions
type Manager struct {
	browser    playwright.Browser
	page       playwright.Page
	context    playwright.BrowserContext
	playwright *playwright.Playwright

	pageListeners    map[string]struct{}
	contextListeners map[string]struct{}
	pages            map[string]playwright.Page
	pageOrder        []string
	activePageID     string
}

// NewManager initializes a new browser manager
func NewManager(ctx context.Context) (*Manager, error) {
	pw, err := playwright.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to run playwright: %w", err)
	}

	// Persistent session: use a user-data-dir so manual logins persist across restarts
	userDataDir := os.Getenv("BROWSER_USER_DATA_DIR")
	if userDataDir == "" {
		userDataDir = ".pw_user_data"
	}
	if err := os.MkdirAll(userDataDir, 0o755); err != nil {
		log.Printf("Warning: failed to ensure user data dir: %v\n", err)
	}

	browserCtx, err := launchPersistentWithFallback(pw, userDataDir, defaultLaunchArgs())
	if err != nil {
		return nil, err
	}

	// Ensure at least one page exists
	if len(browserCtx.Pages()) == 0 {
		if _, err := browserCtx.NewPage(); err != nil {
			return nil, fmt.Errorf("failed to create initial page: %w", err)
		}
	}

	manager := &Manager{
		browser:          nil,
		context:          browserCtx,
		playwright:       pw,
		pageListeners:    make(map[string]struct{}),
		contextListeners: make(map[string]struct{}),
		pages:            make(map[string]playwright.Page),
	}
	manager.attachContextListeners(browserCtx)
	manager.rebuildPageTracking(browserCtx)
	return manager, nil
}

// IsBrowserAlive checks if the browser/page is still alive
func (m *Manager) IsBrowserAlive(ctx context.Context) bool {
	if m.page == nil || m.context == nil {
		return false
	}

	// Try a simple operation to check if page is alive
	_, err := m.page.Title()
	return err == nil
}

// RecoverBrowser attempts to recover from a crashed browser
func (m *Manager) RecoverBrowser(ctx context.Context) error {
	log.Printf("Attempting to recover browser...\n")

	// Close old resources
	m.cleanupCurrentContext()

	// Reinitialize
	pw := m.playwright
	userDataDir := os.Getenv("BROWSER_USER_DATA_DIR")
	if userDataDir == "" {
		userDataDir = ".pw_user_data"
	}

	// Try to create new context
	browserCtx, err := launchPersistentWithFallback(pw, userDataDir, defaultLaunchArgs())
	if err != nil {
		return fmt.Errorf("failed to recover browser: %w", err)
	}

	m.context = browserCtx
	m.attachContextListeners(browserCtx)
	if len(browserCtx.Pages()) == 0 {
		if _, err := browserCtx.NewPage(); err != nil {
			return fmt.Errorf("failed to create page during recovery: %w", err)
		}
	}
	m.rebuildPageTracking(browserCtx)
	log.Printf("✅ Browser recovered successfully\n")
	return nil
}

// ensurePlaywright makes sure the playwright runtime is running; if not, it starts a new one.
func (m *Manager) ensurePlaywright(ctx context.Context) error {
	if m.playwright != nil {
		return nil
	}

	pw, err := playwright.Run()
	if err != nil {
		return fmt.Errorf("failed to start playwright: %w", err)
	}
	m.playwright = pw
	return nil
}

// ensureBrowser ensures there is a usable browser context and page. It first
// attempts a light recovery, then fully restarts playwright/context if needed.
func (m *Manager) ensureBrowser(ctx context.Context) error {
	if m.IsBrowserAlive(ctx) {
		return nil
	}

	// Try lightweight recovery first
	if err := m.RecoverBrowser(ctx); err == nil {
		return nil
	}

	// If lightweight recovery failed, try restarting playwright and creating a fresh context
	if m.playwright != nil {
		_ = m.playwright.Stop()
		m.playwright = nil
	}

	if err := m.ensurePlaywright(ctx); err != nil {
		return err
	}

	// Launch a persistent context similar to NewManager
	userDataDir := os.Getenv("BROWSER_USER_DATA_DIR")
	if userDataDir == "" {
		userDataDir = ".pw_user_data"
	}
	browserCtx, err := launchPersistentWithFallback(m.playwright, userDataDir, defaultLaunchArgs())
	if err != nil {
		return fmt.Errorf("failed to restart browser context: %w", err)
	}

	m.context = browserCtx
	m.attachContextListeners(browserCtx)
	if len(browserCtx.Pages()) == 0 {
		if _, err := browserCtx.NewPage(); err != nil {
			return fmt.Errorf("failed to create page during restart: %w", err)
		}
	}
	m.rebuildPageTracking(browserCtx)
	log.Printf("✅ Browser restarted successfully\n")
	return nil
}

// Navigate goes to a specific URL
// If the page closes (e.g., due to CAPTCHA), it gracefully handles the error
func (m *Manager) Navigate(ctx context.Context, url string) error {
	if err := m.ensureBrowser(ctx); err != nil {
		return fmt.Errorf("browser not available: %w", err)
	}

	url = normalizeURL(url)
	if _, err := m.page.Goto(url); err != nil {
		// Check if error is due to page closure (common with CAPTCHA challenges)
		errMsg := err.Error()
		if strings.Contains(errMsg, "Page closed") || strings.Contains(errMsg, "page closed") {
			// Page closed, likely due to CAPTCHA or security challenge
			// Return a recoverable error that the agent can handle and log for diagnostics
			log.Printf("Warning: page closed during navigation to %s: %v\n", url, err)
			return fmt.Errorf("page closed during navigation (possibly due to CAPTCHA) - retrying may help")
		}
		return fmt.Errorf("failed to navigate to %s: %w", url, err)
	}
	return nil
}

// GetPageContent extracts structured information from the current page
func (m *Manager) GetPageContent(ctx context.Context) (PageContent, error) {
	if err := m.ensureBrowser(ctx); err != nil {
		return PageContent{}, fmt.Errorf("browser not available: %w", err)
	}

	// Get title
	title, err := m.page.Title()
	if err != nil {
		title = "Unknown"
	}

	// Get URL
	url := m.page.URL()

	// Extract all interactive elements
	elements, err := m.extractElements(ctx)
	if err != nil {
		log.Printf("Warning: failed to extract elements: %v\n", err)
		elements = []ElementInfo{}
	}

	// Get main text content
	mainText, err := m.page.TextContent("body")
	if err != nil {
		mainText = ""
	}

	return PageContent{
		Title:    title,
		URL:      url,
		Elements: elements,
		MainText: mainText,
	}, nil
}

// extractElements finds all interactive elements on the page
func (m *Manager) extractElements(ctx context.Context) ([]ElementInfo, error) {
	elements := []ElementInfo{}

	// Find all buttons
	buttons, _ := m.page.QuerySelectorAll("button")
	for i, btn := range buttons {
		text, _ := btn.TextContent()
		selector, _ := m.getSelector(ctx, btn)
		if text != "" {
			elements = append(elements, ElementInfo{
				Type:     "button",
				Text:     text,
				Selector: selector,
				Index:    i,
			})
		}
	}

	// Find all clickable links
	links, _ := m.page.QuerySelectorAll("a[href]")
	for i, link := range links {
		text, _ := link.TextContent()
		href, _ := link.GetAttribute("href")
		selector, _ := m.getSelector(ctx, link)
		if text != "" {
			elements = append(elements, ElementInfo{
				Type:     "link",
				Text:     text,
				Href:     href,
				Selector: selector,
				Index:    i,
			})
		}
	}

	// Find form inputs
	inputs, _ := m.page.QuerySelectorAll("input")
	for i, input := range inputs {
		placeholder, _ := input.GetAttribute("placeholder")
		inputType, _ := input.GetAttribute("type")
		selector, _ := m.getSelector(ctx, input)
		label := placeholder
		if label == "" {
			label = inputType
		}
		elements = append(elements, ElementInfo{
			Type:     "input",
			Text:     label,
			Selector: selector,
			Index:    i,
		})
	}

	// Textareas behave like inputs for most sites
	textareas, _ := m.page.QuerySelectorAll("textarea")
	for i, ta := range textareas {
		placeholder, _ := ta.GetAttribute("placeholder")
		selector, _ := m.getSelector(ctx, ta)
		label := placeholder
		if label == "" {
			label = "textarea"
		}
		elements = append(elements, ElementInfo{
			Type:     "textarea",
			Text:     label,
			Selector: selector,
			Index:    i,
		})
	}

	// Some complex UIs (e.g., Yandex Maps) use contenteditable divs instead of inputs
	contentEditable, _ := m.page.QuerySelectorAll("[contenteditable], [role=\"textbox\"]")
	for i, elem := range contentEditable {
		selector, _ := m.getSelector(ctx, elem)
		label, _ := elem.GetAttribute("aria-label")
		if label == "" {
			label, _ = elem.GetAttribute("placeholder")
		}
		if label == "" {
			label = "text field"
		}
		elements = append(elements, ElementInfo{
			Type:     "editable",
			Text:     label,
			Selector: selector,
			Index:    i,
		})
	}

	return elements, nil
}

// getSelector generates a CSS selector for an element
func (m *Manager) getSelector(ctx context.Context, element playwright.ElementHandle) (string, error) {
	if element == nil {
		return "", fmt.Errorf("nil element handle")
	}

	if id, err := element.GetAttribute("id"); err == nil && id != "" {
		return fmt.Sprintf(`[id="%s"]`, cssEscapeAttrValue(id)), nil
	}

	if name, err := element.GetAttribute("name"); err == nil && name != "" {
		tagName := getTagName(element)
		if tagName == "" {
			tagName = "*"
		}
		return fmt.Sprintf(`%s[name="%s"]`, tagName, cssEscapeAttrValue(name)), nil
	}

	selector, err := m.page.Evaluate(`(element) => {
		let path = [];
		let current = element;
		while (current && current.tagName !== 'BODY') {
			let index = 0;
			let sibling = current.previousElementSibling;
			while (sibling) {
				if (sibling.tagName === current.tagName) index++;
				sibling = sibling.previousElementSibling;
			}
			path.unshift(current.tagName.toLowerCase() + ':nth-of-type(' + (index + 1) + ')');
			current = current.parentElement;
		}
		return path.join(' > ');
	}`, element)

	if err == nil {
		if selectorStr, ok := selector.(string); ok {
			return selectorStr, nil
		}
	}

	return "", fmt.Errorf("failed to get selector")
}

// Click clicks on an element by selector
func (m *Manager) Click(ctx context.Context, selector string) error {
	if err := m.ensureBrowser(ctx); err != nil {
		return fmt.Errorf("browser not available: %w", err)
	}

	if err := m.page.Click(selector); err != nil {
		// If page closed while clicking, attempt non-fatal behavior
		if strings.Contains(err.Error(), "Page closed") || strings.Contains(err.Error(), "page closed") {
			log.Printf("Warning: page closed during click (possibly CAPTCHA): %v\n", err)
			return nil
		}
		return fmt.Errorf("failed to click element: %w", err)
	}
	return nil
}

// Fill fills a form field
func (m *Manager) Fill(ctx context.Context, selector, text string) error {
	if err := m.ensureBrowser(ctx); err != nil {
		return fmt.Errorf("browser not available: %w", err)
	}

	if err := m.page.Fill(selector, text); err != nil {
		if strings.Contains(err.Error(), "Page closed") || strings.Contains(err.Error(), "page closed") {
			log.Printf("Warning: page closed during fill (possibly CAPTCHA): %v\n", err)
			return nil
		}
		return fmt.Errorf("failed to fill form: %w", err)
	}
	return nil
}

// Focus brings focus to an element
func (m *Manager) Focus(ctx context.Context, selector string) error {
	if err := m.ensureBrowser(ctx); err != nil {
		return fmt.Errorf("browser not available: %w", err)
	}

	if err := m.page.Focus(selector); err != nil {
		if strings.Contains(err.Error(), "Page closed") || strings.Contains(err.Error(), "page closed") {
			log.Printf("Warning: page closed during focus (possibly CAPTCHA): %v\n", err)
			return nil
		}
		return fmt.Errorf("failed to focus element: %w", err)
	}
	return nil
}

// TypeText types into an element (character-by-character)
func (m *Manager) TypeText(ctx context.Context, selector, text string) error {
	if err := m.ensureBrowser(ctx); err != nil {
		return fmt.Errorf("browser not available: %w", err)
	}

	if err := m.page.Type(selector, text); err != nil {
		if strings.Contains(err.Error(), "Page closed") || strings.Contains(err.Error(), "page closed") {
			log.Printf("Warning: page closed during type (possibly CAPTCHA): %v\n", err)
			return nil
		}
		return fmt.Errorf("failed to type text: %w", err)
	}
	return nil
}

// PressKey sends a keyboard key press (e.g., Enter)
func (m *Manager) PressKey(ctx context.Context, key string) error {
	if err := m.ensureBrowser(ctx); err != nil {
		return fmt.Errorf("browser not available: %w", err)
	}

	if err := m.page.Keyboard().Press(key); err != nil {
		if strings.Contains(err.Error(), "Page closed") || strings.Contains(err.Error(), "page closed") {
			log.Printf("Warning: page closed during key press (possibly CAPTCHA): %v\n", err)
			return nil
		}
		return fmt.Errorf("failed to press key: %w", err)
	}
	return nil
}

// Wait waits for navigation or element
// If the page closes during waiting (e.g., due to CAPTCHA), it gracefully handles it
func (m *Manager) WaitForNavigation(ctx context.Context) error {
	if err := m.page.WaitForLoadState(); err != nil {
		// Check if error is due to page closure (common with CAPTCHA challenges)
		errMsg := err.Error()
		if strings.Contains(errMsg, "Page closed") || strings.Contains(errMsg, "page closed") {
			// Page closed, likely due to CAPTCHA or security challenge
			// This is not necessarily a fatal error - just log and continue
			log.Printf("Warning: page closed during wait (possibly due to CAPTCHA): %v\n", err)
			return nil
		}
		return fmt.Errorf("failed to wait for navigation: %w", err)
	}
	return nil
}

// Close closes the browser
func (m *Manager) Close(ctx context.Context) error {
	if m.page != nil {
		_ = m.page.Close()
	}
	if m.context != nil {
		_ = m.context.Close()
	}
	// persistent context is closed above; no explicit browser.Close needed
	if m.playwright != nil {
		return m.playwright.Stop()
	}
	return nil
}

// SaveStorageState writes the current context storage state to a file
func (m *Manager) SaveStorageState(path string) error {
	if m.context == nil {
		return fmt.Errorf("no context available")
	}
	stateStr, err := m.context.StorageState()
	if err != nil {
		return fmt.Errorf("failed to get storage state: %w", err)
	}
	stateBytes, err := json.Marshal(stateStr)
	if err != nil {
		return fmt.Errorf("failed to marshal storage state: %w", err)
	}
	if err := os.WriteFile(path, stateBytes, 0o600); err != nil {
		return fmt.Errorf("failed to write storage state file: %w", err)
	}
	return nil
}

func defaultLaunchArgs() []string {
	return []string{
		"--disable-gpu",
		"--disable-features=IsolatedSiteInstances",
	}
}

func launchPersistentWithFallback(pw *playwright.Playwright, userDataDir string, args []string) (playwright.BrowserContext, error) {
	if pw == nil {
		return nil, fmt.Errorf("playwright not initialized")
	}

	requestedBrowser := strings.ToLower(strings.TrimSpace(os.Getenv("PLAYWRIGHT_BROWSER")))
	attempts := []string{}

	launch := func(browserType string) (playwright.BrowserContext, error) {
		opts := playwright.BrowserTypeLaunchPersistentContextOptions{
			Headless: playwright.Bool(false),
			Args:     args,
		}
		switch browserType {
		case "firefox":
			return pw.Firefox.LaunchPersistentContext(userDataDir, opts)
		case "webkit":
			return pw.WebKit.LaunchPersistentContext(userDataDir, opts)
		default:
			return pw.Chromium.LaunchPersistentContext(userDataDir, opts)
		}
	}

	if requestedBrowser != "" {
		attempts = append(attempts, requestedBrowser)
	}
	attempts = append(attempts, "chromium", "firefox", "webkit")

	seen := make(map[string]struct{})
	for _, browserName := range attempts {
		if _, exists := seen[browserName]; exists {
			continue
		}
		seen[browserName] = struct{}{}

		ctx, err := launch(browserName)
		if err == nil {
			if requestedBrowser != "" && requestedBrowser != browserName {
				log.Printf("Requested browser %s unavailable, using %s fallback\n", requestedBrowser, browserName)
			}
			return ctx, nil
		}
		log.Printf("%s launch failed: %v\n", strings.Title(browserName), err)
	}

	return nil, fmt.Errorf("failed to launch persistent browser context (tried %v)", attempts)
}

// ListOpenPages returns metadata about all tracked tabs.
func (m *Manager) ListOpenPages() []TabInfo {
	pages := []TabInfo{}
	for idx, pageID := range m.pageOrder {
		page, ok := m.pages[pageID]
		if !ok {
			continue
		}
		title, _ := page.Title()
		if title == "" {
			title = "Unknown"
		}
		url := page.URL()
		pages = append(pages, TabInfo{
			Index:  idx + 1,
			Title:  title,
			URL:    url,
			Active: pageID == m.activePageID,
		})
	}
	return pages
}

// SwitchToPage selects a browser tab either by index or substring match on title/URL.
func (m *Manager) SwitchToPage(ctx context.Context, target string) error {
	if err := m.ensureBrowser(ctx); err != nil {
		return fmt.Errorf("browser not available: %w", err)
	}
	if len(m.pageOrder) == 0 {
		return fmt.Errorf("no open pages to switch")
	}

	target = strings.TrimSpace(target)
	if target == "" {
		nextID := m.pageOrder[len(m.pageOrder)-1]
		m.setActivePage(nextID, true)
		return nil
	}

	if idx, err := strconv.Atoi(target); err == nil {
		if idx < 1 || idx > len(m.pageOrder) {
			return fmt.Errorf("tab index %d out of range", idx)
		}
		m.setActivePage(m.pageOrder[idx-1], true)
		return nil
	}

	lower := strings.ToLower(target)
	for _, id := range m.pageOrder {
		page := m.pages[id]
		title, _ := page.Title()
		url := page.URL()
		if strings.Contains(strings.ToLower(title), lower) || strings.Contains(strings.ToLower(url), lower) {
			m.setActivePage(id, true)
			return nil
		}
	}

	return fmt.Errorf("no page matches target %q", target)
}
func (m *Manager) attachContextListeners(browserCtx playwright.BrowserContext) {
	if browserCtx == nil {
		return
	}
	if m.contextListeners == nil {
		m.contextListeners = make(map[string]struct{})
	}
	key := fmt.Sprintf("%p", browserCtx)
	if _, exists := m.contextListeners[key]; exists {
		return
	}
	m.contextListeners[key] = struct{}{}

	browserCtx.OnClose(func(playwright.BrowserContext) {
		log.Printf("Browser context closed (window terminated or Playwright restarted).")
	})

	browserCtx.OnPage(func(p playwright.Page) {
		log.Printf("Browser emitted a new page event (URL: %s)\n", safePageURL(p))
		m.registerPage(p, true)
	})
}

func (m *Manager) attachPageListeners(page playwright.Page) {
	if page == nil {
		return
	}
	if m.pageListeners == nil {
		m.pageListeners = make(map[string]struct{})
	}
	key := fmt.Sprintf("%p", page)
	if _, exists := m.pageListeners[key]; exists {
		return
	}
	m.pageListeners[key] = struct{}{}

	page.OnClose(func(p playwright.Page) {
		log.Printf("⚠️  Page close event: title=%q url=%s\n", safePageTitle(p), safePageURL(p))
		m.handlePageClosed(p)
	})

	page.OnCrash(func(p playwright.Page) {
		log.Printf("❌ Page crash event: title=%q url=%s\n", safePageTitle(p), safePageURL(p))
	})
}

func safePageTitle(page playwright.Page) string {
	if page == nil {
		return "unknown"
	}
	title, err := page.Title()
	if err != nil || title == "" {
		return "unknown"
	}
	return title
}

func safePageURL(page playwright.Page) string {
	if page == nil {
		return "unknown"
	}
	url := page.URL()
	if url == "" {
		return "unknown"
	}
	return url
}

func getTagName(element playwright.ElementHandle) string {
	if element == nil {
		return ""
	}
	tag, err := element.Evaluate(`(el) => el.tagName.toLowerCase()`)
	if err == nil {
		if tagStr, ok := tag.(string); ok {
			return tagStr
		}
	}
	return ""
}

func cssEscapeAttrValue(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, `"`, `\"`)
	return value
}

func normalizeURL(url string) string {
	url = strings.TrimSpace(url)
	if url == "" {
		return url
	}
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		return url
	}
	return "https://" + url
}

// PageContent represents extracted page information
type PageContent struct {
	Title    string
	URL      string
	Elements []ElementInfo
	MainText string
}

// ElementInfo represents a single interactive element
type ElementInfo struct {
	Type     string // button, link, input, etc.
	Text     string
	Href     string
	Selector string
	Index    int
}

// TabInfo describes an open browser tab.
type TabInfo struct {
	Index  int
	Title  string
	URL    string
	Active bool
}

func (m *Manager) rebuildPageTracking(browserCtx playwright.BrowserContext) {
	if browserCtx == nil {
		return
	}
	m.pages = make(map[string]playwright.Page)
	m.pageOrder = nil
	m.page = nil
	m.activePageID = ""
	if m.pageListeners == nil {
		m.pageListeners = make(map[string]struct{})
	} else {
		m.pageListeners = make(map[string]struct{})
	}

	for _, pg := range browserCtx.Pages() {
		activate := len(m.pageOrder) == 0 && m.activePageID == ""
		m.registerPage(pg, activate)
	}
	if len(m.pageOrder) > 0 && m.activePageID == "" {
		m.setActivePage(m.pageOrder[0], false)
	}
}

func (m *Manager) registerPage(page playwright.Page, activate bool) {
	if page == nil {
		return
	}
	if m.pages == nil {
		m.pages = make(map[string]playwright.Page)
	}

	id := pageIdentifier(page)
	if _, exists := m.pages[id]; exists {
		return
	}

	m.pages[id] = page
	m.pageOrder = append(m.pageOrder, id)
	m.attachPageListeners(page)
	if activate || m.activePageID == "" {
		m.setActivePage(id, activate)
	}
}

func (m *Manager) handlePageClosed(page playwright.Page) {
	if page == nil {
		return
	}
	id := pageIdentifier(page)
	delete(m.pageListeners, id)
	delete(m.pages, id)

	for idx, pageID := range m.pageOrder {
		if pageID == id {
			m.pageOrder = append(m.pageOrder[:idx], m.pageOrder[idx+1:]...)
			break
		}
	}

	if m.activePageID == id {
		m.activePageID = ""
		m.page = nil
		if len(m.pageOrder) > 0 {
			m.setActivePage(m.pageOrder[len(m.pageOrder)-1], true)
		}
	}
}

func (m *Manager) cleanupCurrentContext() {
	if m.page != nil {
		_ = m.page.Close()
	}
	if m.context != nil {
		_ = m.context.Close()
	}
	m.page = nil
	m.activePageID = ""
	m.pageOrder = nil
	m.pages = make(map[string]playwright.Page)
	m.pageListeners = make(map[string]struct{})
}

func (m *Manager) setActivePage(pageID string, bringToFront bool) {
	page, ok := m.pages[pageID]
	if !ok {
		return
	}
	m.page = page
	m.activePageID = pageID
	if bringToFront && page != nil {
		if err := page.BringToFront(); err != nil {
			log.Printf("Warning: failed to bring page to front: %v\n", err)
		}
	}
}

func pageIdentifier(page playwright.Page) string {
	return fmt.Sprintf("%p", page)
}
