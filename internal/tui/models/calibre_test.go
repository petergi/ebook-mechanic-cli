package models

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewCalibreModel(t *testing.T) {
	m := NewCalibreModel("/test/library", 80, 24)
	if m.libraryPath != "/test/library" {
		t.Errorf("expected libraryPath /test/library, got %s", m.libraryPath)
	}
	if m.width != 80 {
		t.Errorf("expected width 80, got %d", m.width)
	}
	if !m.scanning {
		t.Error("expected scanning to be true initially")
	}
}

func TestCalibreModel_GetScanResult(t *testing.T) {
	m := NewCalibreModel("/test", 80, 24)
	if m.GetScanResult() != nil {
		t.Error("expected nil scan result initially")
	}
}

func TestScanCalibreLibrary(t *testing.T) {
	// Create a temp Calibre-like library structure
	tmpDir, err := os.MkdirTemp("", "calibre-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create author directory
	authorDir := filepath.Join(tmpDir, "Test Author")
	if err := os.MkdirAll(authorDir, 0755); err != nil {
		t.Fatalf("failed to create author dir: %v", err)
	}

	// Create book with ebook file
	bookWithFile := filepath.Join(authorDir, "Good Book (123)")
	if err := os.MkdirAll(bookWithFile, 0755); err != nil {
		t.Fatalf("failed to create book dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(bookWithFile, "test.epub"), []byte("fake"), 0644); err != nil {
		t.Fatalf("failed to create epub: %v", err)
	}
	if err := os.WriteFile(filepath.Join(bookWithFile, "metadata.opf"), []byte("<opf/>"), 0644); err != nil {
		t.Fatalf("failed to create opf: %v", err)
	}

	// Create book with only metadata (no ebook)
	bookWithoutFile := filepath.Join(authorDir, "Orphan Book (456)")
	if err := os.MkdirAll(bookWithoutFile, 0755); err != nil {
		t.Fatalf("failed to create book dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(bookWithoutFile, "metadata.opf"), []byte("<opf/>"), 0644); err != nil {
		t.Fatalf("failed to create opf: %v", err)
	}
	if err := os.WriteFile(filepath.Join(bookWithoutFile, "cover.jpg"), []byte("fake"), 0644); err != nil {
		t.Fatalf("failed to create cover: %v", err)
	}

	// Create book with ebook but no metadata
	bookWithoutMeta := filepath.Join(authorDir, "No Meta Book (789)")
	if err := os.MkdirAll(bookWithoutMeta, 0755); err != nil {
		t.Fatalf("failed to create book dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(bookWithoutMeta, "book.pdf"), []byte("fake"), 0644); err != nil {
		t.Fatalf("failed to create pdf: %v", err)
	}

	// Scan the library
	result := ScanCalibreLibrary(tmpDir, nil)

	if result.TotalAuthors != 1 {
		t.Errorf("expected 1 author, got %d", result.TotalAuthors)
	}
	if result.TotalBooks != 3 {
		t.Errorf("expected 3 books, got %d", result.TotalBooks)
	}
	if len(result.BooksWithoutFiles) != 1 {
		t.Errorf("expected 1 book without files, got %d", len(result.BooksWithoutFiles))
	}
	if len(result.BooksWithoutMeta) != 1 {
		t.Errorf("expected 1 book without metadata, got %d", len(result.BooksWithoutMeta))
	}
}

func TestCleanupCalibreLibrary(t *testing.T) {
	// Create a temp Calibre-like library structure
	tmpDir, err := os.MkdirTemp("", "calibre-cleanup-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create author directory
	authorDir := filepath.Join(tmpDir, "Test Author")
	if err := os.MkdirAll(authorDir, 0755); err != nil {
		t.Fatalf("failed to create author dir: %v", err)
	}

	// Create first orphan book with only metadata (no ebook)
	orphanBook1 := filepath.Join(authorDir, "Orphan Book 1 (456)")
	if err := os.MkdirAll(orphanBook1, 0755); err != nil {
		t.Fatalf("failed to create book dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(orphanBook1, "metadata.opf"), []byte("<opf/>"), 0644); err != nil {
		t.Fatalf("failed to create opf: %v", err)
	}

	// Create second orphan book by the same author
	orphanBook2 := filepath.Join(authorDir, "Orphan Book 2 (789)")
	if err := os.MkdirAll(orphanBook2, 0755); err != nil {
		t.Fatalf("failed to create book dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(orphanBook2, "metadata.opf"), []byte("<opf/>"), 0644); err != nil {
		t.Fatalf("failed to create opf: %v", err)
	}
	if err := os.WriteFile(filepath.Join(orphanBook2, "cover.jpg"), []byte("fake"), 0644); err != nil {
		t.Fatalf("failed to create cover: %v", err)
	}

	// Scan the library first
	result := ScanCalibreLibrary(tmpDir, nil)

	if len(result.BooksWithoutFiles) != 2 {
		t.Errorf("expected 2 books without files, got %d", len(result.BooksWithoutFiles))
	}

	// Run cleanup
	cleanedDirs, cleanedFiles, err := CleanupCalibreLibrary(result)
	if err != nil {
		t.Fatalf("cleanup failed: %v", err)
	}

	if len(cleanedFiles) != 3 {
		t.Errorf("expected 3 files to be cleaned (2 metadata.opf + 1 cover.jpg), got %d", len(cleanedFiles))
	}
	// Should clean up: 2 book dirs + 1 author dir = 3 dirs
	if len(cleanedDirs) != 3 {
		t.Errorf("expected 3 directories to be cleaned (2 book dirs + 1 author dir), got %d", len(cleanedDirs))
	}

	// Verify the orphan book directories were removed
	if _, err := os.Stat(orphanBook1); !os.IsNotExist(err) {
		t.Error("orphan book 1 directory should have been removed")
	}
	if _, err := os.Stat(orphanBook2); !os.IsNotExist(err) {
		t.Error("orphan book 2 directory should have been removed")
	}

	// Verify the author directory was also removed (since all books are gone)
	if _, err := os.Stat(authorDir); !os.IsNotExist(err) {
		t.Error("author directory should have been removed when all books are gone")
	}
}

func TestScanCalibreLibrary_EmptyAuthors(t *testing.T) {
	// Create a temp Calibre-like library structure
	tmpDir, err := os.MkdirTemp("", "calibre-empty-authors-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create an author with books
	authorWithBooks := filepath.Join(tmpDir, "Author With Books")
	if err := os.MkdirAll(authorWithBooks, 0755); err != nil {
		t.Fatalf("failed to create author dir: %v", err)
	}
	bookDir := filepath.Join(authorWithBooks, "Some Book (123)")
	if err := os.MkdirAll(bookDir, 0755); err != nil {
		t.Fatalf("failed to create book dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(bookDir, "test.epub"), []byte("fake"), 0644); err != nil {
		t.Fatalf("failed to create epub: %v", err)
	}

	// Create an empty author directory (no book subfolders)
	emptyAuthor := filepath.Join(tmpDir, "Empty Author")
	if err := os.MkdirAll(emptyAuthor, 0755); err != nil {
		t.Fatalf("failed to create empty author dir: %v", err)
	}

	// Create another empty author with just loose files (not book subfolders)
	emptyAuthorWithFiles := filepath.Join(tmpDir, "Empty Author With Files")
	if err := os.MkdirAll(emptyAuthorWithFiles, 0755); err != nil {
		t.Fatalf("failed to create empty author dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(emptyAuthorWithFiles, "random.txt"), []byte("junk"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	// Scan the library
	result := ScanCalibreLibrary(tmpDir, nil)

	if result.TotalAuthors != 3 {
		t.Errorf("expected 3 authors, got %d", result.TotalAuthors)
	}
	if result.TotalBooks != 1 {
		t.Errorf("expected 1 book, got %d", result.TotalBooks)
	}
	// Both empty authors should be detected (even one with loose files but no book subfolders)
	if len(result.EmptyAuthors) != 2 {
		t.Errorf("expected 2 empty authors, got %d", len(result.EmptyAuthors))
	}
}

func TestCleanupCalibreLibrary_EmptyAuthors(t *testing.T) {
	// Create a temp Calibre-like library structure
	tmpDir, err := os.MkdirTemp("", "calibre-cleanup-empty-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create an author with books (should not be touched)
	authorWithBooks := filepath.Join(tmpDir, "Author With Books")
	if err := os.MkdirAll(authorWithBooks, 0755); err != nil {
		t.Fatalf("failed to create author dir: %v", err)
	}
	bookDir := filepath.Join(authorWithBooks, "Some Book (123)")
	if err := os.MkdirAll(bookDir, 0755); err != nil {
		t.Fatalf("failed to create book dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(bookDir, "test.epub"), []byte("fake"), 0644); err != nil {
		t.Fatalf("failed to create epub: %v", err)
	}

	// Create an empty author directory
	emptyAuthor := filepath.Join(tmpDir, "Empty Author")
	if err := os.MkdirAll(emptyAuthor, 0755); err != nil {
		t.Fatalf("failed to create empty author dir: %v", err)
	}

	// Scan the library first
	result := ScanCalibreLibrary(tmpDir, nil)

	if len(result.EmptyAuthors) != 1 {
		t.Errorf("expected 1 empty author, got %d", len(result.EmptyAuthors))
	}

	// Run cleanup
	cleanedDirs, _, err := CleanupCalibreLibrary(result)
	if err != nil {
		t.Fatalf("cleanup failed: %v", err)
	}

	// Should clean up the empty author directory
	if len(cleanedDirs) != 1 {
		t.Errorf("expected 1 directory to be cleaned, got %d", len(cleanedDirs))
	}

	// Verify the empty author directory was removed
	if _, err := os.Stat(emptyAuthor); !os.IsNotExist(err) {
		t.Error("empty author directory should have been removed")
	}

	// Verify the author with books is still there
	if _, err := os.Stat(authorWithBooks); os.IsNotExist(err) {
		t.Error("author with books should NOT have been removed")
	}
}
