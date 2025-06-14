package login

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/theotruvelot/g0s/internal/cli/config"
	"github.com/theotruvelot/g0s/internal/cli/services"
	"github.com/theotruvelot/g0s/internal/cli/styles"
	"github.com/theotruvelot/g0s/pkg/logger"
	"github.com/theotruvelot/g0s/pkg/proto/auth"
	"go.uber.org/zap"
)

// loginTestCompleteMsg is sent when the login test is complete
type loginTestCompleteMsg struct {
	success      bool
	error        error
	jwtToken     string
	refreshToken string
}

// Model represents the login page model
type Model struct {
	focusIndex    int
	inputs        []textinput.Model
	cursorMode    cursor.Mode
	shouldProceed bool
	error         error
	width         int
	height        int
	spinner       spinner.Model
	isLoading     bool
	serverURL     string
	username      string
	apiToken      string
	authService   *services.AuthService
}

// NewModel creates a new login model
func NewModel(authService *services.AuthService) Model {
	m := Model{
		inputs:      make([]textinput.Model, 3),
		authService: authService,
	}

	// Initialize spinner
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(styles.Primary))
	m.spinner = s

	var t textinput.Model
	for i := range m.inputs {
		t = textinput.New()
		t.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(styles.Primary))
		t.CharLimit = 64
		t.Width = 50

		switch i {
		case 0:
			t.Placeholder = "Server URL (e.g., localhost:50051)"
			t.Focus()
			t.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(styles.Primary))
			t.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(styles.Text))
			t.PlaceholderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(styles.TextMuted))
		case 1:
			t.Placeholder = "Username"
			t.PromptStyle = lipgloss.NewStyle()
			t.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(styles.Text))
			t.PlaceholderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(styles.TextMuted))
		case 2:
			t.Placeholder = "API Token"
			t.EchoMode = textinput.EchoPassword
			t.EchoCharacter = '•'
			t.PromptStyle = lipgloss.NewStyle()
			t.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(styles.Text))
			t.PlaceholderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(styles.TextMuted))
		}

		m.inputs[i] = t
	}

	return m
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, m.spinner.Tick)
}

// performLogin performs the actual login via gRPC
func (m Model) performLogin(serverURL, username, apiToken string) tea.Cmd {
	return func() tea.Msg {
		logger.Debug("Performing login", zap.String("server", serverURL), zap.String("username", username))

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		response, err := m.authService.Login(ctx, serverURL, username, apiToken)
		if err != nil {
			logger.Error("Authentication failed", zap.Error(err))
			return loginTestCompleteMsg{
				success: false,
				error:   fmt.Errorf("authentication failed: %w", err),
			}
		}

		if response.GetStatus() != auth.AuthenticateResponse_OK {
			logger.Error("Authentication failed - invalid credentials")
			return loginTestCompleteMsg{
				success: false,
				error:   fmt.Errorf("invalid credentials"),
			}
		}
		logger.Info("Response", zap.Any("response", response))
		logger.Info("Authentication response received", zap.String("jwtToken", response.GetJwtToken()), zap.String("refreshToken", response.GetJwtRefreshToken()))

		logger.Info("Authentication successful", zap.String("username", username))
		return loginTestCompleteMsg{
			success:      true,
			jwtToken:     response.GetJwtToken(),
			refreshToken: response.GetJwtRefreshToken(),
		}
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case loginTestCompleteMsg:
		m.isLoading = false

		if !msg.success {
			m.error = msg.error
			return m, nil
		}

		cfg := &config.Config{
			ServerURL:    m.serverURL,
			Username:     m.username,
			JWTToken:     msg.jwtToken,
			RefreshToken: msg.refreshToken,
		}

		if err := config.SaveConfig(cfg); err != nil {
			logger.Error("Failed to save config", zap.Error(err))
			m.error = fmt.Errorf("error saving configuration: %w", err)
			return m, nil
		}

		logger.Info("Configuration saved successfully")
		m.shouldProceed = true
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		if m.isLoading {
			switch msg.String() {
			case "ctrl+c", "esc":
				return m, tea.Quit
			}
			return m, nil
		}

		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit

		case "ctrl+r":
			m.cursorMode++
			if m.cursorMode > cursor.CursorHide {
				m.cursorMode = cursor.CursorBlink
			}
			cmds := make([]tea.Cmd, len(m.inputs))
			for i := range m.inputs {
				cmds[i] = m.inputs[i].Cursor.SetMode(m.cursorMode)
			}
			return m, tea.Batch(cmds...)

		case "tab", "shift+tab", "enter", "up", "down":
			s := msg.String()

			if s == "enter" && m.focusIndex == len(m.inputs) {
				return m.handleSubmit()
			}

			if s == "up" || s == "shift+tab" {
				m.focusIndex--
			} else {
				m.focusIndex++
			}

			if m.focusIndex > len(m.inputs) {
				m.focusIndex = 0
			} else if m.focusIndex < 0 {
				m.focusIndex = len(m.inputs)
			}

			cmds := make([]tea.Cmd, len(m.inputs))
			for i := 0; i <= len(m.inputs)-1; i++ {
				if i == m.focusIndex {
					cmds[i] = m.inputs[i].Focus()
					m.inputs[i].PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(styles.Primary))
					m.inputs[i].TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(styles.Text))
					continue
				}
				m.inputs[i].Blur()
				m.inputs[i].PromptStyle = lipgloss.NewStyle()
				m.inputs[i].TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(styles.Text))
			}

			return m, tea.Batch(cmds...)
		}
	}

	// Handle character input and blinking
	var cmd tea.Cmd
	if m.isLoading {
		// Only update spinner when loading
		m.spinner, cmd = m.spinner.Update(msg)
	} else {
		cmd = m.updateInputs(msg)
	}

	return m, cmd
}

// handleSubmit gère la soumission du formulaire
func (m Model) handleSubmit() (tea.Model, tea.Cmd) {
	serverURL := strings.TrimSpace(m.inputs[0].Value())
	username := strings.TrimSpace(m.inputs[1].Value())
	apiToken := strings.TrimSpace(m.inputs[2].Value())

	// Validation simple
	if serverURL == "" || username == "" || apiToken == "" {
		m.error = fmt.Errorf("please fill in all fields")
		for i := range m.inputs {
			m.inputs[i].PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(styles.Error))
			m.inputs[i].PlaceholderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(styles.Error))
		}
		return m, nil
	}

	// Store values and start loading/testing
	m.serverURL = serverURL
	m.username = username
	m.apiToken = apiToken
	m.isLoading = true
	m.error = nil

	// Start the spinner and perform the real login
	return m, tea.Batch(m.spinner.Tick, m.performLogin(serverURL, username, apiToken))
}

func (m *Model) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))

	// Only text inputs with Focus() set will respond, so it's safe to simply
	// update all of them here without any further logic.
	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

func (m Model) View() string {
	var content strings.Builder

	// Logo
	content.WriteString(styles.LogoStyle.Render(`
 _______  _______  _______ 
|       ||  _    ||       |
|    ___|| | |   ||  _____|
|   | __ | | |   || |_____ 
|   ||  || |_|   ||_____  |
|   |_| ||       | _____| |
|_______||_______||_______|
`))

	content.WriteString("\n\n")

	// Title
	content.WriteString(styles.TitleStyle.Render("g0s Authentication"))
	content.WriteString("\n")
	content.WriteString(styles.MutedStyle.Render("Connect to your g0s server"))
	content.WriteString("\n\n")

	// Form container
	var formContent strings.Builder

	if m.isLoading {
		// Show loading state
		formContent.WriteString(lipgloss.NewStyle().Align(lipgloss.Center).Render(
			fmt.Sprintf("%s Authenticating...", m.spinner.View()),
		))
		formContent.WriteString("\n\n")
		formContent.WriteString(styles.MutedStyle.Render(fmt.Sprintf("Server: %s", m.serverURL)))
		formContent.WriteString("\n")
		formContent.WriteString(styles.MutedStyle.Render(fmt.Sprintf("User: %s", m.username)))
	} else {
		// Show form inputs
		for i := range m.inputs {
			var inputStyle lipgloss.Style
			if i == m.focusIndex {
				inputStyle = styles.InputFocusedStyle
			} else {
				inputStyle = styles.InputStyle
			}

			formContent.WriteString(inputStyle.Render(m.inputs[i].View()))
			formContent.WriteString("\n")
		}

		// Submit button
		var button string
		if m.focusIndex == len(m.inputs) {
			button = styles.FormButtonFocusedStyle.Render("[ Authenticate ]")
		} else {
			button = styles.FormButtonStyle.Render("[ Authenticate ]")
		}
		formContent.WriteString("\n")
		formContent.WriteString(lipgloss.NewStyle().Align(lipgloss.Center).Render(button))

		// Error message
		if m.error != nil {
			formContent.WriteString("\n\n")
			formContent.WriteString(styles.ErrorStyle.Render(fmt.Sprintf("❌ %s", m.error.Error())))
		}
	}

	// Wrap form in container
	formContainer := styles.FormContainerStyle.Render(formContent.String())
	content.WriteString(lipgloss.NewStyle().Align(lipgloss.Center).Render(formContainer))

	content.WriteString("\n\n")

	// Help text
	var helpText string
	if m.isLoading {
		helpText = "Authenticating... • Ctrl+C to quit"
	} else {
		helpText = "Use Tab/Shift+Tab to navigate • Enter to submit • Ctrl+C to quit"
		if m.cursorMode != cursor.CursorBlink {
			helpText += fmt.Sprintf(" • Cursor mode: %s (Ctrl+R to change)", m.cursorMode.String())
		}
	}
	content.WriteString(styles.HelpTextStyle.Render(helpText))

	// Always center content and use full terminal dimensions
	return lipgloss.NewStyle().
		Align(lipgloss.Center).
		AlignVertical(lipgloss.Center).
		Width(m.width).
		Height(m.height).
		Render(content.String())
}

func (m Model) ShouldProceed() bool {
	return m.shouldProceed
}
