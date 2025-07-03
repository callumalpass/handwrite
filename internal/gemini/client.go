package gemini

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"log"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type StructuredResponse struct {
	Content string   `json:"content"`
	Tags    []string `json:"tags"`
}

type Client struct {
	client *genai.Client
	model  *genai.GenerativeModel
}

func NewClient(apiKey string, modelName string) (*Client, error) {
	ctx := context.Background()

	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	model := client.GenerativeModel(modelName)
	
	// Configure for structured JSON output
	model.ResponseMIMEType = "application/json"
	model.ResponseSchema = &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"content": {Type: genai.TypeString},
			"tags": {
				Type:  genai.TypeArray,
				Items: &genai.Schema{Type: genai.TypeString},
			},
		},
		Required: []string{"content", "tags"},
	}

	return &Client{
		client: client,
		model:  model,
	}, nil
}

func (c *Client) Close() {
	if c.client != nil {
		c.client.Close()
	}
}

func (c *Client) ExtractTextFromImage(ctx context.Context, img image.Image, prompt string) (string, error) {
	// Convert image to JPEG bytes
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90}); err != nil {
		return "", fmt.Errorf("failed to encode image: %w", err)
	}

	// Create blob from image data
	blob := genai.Blob{
		MIMEType: "image/jpeg",
		Data:     buf.Bytes(),
	}

	// Generate content
	resp, err := c.model.GenerateContent(ctx, genai.Text(prompt), blob)
	if err != nil {
		return "", fmt.Errorf("failed to generate content: %w", err)
	}

	if len(resp.Candidates) == 0 {
		return "", fmt.Errorf("no candidates returned from Gemini")
	}

	if len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no content parts returned from Gemini")
	}

	// Extract text from the first part
	if textPart, ok := resp.Candidates[0].Content.Parts[0].(genai.Text); ok {
		return string(textPart), nil
	}

	return "", fmt.Errorf("unexpected content type returned from Gemini")
}

func (c *Client) ExtractStructuredTextFromImage(ctx context.Context, img image.Image, prompt string) (*StructuredResponse, error) {
	// Convert image to JPEG bytes
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90}); err != nil {
		return nil, fmt.Errorf("failed to encode image: %w", err)
	}

	// Create blob from image data
	blob := genai.Blob{
		MIMEType: "image/jpeg",
		Data:     buf.Bytes(),
	}

	// Generate content with JSON prompt
	resp, err := c.model.GenerateContent(ctx, genai.Text(prompt), blob)
	if err != nil {
		return nil, fmt.Errorf("failed to generate content: %w", err)
	}

	if len(resp.Candidates) == 0 {
		return nil, fmt.Errorf("no candidates returned from Gemini")
	}

	if len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no content parts returned from Gemini")
	}

	// Extract and parse JSON from the first part
	if textPart, ok := resp.Candidates[0].Content.Parts[0].(genai.Text); ok {
		text := string(textPart)
		log.Printf("Raw response text: %q", text)
		
		// Try to extract JSON from response (handle cases where it might be wrapped in markdown)
		jsonStr := text
		if strings.Contains(text, "```json") {
			// Extract JSON from markdown code block
			start := strings.Index(text, "```json") + 7
			end := strings.Index(text[start:], "```")
			if end > 0 {
				jsonStr = strings.TrimSpace(text[start : start+end])
			}
		} else if strings.Contains(text, "```") {
			// Try to extract from any code block
			start := strings.Index(text, "```") + 3
			end := strings.Index(text[start:], "```")
			if end > 0 {
				jsonStr = strings.TrimSpace(text[start : start+end])
			}
		}
		
		log.Printf("Extracted JSON: %q", jsonStr)
		
		var result StructuredResponse
		if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
			log.Printf("Failed to parse JSON response: %s", jsonStr)
			return nil, fmt.Errorf("failed to parse JSON response: %w", err)
		}
		
		log.Printf("Parsed content length: %d, content preview: %q", len(result.Content), result.Content[:min(100, len(result.Content))])
		return &result, nil
	}

	return nil, fmt.Errorf("unexpected content type returned from Gemini")
}

func (c *Client) ExtractTextFromPDF(ctx context.Context, pdfData []byte, prompt string) (string, error) {
	log.Printf("Processing PDF of size: %d bytes", len(pdfData))

	// Create blob from PDF data (supports PDFs up to 20MB)
	blob := genai.Blob{
		MIMEType: "application/pdf",
		Data:     pdfData,
	}

	log.Printf("Sending request to Gemini with prompt: %s", prompt[:min(50, len(prompt))])

	// Generate content from the entire PDF
	resp, err := c.model.GenerateContent(ctx, genai.Text(prompt), blob)
	if err != nil {
		return "", fmt.Errorf("failed to generate content: %w", err)
	}

	log.Printf("Received response with %d candidates", len(resp.Candidates))

	if len(resp.Candidates) == 0 {
		return "", fmt.Errorf("no candidates returned from Gemini")
	}

	if len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no content parts returned from Gemini")
	}

	log.Printf("Response has %d parts", len(resp.Candidates[0].Content.Parts))

	// Extract text from the first part
	if textPart, ok := resp.Candidates[0].Content.Parts[0].(genai.Text); ok {
		result := string(textPart)
		log.Printf("Extracted text length: %d", len(result))
		return result, nil
	}

	return "", fmt.Errorf("unexpected content type returned from Gemini")
}

func (c *Client) ExtractStructuredTextFromPDF(ctx context.Context, pdfData []byte, prompt string) (*StructuredResponse, error) {
	log.Printf("Processing PDF of size: %d bytes", len(pdfData))

	// Create blob from PDF data (supports PDFs up to 20MB)
	blob := genai.Blob{
		MIMEType: "application/pdf",
		Data:     pdfData,
	}

	log.Printf("Sending request to Gemini with prompt: %s", prompt[:min(50, len(prompt))])

	// Generate content from the entire PDF
	resp, err := c.model.GenerateContent(ctx, genai.Text(prompt), blob)
	if err != nil {
		return nil, fmt.Errorf("failed to generate content: %w", err)
	}

	log.Printf("Received response with %d candidates", len(resp.Candidates))

	if len(resp.Candidates) == 0 {
		return nil, fmt.Errorf("no candidates returned from Gemini")
	}

	if len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no content parts returned from Gemini")
	}

	log.Printf("Response has %d parts", len(resp.Candidates[0].Content.Parts))

	// Extract and parse JSON from the first part
	if textPart, ok := resp.Candidates[0].Content.Parts[0].(genai.Text); ok {
		text := string(textPart)
		log.Printf("Raw response text: %q", text)
		
		// Try to extract JSON from response (handle cases where it might be wrapped in markdown)
		jsonStr := text
		if strings.Contains(text, "```json") {
			// Extract JSON from markdown code block
			start := strings.Index(text, "```json") + 7
			end := strings.Index(text[start:], "```")
			if end > 0 {
				jsonStr = strings.TrimSpace(text[start : start+end])
			}
		} else if strings.Contains(text, "```") {
			// Try to extract from any code block
			start := strings.Index(text, "```") + 3
			end := strings.Index(text[start:], "```")
			if end > 0 {
				jsonStr = strings.TrimSpace(text[start : start+end])
			}
		}
		
		log.Printf("Extracted JSON: %q", jsonStr)
		
		var result StructuredResponse
		if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
			log.Printf("Failed to parse JSON response: %s", jsonStr)
			return nil, fmt.Errorf("failed to parse JSON response: %w", err)
		}
		
		log.Printf("Parsed content length: %d, content preview: %q", len(result.Content), result.Content[:min(100, len(result.Content))])
		log.Printf("Extracted structured text with %d tags", len(result.Tags))
		return &result, nil
	}

	return nil, fmt.Errorf("unexpected content type returned from Gemini")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (c *Client) ExtractTextFromImageWithRetry(ctx context.Context, img image.Image, prompt string, maxRetries int) (string, error) {
	var lastErr error

	for i := 0; i < maxRetries; i++ {
		result, err := c.ExtractTextFromImage(ctx, img, prompt)
		if err == nil {
			return result, nil
		}

		lastErr = err
		log.Printf("Attempt %d failed: %v", i+1, err)

		if i < maxRetries-1 {
			log.Printf("Retrying...")
		}
	}

	return "", fmt.Errorf("failed after %d attempts: %w", maxRetries, lastErr)
}

func (c *Client) ExtractTextFromPDFWithRetry(ctx context.Context, pdfData []byte, prompt string, maxRetries int) (string, error) {
	var lastErr error

	for i := 0; i < maxRetries; i++ {
		result, err := c.ExtractTextFromPDF(ctx, pdfData, prompt)
		if err == nil {
			return result, nil
		}

		lastErr = err
		log.Printf("Attempt %d failed: %v", i+1, err)

		if i < maxRetries-1 {
			log.Printf("Retrying...")
		}
	}

	return "", fmt.Errorf("failed after %d attempts: %w", maxRetries, lastErr)
}
