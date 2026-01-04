package models

import (
	"fmt"
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
	currentDir    string
	items         []FileItem
	selected      int
	selectedFiles map[string]bool // Multi-select support: path -> selected
	width         int
	height        int
	errorMsg      string
	showHidden    bool
	filterExts    []string // Filter by extensions (.epub, .pdf)
	mode          BrowserMode
	viewportTop   int // For scrolling
	viewportSize  int // Number of visible items
}

// BrowserMode configures how selection behaves.
type BrowserMode int

const (
	// BrowserModeFile selects individual files and navigates directories.
	BrowserModeFile BrowserMode = iota
	// BrowserModeBatch selects directories or files for batch operations.
	BrowserModeBatch
	// BrowserModeMultiSelect selects multiple files for operations.
	BrowserModeMultiSelect
)

// NewBrowserModel creates a new file browser starting at the given directory
func NewBrowserModel(startDir string, width, height int) BrowserModel {
	return newBrowserModel(startDir, BrowserModeFile, width, height)
}

// NewBatchBrowserModel creates a browser configured for batch selection.
func NewBatchBrowserModel(startDir string, width, height int) BrowserModel {
	return newBrowserModel(startDir, BrowserModeBatch, width, height)
}

// NewMultiSelectBrowserModel creates a browser configured for multiple file selection.
func NewMultiSelectBrowserModel(startDir string, width, height int) BrowserModel {
	return newBrowserModel(startDir, BrowserModeMultiSelect, width, height)
}

func newBrowserModel(startDir string, mode BrowserMode, width, height int) BrowserModel {
	if startDir == "" {
		startDir, _ = os.Getwd()
	}

	if width == 0 {
		width = 80
	}
	if height == 0 {
		height = 24
	}

	// Calculate viewport size to fit all UI chrome
	// Title (1) + pathBox (2) + fileListBox borders (2) + filterInfo (2) + helpBox (2) + spacing (3) = 12 lines overhead
	viewportSize := height - 15
	if viewportSize < 5 {
		viewportSize = 5 // Minimum 5 items visible
	}

	m := BrowserModel{
		currentDir:    startDir,
		selected:      0,
		selectedFiles: make(map[string]bool),
		width:         width,
		height:        height,
		showHidden:    false,
		filterExts:    []string{".epub", ".pdf"},
		mode:          mode,
		viewportSize:  viewportSize,
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
		// Recalculate viewport size to fit UI chrome
		m.viewportSize = m.height - 15
		if m.viewportSize < 5 {
			m.viewportSize = 5
		}
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
					if m.mode == BrowserModeBatch {
						return m, func() tea.Msg {
							return DirectorySelectMsg{Path: item.Path}
						}
					}
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

		case "right", "l":
			if m.mode == BrowserModeBatch {
				if m.selected >= 0 && m.selected < len(m.items) {
					item := m.items[m.selected]
					if item.IsDir {
						m.currentDir = item.Path
						m.selected = 0
						m.viewportTop = 0
						m.loadDirectory()
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

		case " ":
			// Toggle selection of current item
			if m.selected >= 0 && m.selected < len(m.items) {
				item := m.items[m.selected]
				if m.selectedFiles[item.Path] {
					delete(m.selectedFiles, item.Path)
				} else {
					m.selectedFiles[item.Path] = true
				}
				// Move to next item after toggling
				m.selected++
				if m.selected >= len(m.items) {
					m.selected = len(m.items) - 1
				}
				m.updateViewport()
			}

		case "a", "ctrl+a":
			// Select all files (not directories)
			for _, item := range m.items {
				if !item.IsDir {
					m.selectedFiles[item.Path] = true
				}
			}

		case "A":
			// Deselect all
			m.selectedFiles = make(map[string]bool)

		case "s":
			// Submit selected files
			if len(m.selectedFiles) > 0 {
				paths := make([]string, 0, len(m.selectedFiles))
				for path := range m.selectedFiles {
					paths = append(paths, path)
				}
				return m, func() tea.Msg {
					return MultiFileSelectMsg{Paths: paths}
				}
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
	// Title and current path header
	title := styles.RenderTitle("📂 File Browser")

	// Current path in a subtle box
	pathBox := lipgloss.NewStyle().
		Foreground(styles.ColorInfo).
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(styles.ColorMuted).
		Padding(0, 1).
		Width(m.width - 8).
		Render(m.currentDir)

	// Error message if any
	var errorDisplay string
	if m.errorMsg != "" {
		errorDisplay = lipgloss.NewStyle().
			Foreground(styles.ColorError).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(styles.ColorError).
			Padding(0, 1).
			Width(m.width - 8).
			Render(styles.IconCross + " " + m.errorMsg)
	}

	// Render file list with scrolling in a bordered box
	var itemsDisplay string
	visibleStart := m.viewportTop
	visibleEnd := m.viewportTop + m.viewportSize
	if visibleEnd > len(m.items) {
		visibleEnd = len(m.items)
	}

	for i := visibleStart; i < visibleEnd; i++ {
		item := m.items[i]
		cursor := "  "
		icon := "📄"
		checkbox := ""

		if item.IsDir {
			icon = "📁"
		}

		// Show checkbox if in multi-select mode
		if m.mode == BrowserModeMultiSelect {
			if m.selectedFiles[item.Path] {
				checkbox = "[✓] "
			} else if !item.IsDir {
				checkbox = "[ ] "
			} else {
				checkbox = "    " // Indent directories to match files
			}
		}

		line := checkbox + icon + " " + item.Name
		if item.IsDir {
			line += "/"
		}

		if i == m.selected {
			cursor = styles.IconArrow + " "
			itemsDisplay += styles.SelectedListItemStyle.Render(cursor+line) + "\n"
		} else {
			itemsDisplay += styles.ListItemStyle.Render(cursor+line) + "\n"
		}
	}

	// Wrap file list in bordered box
	fileListBox := styles.BorderStyle.
		Width(m.width - 8).
		Height(m.viewportSize + 2).
		Render(itemsDisplay)

	// Scroll indicators
	var scrollInfo string
	if len(m.items) > m.viewportSize {
		if m.viewportTop > 0 && m.viewportTop+m.viewportSize < len(m.items) {
			scrollInfo = styles.MutedStyle.Render("↑ more above and below ↓")
		} else if m.viewportTop > 0 {
			scrollInfo = styles.MutedStyle.Render("↑ more above")
		} else {
			scrollInfo = styles.MutedStyle.Render("↓ more below")
		}
	}

	// File count, filter info, and selection count in a status bar
	statusParts := []string{
		"Showing: ",
		lipgloss.NewStyle().Foreground(styles.ColorInfo).Render(strings.Join(m.filterExts, ", ")),
		"  │  ",
		"Items: ",
		lipgloss.NewStyle().Foreground(styles.ColorInfo).Render(fmt.Sprintf("%d", len(m.items))),
	}
	if len(m.selectedFiles) > 0 {
		statusParts = append(statusParts,
			"  │  ",
			"Selected: ",
			lipgloss.NewStyle().Foreground(styles.ColorSuccess).Render(fmt.Sprintf("%d", len(m.selectedFiles))),
		)
	}

	filterInfo := lipgloss.NewStyle().
		Foreground(styles.ColorMuted).
		Border(lipgloss.NormalBorder(), true, false, false, false).
		BorderForeground(styles.ColorMuted).
		Padding(0, 1).
		Width(m.width - 8).
		Render(lipgloss.JoinHorizontal(lipgloss.Left, statusParts...))

	// Help text
	helpBox := lipgloss.NewStyle().
		Foreground(styles.ColorMuted).
		Border(lipgloss.NormalBorder(), true, false, false, false).
		BorderForeground(styles.ColorMuted).
		Padding(1, 1).
		Width(m.width - 8).
		Render(m.helpText())

	// Combine all parts
	var parts []string
	parts = append(parts, title, pathBox)
	if errorDisplay != "" {
		parts = append(parts, "", errorDisplay)
	}
	parts = append(parts, "", fileListBox)
	if scrollInfo != "" {
		parts = append(parts, scrollInfo)
	}
	parts = append(parts, filterInfo, helpBox)

	content := lipgloss.JoinVertical(lipgloss.Left, parts...)

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Top,
		content,
	)
}

func (m BrowserModel) helpText() string {
	line1 := styles.RenderKeyBinding("↑/↓/j/k", "navigate") + "  " +
		styles.RenderKeyBinding("space", "toggle select") + "  " +
		styles.RenderKeyBinding("a", "select all") + "  " +
		styles.RenderKeyBinding("A", "deselect all")

	line2 := ""
	if m.mode == BrowserModeBatch {
		line2 = styles.RenderKeyBinding("enter", "select dir") + "  " +
			styles.RenderKeyBinding("l/right", "open dir") + "  " +
			styles.RenderKeyBinding("backspace/h", "parent") + "  " +
			styles.RenderKeyBinding("s", "submit") + "  " +
			styles.RenderKeyBinding("esc", "back")
	} else {
		line2 = styles.RenderKeyBinding("enter", "select/open") + "  " +
			styles.RenderKeyBinding("backspace/h", "parent") + "  " +
			styles.RenderKeyBinding("s", "submit selected") + "  " +
			styles.RenderKeyBinding(".", "toggle hidden") + "  " +
			styles.RenderKeyBinding("esc", "back")
	}

	return line1 + "\n" + line2
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

// MultiFileSelectMsg is sent when multiple files are selected
type MultiFileSelectMsg struct {
	Paths []string
}

// DirectorySelectMsg is sent when a directory is selected for batch operations.
type DirectorySelectMsg struct {
	Path string
}

// BackToMenuMsg is sent when the user wants to go back to the menu
type BackToMenuMsg struct{}
