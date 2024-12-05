package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

type ServerInfo struct {
	URL      string
	Username string
	Password string
}
type FileInfo struct {
	FilePath string
	FileName string
	FileType string
}

func main() {
	// set up
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	err := godotenv.Load()
	if err != nil {
		logger.Error("cant load dotenv file", "detail", err)
		os.Exit(1)
	}
	// set Env data
	info := ServerInfo{
		URL:      os.Getenv("WEBDAV_URL"),
		Username: os.Getenv("BASIC_USERNAME"),
		Password: os.Getenv("BASIC_PASSWORD"),
	}

	// Get file
	fInfo := FileInfo{}
	fmt.Println("Movie file path.....")
	if _, err := fmt.Scan(&fInfo.FilePath); err != nil {
		logger.Error("cant scan file path", "detail", err)
		os.Exit(1)
	}
	file, err := os.Open(fInfo.FilePath)
	if err != nil {
		logger.Error("cant open file", "detail", err)
	}
	defer file.Close()

	// Read file
	fileContents, err := io.ReadAll(file)
	if err != nil {
		logger.Error("cant read file", "detail", err)
	}
	fInfo.FileType = http.DetectContentType(fileContents)
	_, fInfo.FileName = filepath.Split(fInfo.FilePath)

	// Echo infos
	logger.Info("server info", "ServerInfo", info)
	logger.Info("upload file info", "FileInfo", fInfo)

	// Create http Request
	req, err := http.NewRequest("PUT", info.URL, bytes.NewReader(fileContents))
	if err != nil {
		logger.Error("NewRequest error", "detail", err)
	}

	// Create BasicAuth
	auth := info.Username + ":" + info.Password
	encodedAuth := base64.StdEncoding.EncodeToString([]byte(auth))
	req.Header.Add("Authorization", "Basic "+encodedAuth)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.Error("cant send file", "detail", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusOK {
		fmt.Println("ファイルが正常にアップロードされました")
	} else {
		fmt.Printf("ファイルのアップロードに失敗しました: %s\n", resp.Status)
	}
}
