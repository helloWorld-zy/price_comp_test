package llm

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strings"
)

// WordExtractor handles text extraction from Word documents (.docx)
type WordExtractor struct{}

// NewWordExtractor creates a new Word extractor
func NewWordExtractor() *WordExtractor {
	return &WordExtractor{}
}

// ExtractText extracts text from a .docx file
// .docx files are actually ZIP archives containing XML files
func (e *WordExtractor) ExtractText(filePath string) (string, error) {
	// Open the .docx file as a ZIP archive
	r, err := zip.OpenReader(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open .docx file: %w", err)
	}
	defer r.Close()

	// Find and read the document.xml file
	var documentXML []byte
	for _, f := range r.File {
		if f.Name == "word/document.xml" {
			rc, err := f.Open()
			if err != nil {
				return "", fmt.Errorf("failed to open document.xml: %w", err)
			}
			defer rc.Close()

			documentXML, err = io.ReadAll(rc)
			if err != nil {
				return "", fmt.Errorf("failed to read document.xml: %w", err)
			}
			break
		}
	}

	if documentXML == nil {
		return "", fmt.Errorf("document.xml not found in .docx file")
	}

	// Parse the XML and extract text
	text, err := e.parseDocumentXML(documentXML)
	if err != nil {
		return "", fmt.Errorf("failed to parse document.xml: %w", err)
	}

	return text, nil
}

// ExtractTextFromReader extracts text from a Word document reader
func (e *WordExtractor) ExtractTextFromReader(r io.ReaderAt, size int64) (string, error) {
	// Open the reader as a ZIP archive
	zipReader, err := zip.NewReader(r, size)
	if err != nil {
		return "", fmt.Errorf("failed to open .docx reader: %w", err)
	}

	// Find and read the document.xml file
	var documentXML []byte
	for _, f := range zipReader.File {
		if f.Name == "word/document.xml" {
			rc, err := f.Open()
			if err != nil {
				return "", fmt.Errorf("failed to open document.xml: %w", err)
			}
			defer rc.Close()

			documentXML, err = io.ReadAll(rc)
			if err != nil {
				return "", fmt.Errorf("failed to read document.xml: %w", err)
			}
			break
		}
	}

	if documentXML == nil {
		return "", fmt.Errorf("document.xml not found in .docx file")
	}

	// Parse the XML and extract text
	text, err := e.parseDocumentXML(documentXML)
	if err != nil {
		return "", fmt.Errorf("failed to parse document.xml: %w", err)
	}

	return text, nil
}

// parseDocumentXML parses the document.xml and extracts all text
func (e *WordExtractor) parseDocumentXML(xmlData []byte) (string, error) {
	// Define structures for the relevant parts of the XML
	type Text struct {
		Content string `xml:",chardata"`
		Space   string `xml:"space,attr"`
	}

	type Run struct {
		Texts []Text `xml:"t"`
	}

	type Paragraph struct {
		Runs []Run `xml:"r"`
	}

	type Body struct {
		Paragraphs []Paragraph `xml:"p"`
	}

	type Document struct {
		XMLName xml.Name `xml:"document"`
		Body    Body     `xml:"body"`
	}

	// Parse the XML
	var doc Document
	decoder := xml.NewDecoder(bytes.NewReader(xmlData))
	decoder.CharsetReader = func(charset string, input io.Reader) (io.Reader, error) {
		return input, nil // Simple charset handling
	}

	if err := decoder.Decode(&doc); err != nil {
		return "", fmt.Errorf("failed to decode XML: %w", err)
	}

	// Extract text from paragraphs
	var paragraphs []string
	for _, para := range doc.Body.Paragraphs {
		var paraText []string
		for _, run := range para.Runs {
			for _, text := range run.Texts {
				paraText = append(paraText, text.Content)
			}
		}

		paraStr := strings.TrimSpace(strings.Join(paraText, ""))
		if paraStr != "" {
			paragraphs = append(paragraphs, paraStr)
		}
	}

	return strings.Join(paragraphs, "\n"), nil
}

// ExtractTextFromFile is a helper that works with any file
func (e *WordExtractor) ExtractTextFromFile(filePath string) (string, error) {
	// Check if file exists
	info, err := os.Stat(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to stat file: %w", err)
	}

	// Open file for reading
	f, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	return e.ExtractTextFromReader(f, info.Size())
}

// GetMetadata extracts basic metadata from a .docx file
func (e *WordExtractor) GetMetadata(filePath string) (map[string]string, error) {
	r, err := zip.OpenReader(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open .docx file: %w", err)
	}
	defer r.Close()

	// Look for core.xml which contains metadata
	var coreXML []byte
	for _, f := range r.File {
		if f.Name == "docProps/core.xml" {
			rc, err := f.Open()
			if err != nil {
				return nil, fmt.Errorf("failed to open core.xml: %w", err)
			}
			defer rc.Close()

			coreXML, err = io.ReadAll(rc)
			if err != nil {
				return nil, fmt.Errorf("failed to read core.xml: %w", err)
			}
			break
		}
	}

	if coreXML == nil {
		return map[string]string{}, nil // No metadata found
	}

	// Parse metadata (simplified)
	metadata := make(map[string]string)

	// Extract creator
	if start := bytes.Index(coreXML, []byte("<dc:creator>")); start != -1 {
		end := bytes.Index(coreXML[start:], []byte("</dc:creator>"))
		if end != -1 {
			metadata["creator"] = string(coreXML[start+12 : start+end])
		}
	}

	// Extract title
	if start := bytes.Index(coreXML, []byte("<dc:title>")); start != -1 {
		end := bytes.Index(coreXML[start:], []byte("</dc:title>"))
		if end != -1 {
			metadata["title"] = string(coreXML[start+10 : start+end])
		}
	}

	// Extract created date
	if start := bytes.Index(coreXML, []byte("<dcterms:created")); start != -1 {
		end := bytes.Index(coreXML[start:], []byte("</dcterms:created>"))
		if end != -1 {
			// Find the content between > and </
			content := coreXML[start : start+end]
			if contentStart := bytes.IndexByte(content, '>'); contentStart != -1 {
				metadata["created"] = string(content[contentStart+1:])
			}
		}
	}

	return metadata, nil
}
