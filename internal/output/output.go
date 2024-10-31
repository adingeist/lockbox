package output

import (
	"fmt"
	"github.com/fatih/color"
)

var (
	Success = color.New(color.FgGreen).SprintFunc()
	Error   = color.New(color.FgRed).SprintFunc()
	Info    = color.New(color.FgCyan).SprintFunc()
	Warning = color.New(color.FgYellow).SprintFunc()
	Header  = color.New(color.FgMagenta, color.Bold).SprintFunc()
)

func SuccessIcon() string {
	return Success("✓")
}

func ErrorIcon() string {
	return Error("✗")
}

func WarningIcon() string {
	return Warning("!")
}

func InfoIcon() string {
	return Info("i")
}

func Successf(format string, a ...interface{}) {
	fmt.Printf("%s %s\n", SuccessIcon(), fmt.Sprintf(format, a...))
}

func Errorf(format string, a ...interface{}) {
	fmt.Printf("%s %s\n", ErrorIcon(), fmt.Sprintf(format, a...))
}

func Warnf(format string, a ...interface{}) {
	fmt.Printf("%s %s\n", WarningIcon(), fmt.Sprintf(format, a...))
}

func Infof(format string, a ...interface{}) {
	fmt.Printf("%s %s\n", InfoIcon(), fmt.Sprintf(format, a...))
}

func ListItem(text string) {
	fmt.Printf("  %s %s\n", Info("•"), text)
}

func Section(title string) {
	fmt.Printf("\n%s:\n", Header(title))
} 