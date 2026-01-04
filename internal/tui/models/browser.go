package models

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/petergi/ebook-mechanic-cli/internal/tui/styles"
)

// FileItem represents a file or directory in the browser
type FileItem struct {
	Name  string
	Path  string
	IsDir bool
	Size  int64
}

// BrowserModel represents the file browser state
type BrowserModel struct {
	currentDir   string
	items        []FileItem
	selected     int
	width        int
	height       int
	errorMsg     string
	showHidden   bool
	filterExts   []string // Filter by extensions (.epub, .pdf)
	viewportTop  int      // For scrolling
	viewportSize int      // Number of visible items
}

// NewBrowserModel creates a new file browser starting at the given directory
func NewBrowserModel(startDir string) BrowserModel {
	if startDir == "" {
		startDir, _ = os.Getwd()
	}

	m := BrowserModel{
		currentDir:   startDir,
		selected:     0,
		width:        80,
		height:       24,
		showHidden:   false,
		filterExts:   []string{".epub", ".pdf"},
		viewportSize: 15,
	}

	m.loadDirectory()
	return m
}

// Init initializes the model
func (m BrowserModel) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model state
func (m BrowserModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewportSize = m.height - 10 // Reserve space for header/footer
		styles.AdaptToTerminal(m.width, m.height)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			m.selected--
			if m.selected < 0 {
				m.selected = len(m.items) - 1 // Wrap to bottom
			}
			m.updateViewport()

		case "down", "j":
			m.selected++
			if m.selected >= len(m.items) {
				m.selected = 0 // Wrap to top
			}
			m.updateViewport()

		case "enter":
			// If directory, navigate into it; if file, select it
			if m.selected >= 0 && m.selected < len(m.items) {
				item := m.items[m.selected]
				if item.IsDir {
					m.currentDir = item.Path
					m.selected = 0
					m.viewportTop = 0
					m.loadDirectory()
				} else {
					// File selected, return message
					return m, func() tea.Msg {
						return FileSelectMsg{Path: item.Path}
					}
				}
			}

		case "backspace", "h":
			// Go to parent directory
			parent := filepath.Dir(m.currentDir)
			if parent != m.currentDir {
				m.currentDir = parent
				m.selected = 0
				m.viewportTop = 0
				m.loadDirectory()
			}

		case ".":
			// Toggle hidden files
			m.showHidden = !m.showHidden
			m.selected = 0
			m.viewportTop = 0
			m.loadDirectory()

		case "esc", "q":
			// Go back to menu
			return m, func() tea.Msg {
				return BackToMenuMsg{}
			}

		case "ctrl+c":
			return m, tea.Quit
		}
	}

	return m, nil
}

// View renders the file browser
func (m BrowserModel) View() string {
	// Title showing current directory
	title := styles.RenderTitle("📂 File Browser")
	currentPath := styles.InfoStyle.Render("Current: " + m.currentDir)

	// Error message if any
	var errorDisplay string
	if m.errorMsg != "" {
		errorDisplay = styles.RenderError(m.errorMsg) + "\n"
	}

	// Render file list with scrolling
	var itemsDisplay string
	visibleStart := m.viewportTop
	visibleEnd := m.viewportTop + m.viewportSize
	if visibleEnd > len(m.items) {
		visibleEnd = len(m.items)
	}

	for i := visibleStart; i < visibleEnd; i++ {
		item := m.items[i]
		cursor := " "
		icon := "📄"
		if item.IsDir {
			icon = "📁"
		}

		line := icon + " " + item.Name
		if item.IsDir {
			line += "/"
		}

		if i == m.selected {
			cursor = styles.IconArrow
			itemsDisplay += styles.SelectedListItemStyle.Render(cursor+" "+line) + "\n"
		} else {
			itemsDisplay += styles.ListItemStyle.Render(cursor+" "+line) + "\n"
		}
	}

	// Scroll indicators
	var scrollIndicator string
	if len(m.items) > m.viewportSize {
		scrollIndicator = styles.MutedStyle.Render(
			lipgloss.PlaceHorizontal(
				m.width-4,
				lipgloss.Right,
				"↓ more ↓",
			),
		)
		if m.viewportTop > 0 {
			scrollIndicator = styles.MutedStyle.Render(
				lipgloss.PlaceHorizontal(
					m.width-4,
					lipgloss.Right,
					"↑ more ↑",
				),
			)
		}
	}

	// Help text
	help := styles.HelpStyle.Render(
		styles.RenderKeyBinding("↑/↓", "navigate") + "  " +
			styles.RenderKeyBinding("enter", "select/open") + "  " +
			styles.RenderKeyBinding("backspace", "parent dir") + "  " +
			styles.RenderKeyBinding(".", "toggle hidden") + "  " +
			styles.RenderKeyBinding("esc", "back"),
	)

	// File count and filter info
	filterInfo := styles.MutedStyle.Render(
		lipgloss.JoinHorizontal(
			lipgloss.Left,
			"Showing: ",
			styles.InfoStyle.Render(strings.Join(m.filterExts, ", ")),
			"  |  ",
			"Total: ",
			styles.InfoStyle.Render(lipgloss.PlaceHorizontal(4, lipgloss.Right, lipgloss.NewStyle().Render(lipgloss.PlaceHorizontal(4, lipgloss.Right, lipgloss.NewStyle().Render(string(rune(len(m.items)))))))),
		),
	)

	// Combine all parts
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		currentPath,
		"",
		errorDisplay,
		itemsDisplay,
		scrollIndicator,
		"",
		filterInfo,
		help,
	)

	return styles.DocStyle.
		Width(m.width).
		Height(m.height).
		Render(content)
}

// loadDirectory loads files and directories from the current directory
func (m *BrowserModel) loadDirectory() {
	m.items = []FileItem{}
	m.errorMsg = ""

	// Add parent directory entry if not at root
	if m.currentDir != "/" && m.currentDir != filepath.VolumeName(m.currentDir)+string(filepath.Separator) {
		m.items = append(m.items, FileItem{
			Name:  "..",
			Path:  filepath.Dir(m.currentDir),
			IsDir: true,
		})
	}

	entries, err := os.ReadDir(m.currentDir)
	if err != nil {
		m.errorMsg = "Error reading directory: " + err.Error()
		return
	}

	// Separate directories and files
	var dirs []FileItem
	var files []FileItem

	for _, entry := range entries {
		name := entry.Name()

		// Skip hidden files unless showHidden is true
		if !m.showHidden && strings.HasPrefix(name, ".") {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		item := FileItem{
			Name:  name,
			Path:  filepath.Join(m.currentDir, name),
			IsDir: entry.IsDir(),
			Size:  info.Size(),
		}

		if entry.IsDir() {
			dirs = append(dirs, item)
		} else {
			// Filter by extensions
			if m.matchesFilter(name) {
				files = append(files, item)
			}
		}
	}

	// Sort directories and files separately
	sort.Slice(dirs, func(i, j int) bool {
		return strings.ToLower(dirs[i].Name) < strings.ToLower(dirs[j].Name)
	})
	sort.Slice(files, func(i, j int) bool {
		return strings.ToLower(files[i].Name) < strings.ToLower(files[j].Name)
	})

	// Add sorted items (directories first, then files)
	m.items = append(m.items, dirs...)
	m.items = append(m.items, files...)
}

// matchesFilter checks if a filename matches the filter extensions
func (m *BrowserModel) matchesFilter(filename string) bool {
	if len(m.filterExts) == 0 {
		return true // No filter, show all files
	}

	ext := strings.ToLower(filepath.Ext(filename))
	for _, filterExt := range m.filterExts {
		if ext == strings.ToLower(filterExt) {
			return true
		}
	}
	return false
}

// updateViewport adjusts the viewport to keep the selected item visible
func (m *BrowserModel) updateViewport() {
	// If selected is above viewport, scroll up
	if m.selected < m.viewportTop {
		m.viewportTop = m.selected
	}

	// If selected is below viewport, scroll down
	if m.selected >= m.viewportTop+m.viewportSize {
		m.viewportTop = m.selected - m.viewportSize + 1
	}

	// Ensure viewport doesn't go negative
	if m.viewportTop < 0 {
		m.viewportTop = 0
	}
}

// SelectedPath returns the path of the currently selected item
func (m BrowserModel) SelectedPath() string {
	if m.selected >= 0 && m.selected < len(m.items) {
		return m.items[m.selected].Path
	}
	return ""
}

// FileSelectMsg is sent when a file is selected
type FileSelectMsg struct {
	Path string
}

// BackToMenuMsg is sent when the user wants to go back to the menu
type BackToMenuMsg struct{}
