package models

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/theotruvelot/g0s/internal/cli/pages/loading"
	"github.com/theotruvelot/g0s/internal/cli/styles"
	"github.com/theotruvelot/g0s/pkg/client"
	"go.uber.org/zap"
)

// RootModel manages the overall application state and navigation
type RootModel struct {
	config      AppConfig
	currentPage PageType
	pages       map[PageType]PageModel
	width       int
	height      int
	ready       bool
	err         error
	log         *zap.Logger
}

// NewRootModel creates a new root model with the given HTTP client and logger
func NewRootModel(httpClient *client.Client, log *zap.Logger) RootModel {
	config := AppConfig{
		HTTPClient: httpClient,
		Logger:     log,
	}

	model := RootModel{
		config:      config,
		currentPage: LoadingPage,
		pages:       make(map[PageType]PageModel),
		log:         log,
	}

	// Initialize pages
	model.initializePages()

	return model
}

// initializePages creates all page models
func (m *RootModel) initializePages() {
	// Initialize loading page with logger
	m.pages[LoadingPage] = loading.NewModel(m.config.HTTPClient, m.config.Logger)

	// Initialize dashboard page with logger

	// Add other pages here as they are created
	// m.pages[SettingsPage] = settings.NewModel(m.config)
}

// Init initializes the root model
func (m RootModel) Init() tea.Cmd {
	m.log.Debug("Initializing root model")

	// Start with the loading page
	if page, exists := m.pages[m.currentPage]; exists {
		return page.OnEnter()
	}

	return nil
}

// Update handles messages and updates the model
func (m RootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true

		// Update all pages with new size
		for pageType, page := range m.pages {
			updatedPage, cmd := page.Update(msg)
			m.pages[pageType] = updatedPage.(PageModel)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.currentPage != LoadingPage {
				m.log.Info("User requested exit")
				return m, tea.Quit
			}
		case "esc":
			if m.currentPage != LoadingPage {
				return m.navigateToPage(LoadingPage, nil)
			}
		}

		// Pass key message to current page
		if page, exists := m.pages[m.currentPage]; exists {
			updatedPage, pageCmd := page.Update(msg)
			m.pages[m.currentPage] = updatedPage.(PageModel)
			cmds = append(cmds, pageCmd)
		}

	case NavigateMsg:
		return m.navigateToPage(msg.Page, msg.Data)

	case ErrorMsg:
		m.log.Error("Received error message",
			zap.String("message", msg.Message),
			zap.Error(msg.Err),
			zap.Bool("fatal", msg.Fatal))

		if msg.Fatal {
			m.err = msg.Err
			return m, tea.Quit
		}

		// Handle non-fatal errors by staying on current page or navigating
		if page, exists := m.pages[m.currentPage]; exists {
			updatedPage, pageCmd := page.Update(msg)
			m.pages[m.currentPage] = updatedPage.(PageModel)
			cmds = append(cmds, pageCmd)
		}

	case HealthCheckMsg:

		if msg.Success {
			// Navigate to dashboard after successful health check
		}

		// Pass to current page to handle failed health check
		if page, exists := m.pages[m.currentPage]; exists {
			updatedPage, pageCmd := page.Update(msg)
			m.pages[m.currentPage] = updatedPage.(PageModel)
			cmds = append(cmds, pageCmd)
		}

	default:
		if page, exists := m.pages[m.currentPage]; exists {
			updatedPage, pageCmd := page.Update(msg)
			m.pages[m.currentPage] = updatedPage.(PageModel)
			cmds = append(cmds, pageCmd)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m RootModel) navigateToPage(pageType PageType, data interface{}) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	m.log.Debug("Navigating to page",
		zap.String("from", m.currentPage.String()),
		zap.String("to", pageType.String()))

	// Call OnExit for current page
	if currentPage, exists := m.pages[m.currentPage]; exists {
		if exitCmd := currentPage.OnExit(); exitCmd != nil {
			cmds = append(cmds, exitCmd)
		}
	}

	m.currentPage = pageType

	if newPage, exists := m.pages[pageType]; exists {
		if enterCmd := newPage.OnEnter(); enterCmd != nil {
			cmds = append(cmds, enterCmd)
		}
	} else {
		m.log.Warn("Attempted to navigate to non-existent page",
			zap.String("page", pageType.String()))
	}

	return m, tea.Batch(cmds...)
}

// View renders the current page
func (m RootModel) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}

	// Render current page
	if page, exists := m.pages[m.currentPage]; exists {
		return page.View()
	}

	// Fallback view
	return styles.ErrorStyle.Render(
		fmt.Sprintf("Page not found: %s", m.currentPage.String()),
	)
}

// HasError returns true if the model has an error
func (m RootModel) HasError() bool {
	return m.err != nil
}

// GetError returns the current error
func (m RootModel) GetError() string {
	if m.err != nil {
		return m.err.Error()
	}
	return ""
}
