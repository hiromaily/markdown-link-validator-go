package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sync"
)

// get all markdown files
func walkDir(root string) ([]string, error) {
	var mdFiles []string

	// find all markdown in directory recurively using filepath.Walk()
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".md" {
			mdFiles = append(mdFiles, path)
		}
		return nil
	})
	return mdFiles, err
}

// Read markdown file
func readFile(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// get file path from markdown file
func extractImagePaths(markdownContent string) []string {
	imageLinkPattern := `!\[.*?\]\((https?://.*?\.(?:jpg|jpeg|png|gif|svg|webp))(?:\s+".*?")?\)`
	re := regexp.MustCompile(imageLinkPattern)

	matches := re.FindAllStringSubmatch(markdownContent, -1)

	var imagePaths []string
	for _, match := range matches {
		if len(match) > 1 {
			//fmt.Println("[Debug] imagePath:", match[1])
			imagePaths = append(imagePaths, match[1])
		}
	}
	return imagePaths
}

// check if an image exists
func checkImagePaths(markdownFile string, imagePaths []string) {
	var wg sync.WaitGroup

	// channel for result
	results := make(chan string, len(imagePaths))

	// run goroutine each imgpath
	for _, imgpath := range imagePaths {
		wg.Add(1)
		go func(p string) {
			defer wg.Done()
			checkURL(p, results)
		}(imgpath)
	}

	// wait all goroutine for WaitGroup completed
	wg.Wait()

	close(results)

	// display result
	for result := range results {
		fmt.Printf("[File]: %s, [URL] %s\n", markdownFile, result)
	}
}

func checkURL(url string, results chan<- string) {
	resp, err := http.Get(url)
	if err != nil {
		results <- fmt.Sprintf("Error fetching %s: %v", url, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		results <- fmt.Sprintf("%s is up with status code 200.", url)
	} else {
		results <- fmt.Sprintf("%s returned status code %d.", url, resp.StatusCode)
	}
}

func main() {
	rootDir := "."

	var markdownFiles []string
	markdownFiles, err := walkDir(rootDir)
	if err != nil {
		fmt.Printf("Error walking the directory: %v\n", err)
		return
	}
	fmt.Printf("%d files found\n", len(markdownFiles))

	for _, markdownFile := range markdownFiles {
		//fmt.Printf("Read file: %s\n", markdownFile)
		markdownContent, err := readFile(markdownFile)
		if err != nil {
			fmt.Printf("Error reading file %s: %v\n", markdownFile, err)
			continue
		}

		imagePaths := extractImagePaths(markdownContent)
		checkImagePaths(markdownFile, imagePaths)
	}
}
