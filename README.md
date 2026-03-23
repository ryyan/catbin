# catbin

Anonymous zero-knowledge encrypted pastebin

https://github.com/user-attachments/assets/2c3ec0bd-61bd-4ce1-825c-476919376d5f

## Getting started

```sh
cd api
go build
./catbin
# run as background process
# nohup ./catbin > log &
```

## Running tests

```sh
cd api
go test -v -cover
```

## Security

*   **Zero-Knowledge:** Encryption/decryption happens entirely in the browser.
*   **Privacy:** No IP logging or tracking.
*   **Hardened:** 1,000,000 PBKDF2 iterations (SHA-512) and AES-256-GCM.
*   **Ephemeral:** Supports "Burn on Read" and timed expirations.
