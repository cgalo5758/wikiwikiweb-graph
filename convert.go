package main

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"regexp"
	"strings"
	"time"

	md "github.com/JohannesKaufmann/html-to-markdown"
)

func html2md(sourceDir, destDir string) {
	// Check if source directory exists
	if _, err := os.Stat(sourceDir); os.IsNotExist(err) {
		fmt.Println("Error: source directory does not exist:", sourceDir)
		return
	}
	// Check if destination directory exists.
	if _, err := os.Stat(destDir); os.IsNotExist(err) {
		err = os.Mkdir(destDir, 0755)
		if err != nil {
			fmt.Println("Error creating destination directory:", err)
			os.Exit(1)
		}
	} else {
		// Clear destination directory
		err = os.RemoveAll(destDir)
		if err != nil {
			fmt.Println("Error clearing destination directory:", err)
			os.Exit(1)
		}
		// Create destination directory
		err = os.Mkdir(destDir, 0755)
		if err != nil {
			fmt.Println("Error creating destination directory:", err)
			os.Exit(1)
		}
	}

	// Get list of HTML files in source directory
	files, err := getFilesList(sourceDir)
	if err != nil {
		fmt.Println("Error getting files:", err)
		os.Exit(1)
	}

	// // Get a list of 256 random files to convert
	// files = getRandomFiles(files, 1024)

	// Read HTML files to memory and convert to Markdown
	for _, file := range files {
		title, body, err := getTitleAndBody(sourceDir, file)
		if err != nil {
			fmt.Println("Error getting title and body:", err)
			os.Exit(1)
		}
		if err := writeMarkdown(destDir, file, title, body); err != nil {
			fmt.Println("Error writing markdown:", err)
			os.Exit(1)
		}
	}
}

// getFilesList returns a list of HTML files in a directory.
func getFilesList(dir string) ([]string, error) {
	files, err := getFilesRecursive(dir)
	if err != nil {
		fmt.Println("Error getting files:", err)
		return nil, err
	}

	var htmlFiles []string
	for _, file := range files {
		if isHtml(file) {
			htmlFiles = append(htmlFiles, file)
		}
	}

	return htmlFiles, nil
}

// getFilesRecursive returns a list of files in a directory and its subdirectories.
func getFilesRecursive(dir string) ([]string, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		fmt.Println("Error reading directory:", err)
		return nil, err
	}

	// Get files in current directory
	var allFiles []string
	for _, file := range files {
		if file.IsDir() {
			subFiles, err := getFilesRecursive(file.Name())
			if err != nil {
				fmt.Println("Error getting all files:", err)
				return nil, err
			}
			allFiles = append(allFiles, subFiles...)
		} else {
			allFiles = append(allFiles, file.Name())
		}
	}

	return allFiles, nil
}

// isHtml returns true if the file extension is .html.
func isHtml(file string) bool {
	return strings.HasSuffix(file, ".html")
}

// getTitleAndBody returns the title and body of an HTML file.
func getTitleAndBody(sourceDir, file string) (string, string, error) {
	content, err := ioutil.ReadFile(sourceDir + "/" + file)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return "", "", err
	}

	title := getHtmlTitle(content)
	body := getBody(content)

	return title, body, nil
}

// getHtmlTitle returns the extracted title between <title> HTML element tags.
func getHtmlTitle(content []byte) string {
	title := string(content)
	title = strings.Split(title, "<title>")[1]
	title = strings.Split(title, "</title>")[0]
	title = "# " + title

	return title
}

// getBody returns the extracted body between <div id="wiki"> HTML element tags.
func getBody(content []byte) string {
	body := string(content)
	body = strings.Split(body, "<div id=\"wiki\">")[1]
	body = strings.Split(body, "</div>")[0]

	return htmlToMarkdown(body)
}

// Convert HTML string to Markdown string.
func htmlToMarkdown(html string) string {
	converter := md.NewConverter("", true, nil)
	markdown, err := converter.ConvertString(html)
	if err != nil {
		fmt.Println("Error converting HTML to Markdown:", err)
		return ""
	}

	re := regexp.MustCompile(`(?m)\(wiki\%3F(.*?)\.html\)`)
	substitution := "($1)"

	markdown = re.ReplaceAllString(markdown, substitution)
	return markdown
}

// writeMarkdown writes a title and body to a markdown file.
func writeMarkdown(dir, fileName, title, body string) error {
	//fileName := title
	fileName = strings.TrimSuffix(fileName, ".html")
	fileName = fileName + ".md"

	// Remove "wiki?" from file name
	fileName = strings.ReplaceAll(fileName, "wiki?", "")

	filePath := dir + "/" + fileName

	file, err := os.Create(filePath)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return err
	}
	defer file.Close()

	_, err = file.WriteString(title + "\n\n" + body)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return err
	}

	return nil
}

// getRandomFiles returns a list of random files from a list of files.
func getRandomFiles(files []string, num int) []string {
	if num > len(files) {
		num = len(files)
	}

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(files), func(i, j int) { files[i], files[j] = files[j], files[i] })

	return files[:num]
}
