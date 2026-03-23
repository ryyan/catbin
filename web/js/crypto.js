const ivLength = 12;
const saltLength = 64;
const pbkdf2Iterations = 1000000;

function getCrypto() {
  const cryptoObj = self.crypto || self.msCrypto;
  if (!cryptoObj || !cryptoObj.subtle) {
    throw new Error("Web Crypto API is not available. Please ensure you are using a secure context (HTTPS or localhost).");
  }
  return cryptoObj;
}

function toBase64(bytes) {
  let binary = '';
  const len = bytes.byteLength;
  for (let i = 0; i < len; i++) {
    binary += String.fromCharCode(bytes[i]);
  }
  return btoa(binary);
}

function fromBase64(base64) {
  const binary = atob(base64);
  const bytes = new Uint8Array(binary.length);
  for (let i = 0; i < binary.length; i++) {
    bytes[i] = binary.charCodeAt(i);
  }
  return bytes;
}

async function hashKey(keyStr, salt) {
  const crypto = getCrypto();
  const enc = new TextEncoder();
  const keyMaterial = await crypto.subtle.importKey(
    "raw",
    enc.encode(keyStr),
    { name: "PBKDF2" },
    false,
    ["deriveBits", "deriveKey"]
  );
  
  return crypto.subtle.deriveKey(
    {
      name: "PBKDF2",
      salt: salt,
      iterations: pbkdf2Iterations,
      hash: "SHA-512"
    },
    keyMaterial,
    { name: "AES-GCM", length: 256 },
    true,
    ["encrypt", "decrypt"]
  );
}

export async function encrypt(keyStr, text) {
  const crypto = getCrypto();
  const iv = crypto.getRandomValues(new Uint8Array(ivLength));
  const salt = crypto.getRandomValues(new Uint8Array(saltLength));
  
  const key = await hashKey(keyStr, salt);
  const enc = new TextEncoder();
  
  const encryptedBuf = await crypto.subtle.encrypt(
    { name: "AES-GCM", iv: iv },
    key,
    enc.encode(text)
  );
  
  const encryptedBytes = new Uint8Array(encryptedBuf);
  const result = new Uint8Array(salt.length + iv.length + encryptedBytes.length);
  result.set(salt, 0);
  result.set(iv, salt.length);
  result.set(encryptedBytes, salt.length + iv.length);
  
  return toBase64(result);
}

export async function decrypt(keyStr, base64Text) {
  const crypto = getCrypto();
  const bytes = fromBase64(base64Text);
  const salt = bytes.slice(0, saltLength);
  const iv = bytes.slice(saltLength, saltLength + ivLength);
  const data = bytes.slice(saltLength + ivLength);
  
  const key = await hashKey(keyStr, salt);
  
  const decryptedBuf = await crypto.subtle.decrypt(
    { name: "AES-GCM", iv: iv },
    key,
    data
  );
  
  const dec = new TextDecoder();
  return dec.decode(decryptedBuf);
}
