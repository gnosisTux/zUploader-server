package internal

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	IndexTmpl   = template.Must(template.ParseFiles("templates/index.html"))
	DecryptTmpl = template.Must(template.ParseFiles("templates/decrypt.html"))
)

func GetClientIP(r *http.Request) string {
	xff := r.Header.Get("X-Forwarded-For")
	if xff == "" {
		Error.Printf("[WARN] X-Forwarded-For missing! Request from %s will be logged without real IP.", r.RemoteAddr)
		return r.RemoteAddr
	}
	ip := strings.TrimSpace(strings.Split(xff, ",")[0])
	if ip == "" {
		Error.Printf("[WARN] Empty IP in X-Forwarded-For! Request from %s will be logged without real IP.", r.RemoteAddr)
		return r.RemoteAddr
	}
	return ip
}

func HandleFileUpload(w http.ResponseWriter, r *http.Request) {
	const pgpHeader = "-----BEGIN PGP MESSAGE-----"

	ip := GetClientIP(r)
	ua := r.UserAgent()
	timestamp := time.Now().UTC().Format(time.RFC3339)

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
		Error.Printf("[REJECTED] timestamp=%s ip=%s reason=not_pgp ua=\"%s\"", timestamp, ip, ua)
		AuditLog("REJECTED", ip, "", 0, ua, "", "not_pgp")
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
		Error.Printf("[ERROR] timestamp=%s ip=%s reason=create_failed err=%v", timestamp, ip, err)
		http.Error(w, "Error saving file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		Error.Printf("[ERROR] timestamp=%s ip=%s reason=write_failed err=%v", timestamp, ip, err)
		http.Error(w, "Error writing file content", http.StatusInternalServerError)
		return
	}

	proto := r.Header.Get("X-Forwarded-Proto")
	if proto == "" {
		proto = "http"
	}

	downloadURL := fmt.Sprintf("%s://%s/uploads/%s", proto, r.Host, randomName)
	fmt.Fprintf(w, "File uploaded successfully. Download at: %s", downloadURL)

	Access.Printf("[UPLOAD] timestamp=%s ip=%s file=%s size_bytes=%d ua=\"%s\" url=%s",
		timestamp, ip, randomName, header.Size, ua, downloadURL)
	AuditLog("UPLOAD", ip, randomName, header.Size, ua, "", "")
}

func HandleFileDownload(w http.ResponseWriter, r *http.Request) {
	ip := GetClientIP(r)
	ua := r.UserAgent()
	timestamp := time.Now().UTC().Format(time.RFC3339)

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
		Error.Printf("[REJECTED] timestamp=%s ip=%s reason=path_traversal path=%s ua=\"%s\"", timestamp, ip, path, ua)
		AuditLog("REJECTED", ip, path, 0, ua, "", "path_traversal")
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	if _, err := os.Stat(absPath); err != nil {
		Error.Printf("[NOT_FOUND] timestamp=%s ip=%s file=%s ua=\"%s\"", timestamp, ip, path, ua)
		http.NotFound(w, r)
		return
	}

	mode := "view"
	if raw {
		mode = "raw"
	}

	Access.Printf("[DOWNLOAD] timestamp=%s ip=%s file=%s ua=\"%s\" mode=%s",
		timestamp, ip, path, ua, mode)
	AuditLog("DOWNLOAD", ip, path, 0, ua, mode, "")

	if raw {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", "attachment")
		http.ServeFile(w, r, absPath)
		return
	}

	DecryptTmpl.Execute(w, map[string]string{"File": path})
}

func HandleIndex(w http.ResponseWriter, r *http.Request) {
	clientIP := GetClientIP(r)

	data := map[string]interface{}{
		"IP":        clientIP,
		"MaxUpload": MaxUploadSize / (1024 * 1024),
	}

	IndexTmpl.Execute(w, data)
}
