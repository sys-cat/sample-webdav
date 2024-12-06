package main

import (
	"encoding/base64"
	"fmt"
	"log/slog"
	"net/http"
	"os"

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
	fileInfo, err := file.Stat()
	if err != nil {
		logger.Error("cant read file", "detail", err)
	}
	{
		// pick 512Bytes for DetectContentType
		buffer := make([]byte, 512)
		file.Read(buffer)
		// Check mimetype
		fInfo.FileType = http.DetectContentType(buffer)
		// reset read point
		file.Seek(0, 0)
	}
	// set file name. ps: not use filepath.Split
	fInfo.FileName = fileInfo.Name()

	// Echo infos
	logger.Info("server info", "ServerInfo", info)
	logger.Info("upload file info", "FileInfo", fInfo)

	// Create http Request
	req, err := http.NewRequest("PUT", info.URL, file)
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
