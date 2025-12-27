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

	"github.com/BurntSushi/toml"
)

// ----------------------------
// Configuration
// ----------------------------
type Config struct {
	UploadDir   string `toml:"upload_dir"`
	Host        string `toml:"host"`
	Port        int    `toml:"port"`
	MaxUploadMB int    `toml:"max_upload_mb"`
}

var config Config
var maxUploadSize int64

// ----------------------------
// Templates
// ----------------------------
var (
	indexTmpl   = template.Must(template.ParseFiles("templates/index.html"))
	decryptTmpl = template.Must(template.ParseFiles("templates/decrypt.html"))
)

// ----------------------------
// Utilities
// ----------------------------
func generateRandomName(n int) string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		if err != nil {
			panic(err)
		}
		b[i] = chars[num.Int64()]
	}
	return string(b)
}

// ----------------------------
// Handlers
// ----------------------------
func handleFileUpload(w http.ResponseWriter, r *http.Request) {
	const pgpHeader = "-----BEGIN PGP MESSAGE-----"

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed. Use POST.", http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

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
	randomName := generateRandomName(16) + ext

	os.MkdirAll(config.UploadDir, os.ModePerm)
	dstPath := filepath.Join(config.UploadDir, randomName)

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

	filePath := filepath.Join(config.UploadDir, filepath.Clean(path))

	absPath, _ := filepath.Abs(filePath)
	absUploadDir, _ := filepath.Abs(config.UploadDir)
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
	// Load configuration from TOML
	if _, err := toml.DecodeFile("config.toml", &config); err != nil {
		fmt.Println("Error loading config.toml:", err)
		os.Exit(1)
	}

	maxUploadSize = int64(config.MaxUploadMB) * 1024 * 1024

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.HandleFunc("/upload", handleFileUpload)
	http.HandleFunc("/uploads/", handleFileDownload)
	http.HandleFunc("/", handler)

	address := fmt.Sprintf("%s:%d", config.Host, config.Port)
	fmt.Println("Server running on", address)
	http.ListenAndServe(address, nil)
}
