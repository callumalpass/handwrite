package processor

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"
)

type ImageData struct {
	Image    image.Image
	PageNum  int
	Filename string
}

type PDFData struct {
	Data     []byte
	Filename string
}

func GetPDFData(inputPath string) (*PDFData, error) {
	ext := strings.ToLower(filepath.Ext(inputPath))
	
	if ext != ".pdf" {
		return nil, fmt.Errorf("not a PDF file: %s", ext)
	}
	
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read PDF file: %w", err)
	}
	
	return &PDFData{
		Data:     data,
		Filename: filepath.Base(inputPath),
	}, nil
}

func GetImagesFromFile(inputPath string) ([]ImageData, error) {
	ext := strings.ToLower(filepath.Ext(inputPath))
	
	switch ext {
	case ".pdf":
		return nil, fmt.Errorf("PDF files should be processed with GetPDFData")
	case ".png", ".jpg", ".jpeg":
		return loadImageFile(inputPath)
	default:
		return nil, fmt.Errorf("unsupported file type: %s", ext)
	}
}

func loadImageFile(imagePath string) ([]ImageData, error) {
	file, err := os.Open(imagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open image file: %w", err)
	}
	defer file.Close()

	var img image.Image
	ext := strings.ToLower(filepath.Ext(imagePath))
	
	switch ext {
	case ".png":
		img, err = png.Decode(file)
	case ".jpg", ".jpeg":
		img, err = jpeg.Decode(file)
	default:
		return nil, fmt.Errorf("unsupported image format: %s", ext)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	return []ImageData{{
		Image:    img,
		PageNum:  1,
		Filename: filepath.Base(imagePath),
	}}, nil
}

func GetSupportedFiles(inputPath string) ([]string, error) {
	var files []string
	
	info, err := os.Stat(inputPath)
	if err != nil {
		return nil, fmt.Errorf("path does not exist: %w", err)
	}

	if !info.IsDir() {
		// Single file
		if isSupportedFile(inputPath) {
			files = append(files, inputPath)
		}
		return files, nil
	}

	// Directory - walk through recursively
	err = filepath.Walk(inputPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if !info.IsDir() && isSupportedFile(path) {
			files = append(files, path)
		}
		
		return nil
	})

	return files, err
}

func isSupportedFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	supportedExts := []string{".pdf", ".png", ".jpg", ".jpeg"}
	
	for _, supportedExt := range supportedExts {
		if ext == supportedExt {
			return true
		}
	}
	
	return false
}