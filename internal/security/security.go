package security

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type DestructiveAction struct {
	Type        string
	Description string
	Target      string
	Severity    string
}

type Validator struct {
	reader *bufio.Reader
}

func NewValidator() *Validator {
	return &Validator{
		reader: bufio.NewReader(os.Stdin),
	}
}

func (v *Validator) IsDestructive(action string) bool {
	destructiveKeywords := []string{
		"delete", "remove", "destroy",
		"payment", "purchase", "checkout", "pay",
		"logout", "sign out",
		"clear", "reset", "wipe",
		"disable", "close account",
	}

	actionLower := strings.ToLower(action)
	for _, keyword := range destructiveKeywords {
		if strings.Contains(actionLower, keyword) {
			return true
		}
	}
	return false
}

func (v *Validator) RequestConfirmation(action DestructiveAction) (bool, error) {
	fmt.Println("\n⚠️  SECURITY CONFIRMATION REQUIRED")
	fmt.Printf("Action Type: %s (%s severity)\n", action.Type, action.Severity)
	fmt.Printf("Description: %s\n", action.Description)
	if action.Target != "" {
		fmt.Printf("Target: %s\n", action.Target)
	}
	fmt.Print("\nDo you want to proceed? (yes/no): ")

	response, err := v.reader.ReadString('\n')
	if err != nil {
		return false, err
	}

	response = strings.ToLower(strings.TrimSpace(response))
	return response == "yes" || response == "y", nil
}

func LogAction(actionType, description string, approved bool) {
	status := "DENIED"
	if approved {
		status = "APPROVED"
	}
	fmt.Printf("[SECURITY LOG] %s - Type: %s, Description: %s\n", status, actionType, description)
}
