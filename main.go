package main

import (
	"fmt"
	"html/template"
	"io"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ----------------------------
// Configuration
// ----------------------------
const uploadDir = "./uploads/"
const maxUploadSize = 500 * 1024 * 1024 // 500 MB
const pgpHeader = "-----BEGIN PGP MESSAGE-----"

// ----------------------------
// Utilities
// ----------------------------
func generateRandomName(n int) string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, n)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
}

// ----------------------------
// Handlers
// ----------------------------
func handleFileUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed. Use POST.", http.StatusMethodNotAllowed)
		return
	}

	// Limit upload size
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error receiving file: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// 🔐 Check PGP header
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
		http.Error(
			w,
			"Upload rejected: file is not PGP encrypted",
			http.StatusBadRequest,
		)
		return
	}

	// Reset file pointer after inspection
	_, err = seeker.Seek(0, io.SeekStart)
	if err != nil {
		http.Error(w, "Error resetting file pointer", http.StatusInternalServerError)
		return
	}

	// Keep original extension (.gpg / .pgp / whatever)
	ext := strings.ToLower(filepath.Ext(header.Filename))
	randomName := generateRandomName(12) + ext

	os.MkdirAll(uploadDir, os.ModePerm)
	dstPath := filepath.Join(uploadDir, randomName)

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
	fmt.Fprintf(
		w,
		"File uploaded successfully. Download at: http://%s/uploads/%s",
		host,
		randomName,
	)
}

func handleFileDownload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed. Use GET.", http.StatusMethodNotAllowed)
		return
	}

	filename := strings.TrimPrefix(r.URL.Path, "/uploads/")
	if filename == "" {
		http.Error(w, "No file specified", http.StatusBadRequest)
		return
	}

	filePath := filepath.Join(uploadDir, filepath.Clean(filename))
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.NotFound(w, r)
		return
	}

	http.ServeFile(w, r, filePath)
}

func handler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/index.html"))

	clientIP := r.RemoteAddr
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		clientIP = forwarded
	}

	data := map[string]interface{}{
		"IP":        clientIP,
		"MaxUpload": maxUploadSize / (1024 * 1024),
	}

	tmpl.Execute(w, data)
}

// ----------------------------
// Main
// ----------------------------
func main() {
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.HandleFunc("/upload", handleFileUpload)
	http.HandleFunc("/uploads/", handleFileDownload)
	http.HandleFunc("/", handler)

	fmt.Println("Server running on :8001")
	http.ListenAndServe(":8001", nil)
}
