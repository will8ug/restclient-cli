package tui

import "github.com/charmbracelet/lipgloss"

var (
	activeBorderColor   = lipgloss.Color("205")
	inactiveBorderColor = lipgloss.Color("240")

	activePanelStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(activeBorderColor)

	inactivePanelStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(inactiveBorderColor)

	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("230")).
			Background(lipgloss.Color("236")).
			Padding(0, 1)

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("230"))

	methodStyles = map[string]lipgloss.Style{
		"GET":     lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("42")),
		"POST":    lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("33")),
		"PUT":     lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("214")),
		"PATCH":   lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("214")),
		"DELETE":  lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("196")),
		"HEAD":    lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("247")),
		"OPTIONS": lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("247")),
	}

	statusOKStyle       = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("42"))
	statusRedirectStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("214"))
	statusErrorStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("196"))
	errorTextStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	dimStyle            = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
)

func methodStyle(method string) lipgloss.Style {
	if s, ok := methodStyles[method]; ok {
		return s
	}
	return lipgloss.NewStyle().Bold(true)
}

func statusStyle(code int) lipgloss.Style {
	switch {
	case code >= 200 && code < 300:
		return statusOKStyle
	case code >= 300 && code < 400:
		return statusRedirectStyle
	default:
		return statusErrorStyle
	}
}
