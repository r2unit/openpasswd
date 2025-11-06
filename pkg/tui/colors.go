package tui

import "fmt"

const (
	colorReset = "\033[0m"

	colorRedStart    = "\033[38;5;196m"
	colorRedEnd      = "\033[38;5;124m"
	colorGreenStart  = "\033[38;5;46m"
	colorGreenEnd    = "\033[38;5;22m"
	colorOrangeStart = "\033[38;5;214m"
	colorOrangeEnd   = "\033[38;5;166m"
	colorBlueStart   = "\033[38;5;51m"
	colorBlueEnd     = "\033[38;5;21m"
)

func ColorSuccess(text string) string {
	return fmt.Sprintf("%s%s%s", colorGreenStart, text, colorReset)
}

func ColorError(text string) string {
	return fmt.Sprintf("%s%s%s", colorRedStart, text, colorReset)
}

func ColorWarning(text string) string {
	return fmt.Sprintf("%s%s%s", colorOrangeStart, text, colorReset)
}

func ColorInfo(text string) string {
	return fmt.Sprintf("%s%s%s", colorBlueStart, text, colorReset)
}
