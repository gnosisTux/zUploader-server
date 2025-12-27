package internal

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var (
	IndexTmpl   = template.Must(template.ParseFiles("templates/index.html"))
	DecryptTmpl = template.Must(template.ParseFiles("templates/decrypt.html"))
)

func HandleFileUpload(w http.ResponseWriter, r *http.Request) {
	const pgpHeader = "-----BEGIN PGP MESSAGE-----"

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed. Use POST.", http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, MaxUploadSize)

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error receiving file: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	seeker, ok := file.(io.ReadSeeker)
	if !ok {
		http.Error(w, "Unable to inspect uploaded file", http.StatusInternalServerError)
		return
	}

	buf := make([]byte, len(pgpHeader))
	_, err = seeker.Read(buf)
	if err != nil {
		http.Error(w, "Error reading file", http.StatusBadRequest)
		return
	}

	if string(buf) != pgpHeader {
		http.Error(w, "Upload rejected: file is not PGP encrypted", http.StatusBadRequest)
		return
	}

	_, err = seeker.Seek(0, io.SeekStart)
	if err != nil {
		http.Error(w, "Error resetting file pointer", http.StatusInternalServerError)
		return
	}

	ext := strings.ToLower(filepath.Ext(header.Filename))
	randomName := GenerateRandomName(16) + ext

	os.MkdirAll(ConfigData.UploadDir, os.ModePerm)
	dstPath := filepath.Join(ConfigData.UploadDir, randomName)

	dst, err := os.Create(dstPath)
	if err != nil {
		http.Error(w, "Error saving file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		http.Error(w, "Error writing file content", http.StatusInternalServerError)
		return
	}

	host := r.Host
	fmt.Fprintf(w, "File uploaded successfully. Download at: http://%s/uploads/%s", host, randomName)
}

func HandleFileDownload(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/uploads/")
	if path == "" {
		http.Error(w, "No file specified", http.StatusBadRequest)
		return
	}

	raw := false
	if strings.HasSuffix(path, "/raw") {
		raw = true
		path = strings.TrimSuffix(path, "/raw")
	}

	filePath := filepath.Join(ConfigData.UploadDir, filepath.Clean(path))
	absPath, _ := filepath.Abs(filePath)
	absUploadDir, _ := filepath.Abs(ConfigData.UploadDir)
	if !strings.HasPrefix(absPath, absUploadDir) {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	if _, err := os.Stat(absPath); err != nil {
		http.NotFound(w, r)
		return
	}

	if raw {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", "attachment")
		http.ServeFile(w, r, absPath)
		return
	}

	DecryptTmpl.Execute(w, map[string]string{"File": path})
}

func HandleIndex(w http.ResponseWriter, r *http.Request) {
	clientIP := r.RemoteAddr
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		clientIP = forwarded
	}

	data := map[string]interface{}{
		"IP":        clientIP,
		"MaxUpload": MaxUploadSize / (1024 * 1024),
	}

	IndexTmpl.Execute(w, data)
}
