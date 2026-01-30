package models

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/petergi/ebook-mechanic-cli/internal/tui/styles"
)

// CalibreBook represents a book folder in a Calibre library
type CalibreBook struct {
	AuthorPath string   // Full path to author directory
	Author     string   // Author name
	BookPath   string   // Full path to book directory
	BookTitle  string   // Book title (folder name)
	EbookFiles []string // List of ebook files found
	HasCover   bool     // Has cover.jpg/jpeg/png
	HasOPF     bool     // Has metadata.opf
}

// CalibreAuthor represents an author in a Calibre library
type CalibreAuthor struct {
	Path  string        // Full path to author directory
	Name  string        // Author name
	Books []CalibreBook // Books by this author
}

// CalibreScanResult holds the results of a Calibre library scan
type CalibreScanResult struct {
	LibraryPath       string          // Root path of the library
	Authors           []CalibreAuthor // All authors found
	BooksWithoutFiles []CalibreBook   // Books with no ebook files (metadata only)
	BooksWithoutMeta  []CalibreBook   // Books with ebook files but no metadata
	EmptyAuthors      []string        // Author directories with no book subdirectories
	TotalAuthors      int
	TotalBooks        int
	ScanDuration      time.Duration
	CleanedDirs       []string // Directories that were cleaned up
	CleanedFiles      []string // Files that were removed during cleanup
	CleanupError      error    // Error during cleanup if any
}

// CalibreModel handles the Calibre library cleanup workflow
type CalibreModel struct {
	libraryPath    string
	state          calibreState
	scanResult     *CalibreScanResult
	selectedFilter int // 0: without files, 1: without metadata
	viewportTop    int
	viewportSize   int
	width          int
	height         int
	spinner        int
	scanning       bool
	scanProgress   int
	scanTotal      int
	currentAuthor  string
	cleanupDone    bool // Cleanup completed
}

type calibreState int

const (
	calibreStateScanning calibreState = iota
	calibreStateResults
	calibreStateCleanupConfirm
	calibreStateCleanupProgress
	calibreStateCleanupDone
)

// CalibreScanMsg signals that a Calibre library scan should start
type CalibreScanMsg struct {
	LibraryPath string
}

// CalibreScanProgressMsg reports scan progress
type CalibreScanProgressMsg struct {
	Current int
	Total   int
	Author  string
}

// CalibreScanCompleteMsg signals scan completion
type CalibreScanCompleteMsg struct {
	Result *CalibreScanResult
}

// CalibreCleanupMsg signals that cleanup should be performed
type CalibreCleanupMsg struct{}

// CalibreCleanupCompleteMsg signals cleanup completion
type CalibreCleanupCompleteMsg struct {
	CleanedDirs  []string
	CleanedFiles []string
	Error        error
}

// NewCalibreModel creates a new Calibre library cleanup model
func NewCalibreModel(libraryPath string, width, height int) CalibreModel {
	if width == 0 {
		width = 80
	}
	if height == 0 {
		height = 24
	}

	return CalibreModel{
		libraryPath:  libraryPath,
		state:        calibreStateScanning,
		viewportSize: height - 20,
		width:        width,
		height:       height,
		scanning:     true,
	}
}

// Init initializes the model
func (m CalibreModel) Init() tea.Cmd {
	return tea.Batch(
		m.tick(),
		m.startScan(),
	)
}

func (m CalibreModel) tick() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return TickMsg{}
	})
}

func (m CalibreModel) startScan() tea.Cmd {
	return func() tea.Msg {
		return CalibreScanMsg{LibraryPath: m.libraryPath}
	}
}

// Update handles messages and updates the model state
func (m CalibreModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewportSize = msg.Height - 20
		styles.AdaptToTerminal(m.width, m.height)
		return m, nil

	case TickMsg:
		if m.scanning {
			m.spinner = (m.spinner + 1) % len(styles.IconSpinner)
			return m, m.tick()
		}
		return m, nil

	case CalibreScanProgressMsg:
		m.scanProgress = msg.Current
		m.scanTotal = msg.Total
		m.currentAuthor = msg.Author
		return m, nil

	case CalibreScanCompleteMsg:
		m.scanning = false
		m.scanResult = msg.Result
		m.state = calibreStateResults
		return m, nil

	case CalibreCleanupCompleteMsg:
		m.scanResult.CleanedDirs = msg.CleanedDirs
		m.scanResult.CleanedFiles = msg.CleanedFiles
		m.scanResult.CleanupError = msg.Error
		m.state = calibreStateCleanupDone
		m.cleanupDone = true
		return m, nil

	case tea.KeyMsg:
		switch m.state {
		case calibreStateResults:
			return m.handleResultsKeys(msg)
		case calibreStateCleanupConfirm:
			return m.handleCleanupConfirmKeys(msg)
		case calibreStateCleanupDone:
			return m.handleCleanupDoneKeys(msg)
		default:
			// During scanning, only allow cancel
			if msg.String() == "ctrl+c" || msg.String() == "q" {
				return m, func() tea.Msg {
					return BackToMenuMsg{}
				}
			}
		}
	}

	return m, nil
}

func (m CalibreModel) handleResultsKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "tab", "left", "right":
		// Toggle between filters (3 categories now)
		m.selectedFilter = (m.selectedFilter + 1) % 3
		m.viewportTop = 0
		return m, nil

	case "up", "k":
		if m.viewportTop > 0 {
			m.viewportTop--
		}
		return m, nil

	case "down", "j":
		items := m.currentItems()
		if m.viewportTop < len(items)-m.viewportSize {
			m.viewportTop++
		}
		return m, nil

	case "c":
		// Show cleanup confirmation
		if len(m.scanResult.BooksWithoutFiles) > 0 || len(m.scanResult.EmptyAuthors) > 0 {
			m.state = calibreStateCleanupConfirm
		}
		return m, nil

	case "s":
		// Save report
		return m, m.saveReport()

	case "esc", "q":
		return m, func() tea.Msg {
			return BackToMenuMsg{}
		}
	}

	return m, nil
}

func (m CalibreModel) handleCleanupConfirmKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		m.state = calibreStateCleanupProgress
		return m, func() tea.Msg {
			return CalibreCleanupMsg{}
		}
	case "n", "N", "esc":
		m.state = calibreStateResults
		return m, nil
	}
	return m, nil
}

func (m CalibreModel) handleCleanupDoneKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter", "esc", "q":
		return m, func() tea.Msg {
			return BackToMenuMsg{}
		}
	}
	return m, nil
}

func (m CalibreModel) currentItems() []CalibreBook {
	if m.scanResult == nil {
		return nil
	}
	if m.selectedFilter == 0 {
		return m.scanResult.BooksWithoutFiles
	} else if m.selectedFilter == 1 {
		return m.scanResult.BooksWithoutMeta
	}
	// Filter 2 is empty authors - handled separately in renderResults
	return nil
}

// currentEmptyAuthors returns the list of empty author directories for display
func (m CalibreModel) currentEmptyAuthors() []string {
	if m.scanResult == nil {
		return nil
	}
	return m.scanResult.EmptyAuthors
}

// GetScanResult returns the scan result for cleanup operations
func (m CalibreModel) GetScanResult() *CalibreScanResult {
	return m.scanResult
}

func (m CalibreModel) saveReport() tea.Cmd {
	return func() tea.Msg {
		if m.scanResult == nil {
			return nil
		}

		// Generate report filename
		timestamp := time.Now().Format("20060102-150405")
		filename := fmt.Sprintf("calibre-library-report-%s.md", timestamp)

		f, err := os.Create(filename)
		if err != nil {
			return nil
		}
		defer func() { _ = f.Close() }()

		// Write markdown report
		_, _ = fmt.Fprintf(f, "# Calibre Library Cleanup Report\n\n")
		_, _ = fmt.Fprintf(f, "**Library Path:** `%s`\n\n", m.scanResult.LibraryPath)
		_, _ = fmt.Fprintf(f, "**Scan Date:** %s\n\n", time.Now().Format("2006-01-02 15:04:05"))
		_, _ = fmt.Fprintf(f, "**Scan Duration:** %s\n\n", m.scanResult.ScanDuration.Round(time.Millisecond))

		_, _ = fmt.Fprintf(f, "## Summary\n\n")
		_, _ = fmt.Fprintf(f, "| Metric | Count |\n")
		_, _ = fmt.Fprintf(f, "|--------|-------|\n")
		_, _ = fmt.Fprintf(f, "| Total Authors | %d |\n", m.scanResult.TotalAuthors)
		_, _ = fmt.Fprintf(f, "| Total Books | %d |\n", m.scanResult.TotalBooks)
		_, _ = fmt.Fprintf(f, "| Books without ebook files | %d |\n", len(m.scanResult.BooksWithoutFiles))
		_, _ = fmt.Fprintf(f, "| Books without metadata | %d |\n", len(m.scanResult.BooksWithoutMeta))
		_, _ = fmt.Fprintf(f, "| Empty author directories | %d |\n", len(m.scanResult.EmptyAuthors))
		_, _ = fmt.Fprintf(f, "\n")

		if len(m.scanResult.BooksWithoutFiles) > 0 {
			_, _ = fmt.Fprintf(f, "## Books Without Ebook Files (Metadata Only)\n\n")
			_, _ = fmt.Fprintf(f, "These folders contain metadata files but no actual ebook files.\n\n")
			_, _ = fmt.Fprintf(f, "| Author | Book Title | Path |\n")
			_, _ = fmt.Fprintf(f, "|--------|------------|------|\n")
			for _, book := range m.scanResult.BooksWithoutFiles {
				_, _ = fmt.Fprintf(f, "| %s | %s | `%s` |\n", book.Author, book.BookTitle, book.BookPath)
			}
			_, _ = fmt.Fprintf(f, "\n")
		}

		if len(m.scanResult.BooksWithoutMeta) > 0 {
			_, _ = fmt.Fprintf(f, "## Books Without Metadata Files\n\n")
			_, _ = fmt.Fprintf(f, "These folders contain ebook files but are missing metadata (metadata.opf, cover).\n\n")
			_, _ = fmt.Fprintf(f, "| Author | Book Title | Ebook Files | Has Cover | Has OPF |\n")
			_, _ = fmt.Fprintf(f, "|--------|------------|-------------|-----------|----------|\n")
			for _, book := range m.scanResult.BooksWithoutMeta {
				files := strings.Join(book.EbookFiles, ", ")
				cover := "❌"
				if book.HasCover {
					cover = "✓"
				}
				opf := "❌"
				if book.HasOPF {
					opf = "✓"
				}
				_, _ = fmt.Fprintf(f, "| %s | %s | %s | %s | %s |\n", book.Author, book.BookTitle, files, cover, opf)
			}
			_, _ = fmt.Fprintf(f, "\n")
		}

		if len(m.scanResult.EmptyAuthors) > 0 {
			_, _ = fmt.Fprintf(f, "## Empty Author Directories\n\n")
			_, _ = fmt.Fprintf(f, "These author folders contain no book subfolders.\n\n")
			for _, authorPath := range m.scanResult.EmptyAuthors {
				_, _ = fmt.Fprintf(f, "- `%s`\n", authorPath)
			}
			_, _ = fmt.Fprintf(f, "\n")
		}

		if len(m.scanResult.CleanedDirs) > 0 {
			_, _ = fmt.Fprintf(f, "## Cleanup Actions\n\n")
			_, _ = fmt.Fprintf(f, "**Directories Removed:** %d\n\n", len(m.scanResult.CleanedDirs))
			for _, dir := range m.scanResult.CleanedDirs {
				_, _ = fmt.Fprintf(f, "- `%s`\n", dir)
			}
			_, _ = fmt.Fprintf(f, "\n")
		}

		return nil
	}
}

// View renders the model
func (m CalibreModel) View() string {
	switch m.state {
	case calibreStateScanning:
		return m.renderScanning()
	case calibreStateResults:
		return m.renderResults()
	case calibreStateCleanupConfirm:
		return m.renderCleanupConfirm()
	case calibreStateCleanupProgress:
		return m.renderCleanupProgress()
	case calibreStateCleanupDone:
		return m.renderCleanupDone()
	}
	return ""
}

func (m CalibreModel) renderScanning() string {
	spinnerChar := string(styles.IconSpinner[m.spinner])

	title := styles.RenderTitle("📚 Calibre Library Cleanup")
	subtitle := styles.RenderSubtitle("Scanning library structure...")

	var progress string
	if m.scanTotal > 0 {
		progress = fmt.Sprintf("%s Scanning author %d of %d: %s",
			spinnerChar, m.scanProgress, m.scanTotal, m.currentAuthor)
	} else {
		progress = fmt.Sprintf("%s Discovering authors...", spinnerChar)
	}

	progressBox := styles.BorderStyle.
		Width(60).
		Render(progress)

	help := styles.MutedStyle.Render("Press Ctrl+C to cancel")

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		subtitle,
		"",
		progressBox,
		"",
		help,
	)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

func (m CalibreModel) renderResults() string {
	if m.scanResult == nil {
		return "No results"
	}

	title := styles.RenderTitle("📚 Calibre Library Cleanup")

	// Summary stats
	summary := fmt.Sprintf(
		"Library: %s\n"+
			"Authors: %d  |  Books: %d  |  Scan time: %s\n"+
			"\n"+
			"📁 Books without ebook files: %d\n"+
			"📋 Books without metadata: %d\n"+
			"🗂️  Empty author folders: %d",
		m.scanResult.LibraryPath,
		m.scanResult.TotalAuthors,
		m.scanResult.TotalBooks,
		m.scanResult.ScanDuration.Round(time.Millisecond),
		len(m.scanResult.BooksWithoutFiles),
		len(m.scanResult.BooksWithoutMeta),
		len(m.scanResult.EmptyAuthors),
	)

	summaryBox := styles.BorderStyle.Width(70).Render(summary)

	// Tab selector (3 tabs now)
	tab1 := "Without Files"
	tab2 := "Without Metadata"
	tab3 := "Empty Authors"
	switch m.selectedFilter {
	case 0:
		tab1 = styles.SelectedListItemStyle.Render("[" + tab1 + "]")
		tab2 = styles.MutedStyle.Render(" " + tab2 + " ")
		tab3 = styles.MutedStyle.Render(" " + tab3 + " ")
	case 1:
		tab1 = styles.MutedStyle.Render(" " + tab1 + " ")
		tab2 = styles.SelectedListItemStyle.Render("[" + tab2 + "]")
		tab3 = styles.MutedStyle.Render(" " + tab3 + " ")
	case 2:
		tab1 = styles.MutedStyle.Render(" " + tab1 + " ")
		tab2 = styles.MutedStyle.Render(" " + tab2 + " ")
		tab3 = styles.SelectedListItemStyle.Render("[" + tab3 + "]")
	}
	tabs := tab1 + "  " + tab2 + "  " + tab3

	// Items list - different handling for empty authors
	var itemsContent string
	if m.selectedFilter == 2 {
		// Empty authors view
		emptyAuthors := m.currentEmptyAuthors()
		if len(emptyAuthors) == 0 {
			itemsContent = styles.SuccessStyle.Render("✓ No empty author folders found!")
		} else {
			start := m.viewportTop
			end := start + m.viewportSize
			if end > len(emptyAuthors) {
				end = len(emptyAuthors)
			}

			for i := start; i < end; i++ {
				// Just show the author name (last component of path)
				authorName := filepath.Base(emptyAuthors[i])
				itemsContent += authorName + "\n"
			}

			if len(emptyAuthors) > m.viewportSize {
				itemsContent += styles.MutedStyle.Render(fmt.Sprintf("\n... showing %d-%d of %d", start+1, end, len(emptyAuthors)))
			}
		}
	} else {
		// Books view
		items := m.currentItems()
		if len(items) == 0 {
			itemsContent = styles.SuccessStyle.Render("✓ No issues found in this category!")
		} else {
			start := m.viewportTop
			end := start + m.viewportSize
			if end > len(items) {
				end = len(items)
			}

			for i := start; i < end; i++ {
				book := items[i]
				line := fmt.Sprintf("%s / %s", book.Author, book.BookTitle)
				if m.selectedFilter == 1 && len(book.EbookFiles) > 0 {
					line += fmt.Sprintf(" [%s]", strings.Join(book.EbookFiles, ", "))
				}
				itemsContent += line + "\n"
			}

			if len(items) > m.viewportSize {
				itemsContent += styles.MutedStyle.Render(fmt.Sprintf("\n... showing %d-%d of %d", start+1, end, len(items)))
			}
		}
	}

	listBox := styles.BorderStyle.Width(70).Height(m.viewportSize + 2).Render(itemsContent)

	// Help
	helpParts := []string{
		styles.RenderKeyBinding("tab", "switch category"),
		styles.RenderKeyBinding("↑/↓", "scroll"),
		styles.RenderKeyBinding("s", "save report"),
	}
	if len(m.scanResult.BooksWithoutFiles) > 0 || len(m.scanResult.EmptyAuthors) > 0 {
		helpParts = append(helpParts, styles.RenderKeyBinding("c", "cleanup"))
	}
	helpParts = append(helpParts, styles.RenderKeyBinding("q", "back"))
	help := strings.Join(helpParts, "  ")

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		summaryBox,
		"",
		tabs,
		listBox,
		"",
		help,
	)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

func (m CalibreModel) renderCleanupConfirm() string {
	title := styles.RenderTitle("⚠️  Confirm Cleanup")

	bookCount := len(m.scanResult.BooksWithoutFiles)
	authorCount := len(m.scanResult.EmptyAuthors)

	var warningParts []string
	if bookCount > 0 {
		warningParts = append(warningParts, fmt.Sprintf(
			"• %d book folder(s) that contain only metadata files\n  (no actual ebook files)",
			bookCount,
		))
	}
	if authorCount > 0 {
		warningParts = append(warningParts, fmt.Sprintf(
			"• %d empty author folder(s) with no books",
			authorCount,
		))
	}

	warning := fmt.Sprintf(
		"This will permanently delete:\n\n%s\n\nThis action cannot be undone!",
		strings.Join(warningParts, "\n\n"),
	)

	warningBox := lipgloss.NewStyle().
		Foreground(styles.ColorWarning).
		Border(lipgloss.RoundedBorder()).
		Padding(1, 2).
		Width(50).
		Render(warning)

	help := styles.RenderKeyBinding("y", "yes, delete") + "  " +
		styles.RenderKeyBinding("n", "no, cancel")

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		"",
		warningBox,
		"",
		help,
	)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

func (m CalibreModel) renderCleanupProgress() string {
	title := styles.RenderTitle("🧹 Cleaning Up...")
	spinnerChar := string(styles.IconSpinner[m.spinner])
	progress := fmt.Sprintf("%s Removing metadata-only folders...", spinnerChar)

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		"",
		progress,
	)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

func (m CalibreModel) renderCleanupDone() string {
	title := styles.RenderTitle("✅ Cleanup Complete")

	var resultText string
	if m.scanResult.CleanupError != nil {
		resultText = styles.ErrorStyle.Render(fmt.Sprintf("Error: %v", m.scanResult.CleanupError))
	} else {
		resultText = fmt.Sprintf(
			"Removed %d directories\nRemoved %d files",
			len(m.scanResult.CleanedDirs),
			len(m.scanResult.CleanedFiles),
		)
	}

	resultBox := styles.BorderStyle.Width(50).Render(resultText)

	help := styles.RenderKeyBinding("enter", "back to menu")

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		"",
		resultBox,
		"",
		help,
	)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

// ScanCalibreLibrary scans a Calibre library and returns the results
func ScanCalibreLibrary(libraryPath string, progressCh chan<- CalibreScanProgressMsg) *CalibreScanResult {
	startTime := time.Now()
	result := &CalibreScanResult{
		LibraryPath: libraryPath,
	}

	// Supported ebook extensions
	ebookExts := map[string]bool{
		".epub": true, ".pdf": true, ".mobi": true, ".azw": true, ".azw3": true,
		".azw4": true, ".kfx": true, ".cbz": true, ".cbr": true, ".fb2": true,
		".lit": true, ".pdb": true, ".txt": true, ".rtf": true, ".djvu": true,
	}

	// First, discover all author directories
	authorDirs, err := os.ReadDir(libraryPath)
	if err != nil {
		return result
	}

	var authors []string
	for _, entry := range authorDirs {
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			authors = append(authors, entry.Name())
		}
	}

	result.TotalAuthors = len(authors)

	// Scan each author
	for i, authorName := range authors {
		authorPath := filepath.Join(libraryPath, authorName)

		if progressCh != nil {
			progressCh <- CalibreScanProgressMsg{
				Current: i + 1,
				Total:   len(authors),
				Author:  authorName,
			}
		}

		author := CalibreAuthor{
			Path: authorPath,
			Name: authorName,
		}

		// Scan book directories within author
		bookDirs, err := os.ReadDir(authorPath)
		if err != nil {
			continue
		}

		// Filter to only book subdirectories
		var bookSubdirs []os.DirEntry
		for _, entry := range bookDirs {
			if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
				bookSubdirs = append(bookSubdirs, entry)
			}
		}

		// Check if author directory is empty (no book subdirectories)
		if len(bookSubdirs) == 0 {
			result.EmptyAuthors = append(result.EmptyAuthors, authorPath)
			result.Authors = append(result.Authors, author)
			continue
		}

		for _, bookEntry := range bookSubdirs {

			bookPath := filepath.Join(authorPath, bookEntry.Name())
			book := CalibreBook{
				AuthorPath: authorPath,
				Author:     authorName,
				BookPath:   bookPath,
				BookTitle:  bookEntry.Name(),
			}

			// Scan book contents
			bookContents, err := os.ReadDir(bookPath)
			if err != nil {
				continue
			}

			for _, item := range bookContents {
				if item.IsDir() {
					continue
				}

				name := strings.ToLower(item.Name())
				ext := strings.ToLower(filepath.Ext(item.Name()))

				// Check for ebook files
				if ebookExts[ext] {
					book.EbookFiles = append(book.EbookFiles, item.Name())
				}

				// Check for cover
				if name == "cover.jpg" || name == "cover.jpeg" || name == "cover.png" {
					book.HasCover = true
				}

				// Check for metadata
				if name == "metadata.opf" {
					book.HasOPF = true
				}
			}

			author.Books = append(author.Books, book)
			result.TotalBooks++

			// Categorize
			if len(book.EbookFiles) == 0 {
				result.BooksWithoutFiles = append(result.BooksWithoutFiles, book)
			} else if !book.HasOPF && !book.HasCover {
				result.BooksWithoutMeta = append(result.BooksWithoutMeta, book)
			}
		}

		result.Authors = append(result.Authors, author)
	}

	result.ScanDuration = time.Since(startTime)
	return result
}

// CleanupCalibreLibrary removes book folders that have no ebook files
// and removes empty author directories
func CleanupCalibreLibrary(result *CalibreScanResult) ([]string, []string, error) {
	var cleanedDirs []string
	var cleanedFiles []string

	// Track author directories that may become empty after removing books
	authorPaths := make(map[string]bool)

	for _, book := range result.BooksWithoutFiles {
		// Track the author path for later cleanup
		authorPaths[book.AuthorPath] = true

		// Remove all files in the book directory
		entries, err := os.ReadDir(book.BookPath)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				filePath := filepath.Join(book.BookPath, entry.Name())
				if err := os.Remove(filePath); err == nil {
					cleanedFiles = append(cleanedFiles, filePath)
				}
			}
		}

		// Remove the book directory
		if err := os.Remove(book.BookPath); err == nil {
			cleanedDirs = append(cleanedDirs, book.BookPath)
		}
	}

	// After all books are removed, check and remove empty author directories
	for authorPath := range authorPaths {
		entries, err := os.ReadDir(authorPath)
		if err == nil && len(entries) == 0 {
			if err := os.Remove(authorPath); err == nil {
				cleanedDirs = append(cleanedDirs, authorPath)
			}
		}
	}

	// Also remove pre-existing empty author directories found during scan
	for _, authorPath := range result.EmptyAuthors {
		// Skip if we already tried to remove this one above
		if authorPaths[authorPath] {
			continue
		}
		// Double-check it's still empty before removing
		entries, err := os.ReadDir(authorPath)
		if err == nil && len(entries) == 0 {
			if err := os.Remove(authorPath); err == nil {
				cleanedDirs = append(cleanedDirs, authorPath)
			}
		}
	}

	return cleanedDirs, cleanedFiles, nil
}
