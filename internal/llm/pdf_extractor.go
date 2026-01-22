package llm

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

// PDFExtractor handles text extraction from PDF files
type PDFExtractor struct {
	// Using pdftotext command-line tool (from poppler-utils)
	// This is a common approach for production systems
	pdfToTextPath string
}

// NewPDFExtractor creates a new PDF extractor
func NewPDFExtractor() *PDFExtractor {
	return &PDFExtractor{
		pdfToTextPath: "pdftotext", // Assumes pdftotext is in PATH
	}
}

// ExtractText extracts text from a PDF file
func (e *PDFExtractor) ExtractText(filePath string) (string, error) {
	// Try using pdftotext command-line tool first
	text, err := e.extractWithPdfToText(filePath)
	if err == nil {
		return text, nil
	}

	// Fallback: return error with instructions
	return "", fmt.Errorf("failed to extract PDF text: %w. Please ensure 'pdftotext' (poppler-utils) is installed", err)
}

// extractWithPdfToText uses the pdftotext command-line tool
func (e *PDFExtractor) extractWithPdfToText(filePath string) (string, error) {
	// pdftotext options:
	// -layout: maintain original physical layout
	// -enc UTF-8: output encoding
	// - (dash): write to stdout
	cmd := exec.Command(e.pdfToTextPath, "-layout", "-enc", "UTF-8", filePath, "-")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("pdftotext command failed: %w, stderr: %s", err, stderr.String())
	}

	text := stdout.String()

	// Clean up the extracted text
	text = e.cleanText(text)

	return text, nil
}

// cleanText cleans up extracted text
func (e *PDFExtractor) cleanText(text string) string {
	// Remove excessive whitespace
	lines := strings.Split(text, "\n")
	var cleaned []string

	for _, line := range lines {
		// Trim leading/trailing whitespace
		line = strings.TrimSpace(line)

		// Skip empty lines
		if line == "" {
			continue
		}

		// Normalize multiple spaces to single space
		line = strings.Join(strings.Fields(line), " ")

		cleaned = append(cleaned, line)
	}

	return strings.Join(cleaned, "\n")
}

// ExtractTextFromReader extracts text from a PDF reader (for streaming)
// This is a placeholder for future implementation with pure Go libraries
func (e *PDFExtractor) ExtractTextFromReader(r io.Reader) (string, error) {
	// TODO: Implement pure Go PDF parsing using libraries like:
	// - github.com/ledongthuc/pdf
	// - github.com/pdfcpu/pdfcpu
	// For now, this requires writing to a temp file first
	return "", fmt.Errorf("streaming PDF extraction not yet implemented")
}

// GetMetadata extracts metadata from PDF
func (e *PDFExtractor) GetMetadata(filePath string) (map[string]string, error) {
	// Use pdfinfo command to get metadata
	cmd := exec.Command("pdfinfo", filePath)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("pdfinfo command failed: %w, stderr: %s", err, stderr.String())
	}

	metadata := make(map[string]string)
	lines := strings.Split(stdout.String(), "\n")

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			metadata[key] = value
		}
	}

	return metadata, nil
}
