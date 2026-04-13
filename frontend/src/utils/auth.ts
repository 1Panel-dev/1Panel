import CryptoJS from 'crypto-js';
import JSEncrypt from 'jsencrypt';

export function getCookie(name: string) {
    const value = `; ${document.cookie}`;
    const parts = value.split(`; ${name}=`);
    if (parts.length === 2) return parts.pop()?.split(';').shift();
}

function rsaEncrypt(data: string, publicKey: string) {
    if (!data) {
        return data;
    }
    const jsEncrypt = new JSEncrypt();
    jsEncrypt.setPublicKey(publicKey);
    return jsEncrypt.encrypt(data);
}

function aesEncrypt(data: string, key: string) {
    const keyBytes = CryptoJS.enc.Utf8.parse(key);
    const iv = CryptoJS.lib.WordArray.random(16);
    const encrypted = CryptoJS.AES.encrypt(data, keyBytes, {
        iv: iv,
        mode: CryptoJS.mode.CBC,
        padding: CryptoJS.pad.Pkcs7,
    });
    return iv.toString(CryptoJS.enc.Base64) + ':' + encrypted.toString();
}

function urlDecode(value: string): string {
    return decodeURIComponent(value.replace(/\+/g, ' '));
}

function generateAESKey(): string {
    const keyLength = 16;
    const randomBytes = new Uint8Array(keyLength);
    crypto.getRandomValues(randomBytes);
    return Array.from(randomBytes)
        .map((b) => b.toString(16).padStart(2, '0'))
        .join('');
}

export const encryptPassword = (password: string) => {
    if (!password) {
        return '';
    }
    let rsaPublicKeyText = getCookie('panel_public_key');
    if (!rsaPublicKeyText) {
        return password;
    }
    rsaPublicKeyText = urlDecode(rsaPublicKeyText);

    const aesKey = generateAESKey();
    rsaPublicKeyText = rsaPublicKeyText.replaceAll('"', '');
    const rsaPublicKey = atob(rsaPublicKeyText);
    const keyCipher = rsaEncrypt(aesKey, rsaPublicKey);
    const passwordCipher = aesEncrypt(password, aesKey);
    return `${keyCipher}:${passwordCipher}`;
};

export function base64UrlToBuffer(base64url: string): ArrayBuffer {
    let base64 = base64url.replace(/-/g, '+').replace(/_/g, '/');
    const pad = base64.length % 4;
    if (pad) {
        base64 += '='.repeat(4 - pad);
    }
    const binary = window.atob(base64);
    const bytes = new Uint8Array(binary.length);
    for (let i = 0; i < binary.length; i++) {
        bytes[i] = binary.charCodeAt(i);
    }
    return bytes.buffer;
}

export function bufferToBase64Url(buffer: ArrayBuffer): string {
    const bytes = new Uint8Array(buffer);
    let binary = '';
    bytes.forEach((byte) => {
        binary += String.fromCharCode(byte);
    });
    const base64 = window.btoa(binary);
    return base64.replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '');
}
