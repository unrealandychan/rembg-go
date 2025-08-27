package models

import (
    "fmt"
    "net/http"
    "os"
    "path/filepath"
)

// DownloadModel downloads a model from URL to the models directory.
func DownloadModel(url, destDir string) (string, error) {
    if url == "" {
        return "", fmt.Errorf("url required")
    }
    if destDir == "" {
        destDir = "models"
    }
    if err := os.MkdirAll(destDir, 0o755); err != nil {
        return "", err
    }
    // naive downloader (no progress)
    resp, err := http.Get(url)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()
    name := filepath.Base(resp.Request.URL.Path)
    outPath := filepath.Join(destDir, name)
    f, err := os.Create(outPath)
    if err != nil {
        return "", err
    }
    defer f.Close()
    if _, err := f.ReadFrom(resp.Body); err != nil {
        return "", err
    }
    return outPath, nil
}
