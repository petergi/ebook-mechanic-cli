package models

import (
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewBrowserModel(t *testing.T) {
	tmpDir := t.TempDir()
	m := NewBrowserModel(tmpDir, 80, 24)

	if m.currentDir != tmpDir {
		t.Errorf("Expected currentDir to be %s, got %s", tmpDir, m.currentDir)
	}

	if m.selected != 0 {
		t.Error("Expected selected to be 0")
	}

	if len(m.filterExts) != 2 {
		t.Errorf("Expected 2 filter extensions, got %d", len(m.filterExts))
	}

	if m.filterExts[0] != ".epub" || m.filterExts[1] != ".pdf" {
		t.Error("Expected filter extensions to be .epub and .pdf")
	}
}

func TestNewBrowserModel_EmptyPath(t *testing.T) {
	m := NewBrowserModel("", 80, 24)

	// Should default to current working directory
	cwd, _ := os.Getwd()
	if m.currentDir != cwd {
		t.Errorf("Expected currentDir to be %s, got %s", cwd, m.currentDir)
	}
}

func TestBrowserModel_Update_NavigateDown(t *testing.T) {
	tmpDir := createTestDirectory(t)
	m := NewBrowserModel(tmpDir, 80, 24)

	if len(m.items) < 2 {
		t.Skip("Need at least 2 items for this test")
	}

	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updatedModel.(BrowserModel)

	if m.selected != 1 {
		t.Errorf("Expected selected to be 1, got %d", m.selected)
	}
}

func TestBrowserModel_Update_NavigateUp(t *testing.T) {
	tmpDir := createTestDirectory(t)
	m := NewBrowserModel(tmpDir, 80, 24)
	m.selected = 1

	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = updatedModel.(BrowserModel)

	if m.selected != 0 {
		t.Errorf("Expected selected to be 0, got %d", m.selected)
	}
}

func TestBrowserModel_Update_WrappingDown(t *testing.T) {
	tmpDir := createTestDirectory(t)
	m := NewBrowserModel(tmpDir, 80, 24)

	if len(m.items) == 0 {
		t.Skip("Need items for this test")
	}

	m.selected = len(m.items) - 1

	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updatedModel.(BrowserModel)

	if m.selected != 0 {
		t.Errorf("Expected selected to wrap to 0, got %d", m.selected)
	}
}

func TestBrowserModel_Update_WrappingUp(t *testing.T) {
	tmpDir := createTestDirectory(t)
	m := NewBrowserModel(tmpDir, 80, 24)

	if len(m.items) == 0 {
		t.Skip("Need items for this test")
	}

	m.selected = 0

	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = updatedModel.(BrowserModel)

	if m.selected != len(m.items)-1 {
		t.Errorf("Expected selected to wrap to %d, got %d", len(m.items)-1, m.selected)
	}
}

func TestBrowserModel_Update_VimKeys(t *testing.T) {
	tmpDir := createTestDirectory(t)

	tests := []struct {
		name     string
		key      string
		initial  int
		expected int
	}{
		{"j moves down", "j", 0, 1},
		{"k moves up", "k", 1, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewBrowserModel(tmpDir, 80, 24)
			if len(m.items) < 2 {
				t.Skip("Need at least 2 items")
			}

			m.selected = tt.initial

			updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{rune(tt.key[0])}})
			m = updatedModel.(BrowserModel)

			if m.selected != tt.expected {
				t.Errorf("Expected selected to be %d, got %d", tt.expected, m.selected)
			}
		})
	}
}

func TestBrowserModel_Update_EnterDirectory(t *testing.T) {
	tmpDir := createTestDirectory(t)
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdir: %v", err)
	}

	m := NewBrowserModel(tmpDir, 80, 24)
	m.loadDirectory()

	// Find the subdirectory in items
	dirIndex := -1
	for i, item := range m.items {
		if item.IsDir && item.Name == "subdir" {
			dirIndex = i
			break
		}
	}

	if dirIndex == -1 {
		t.Skip("Subdirectory not found in items")
	}

	m.selected = dirIndex
	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updatedModel.(BrowserModel)

	if m.currentDir != subDir {
		t.Errorf("Expected currentDir to be %s, got %s", subDir, m.currentDir)
	}
}

func TestBrowserModel_Update_BatchEnterDirectory(t *testing.T) {
	tmpDir := createTestDirectory(t)
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdir: %v", err)
	}

	m := NewBatchBrowserModel(tmpDir, 80, 24)
	m.loadDirectory()

	dirIndex := -1
	for i, item := range m.items {
		if item.IsDir && item.Name == "subdir" {
			dirIndex = i
			break
		}
	}

	if dirIndex == -1 {
		t.Skip("Subdirectory not found in items")
	}

	m.selected = dirIndex
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if cmd == nil {
		t.Fatal("Expected command to be non-nil for directory selection")
	}

	msg := cmd()
	selectMsg, ok := msg.(DirectorySelectMsg)
	if !ok {
		t.Fatalf("Expected DirectorySelectMsg, got %T", msg)
	}

	if selectMsg.Path != subDir {
		t.Errorf("Expected path %s, got %s", subDir, selectMsg.Path)
	}
}

func TestBrowserModel_Update_BatchOpenDirectory(t *testing.T) {
	tmpDir := createTestDirectory(t)
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdir: %v", err)
	}

	m := NewBatchBrowserModel(tmpDir, 80, 24)
	m.loadDirectory()

	dirIndex := -1
	for i, item := range m.items {
		if item.IsDir && item.Name == "subdir" {
			dirIndex = i
			break
		}
	}

	if dirIndex == -1 {
		t.Skip("Subdirectory not found in items")
	}

	m.selected = dirIndex
	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	m = updatedModel.(BrowserModel)

	if m.currentDir != subDir {
		t.Errorf("Expected currentDir to be %s, got %s", subDir, m.currentDir)
	}
}

func TestBrowserModel_Update_SelectFile(t *testing.T) {
	tmpDir := createTestDirectory(t)
	testFile := filepath.Join(tmpDir, "test.epub")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	m := NewBrowserModel(tmpDir, 80, 24)
	m.loadDirectory()

	// Find the file in items
	fileIndex := -1
	for i, item := range m.items {
		if !item.IsDir && item.Name == "test.epub" {
			fileIndex = i
			break
		}
	}

	if fileIndex == -1 {
		t.Skip("Test file not found in items")
	}

	m.selected = fileIndex
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if cmd == nil {
		t.Fatal("Expected command to be non-nil for file selection")
	}

	msg := cmd()
	selectMsg, ok := msg.(FileSelectMsg)
	if !ok {
		t.Fatal("Expected FileSelectMsg")
	}

	if selectMsg.Path != testFile {
		t.Errorf("Expected path %s, got %s", testFile, selectMsg.Path)
	}
}

func TestBrowserModel_Update_BackspaceToParent(t *testing.T) {
	tmpDir := createTestDirectory(t)
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdir: %v", err)
	}

	m := NewBrowserModel(subDir, 80, 24)

	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	m = updatedModel.(BrowserModel)

	if m.currentDir != tmpDir {
		t.Errorf("Expected currentDir to be %s, got %s", tmpDir, m.currentDir)
	}
}

func TestBrowserModel_Update_HKeyToParent(t *testing.T) {
	tmpDir := createTestDirectory(t)
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdir: %v", err)
	}

	m := NewBrowserModel(subDir, 80, 24)

	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	m = updatedModel.(BrowserModel)

	if m.currentDir != tmpDir {
		t.Errorf("Expected currentDir to be %s, got %s", tmpDir, m.currentDir)
	}
}

func TestBrowserModel_Update_ToggleHidden(t *testing.T) {
	tmpDir := createTestDirectory(t)
	m := NewBrowserModel(tmpDir, 80, 24)

	initialShowHidden := m.showHidden

	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'.'}})
	m = updatedModel.(BrowserModel)

	if m.showHidden == initialShowHidden {
		t.Error("Expected showHidden to toggle")
	}
}

func TestBrowserModel_Update_BackToMenu(t *testing.T) {
	tmpDir := createTestDirectory(t)
	m := NewBrowserModel(tmpDir, 80, 24)

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if cmd == nil {
		t.Fatal("Expected command to be non-nil")
	}

	msg := cmd()
	_, ok := msg.(BackToMenuMsg)
	if !ok {
		t.Fatal("Expected BackToMenuMsg")
	}
}

func TestBrowserModel_Update_QuitKey(t *testing.T) {
	tmpDir := createTestDirectory(t)
	m := NewBrowserModel(tmpDir, 80, 24)

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})

	if cmd == nil {
		t.Fatal("Expected command to be non-nil")
	}

	msg := cmd()
	_, ok := msg.(BackToMenuMsg)
	if !ok {
		t.Fatal("Expected BackToMenuMsg for q key")
	}
}

func TestBrowserModel_Update_CtrlC(t *testing.T) {
	tmpDir := createTestDirectory(t)
	m := NewBrowserModel(tmpDir, 80, 24)

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})

	// Should return tea.Quit
	if cmd == nil {
		t.Fatal("Expected quit command to be non-nil")
	}
}

func TestBrowserModel_Update_WindowSize(t *testing.T) {
	tmpDir := createTestDirectory(t)
	m := NewBrowserModel(tmpDir, 80, 24)

	updatedModel, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m = updatedModel.(BrowserModel)

	if m.width != 120 {
		t.Errorf("Expected width to be 120, got %d", m.width)
	}

	if m.height != 40 {
		t.Errorf("Expected height to be 40, got %d", m.height)
	}

	expectedViewportSize := 40 - 15
	if m.viewportSize != expectedViewportSize {
		t.Errorf("Expected viewportSize to be %d, got %d", expectedViewportSize, m.viewportSize)
	}
}

func TestBrowserModel_SelectedPath(t *testing.T) {
	tmpDir := createTestDirectory(t)
	testFile := filepath.Join(tmpDir, "test.epub")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	m := NewBrowserModel(tmpDir, 80, 24)
	m.loadDirectory()

	// Find the file
	fileIndex := -1
	for i, item := range m.items {
		if !item.IsDir && item.Name == "test.epub" {
			fileIndex = i
			break
		}
	}

	if fileIndex == -1 {
		t.Skip("Test file not found")
	}

	m.selected = fileIndex
	path := m.SelectedPath()

	if path != testFile {
		t.Errorf("Expected path %s, got %s", testFile, path)
	}
}

func TestBrowserModel_SelectedPath_OutOfBounds(t *testing.T) {
	tmpDir := createTestDirectory(t)
	m := NewBrowserModel(tmpDir, 80, 24)

	m.selected = -1
	path := m.SelectedPath()
	if path != "" {
		t.Errorf("Expected empty path for out of bounds, got %s", path)
	}

	m.selected = 999
	path = m.SelectedPath()
	if path != "" {
		t.Errorf("Expected empty path for out of bounds, got %s", path)
	}
}

func TestBrowserModel_matchesFilter(t *testing.T) {
	m := NewBrowserModel("", 80, 24)

	tests := []struct {
		filename string
		expected bool
	}{
		{"test.epub", true},
		{"test.EPUB", true},
		{"test.pdf", true},
		{"test.PDF", true},
		{"test.txt", false},
		{"test.docx", false},
		{"test", false},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			result := m.matchesFilter(tt.filename)
			if result != tt.expected {
				t.Errorf("Expected matchesFilter(%s) to be %v, got %v", tt.filename, tt.expected, result)
			}
		})
	}
}

func TestBrowserModel_matchesFilter_NoFilter(t *testing.T) {
	m := NewBrowserModel("", 80, 24)
	m.filterExts = []string{} // No filter

	if !m.matchesFilter("anything.txt") {
		t.Error("Expected to match when no filter is set")
	}
}

func TestBrowserModel_loadDirectory(t *testing.T) {
	tmpDir := createTestDirectory(t)

	// Create test files and directories
	if err := os.WriteFile(filepath.Join(tmpDir, "test.epub"), []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to write test.epub: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "test.pdf"), []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to write test.pdf: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "test.txt"), []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to write test.txt: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, ".hidden.epub"), []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to write .hidden.epub: %v", err)
	}
	if err := os.Mkdir(filepath.Join(tmpDir, "subdir"), 0755); err != nil {
		t.Fatalf("Failed to create subdir: %v", err)
	}

	m := NewBrowserModel(tmpDir, 80, 24)
	m.loadDirectory()

	// Should have: test.epub, test.pdf, subdir (test.txt filtered out, .hidden.epub hidden)
	if len(m.items) < 3 {
		t.Errorf("Expected at least 3 items (2 files + 1 dir), got %d", len(m.items))
	}
}

func TestBrowserModel_loadDirectory_ShowHidden(t *testing.T) {
	tmpDir := createTestDirectory(t)
	if err := os.WriteFile(filepath.Join(tmpDir, ".hidden.epub"), []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to write .hidden.epub: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "visible.epub"), []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to write visible.epub: %v", err)
	}

	m := NewBrowserModel(tmpDir, 80, 24)
	m.showHidden = true
	m.loadDirectory()

	// Should include .hidden.epub
	foundHidden := false
	for _, item := range m.items {
		if item.Name == ".hidden.epub" {
			foundHidden = true
			break
		}
	}

	if !foundHidden {
		t.Error("Expected to find .hidden.epub when showHidden is true")
	}
}

func TestBrowserModel_updateViewport(t *testing.T) {
	tmpDir := createTestDirectory(t)
	m := NewBrowserModel(tmpDir, 80, 24)
	m.viewportSize = 5

	// Create enough items
	for i := 0; i < 10; i++ {
		m.items = append(m.items, FileItem{
			Name:  filepath.Base(tmpDir),
			Path:  tmpDir,
			IsDir: true,
		})
	}

	// Test scrolling down
	m.selected = 6
	m.updateViewport()

	if m.viewportTop > m.selected {
		t.Error("Selected item should be visible in viewport")
	}

	// Test scrolling up
	m.selected = 0
	m.updateViewport()

	if m.viewportTop != 0 {
		t.Errorf("Expected viewportTop to be 0, got %d", m.viewportTop)
	}
}

func TestBrowserModel_View(t *testing.T) {
	tmpDir := createTestDirectory(t)
	m := NewBrowserModel(tmpDir, 80, 24)

	view := m.View()

	if view == "" {
		t.Error("Expected non-empty view")
	}
}

func TestBrowserModel_Init(t *testing.T) {
	tmpDir := createTestDirectory(t)
	m := NewBrowserModel(tmpDir, 80, 24)

	cmd := m.Init()

	if cmd != nil {
		t.Error("Expected Init to return nil command")
	}
}

// Helper function to create a test directory with some files
func createTestDirectory(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()

	// Create a basic structure
	if err := os.WriteFile(filepath.Join(tmpDir, "test1.epub"), []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to write test1.epub: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "test2.pdf"), []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to write test2.pdf: %v", err)
	}

	return tmpDir
}
