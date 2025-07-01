package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "handwrite",
	Short: "A tool to convert handwritten notes from PDF files into organized Markdown documents",
	Long: `Handwrite is a command-line tool that converts handwritten notes from PDF files 
into organized Markdown documents using Google's Gemini AI for OCR processing.

Features:
- PDF to Markdown conversion
- Gemini OCR with customizable prompts
- Template-based output formatting
- Batch processing support
- Obsidian-compatible output`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(processCmd)
	rootCmd.AddCommand(configCmd)
}
