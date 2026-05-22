
package main

import (
	"fmt"
	"net/http"

	"github.com/gnosisTux/zUploader-server/internal"
)

func main() {
	internal.LoadConfig("config.toml")
	internal.Init()

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/upload", internal.HandleFileUpload)
	http.HandleFunc("/uploads/", internal.HandleFileDownload)
	http.HandleFunc("/", internal.HandleIndex)

	address := fmt.Sprintf("%s:%d", internal.ConfigData.Host, internal.ConfigData.Port)

	internal.Access.Printf("Server starting on %s", address)
	fmt.Println("Server running on", address)

	if err := http.ListenAndServe(address, internal.LoggingMiddleware(http.DefaultServeMux)); err != nil {
		internal.Error.Printf("[FATAL] Server crashed: %v", err)
		fmt.Println("Server crashed:", err)
		panic(err)
	}
}
