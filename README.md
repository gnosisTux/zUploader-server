# zUploader - Secure PGP File Uploader (Server)

---

![interface](https://i.imgur.com/d7j7GWb.png)
![encryptedfile](https://i.imgur.com/eN4HCn2.png)

zUploader is a minimalist server for uploading and **decrypting files with symmetric PGP directly from the browser**.  
This server is designed to work alongside the **zUploader terminal client**, which adds **asymmetric encryption** and optimized CLI usage: [https://github.com/gnosisTux/zUploader](https://github.com/gnosisTux/zUploader)

---

## Features

* Upload files encrypted directly in the browser (symmetric PGP only).
* Decrypt files directly from the browser.
* Configurable maximum file size (default 500 MB).
* Files saved with random names for extra security.
* Direct download via unique URL.
* Minimal and lightweight: only Go and HTML/CSS/JS.
* Progress bar and cooldown to prevent upload spamming.

---

## Installation

1. Clone the repository:

```bash
git clone https://github.com/gnosisTux/zUploader-server.git
cd zUploader-server
````

2. Run the server:
    

```bash
go run main.go
```

The server will start at `http://localhost:8001`.

---

## Directory structure

```
server/
├── LICENSE         # GPLv3 License
├── README.md       # This file
├── main.go         # Go server
├── static/         # Static files (JS, CSS, images)
├── templates/      # HTML templates
└── uploads/        # Uploaded files
```

---

## Web Usage (basic only)

1. Open `http://localhost:8001` in your browser.
    
2. To upload: select a file and enter an encryption password. Click **Encrypt & Upload**.
    
3. To decrypt: open the file link, enter the password, and click **Decrypt & Download**.
    
4. Receive your decrypted file directly in the browser.
    

> ⚠️ The web interface **only supports symmetric encryption**.

---

## Security

- Only files starting with the PGP header (`-----BEGIN PGP MESSAGE-----`) are accepted.
    
- File names are generated randomly.
    
- No passwords or sensitive information are stored on the server.
    

---

## License

This project is licensed under **GPLv3**.  
See the `LICENSE` file for details.
