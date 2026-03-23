import { encrypt, decrypt } from './crypto.js';

// Global reference to DOM elements
const EL = {
  encryptSection: document.getElementById('encrypt-section'),
  encryptForm: document.getElementById('encrypt-form'),
  encryptText: document.getElementById('encrypt-text'),
  encryptPass: document.getElementById('encrypt-secret'),
  encryptExp: document.getElementById('encrypt-expiration'),
  encryptSubmit: document.getElementById('encrypt-submit'),
  encryptStatus: document.getElementById('encrypt-status'),
  charCounter: document.getElementById('char-counter'),

  decryptSection: document.getElementById('decrypt-section'),
  decryptForm: document.getElementById('decrypt-form'),
  decryptText: document.getElementById('decrypt-text'),
  decryptPass: document.getElementById('decrypt-secret'),
  decryptExp: document.getElementById('decrypt-expiration'),
  decryptStatus: document.getElementById('decrypt-status'),
  copyBtn: document.getElementById('copy-link'),
};

document.addEventListener('DOMContentLoaded', () => {
  // If a path exists, we are in decryption/view mode
  const id = window.location.pathname.substring(1);
  if (id) {
    showSection(EL.decryptSection);
    initDecryptMode(id);
  } else {
    showSection(EL.encryptSection);
    initEncryptMode();
  }
});

function initEncryptMode() {
  let plaintext = '';

  // Update char counter and UI state as the user types
  EL.encryptText.addEventListener('input', () => {
    plaintext = EL.encryptText.value;
    const len = plaintext.length;
    EL.charCounter.innerText = `${len.toLocaleString()} / 10,000`;
    
    // Only enable password field if there is content to encrypt
    EL.encryptPass.disabled = len === 0;
    if (len === 0) EL.encryptPass.value = '';
    EL.encryptSubmit.disabled = !(len > 0 && EL.encryptPass.value);
  });

  EL.encryptPass.addEventListener('input', () => {
    EL.encryptSubmit.disabled = !(plaintext && EL.encryptPass.value);
  });

  EL.encryptForm.addEventListener('submit', async (e) => {
    e.preventDefault();
    if (!plaintext || !EL.encryptPass.value) return;

    EL.encryptSubmit.disabled = true;
    EL.encryptStatus.innerText = 'Encrypting (1M iterations)...';
    EL.encryptStatus.classList.remove('error-text', 'success-text');

    try {
      // Step 1: Client-side AES-256-GCM Encryption
      const finalEncrypted = await encrypt(EL.encryptPass.value, plaintext);
      EL.encryptStatus.innerText = 'Uploading...';
      
      const formData = new URLSearchParams();
      formData.append('text', finalEncrypted);
      formData.append('expiration', EL.encryptExp.value);

      // Step 2: Upload to API
      const res = await fetch('/msg', {
        method: 'POST',
        body: formData,
        headers: { 'Content-Type': 'application/x-www-form-urlencoded' }
      });

      if (!res.ok) throw new Error(await res.text());

      const data = await res.json();
      const newId = data.id;
      const actualExp = data.expiration;
      
      // Update URL without reloading to prevent premature "burn" on creator verification
      window.history.pushState({}, '', `/${newId}`);

      // Transition to decryption view using local data to verify password
      hideSection(EL.encryptSection);
      showSection(EL.decryptSection);
      
      EL.decryptText.value = finalEncrypted;
      EL.decryptExp.value = actualExp === 'burn' 
        ? 'Burn on Read (Preview)' 
        : formatExpiration(actualExp);
      
      setupDecryptHandler(finalEncrypted);
      EL.decryptStatus.innerText = 'Paste created! You are now in preview mode.';
      EL.decryptStatus.classList.add('success-text');
    } catch (err) {
      EL.encryptStatus.innerText = `Error: ${err.message}`;
      EL.encryptStatus.classList.add('error-text');
      EL.encryptSubmit.disabled = false;
    }
  });
}

function initDecryptMode(id) {
  fetch(`/msg/${id}`)
    .then(res => {
      if (!res.ok) throw new Error('Text not found or already burned');
      return res.json();
    })
    .then(data => {
      EL.decryptText.value = data.text;
      EL.decryptExp.value = data.expiration === 'burn' 
        ? 'Burn on Read (Burned)' 
        : formatExpiration(data.expiration);
      setupDecryptHandler(data.text);
    })
    .catch(err => {
      alert(err.message);
      window.location.href = window.location.origin;
    });
}

function setupDecryptHandler(encryptedData) {
  let decryptionTimeout = null;
  
  // Replace the password input element to clear all old event listeners
  const newPassEl = EL.decryptPass.cloneNode(true);
  EL.decryptPass.parentNode.replaceChild(newPassEl, EL.decryptPass);
  EL.decryptPass = newPassEl;

  // Real-time decryption as the user types the password
  EL.decryptPass.addEventListener('input', () => {
    EL.decryptStatus.innerText = '';
    EL.decryptStatus.classList.remove('error-text', 'success-text');
    
    if (decryptionTimeout) clearTimeout(decryptionTimeout);
    if (!EL.decryptPass.value) {
      EL.decryptText.value = encryptedData;
      return;
    }

    // Debounce to prevent UI lockup during expensive PBKDF2 calculation
    decryptionTimeout = setTimeout(async () => {
      EL.decryptStatus.innerText = 'Decrypting...';
      try {
        const decrypted = await decrypt(EL.decryptPass.value, encryptedData);
        EL.decryptText.value = decrypted;
        EL.decryptStatus.innerText = 'Decrypted successfully';
        EL.decryptStatus.classList.add('success-text');
      } catch (err) {
        EL.decryptText.value = encryptedData;
        EL.decryptStatus.innerText = (err.name === 'OperationError') ? 'Incorrect password' : 'Decryption failed';
        EL.decryptStatus.classList.add('error-text');
      }
    }, 300);
  });

  // Re-bind the copy button listener
  const newCopyBtn = EL.copyBtn.cloneNode(true);
  EL.copyBtn.parentNode.replaceChild(newCopyBtn, EL.copyBtn);
  EL.copyBtn = newCopyBtn;
  
  EL.copyBtn.addEventListener('click', () => {
    navigator.clipboard.writeText(window.location.href).then(() => {
      const originalText = EL.copyBtn.innerText;
      EL.copyBtn.innerText = 'Copied!';
      setTimeout(() => EL.copyBtn.innerText = originalText, 2000);
    });
  });
}

function showSection(section) {
  section.style.display = 'block';
  section.setAttribute('aria-hidden', 'false');
}

function hideSection(section) {
  section.style.display = 'none';
  section.setAttribute('aria-hidden', 'true');
}

function formatExpiration(dateStr) {
  try {
    const d = new Date(dateStr);
    return d.toLocaleString(undefined, {
      year: 'numeric', month: 'short', day: 'numeric',
      hour: '2-digit', minute: '2-digit'
    });
  } catch (e) {
    return dateStr;
  }
}
