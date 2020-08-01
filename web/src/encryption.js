// Thanks to https://github.com/MauriceButler/cryptr for the encrypt/decrypt code
const crypto = require("crypto");

const ivLength = 16; // initialization vector
const saltLength = 69;
const tagLength = 16;
const tagPosition = saltLength + ivLength;
const encryptedPosition = tagPosition + tagLength;

function hashKey(key, salt) {
  return crypto.pbkdf2Sync(key, salt, 1420, 32, "sha512");
}

module.exports = {
  encrypt: (key, text) => {
    const iv = crypto.randomBytes(ivLength);
    const salt = crypto.randomBytes(saltLength);
    const hashedKey = hashKey(key, salt);
    const cipher = crypto.createCipheriv("aes-256-gcm", hashedKey, iv);
    const encrypted = Buffer.concat([
      cipher.update(String(text), "utf8"),
      cipher.final(),
    ]);
    const tag = cipher.getAuthTag();
    return Buffer.concat([salt, iv, tag, encrypted]).toString("base64");
  },

  decrypt: (key, text) => {
    const str = Buffer.from(String(text), "base64");
    const salt = str.slice(0, saltLength);
    const iv = str.slice(saltLength, tagPosition);
    const tag = str.slice(tagPosition, encryptedPosition);
    const encrypted = str.slice(encryptedPosition);
    const hashedKey = hashKey(key, salt);
    const decipher = crypto.createDecipheriv("aes-256-gcm", hashedKey, iv);
    decipher.setAuthTag(tag);
    return decipher.update(encrypted) + decipher.final("utf8");
  },
};
