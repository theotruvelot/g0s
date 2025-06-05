package styles

import (
	"fmt"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"
)

func TestColors(t *testing.T) {
	// Test that all color constants are defined and not empty
	colors := map[string]string{
		"Primary":    Primary,
		"Secondary":  Secondary,
		"Accent":     Accent,
		"Success":    Success,
		"Warning":    Warning,
		"Error":      Error,
		"Info":       Info,
		"Background": Background,
		"Surface":    Surface,
		"Border":     Border,
		"Text":       Text,
		"TextMuted":  TextMuted,
	}

	for name, color := range colors {
		t.Run(name, func(t *testing.T) {
			assert.NotEmpty(t, color, "Color %s should not be empty", name)
			assert.True(t, len(color) > 0, "Color %s should have content", name)
			// Check that it's a valid hex color (starts with #)
			assert.True(t, color[0] == '#', "Color %s should start with #", name)
			assert.True(t, len(color) == 7, "Color %s should be 7 characters long", name)
		})
	}
}

func TestTextStyles(t *testing.T) {
	tests := []struct {
		name  string
		style lipgloss.Style
	}{
		{"TitleStyle", TitleStyle},
		{"SubtitleStyle", SubtitleStyle},
		{"BodyStyle", BodyStyle},
		{"MutedStyle", MutedStyle},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test that the style is not nil/empty
			rendered := tt.style.Render("test")
			assert.NotEmpty(t, rendered, "Style %s should render content", tt.name)

			// Test that the style can be applied
			assert.Contains(t, rendered, "test", "Style %s should contain the original text", tt.name)
		})
	}
}

func TestStatusStyles(t *testing.T) {
	tests := []struct {
		name  string
		style lipgloss.Style
	}{
		{"SuccessStyle", SuccessStyle},
		{"WarningStyle", WarningStyle},
		{"ErrorStyle", ErrorStyle},
		{"InfoStyle", InfoStyle},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rendered := tt.style.Render("test")
			assert.NotEmpty(t, rendered, "Status style %s should render content", tt.name)
			assert.Contains(t, rendered, "test", "Status style %s should contain the original text", tt.name)
		})
	}
}

func TestLoadingStyles(t *testing.T) {
	tests := []struct {
		name  string
		style lipgloss.Style
	}{
		{"LoadingStyle", LoadingStyle},
		{"SpinnerStyle", SpinnerStyle},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rendered := tt.style.Render("test")
			assert.NotEmpty(t, rendered, "Loading style %s should render content", tt.name)
			assert.Contains(t, rendered, "test", "Loading style %s should contain the original text", tt.name)
		})
	}
}

func TestContainerStyles(t *testing.T) {
	tests := []struct {
		name  string
		style lipgloss.Style
	}{
		{"BoxStyle", BoxStyle},
		{"HighlightBoxStyle", HighlightBoxStyle},
		{"ContentStyle", ContentStyle},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rendered := tt.style.Render("test content")
			assert.NotEmpty(t, rendered, "Container style %s should render content", tt.name)
			assert.Contains(t, rendered, "test content", "Container style %s should contain the original text", tt.name)
		})
	}
}

func TestInteractiveStyles(t *testing.T) {
	tests := []struct {
		name  string
		style lipgloss.Style
	}{
		{"ButtonStyle", ButtonStyle},
		{"ButtonActiveStyle", ButtonActiveStyle},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rendered := tt.style.Render("Button")
			assert.NotEmpty(t, rendered, "Interactive style %s should render content", tt.name)
			assert.Contains(t, rendered, "Button", "Interactive style %s should contain the original text", tt.name)
		})
	}
}

func TestHeaderFooterStyles(t *testing.T) {
	tests := []struct {
		name  string
		style lipgloss.Style
	}{
		{"HeaderStyle", HeaderStyle},
		{"FooterStyle", FooterStyle},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rendered := tt.style.Render("Header/Footer")
			assert.NotEmpty(t, rendered, "Header/Footer style %s should render content", tt.name)
			assert.Contains(t, rendered, "Header/Footer", "Header/Footer style %s should contain the original text", tt.name)
		})
	}
}

func TestMetricStyles(t *testing.T) {
	tests := []struct {
		name  string
		style lipgloss.Style
	}{
		{"MetricLabelStyle", MetricLabelStyle},
		{"MetricValueStyle", MetricValueStyle},
		{"MetricGoodStyle", MetricGoodStyle},
		{"MetricBadStyle", MetricBadStyle},
		{"MetricWarningStyle", MetricWarningStyle},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rendered := tt.style.Render("123")
			assert.NotEmpty(t, rendered, "Metric style %s should render content", tt.name)
			assert.Contains(t, rendered, "123", "Metric style %s should contain the original text", tt.name)
		})
	}
}

func TestStatusStyle_Function(t *testing.T) {
	tests := []struct {
		name           string
		status         string
		expectedResult lipgloss.Style
	}{
		{
			name:           "success status",
			status:         "success",
			expectedResult: SuccessStyle,
		},
		{
			name:           "healthy status",
			status:         "healthy",
			expectedResult: SuccessStyle,
		},
		{
			name:           "ok status",
			status:         "ok",
			expectedResult: SuccessStyle,
		},
		{
			name:           "online status",
			status:         "online",
			expectedResult: SuccessStyle,
		},
		{
			name:           "warning status",
			status:         "warning",
			expectedResult: WarningStyle,
		},
		{
			name:           "degraded status",
			status:         "degraded",
			expectedResult: WarningStyle,
		},
		{
			name:           "error status",
			status:         "error",
			expectedResult: ErrorStyle,
		},
		{
			name:           "unhealthy status",
			status:         "unhealthy",
			expectedResult: ErrorStyle,
		},
		{
			name:           "failed status",
			status:         "failed",
			expectedResult: ErrorStyle,
		},
		{
			name:           "offline status",
			status:         "offline",
			expectedResult: ErrorStyle,
		},
		{
			name:           "unknown status defaults to info",
			status:         "unknown",
			expectedResult: InfoStyle,
		},
		{
			name:           "empty status defaults to info",
			status:         "",
			expectedResult: InfoStyle,
		},
		{
			name:           "custom status defaults to info",
			status:         "custom_status",
			expectedResult: InfoStyle,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StatusStyle(tt.status)

			// Test that the function returns a style
			rendered := result.Render("test")
			assert.NotEmpty(t, rendered, "StatusStyle should render content")
			assert.Contains(t, rendered, "test", "StatusStyle should contain the original text")

			// Test specific color expectations by comparing rendered output
			expectedRendered := tt.expectedResult.Render("test")
			assert.Equal(t, expectedRendered, rendered, "StatusStyle(%s) should match expected style", tt.status)
		})
	}
}

func TestStatusStyle_CaseInsensitive(t *testing.T) {
	// Test that status matching is case-sensitive (current implementation)
	tests := []struct {
		name   string
		status string
	}{
		{"uppercase SUCCESS", "SUCCESS"},
		{"mixed case Success", "Success"},
		{"uppercase ERROR", "ERROR"},
		{"mixed case Error", "Error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StatusStyle(tt.status)

			// These should all default to InfoStyle since the function is case-sensitive
			expectedRendered := InfoStyle.Render("test")
			actualRendered := result.Render("test")

			assert.Equal(t, expectedRendered, actualRendered,
				"StatusStyle should be case-sensitive and default to InfoStyle for %s", tt.status)
		})
	}
}

func TestAllStylesRenderProperly(t *testing.T) {
	// Test that all exported styles can render without panicking
	testText := "Test Content"

	styles := map[string]lipgloss.Style{
		"TitleStyle":         TitleStyle,
		"SubtitleStyle":      SubtitleStyle,
		"BodyStyle":          BodyStyle,
		"MutedStyle":         MutedStyle,
		"SuccessStyle":       SuccessStyle,
		"WarningStyle":       WarningStyle,
		"ErrorStyle":         ErrorStyle,
		"InfoStyle":          InfoStyle,
		"LoadingStyle":       LoadingStyle,
		"SpinnerStyle":       SpinnerStyle,
		"BoxStyle":           BoxStyle,
		"HighlightBoxStyle":  HighlightBoxStyle,
		"ContentStyle":       ContentStyle,
		"ButtonStyle":        ButtonStyle,
		"ButtonActiveStyle":  ButtonActiveStyle,
		"HeaderStyle":        HeaderStyle,
		"FooterStyle":        FooterStyle,
		"MetricLabelStyle":   MetricLabelStyle,
		"MetricValueStyle":   MetricValueStyle,
		"MetricGoodStyle":    MetricGoodStyle,
		"MetricBadStyle":     MetricBadStyle,
		"MetricWarningStyle": MetricWarningStyle,
	}

	for name, style := range styles {
		t.Run(name, func(t *testing.T) {
			// Should not panic
			assert.NotPanics(t, func() {
				rendered := style.Render(testText)
				assert.NotEmpty(t, rendered, "Style %s should produce output", name)
			}, "Style %s should not panic when rendering", name)
		})
	}
}

func TestEmptyTextRendering(t *testing.T) {
	// Test that styles handle empty text gracefully
	styles := []lipgloss.Style{
		TitleStyle,
		SuccessStyle,
		ErrorStyle,
		BoxStyle,
	}

	for i, style := range styles {
		t.Run(fmt.Sprintf("style_%d", i), func(t *testing.T) {
			assert.NotPanics(t, func() {
				rendered := style.Render("")
				// Empty text should still be handled gracefully
				assert.NotNil(t, rendered)
			})
		})
	}
}
