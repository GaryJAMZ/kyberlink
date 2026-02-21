// Copyright 2026 José Antonio Garibay Marcelo
// SPDX-License-Identifier: Apache-2.0

export interface KyberLinkConfig {
    gatewayUrl: string;
    timeout?: number;
}

export interface SecureResponse<T = any> {
    data: T;
    originalResponse?: any;
}

export interface GatewayRequestPayload {
    finalApi: string;
    method: string;
    payload: any;
    timestamp: number;
    nonce: string;
}

export interface KyberLinkRequest {
    v: number;
    sessionID: string;
    clientPublicKey: string;
    secretCiphertext: string;
    encryptedData: string;
}

export interface KyberLinkResponse {
    v: number;
    secretCiphertext: string;
    encryptedData: string;
}
