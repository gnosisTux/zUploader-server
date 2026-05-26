# zUploader - Secure PGP File Uploader (Server)

---

![interface](https://i.imgur.com/d7j7GWb.png)

![encryptedfile](https://i.imgur.com/eN4HCn2.png)

zUploader is a minimalist server for uploading and **decrypting files with symmetric PGP directly from the browser**.  
This server is designed to work alongside the **zUploader terminal client**, which adds **asymmetric encryption** and optimized CLI usage: [https://github.com/gnosisTux/zUploader](https://github.com/gnosisTux/zUploader)

---

## Features

- Upload files encrypted directly in the browser (symmetric PGP only)
- Decrypt files directly from the browser
- Configurable maximum file size (default 500 MB)
- Files saved with random names for extra security
- Direct download via unique URL
- Structured logging: access log to stdout, errors to stderr
- SQLite audit database with automatic purge after 12 months
- Progress bar and cooldown to prevent upload spamming
- Minimal and lightweight: only Go and HTML/CSS/JS
- Deployment support: Docker, systemd, and FreeBSD rc

---

## Directory structure

```
zUploader-server/
├── LICENSE
├── README.md
├── config.toml             # Main configuration file
├── main.go                 # Entry point
├── go.mod
├── go.sum
├── docker/
│   ├── Dockerfile          # Multi-stage Docker build
│   ├── docker-compose.yml
│   └── config.toml         # Config for Docker deployment
├── init/
│   ├── systemd/
│   │   └── zuploader.service
│   └── freebsd-rc/
│       └── zuploader
├── internal/
│   ├── audit.go            # SQLite audit logging
│   ├── config.go           # Config loader
│   ├── handlers.go         # HTTP handlers
│   ├── init.go             # Initialization
│   ├── logger.go           # Structured logger
│   ├── middleware.go       # HTTP middleware
│   ├── purge.go            # Scheduled audit purge
│   └── utils.go            # Utilities (random name generator)
├── static/
│   ├── encrypt.js
│   ├── decrypt.js
│   ├── style.css
│   └── zuploader.png
└── templates/
    ├── index.html
    └── decrypt.html
```

---

## Configuration

All configuration is done via `config.toml`:

```toml
upload_dir   = "./uploads/"     # Directory where uploaded files are stored
host         = "0.0.0.0"       # Bind address
port         = 8002             # Bind port
max_upload_mb = 500             # Maximum upload size in MB
audit_db     = "audit.db"      # Path to the SQLite audit database
```

> For Docker deployments, use `docker/config.toml` which sets absolute paths under `/srv/zuploader/`.

---

## Installation

### Manual (from source)

Requirements: Go 1.21+

```bash
git clone https://github.com/gnosisTux/zUploader-server.git
cd zUploader-server
go build -o zuploader main.go
./zuploader
```

The server will start on the address defined in `config.toml`.

### Docker Compose

```bash
cd docker/
docker compose up -d
```

This will:
- Build the image from the multi-stage `Dockerfile`
- Bind port `8002`
- Mount `/srv/zuploader` for persistent uploads and audit database
- Mount `docker/config.toml` as the configuration file (read-only)

To use a custom config, edit `docker/config.toml` before starting.

---

## Deployment

### systemd (Linux)

1. Build the binary and place it at `/home/zuploader/zuploader`
2. Create a dedicated user:

```bash
useradd --system --no-create-home --shell /usr/sbin/nologin zuploader
```

3. Copy the service file:

```bash
cp init/systemd/zuploader.service /etc/systemd/system/
systemctl daemon-reload
systemctl enable --now zuploader
```

4. Check logs:

```bash
# Access log (stdout)
journalctl -u zuploader -o cat

# Filter by identifier
journalctl -t zuploader
```

### FreeBSD rc

1. Place the binary at `/home/zuploader/zuploader`
2. Create the `zuploader` user
3. Copy the rc script:

```bash
cp init/freebsd-rc/zuploader /usr/local/etc/rc.d/zuploader
chmod +x /usr/local/etc/rc.d/zuploader
```

4. Enable in `/etc/rc.conf`:

```
zuploader_enable="YES"
```

5. Start the service:

```bash
service zuploader start
```

Logs are written to `/var/log/zuploader/access.log`.

---

## Logging

zUploader uses two separate log streams:

| Stream | Content | Destination |
|--------|---------|-------------|
| stdout | Access log: uploads, downloads, requests | journald / log file |
| stderr | Errors, warnings, audit DB events | journald / stderr |

Each log line includes: timestamp, action, IP, file, size, user-agent, and HTTP status/duration where applicable.

Example entries:

```
[UPLOAD]   timestamp=... ip=1.2.3.4 file=aBcD1234.txt size_bytes=204800 ua="curl/8.0" url=https://...
[DOWNLOAD] timestamp=... ip=1.2.3.4 file=aBcD1234.txt ua="Mozilla/5.0" mode=raw
[REJECTED] timestamp=... ip=1.2.3.4 reason=not_pgp ua="curl/8.0"
[REQUEST]  method=POST path=/upload ip=1.2.3.4 status=200 duration=142ms
```

---

## Audit Database

All upload, download, and rejection events are stored in a SQLite database (`audit.db` by default).

Schema:

```sql
CREATE TABLE audit (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp  TEXT NOT NULL,
    action     TEXT NOT NULL,   -- UPLOAD, DOWNLOAD, REJECTED
    ip         TEXT NOT NULL,
    file       TEXT,
    size_bytes INTEGER,
    ua         TEXT,
    mode       TEXT,            -- raw / view (downloads only)
    reason     TEXT             -- not_pgp, path_traversal (rejections only)
);
```

Records older than **12 months** are automatically purged once per day.

---

## Web Usage

1. Open `http://your-server:8002` in your browser
2. **Upload:** select one or more files and enter an encryption password, then click **Encrypt & Upload**
   - Multiple files are automatically bundled into a single encrypted zip
3. **Decrypt:** open the file link, enter the password, click **Decrypt & Download**
4. The file is decrypted entirely in the browser — the server never sees the password

> The web interface only supports **symmetric PGP encryption**.

---

## CLI Usage

For full functionality including asymmetric encryption, use the CLI tools:
[https://github.com/gnosisTux/zUploader](https://github.com/gnosisTux/zUploader)

```bash
# Encrypt & upload asymmetrically
python3 zuploader.py --armor user@example.com path/to/file

# Encrypt & upload symmetrically
python3 zuploader.py --sym-password yourpassword path/to/file

# Download & decrypt
python3 zget.py URL_TO_FILE
```

---

## Reverse Proxy

zUploader **must** be deployed behind a reverse proxy (nginx, Caddy, etc.). It relies on the `X-Forwarded-For` header to log the real client IP — if the proxy does not set this header, or if the server is exposed directly to the internet, any client can spoof their IP in logs and the audit database simply by setting that header themselves.

The proxy must:
1. Set `X-Forwarded-For` to the real client IP
2. Strip any `X-Forwarded-For` header coming from the client before forwarding

Example nginx config:

```nginx
location / {
    proxy_pass http://127.0.0.1:8002;
    proxy_set_header X-Forwarded-For $remote_addr;
    proxy_set_header X-Forwarded-Proto $scheme;
    proxy_set_header Host $host;
}
```

> Note: use `$remote_addr` and not `$proxy_add_x_forwarded_for` — the latter appends to any existing header the client may have sent, which defeats the purpose.

Example Caddy config:

```
your.domain {
    reverse_proxy 127.0.0.1:8002 {
        header_up X-Forwarded-For {remote_host}
        header_up X-Forwarded-Proto {scheme}
    }
}
```

---

## Security

- Only files starting with the PGP header (`-----BEGIN PGP MESSAGE-----`) are accepted
- File names are generated randomly using a cryptographically secure random source
- Path traversal is blocked on the download endpoint
- No passwords or sensitive information are stored on the server
- The upload endpoint has no authentication — access control should be handled at the reverse proxy level (IP allowlist, auth middleware, etc.)

---

## License

This project is licensed under **GPLv3**.  
See the `LICENSE` file for details.
## License

This project is licensed under **GPLv3**.  
See the `LICENSE` file for details.
