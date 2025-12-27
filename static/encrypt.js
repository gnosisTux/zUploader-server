document.addEventListener("DOMContentLoaded", () => {
    const fileInput = document.getElementById('fileInput');
    const passwordInput = document.getElementById('password');
    const uploadBtn = document.getElementById('uploadBtn');
    const progressBar = document.getElementById('progressBar');
    const linkContainer = document.getElementById('linkContainer');

    let cooldown = false;
    let cooldownEnd = 0;
    let cooldownInterval;

    function startCooldown(seconds) {
        cooldown = true;
        cooldownEnd = Date.now() + seconds * 1000;
        uploadBtn.disabled = true;

        if (!document.getElementById("cooldownMessage")) {
            const msg = document.createElement("div");
            msg.id = "cooldownMessage";
            msg.style.color = "yellow";
            msg.style.marginTop = "10px";
            uploadBtn.parentNode.insertBefore(msg, uploadBtn.nextSibling);
        }

        cooldownInterval = setInterval(updateCooldownMessage, 500);
    }

    function updateCooldownMessage() {
        const msg = document.getElementById("cooldownMessage");
        const remaining = Math.ceil((cooldownEnd - Date.now()) / 1000);
        if (remaining > 0) {
            msg.textContent = `Please wait ${remaining} second(s) before uploading again.`;
        } else {
            clearInterval(cooldownInterval);
            cooldown = false;
            uploadBtn.disabled = false;
            msg.remove();
        }
    }

    uploadBtn.addEventListener('click', async () => {
        if (cooldown) {
            updateCooldownMessage();
            return;
        }

        if (!fileInput.files.length) {
            alert("Select at least one file");
            return;
        }
        if (!passwordInput.value) {
            alert("Enter encryption password");
            return;
        }

        startCooldown(60); // 1 min cooldown

        try {
            // Create a zip containing all selected files
            const zip = new JSZip();
            for (const file of fileInput.files) {
                const arrayBuffer = await file.arrayBuffer();
                zip.file(file.name, arrayBuffer); // keep original file name and extension
            }

            const zipContent = await zip.generateAsync({ type: "uint8array" });

            // Encrypt the zip with PGP symmetric encryption
            const encrypted = await openpgp.encrypt({
                message: await openpgp.createMessage({ binary: zipContent }),
                passwords: [passwordInput.value],
                format: 'armored'
            });

            const encryptedBlob = new Blob([encrypted], { type: 'text/plain' });

            // File to upload
            const encryptedFilename = "batch_upload.zip";

            const formData = new FormData();
            formData.append('file', encryptedBlob, encryptedFilename);

            const xhr = new XMLHttpRequest();
            xhr.open('POST', '/upload', true);

            xhr.upload.onprogress = (e) => {
                if (e.lengthComputable) {
                    progressBar.style.width = (e.loaded / e.total) * 100 + "%";
                }
            };

            xhr.onload = () => {
                progressBar.style.width = "0%";
                linkContainer.innerHTML = "";

                if (xhr.status === 200) {
                    const url = xhr.responseText.split("Download at: ")[1];
                    if (url) {
                        const text = document.createElement("span");
                        text.textContent = "Download your encrypted batch file: ";
                        const a = document.createElement("a");
                        a.href = url.trim();
                        a.textContent = url.trim();
                        a.target = "_blank";
                        a.className = "file-link";
                        linkContainer.appendChild(text);
                        linkContainer.appendChild(a);
                    }
                    fileInput.value = "";
                    passwordInput.value = "";
                } else {
                    const err = document.createElement("div");
                    err.textContent = "Upload failed: " + xhr.statusText;
                    err.style.color = "red";
                    linkContainer.appendChild(err);
                }
            };

            xhr.send(formData);

        } catch (err) {
            alert("Encryption or upload failed: " + err.message);
            clearInterval(cooldownInterval);
            cooldown = false;
            uploadBtn.disabled = false;
            const msg = document.getElementById("cooldownMessage");
            if (msg) msg.remove();
        }
    });
});
