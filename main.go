package main

import (
	"crypto/rand"
	"fmt"
	"html/template"
	"io"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// ----------------------------
// Configuration
// ----------------------------
const (
	uploadDir     = "./uploads/"               // Directory to store uploaded files
	maxUploadSize = 500 * 1024 * 1024          // Maximum allowed upload size (500 MB)
)

// ----------------------------
// Templates
// ----------------------------
var (
	indexTmpl   = template.Must(template.ParseFiles("templates/index.html"))    // Main page template
	decryptTmpl = template.Must(template.ParseFiles("templates/decrypt.html"))  // Decrypt page template
)

// ----------------------------
// Utilities
// ----------------------------

// generateRandomName generates a cryptographically secure random string
// of length n using letters and digits.
func generateRandomName(n int) string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		if err != nil {
			panic(err) // Extremely rare error if system entropy is unavailable
		}
		b[i] = chars[num.Int64()]
	}
	return string(b)
}

// ----------------------------
// Handlers
// ----------------------------

// handleFileUpload handles file upload requests.
// It validates that the uploaded file is PGP encrypted, limits its size,
// generates a secure random filename, and saves it to the upload directory.
func handleFileUpload(w http.ResponseWriter, r *http.Request) {
	const pgpHeader = "-----BEGIN PGP MESSAGE-----"

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed. Use POST.", http.StatusMethodNotAllowed)
		return
	}

	// Limit upload size to prevent memory/disk exhaustion
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error receiving file: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Verify that the uploaded file starts with PGP header
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

	// Reset file pointer after reading header
	_, err = seeker.Seek(0, io.SeekStart)
	if err != nil {
		http.Error(w, "Error resetting file pointer", http.StatusInternalServerError)
		return
	}

	// Preserve original extension and generate random filename
	ext := strings.ToLower(filepath.Ext(header.Filename))
	randomName := generateRandomName(16) + ext

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

// handleFileDownload handles requests to download uploaded files.
// It prevents path traversal, serves files as attachment if requested,
// or renders the decryption page.
func handleFileDownload(w http.ResponseWriter, r *http.Request) {
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

	filePath := filepath.Join(uploadDir, filepath.Clean(path))

	// Prevent path traversal by ensuring the absolute path is within uploadDir
	absPath, _ := filepath.Abs(filePath)
	absUploadDir, _ := filepath.Abs(uploadDir)
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

	decryptTmpl.Execute(w, map[string]string{
		"File": path,
	})
}

// handler renders the main index page, displaying the client IP
// and maximum upload size. X-Forwarded-For header is used if present.
func handler(w http.ResponseWriter, r *http.Request) {
	clientIP := r.RemoteAddr
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		clientIP = forwarded
	}

	data := map[string]interface{}{
		"IP":        clientIP,
		"MaxUpload": maxUploadSize / (1024 * 1024),
	}

	indexTmpl.Execute(w, data)
}

// ----------------------------
// Main
// ----------------------------
func main() {
	// Serve static assets
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Define route handlers
	http.HandleFunc("/upload", handleFileUpload)
	http.HandleFunc("/uploads/", handleFileDownload)
	http.HandleFunc("/", handler)

	fmt.Println("Server running on :8001")
	http.ListenAndServe(":8001", nil)
}
