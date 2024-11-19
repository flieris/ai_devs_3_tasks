package main

import (
	"ai_devs_3_tasks/helpers"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"
)

func CategorizeFiles(files []fs.DirEntry, path string) (reportFiles, factFiles map[string]helpers.FileData, err error) {
	reportFiles = make(map[string]helpers.FileData)
	factFiles = make(map[string]helpers.FileData)
	for _, file := range files {
		if file.IsDir() {
			factDir, err := os.ReadDir(path + "/" + file.Name())
			if err != nil {
				return nil, nil, err
			}
			for _, fact := range factDir {

				fileFd, err := os.Open(path + "/facts/" + fact.Name())
				if err != nil {
					return nil, nil, err
				}
				defer fileFd.Close()

				output, err := io.ReadAll(fileFd)
				if err != nil {
					return nil, nil, err
				}
				if !strings.Contains(string(output), "entry deleted") {
					log.Printf("Working on fact file %s\n", fact.Name())
					keywords, err := helpers.GetKeywords(string(output))
					if err != nil {
						log.Fatalf("Error: %v", err)
					}

					factFiles[fact.Name()] = helpers.FileData{Content: string(output), Keywords: keywords}
				}
			}
		} else {
			ext := filepath.Ext(file.Name())
			switch ext {
			case ".txt":
				fileFd, err := os.Open(path + "/" + file.Name())
				if err != nil {
					return nil, nil, err
				}
				defer fileFd.Close()

				output, err := io.ReadAll(fileFd)
				if err != nil {
					return nil, nil, err
				}
				log.Printf("Working on report, %s\n", file.Name())
				keywords, err := helpers.GetKeywords(string(output))
				if err != nil {
					log.Fatalf("Error: %v", err)
				}

				reportFiles[file.Name()] = helpers.FileData{Content: string(output), Keywords: keywords}
				time.Sleep(5 * time.Second)
			}
		}
	}
	return
}

func GetCommonFiles(srcFiles *map[string]helpers.FileData, toCompare map[string]helpers.FileData) {
	for srcName, srcData := range *srcFiles {
		for compareName, compareData := range toCompare {
			commonKeywords := findCommonKeywords(strings.Split(srcData.Keywords, ","), strings.Split(compareData.Keywords, ","))
			if len(commonKeywords) > 0 {
				log.Printf("Fil;e %s correlates with %s based on keywords: %v\n", srcName, compareName, commonKeywords)
				srcData.CommonFiles = append(srcData.CommonFiles, compareName)
			}
		}
	}
	fmt.Println(srcFiles)
}

func findCommonKeywords(keywords1, keywords2 []string) []string {
	common := []string{}
	keywordSet := make(map[string]struct{})

	for _, keyword := range keywords1 {
		keywordSet[keyword] = struct{}{}
	}

	for _, keyword := range keywords2 {
		if _, exists := keywordSet[keyword]; exists {
			common = append(common, keyword)
		}
	}

	return common
}

func parseKeywords(keywords string) []string {
	return strings.Split(keywords, ",")
}

// Helper function to identify if a keyword looks like a name
func isName(keyword string) bool {
	words := strings.Fields(keyword) // Split into words by spaces
	if len(words) < 2 {              // Names usually have at least two words
		return false
	}
	for _, word := range words {
		if len(word) == 0 || !unicode.IsUpper(rune(word[0])) {
			return false // Words in a name typically start with uppercase
		}
	}
	return true
}

// Function to extract names dynamically from a list of keywords
func extractNames(keywords []string) map[string]struct{} {
	names := make(map[string]struct{})
	for _, keyword := range keywords {
		trimmed := strings.TrimSpace(keyword)
		if isName(trimmed) {
			names[trimmed] = struct{}{}
		}
	}
	return names
}

// Main function to correlate srcFiles with compareFiles
func CorrelateFilesBetweenGroups(srcFiles map[string]helpers.FileData, compareFiles map[string]helpers.FileData) {
	for srcFileName, srcFileData := range srcFiles {
		keywords1 := parseKeywords(srcFileData.Keywords)
		names1 := extractNames(keywords1)

		updatedSrcFileData := srcFileData
		for compareFileName, compareFileData := range compareFiles {
			keywords2 := parseKeywords(compareFileData.Keywords)
			names2 := extractNames(keywords2)

			// Check if they share at least one name
			for name := range names1 {
				if _, exists := names2[name]; exists {
					// Add correlation to the srcFile's CommonFiles
					//					srcFileData.CommonFiles = appendUnique(srcFileData.CommonFiles, compareFileName)
					updatedSrcFileData.CommonFiles = appendUnique(srcFiles[srcFileName].CommonFiles, compareFileName)
					break
				}
			}
		}
		srcFiles[srcFileName] = updatedSrcFileData
	}

	// Special case: Ensure compareFiles with multiple matching names link to multiple srcFiles
	for compareFileName, compareFileData := range compareFiles {
		keywords2 := parseKeywords(compareFileData.Keywords)
		names2 := extractNames(keywords2)

		for name := range names2 {
			for srcFileName, srcFileData := range srcFiles {
				keywords1 := parseKeywords(srcFileData.Keywords)
				names1 := extractNames(keywords1)

				if _, exists := names1[name]; exists {
					// Add correlation if srcFile matches this name
					updatedSrcFileData := srcFileData
					updatedSrcFileData.CommonFiles = appendUnique(updatedSrcFileData.CommonFiles, compareFileName)
					srcFiles[srcFileName] = updatedSrcFileData
				}
			}
		}
	}
}

// Helper function to append unique values to a slice
func appendUnique(slice []string, value string) []string {
	for _, v := range slice {
		if v == value {
			return slice // Already exists
		}
	}
	return append(slice, value)
}
