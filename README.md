# zUploader - Secure PGP File Uploader (Server)

---

![interface](https://i.imgur.com/d7j7GWb.png)

zUploader is a minimalist server for uploading files encrypted with **symmetric PGP** from the browser. Perfect for secure and private transfers.

> ⚠️ **Important:** The web interface only supports basic symmetric encryption.
> For full functionality and better security, **always use the terminal client**: [https://github.com/gnosisTux/zUploader](https://github.com/gnosisTux/zUploader)

---

## Features

* Upload files encrypted directly in the browser (symmetric PGP only).
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
```

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
2. Select a file and enter an encryption password.
3. Click **Encrypt & Upload**.
4. Get the download link for your file.

> ⚠️ The web interface **only supports symmetric encryption**. For any other functionality, use the terminal client.

---

## Security

* Only files starting with the PGP header (`-----BEGIN PGP MESSAGE-----`) are accepted.
* File names are generated randomly.
* No passwords or sensitive information are stored on the server.

---

## Recommended Client

For a complete and secure workflow, use the **terminal client**: [https://github.com/gnosisTux/zUploader](https://github.com/gnosisTux/zUploader)

---

## License

This project is licensed under **GPLv3**.
See the `LICENSE` file for details.


