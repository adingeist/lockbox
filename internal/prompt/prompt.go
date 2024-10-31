package prompt

import (
	"github.com/AlecAivazis/survey/v2"
)

// SelectFromList presents a list of options to choose from using arrow keys
func SelectFromList(message string, options []string) (string, error) {
	var selected string
	prompt := &survey.Select{
		Message: message,
		Options: options,
	}
	err := survey.AskOne(prompt, &selected)
	return selected, err
}

// Confirm asks for confirmation with yes/no
func Confirm(message string) (bool, error) {
	var confirmed bool
	prompt := &survey.Confirm{
		Message: message,
	}
	err := survey.AskOne(prompt, &confirmed)
	return confirmed, err
}

// Input asks for text input
func Input(message string) (string, error) {
	var input string
	prompt := &survey.Input{
		Message: message,
	}
	err := survey.AskOne(prompt, &input)
	return input, err
} 