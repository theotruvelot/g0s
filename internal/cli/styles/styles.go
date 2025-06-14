package styles

import (
	"github.com/charmbracelet/lipgloss"
)

// Color palette
const (
	Primary   = "#FF8E00" // Orange
	Secondary = "#10B981" // Green
	Accent    = "#7C3AED" // Purple

	// Status colors
	Success = "#10B981" // Green
	Warning = "#FF8E00" // Orange
	Error   = "#EF4444" // Red
	Info    = "#3B82F6" // Blue

	// Neutral colors
	Background = "#0F172A" // Dark blue
	Surface    = "#1E293B" // Lighter dark blue
	Border     = "#334155" // Gray blue
	Text       = "#F8FAFC" // Light gray
	TextMuted  = "#94A3B8" // Muted gray
)

// Base styles
var (
	// Text styles
	TitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(Primary)).
			Bold(true).
			Margin(1, 0)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(Text)).
			Bold(true)

	BodyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(Text))

	MutedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(TextMuted))

	// Status styles
	SuccessStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(Success)).
			Bold(true)

	WarningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(Warning)).
			Bold(true)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(Error)).
			Bold(true)

	InfoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(Info)).
			Bold(true)

	// Loading styles
	LoadingStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(Primary)).
			Bold(true).
			Align(lipgloss.Center)

	SpinnerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(Primary))

	// Container styles
	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(Border)).
			Padding(1, 2).
			Margin(1, 0)

	HighlightBoxStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color(Primary)).
				Padding(1, 2).
				Margin(1, 0)

	ContentStyle = lipgloss.NewStyle().
			Padding(1, 2)

	// Interactive styles
	ButtonStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(Text)).
			Background(lipgloss.Color(Primary)).
			Padding(0, 2).
			Margin(0, 1)

	ButtonActiveStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(Background)).
				Background(lipgloss.Color(Secondary)).
				Padding(0, 2).
				Margin(0, 1).
				Bold(true)

	// Header and footer styles
	HeaderStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(Text)).
			Background(lipgloss.Color(Surface)).
			Padding(0, 1).
			Width(100).
			Bold(true)

	FooterStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(TextMuted)).
			Background(lipgloss.Color(Surface)).
			Padding(0, 1).
			Width(100)

	// Metrics styles
	MetricLabelStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(TextMuted)).
				Width(15).
				Align(lipgloss.Right)

	MetricValueStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(Text)).
				Bold(true)

	MetricGoodStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(Success)).
			Bold(true)

	MetricBadStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(Error)).
			Bold(true)

	MetricWarningStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(Warning)).
				Bold(true)

	FocusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(Accent))

	// Form styles
	FormContainerStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color(Border)).
				Padding(2, 4).
				Margin(1, 0).
				Width(60)

	InputStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(Text)).
			Margin(0, 0, 1, 0)

	InputFocusedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(Primary)).
				Margin(0, 0, 1, 0)

	InputPlaceholderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(TextMuted))

	FormButtonStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(Background)).
			Background(lipgloss.Color(Primary)).
			Padding(0, 3).
			Margin(1, 0).
			Bold(true).
			Align(lipgloss.Center)

	FormButtonFocusedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(Background)).
				Background(lipgloss.Color(Secondary)).
				Padding(0, 3).
				Margin(1, 0).
				Bold(true).
				Align(lipgloss.Center)

	// Logo style
	LogoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(Primary)).
			Bold(true).
			Align(lipgloss.Center)

	// Help text style
	HelpTextStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(TextMuted)).
			Italic(true).
			Align(lipgloss.Center)
)

// Helper functions for dynamic styling

// Conditional styles based on state
func StatusStyle(status string) lipgloss.Style {
	switch status {
	case "success", "healthy", "ok", "online":
		return SuccessStyle
	case "warning", "degraded":
		return WarningStyle
	case "error", "unhealthy", "failed", "offline":
		return ErrorStyle
	default:
		return InfoStyle
	}
}
