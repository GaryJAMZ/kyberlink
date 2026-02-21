// Copyright 2026 José Antonio Garibay Marcelo
// SPDX-License-Identifier: Apache-2.0

import { MlKem1024 } from 'mlkem';

export const b64e = (bytes: Uint8Array): string => {
    let binary = '';
    const len = bytes.byteLength;
    for (let i = 0; i < len; i++) {
        binary += String.fromCharCode(bytes[i]);
    }
    return btoa(binary);
};

export const b64d = (str: string): Uint8Array => {
    const binaryString = atob(str);
    const len = binaryString.length;
    const bytes = new Uint8Array(len);
    for (let i = 0; i < len; i++) {
        bytes[i] = binaryString.charCodeAt(i);
    }
    return bytes;
};

export const deriveAESKey = async (ss: Uint8Array, salt: Uint8Array, info: Uint8Array): Promise<CryptoKey> => {
    const hkdfKey = await crypto.subtle.importKey('raw', ss as any, 'HKDF', false, ['deriveKey']);
    return await crypto.subtle.deriveKey(
        { name: 'HKDF', hash: 'SHA-256', salt: salt as any, info: info as any },
        hkdfKey,
        { name: 'AES-GCM', length: 256 },
        false,
        ['encrypt', 'decrypt']
    );
};

export const encryptWithDirectKey = async (data: Uint8Array, key: Uint8Array): Promise<Uint8Array> => {
    const aesKey = await crypto.subtle.importKey(
        'raw', key.slice(0, 32) as any, { name: 'AES-GCM' }, false, ['encrypt']
    );

    const iv = new Uint8Array(12);
    crypto.getRandomValues(iv);

    const encrypted = await crypto.subtle.encrypt({ name: 'AES-GCM', iv }, aesKey, data as any);
    const encryptedBytes = new Uint8Array(encrypted);

    const result = new Uint8Array(iv.length + encryptedBytes.length);
    result.set(iv, 0);
    result.set(encryptedBytes, iv.length);
    return result;
};

export const decryptWithDirectKey = async (encryptedData: Uint8Array, key: Uint8Array): Promise<Uint8Array> => {
    const aesKey = await crypto.subtle.importKey(
        'raw', key.slice(0, 32) as any, { name: 'AES-GCM' }, false, ['decrypt']
    );

    const iv = encryptedData.slice(0, 12);
    const ciphertext = encryptedData.slice(12);

    const decrypted = await crypto.subtle.decrypt({ name: 'AES-GCM', iv }, aesKey, ciphertext as any);
    return new Uint8Array(decrypted);
};

export const encryptPayload = async (data: any, ss: Uint8Array, sessionID: string): Promise<string> => {
    const dataBytes = new TextEncoder().encode(JSON.stringify(data));

    const ss1 = ss;
    const ss2 = new Uint8Array(32);
    for (let i = 0; i < 32; i++) ss2[i] = ss[31 - i];

    const salt = new Uint8Array(16);
    const iv = new Uint8Array(12);
    crypto.getRandomValues(salt);
    crypto.getRandomValues(iv);

    const info = new TextEncoder().encode(`kyberlink:v1|session=${sessionID}`);
    const aesKey = await deriveAESKey(ss1, salt, info);
    const aad = new TextEncoder().encode(`kyberlink:v1|session=${sessionID}`);

    const encrypted = await crypto.subtle.encrypt(
        { name: 'AES-GCM', iv, additionalData: aad }, aesKey, dataBytes
    );
    const encryptedData = new Uint8Array(encrypted);

    // Encrypt salt+IV with ss2 (reversed shared secret)
    const saltIvCombined = new Uint8Array(salt.length + iv.length);
    saltIvCombined.set(salt, 0);
    saltIvCombined.set(iv, salt.length);
    const encryptedSaltIv = await encryptWithDirectKey(saltIvCombined, ss2);

    // Length prefix (2 bytes) + encryptedSaltIv + encryptedData
    const combined = new Uint8Array(2 + encryptedSaltIv.length + encryptedData.length);
    combined[0] = (encryptedSaltIv.length >> 8) & 0xFF;
    combined[1] = encryptedSaltIv.length & 0xFF;
    combined.set(encryptedSaltIv, 2);
    combined.set(encryptedData, 2 + encryptedSaltIv.length);

    return b64e(combined);
};

export const decryptResponse = async (
    encryptedDataB64: string,
    secretCiphertextB64: string,
    privateKey: Uint8Array,
    kemInstance: MlKem1024,
    sessionID: string
): Promise<any> => {
    const combinedSecret = b64d(secretCiphertextB64);

    // Kyber-1024 CT = 1568 bytes
    const kyberCTLen = 1568;
    if (combinedSecret.length < kyberCTLen) {
        throw new Error("Invalid combined secret length");
    }

    const kyberCiphertext = combinedSecret.slice(0, kyberCTLen);
    const encryptedSaltIv = combinedSecret.slice(kyberCTLen);

    const sharedSecret = await kemInstance.decap(kyberCiphertext, privateKey);

    const ss1 = sharedSecret;
    const ss2 = new Uint8Array(32);
    for (let i = 0; i < 32; i++) ss2[i] = sharedSecret[31 - i];

    const saltIvBytes = await decryptWithDirectKey(encryptedSaltIv, ss2);
    if (saltIvBytes.length !== 28) throw new Error("Invalid salt/IV length");

    const salt = saltIvBytes.slice(0, 16);
    const iv = saltIvBytes.slice(16, 28);

    const encryptedDataBytes = b64d(encryptedDataB64);
    const info = new TextEncoder().encode(`kyberlink:v1|session=${sessionID}`);
    const aesKey = await deriveAESKey(ss1, salt, info);
    const aad = new TextEncoder().encode(`kyberlink:v1|session=${sessionID}`);

    const decryptedBuffer = await crypto.subtle.decrypt(
        { name: 'AES-GCM', iv, additionalData: aad }, aesKey, encryptedDataBytes as any
    );

    const decryptedData = new TextDecoder().decode(decryptedBuffer);
    try {
        return JSON.parse(decryptedData);
    } catch {
        return decryptedData;
    }
};
