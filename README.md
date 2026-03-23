# catbin

Anonymous encrypted pastebin

https://github.com/ryyan/catbin/assets/4228816/35f1689e-d072-46d8-bb80-ae7f53d5cfb0

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
