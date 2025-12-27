
package main

import (
	"fmt"
	"net/http"

	"github.com/gnosisTux/zUploader-server/internal"
)

func main() {
	internal.LoadConfig("config.toml")

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.HandleFunc("/upload", internal.HandleFileUpload)
	http.HandleFunc("/uploads/", internal.HandleFileDownload)
	http.HandleFunc("/", internal.HandleIndex)

	address := fmt.Sprintf("%s:%d", internal.ConfigData.Host, internal.ConfigData.Port)
	fmt.Println("Server running on", address)
	http.ListenAndServe(address, nil)
}
