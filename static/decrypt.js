async function decrypt() {
    const password = document.getElementById("password").value;
    if (!password) return alert("Please enter the password!");

    try {
        const res = await fetch(`/uploads/${fileName}/raw`);
        const encrypted = await res.text();

        const message = await openpgp.readMessage({
            armoredMessage: encrypted
        });

        const { data } = await openpgp.decrypt({
            message,
            passwords: [password],
            format: 'binary'
        });

        if (!data || data.byteLength === 0) {
            alert("Incorrect password!");
            return;
        }

        alert("Password correct!");
        const blob = new Blob([data]);
        const a = document.createElement("a");
        a.href = URL.createObjectURL(blob);
        a.download = fileName;
        a.click();

    } catch (err) {
        console.error(err);
        alert("Incorrect password or corrupted file!");
    }
}
