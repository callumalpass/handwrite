package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/callumalpass/handwrite/internal/config"
	"github.com/callumalpass/handwrite/internal/gemini"
	"github.com/callumalpass/handwrite/internal/processor"
	"github.com/callumalpass/handwrite/internal/template"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

var (
	configFile string
	workers    int
)

var processCmd = &cobra.Command{
	Use:   "process <input_path> <output_dir>",
	Short: "Process PDF or image files to extract handwritten text",
	Long: `Process PDF or image files to extract handwritten text using Gemini OCR.
	
Input can be a single file or a directory containing multiple files.
Supported formats: PDF, PNG, JPG, JPEG`,
	Args: cobra.ExactArgs(2),
	Run:  runProcess,
}

func init() {
	processCmd.Flags().StringVar(&configFile, "config", "", "Path to configuration file")
	processCmd.Flags().IntVar(&workers, "workers", 4, "Number of concurrent workers")
}

func runProcess(cmd *cobra.Command, args []string) {
	inputPath := args[0]
	outputDir := args[1]

	// Validate output directory
	if info, err := os.Stat(outputDir); err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(outputDir, 0755); err != nil {
				log.Fatalf("Failed to create output directory: %v", err)
			}
		} else {
			log.Fatalf("Error accessing output directory: %v", err)
		}
	} else if !info.IsDir() {
		log.Fatalf("Output path is not a directory: %s", outputDir)
	}

	// Load configuration
	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Get Gemini API key
	apiKey, err := config.GetGeminiAPIKey()
	if err != nil {
		log.Fatalf("Failed to get Gemini API key: %v", err)
	}

	// Create Gemini client
	geminiClient, err := gemini.NewClient(apiKey, cfg.Gemini.Model)
	if err != nil {
		log.Fatalf("Failed to create Gemini client: %v", err)
	}
	defer func() {
		geminiClient.Close()
	}()

	// Get list of input files
	inputFiles, err := processor.GetSupportedFiles(inputPath)
	if err != nil {
		log.Fatalf("Failed to get input files: %v", err)
	}

	if len(inputFiles) == 0 {
		log.Fatalf("No supported files found in: %s", inputPath)
	}

	fmt.Printf("Processing %d file(s) with %d workers...\n", len(inputFiles), workers)

	// Create progress bar
	bar := progressbar.Default(int64(len(inputFiles)))

	// Process files concurrently
	results := processFilesConcurrently(inputFiles, outputDir, cfg, geminiClient, workers, bar)

	// Print results
	fmt.Printf("\nProcessing complete:\n")
	fmt.Printf("  Successful: %d\n", results.successful)
	fmt.Printf("  Failed: %d\n", results.failed)

	if results.failed > 0 {
		os.Exit(1)
	}
}

type ProcessingResults struct {
	successful int
	failed     int
}

func processFilesConcurrently(inputFiles []string, outputDir string, cfg *config.Config, geminiClient *gemini.Client, numWorkers int, bar *progressbar.ProgressBar) ProcessingResults {
	// Create channels
	jobs := make(chan string, len(inputFiles))
	results := make(chan bool, len(inputFiles))

	// Start workers
	var wg sync.WaitGroup
	for w := 1; w <= numWorkers; w++ {
		wg.Add(1)
		go worker(w, jobs, results, outputDir, cfg, geminiClient, bar, &wg)
	}

	// Send jobs
	for _, file := range inputFiles {
		jobs <- file
	}
	close(jobs)

	// Wait for workers to finish
	wg.Wait()
	close(results)

	// Collect results
	var successful, failed int
	for success := range results {
		if success {
			successful++
		} else {
			failed++
		}
	}

	return ProcessingResults{
		successful: successful,
		failed:     failed,
	}
}

func worker(_ int, jobs <-chan string, results chan<- bool, outputDir string, cfg *config.Config, geminiClient *gemini.Client, bar *progressbar.ProgressBar, wg *sync.WaitGroup) {
	defer wg.Done()

	for inputFile := range jobs {
		success := processFile(inputFile, outputDir, cfg, geminiClient)
		results <- success
		_ = bar.Add(1)
	}
}

func processFile(inputPath, outputDir string, cfg *config.Config, geminiClient *gemini.Client) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	ext := strings.ToLower(filepath.Ext(inputPath))
	var fullText string
	var pageCount int = 1 // Default for single files

	if ext == ".pdf" {
		// Handle PDF files - process entire PDF at once
		pdfData, err := processor.GetPDFData(inputPath)
		if err != nil {
			log.Printf("Error reading PDF %s: %v", inputPath, err)
			return false
		}

		// Process entire PDF with OCR
		text, err := geminiClient.ExtractTextFromPDFWithRetry(ctx, pdfData.Data, cfg.Gemini.Prompt, 3)
		if err != nil {
			log.Printf("Error processing PDF %s: %v", inputPath, err)
			return false
		}

		fullText = text
		log.Printf("Full text extracted, length: %d", len(fullText))
		// Note: We don't know exact page count without parsing, but Gemini processes all pages
	} else {
		// Handle image files
		images, err := processor.GetImagesFromFile(inputPath)
		if err != nil {
			log.Printf("Error extracting images from %s: %v", inputPath, err)
			return false
		}

		if len(images) == 0 {
			log.Printf("No images found in: %s", inputPath)
			return false
		}

		pageCount = len(images)
		var pageResults []string

		// Process each image with OCR
		for _, imgData := range images {
			text, err := geminiClient.ExtractTextFromImageWithRetry(ctx, imgData.Image, cfg.Gemini.Prompt, 3)
			if err != nil {
				log.Printf("Error processing page %d of %s: %v", imgData.PageNum, inputPath, err)
				pageResults = append(pageResults, fmt.Sprintf("Error processing page %d: %v", imgData.PageNum, err))
			} else {
				pageResults = append(pageResults, text)
			}
		}

		// Combine all page results
		fullText = strings.Join(pageResults, "\n\n")
	}

	if strings.TrimSpace(fullText) == "" {
		log.Printf("No text extracted from: %s", inputPath)
		return false
	}

	// Create output filename
	baseName := strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath))
	outputFilename := baseName + ".md"
	outputPath := filepath.Join(outputDir, outputFilename)

	// Create template data
	templateData := template.CreateTemplateData(
		fullText,
		filepath.Base(inputPath),
		inputPath,
		outputDir,
		pageCount,
		cfg.Gemini.Model,
		cfg.Template.Variables,
	)

	log.Printf("Template data content length: %d", len(templateData.Content))

	// Render template
	templatePath := cfg.Template.Path
	if !filepath.IsAbs(templatePath) {
		// Relative to executable directory
		execDir, _ := os.Executable()
		templatePath = filepath.Join(filepath.Dir(execDir), templatePath)
	}

	if err := template.RenderTemplate(templatePath, outputPath, templateData); err != nil {
		log.Printf("Error rendering template for %s: %v", inputPath, err)
		return false
	}

	return true
}

