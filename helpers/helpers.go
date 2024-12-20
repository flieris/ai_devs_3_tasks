package helpers

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"
)

type JsonMessage struct {
	MsgID int64  `json:"msgID"`
	Text  string `json:"text"`
}

type JsonAnswer struct {
	Task   string      `json:"task"`
	ApiKey string      `json:"apikey"`
	Answer interface{} `json:"answer"`
}

type JsonResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Hint    string `json:"hint.omitempty"`
	Debug   string `json:"debug,omitempty"`
}

func SendRequest(apiUrl string, requestBody interface{}, responseObj interface{}) error {
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return err
	}

	resp, err := http.Post(apiUrl, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(responseObj); err != nil {
		return err
	}
	return nil
}

func (ja JsonAnswer) String() string {
	jsonData, err := json.Marshal(ja)
	if err != nil {
		return fmt.Sprintf("Error converting JsonAnswer to string: %v", err)
	}
	return string(jsonData)
}

func SendJson(apiUrl string, message JsonMessage) (*JsonMessage, error) {
	var responseObj JsonMessage
	err := SendRequest(apiUrl, message, &responseObj)
	if err != nil {
		return nil, err
	}
	return &responseObj, nil
}

func SendAnswer(apiUrl string, message JsonAnswer) (*JsonResponse, error) {
	var responseObj JsonResponse
	err := SendRequest(apiUrl, message, &responseObj)
	if err != nil {
		return nil, err
	}

	return &responseObj, nil
}

func GetData(apiUrl string) ([]byte, error) {
	// Create a custom HTTP client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Create a request with custom headers
	req, err := http.NewRequest("GET", apiUrl, nil)
	if err != nil {
		return nil, err
	}

	// Make the request
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func DownloadFile(url string) (string, error) {
	// Create a custom HTTP client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Create a request with custom headers
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	// Make the request
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	filename := path.Base(req.URL.Path)
	out, err := os.Create("./" + filename)
	if err != nil {
		return "", nil
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)

	return out.Name(), nil
}

func GetZip(apiUrl string, zipPath string) (err error) {
	zipFileBytes, err := GetData(apiUrl)

	out, err := os.Create(zipPath)
	if err != nil {
		return
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, bytes.NewReader(zipFileBytes))
	if err != nil {
		return
	}
	log.Printf("File downloaded: %s", zipPath)
	return
}

func Unzip(zipFile string, unzipPath string) (err error) {
	archive, err := zip.OpenReader(zipFile)
	if err != nil {
		return
	}
	defer archive.Close()

	for _, f := range archive.File {
		unzipPath := filepath.Join(unzipPath, f.Name)
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(unzipPath, os.ModePerm); err != nil {
				log.Printf("Error: %v", err)
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(unzipPath), os.ModePerm); err != nil {
			log.Fatalf("Error: %v", err)
		}

		dstFile, err := os.OpenFile(unzipPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			continue
		}

		src, err := f.Open()
		if err != nil {
			continue
		}

		if _, err := io.Copy(dstFile, src); err != nil {
			continue
		}

		dstFile.Close()
		src.Close()

	}
	return
}

func UnzipSpecificFile(zipFile string, unzipPath string, specFile string) (err error) {
	archive, err := zip.OpenReader(zipFile)
	if err != nil {
		return
	}
	defer archive.Close()

	for _, f := range archive.File {
		if f.Name != specFile {
			continue
		}
		if err := os.MkdirAll(filepath.Dir(unzipPath), os.ModePerm); err != nil {
			log.Fatalf("Error: %v", err)
		}

		dstFile, err := os.OpenFile(unzipPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			continue
		}

		src, err := f.Open()
		if err != nil {
			continue
		}

		if _, err := io.Copy(dstFile, src); err != nil {
			continue
		}

		dstFile.Close()
		src.Close()
		break
	}
	return
}

func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}
