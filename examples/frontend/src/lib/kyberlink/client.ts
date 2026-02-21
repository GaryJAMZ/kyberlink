// Copyright 2026 José Antonio Garibay Marcelo
// SPDX-License-Identifier: Apache-2.0

import { MlKem1024 } from 'mlkem';
import type {
    KyberLinkConfig,
    KyberLinkRequest,
    GatewayRequestPayload,
    KyberLinkResponse,
    SecureResponse
} from './types';
import {
    b64d,
    b64e,
    encryptPayload,
    decryptResponse
} from './encryption';

export class KyberLinkClient {
    private config: KyberLinkConfig;

    constructor(config: KyberLinkConfig) {
        this.config = config;
    }

    async init(): Promise<{ id: string, publicKey: Uint8Array }> {
        console.log("[PHASE 1] Requesting server public key...");
        const response = await fetch(`${this.config.gatewayUrl}/kempublic`);
        if (!response.ok) {
            throw new Error(`Init failed: ${response.statusText}`);
        }
        const data = await response.json();
        const session = {
            id: data.sessionID,
            publicKey: b64d(data.publicKey)
        };
        console.log(`[PHASE 1] Session: ${session.id.substring(0, 8)}...`);
        return session;
    }

    async send<T = any>(
        finalApi: string,
        method: string = 'POST',
        payload: any = {}
    ): Promise<SecureResponse<T>> {
        const session = await this.init();

        const kem = new MlKem1024();
        const [ct, ss] = await kem.encap(session.publicKey as any);
        const [clientPubKey, clientPrivKey] = await kem.generateKeyPair();

        const requestPayload: GatewayRequestPayload = {
            finalApi,
            method,
            payload,
            timestamp: Math.floor(Date.now() / 1000),
            nonce: crypto.randomUUID()
        };

        const encryptedData = await encryptPayload(requestPayload, ss, session.id);

        const request: KyberLinkRequest = {
            v: 1,
            sessionID: session.id,
            clientPublicKey: b64e(clientPubKey),
            secretCiphertext: b64e(ct),
            encryptedData
        };

        console.log("[PHASE 2] Sending encrypted request...");

        const response = await fetch(`${this.config.gatewayUrl}/gateway`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(request)
        });

        if (!response.ok) {
            const errText = await response.text();
            throw new Error(`Gateway Error ${response.status}: ${errText}`);
        }

        const responseJson: KyberLinkResponse = await response.json();
        console.log("[PHASE 4] Encrypted response received");

        const decryptedData = await decryptResponse(
            responseJson.encryptedData,
            responseJson.secretCiphertext,
            clientPrivKey,
            kem,
            session.id
        );

        console.log("[PHASE 4] Response decrypted");

        return {
            data: decryptedData,
            originalResponse: responseJson
        };
    }
}
