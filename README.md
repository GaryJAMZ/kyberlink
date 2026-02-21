# KyberLink

[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

> A Post-Quantum Security Gateway for Hybrid Key Exchange

> **Personal note:** KyberLink is my personal experiment in implementing a hybrid post-quantum protocol in production. It works, I use it, and I publish it so the community can review, critique, and improve it.

**Version**: 1.0.0 | **License**: Apache 2.0 | **Status**: Experimental | **Go 1.23+** | **Node.js 18+**

---

## ⚠️ Security Notice

**Experimental Software** — KyberLink is in active production but has **NOT** undergone an independent security audit. An independent security audit is strongly recommended before using it in environments that handle critical or sensitive data. Report vulnerabilities via [SECURITY.md](SECURITY.md).

---

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Quick Start](#quick-start)
- [Protocol Specification](#protocol-specification)
- [API Reference](#api-reference)
- [Configuration](#configuration)
- [Security Features](#security-features)
- [Project Structure](#project-structure)
- [Technology Stack](#technology-stack)
- [Contributing](#contributing)
- [License](#license)

---

## Overview

KyberLink is a post-quantum security gateway that implements a hybrid key exchange protocol combining **ML-KEM-1024** (Kyber1024) and **AES-256-GCM** for authenticated encryption. It provides a transparent encryption layer between clients and backend services without requiring cryptographic changes to your business logic.

### Key Capabilities

- **Post-Quantum Ready**: Uses ML-KEM-1024, standardized by NIST as a key encapsulation mechanism resistant to quantum computing threats
- **Hybrid Encryption**: Combines Kyber key encapsulation with AES-256-GCM for robust authenticated encryption
- **Perfect Forward Secrecy**: Fresh keypairs generated per request, ensuring compromised session keys don't expose historical data
- **Replay Protection**: Timestamp validation with configurable windows and nonce-based attack prevention
- **Transparent Integration**: Gateway sits between clients and backends, requiring no cryptographic knowledge from business logic developers
- **Single-Use Sessions**: Keys are burned immediately after consumption, eliminating key reuse vulnerabilities

---

## Architecture

### System Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                          CLIENT APPLICATION                      │
│                       (TypeScript/JavaScript)                     │
│                                                                   │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │         KyberLink Client (mlkem npm + Web Crypto)       │   │
│  │  • ML-KEM-1024 key encapsulation                        │   │
│  │  • AES-256-GCM encryption/decryption                    │   │
│  │  • HKDF-SHA256 key derivation                           │   │
│  │  • Replay protection (nonces + timestamps)             │   │
│  └─────────────────────────────────────────────────────────┘   │
└──────────────────────────┬──────────────────────────────────────┘
                           │
                  HTTPS/TLS (Encrypted)
                           │
        ┌──────────────────┴──────────────────┐
        │                                     │
┌───────▼──────────────────────────────────────────────────────────┐
│              KYBERLINK GATEWAY SERVER (Go/Gin)                   │
│                                                                   │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │         Cryptographic Layer (circl/ML-KEM)             │   │
│  │  • Session Key Management                              │   │
│  │  • Public Key Decapsulation (CT1 → SS1)               │   │
│  │  • Shared Secret Derivation & Key Rotation             │   │
│  └─────────────────────────────────────────────────────────┘   │
│                                                                   │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │         HTTP Request Handler & Router                  │   │
│  │  • GET /kempublic   - Handshake & public key issue    │   │
│  │  • POST /gateway    - Main encrypted request endpoint │   │
│  │  • Rate limiting & request validation                  │   │
│  └─────────────────────────────────────────────────────────┘   │
│                                                                   │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │         Session & Security Management                  │   │
│  │  • 256-bit random session IDs                          │   │
│  │  • Timestamp validation (60s window)                   │   │
│  │  • Nonce tracking & replay detection                   │   │
│  │  • Single-use key enforcement                          │   │
│  └─────────────────────────────────────────────────────────┘   │
└──────────────────────────┬──────────────────────────────────────┘
                           │
                      HTTP (Clear)
                           │
┌───────────────────────────┴──────────────────────────────────────┐
│              BACKEND SERVICE (Business Logic)                    │
│                  (Any language/framework)                         │
│                                                                   │
│  Receives clear-text JSON requests:                              │
│  {                                                               │
│    "finalApi": "/api/users/123",                                │
│    "method": "GET",                                             │
│    "payload": {...},                                            │
│    "timestamp": 1708425000,                                     │
│    "nonce": "abc123..."                                         │
│  }                                                               │
│                                                                   │
│  Returns standard HTTP responses (no encryption overhead)        │
└───────────────────────────────────────────────────────────────────┘
```

### Request/Response Flow

```
PHASE 1: HANDSHAKE
  Client → GET /kempublic
         ← { sessionID (256-bit random hex), pub1 (ML-KEM public key) }

PHASE 2: CLIENT REQUEST
  Client → POST /gateway {
             sessionID (256-bit random hex),
             CT1 (Kyber ciphertext),
             encrypted_payload,
             pub2 (client's ephemeral public key)
           }

PHASE 3: GATEWAY PROCESSING
  Gateway → [decrypt payload with SS1, validate timestamp/nonce]
          → [forward clear request to backend]
         ← [receive backend response]
          → [encapsulate pub2 → CT2, encrypt response with SS2]

PHASE 4: CLIENT RECEPTION
  Client → [decapsulate CT2 → SS2, decrypt response]
         ← [application receives clear response]

Key Lifecycle:
  • Each request uses fresh ephemeral keypairs
  • Shared secrets derived via HKDF-SHA256
  • Sessions burned after single use
  • Keys never reused (perfect forward secrecy)
```

---

## Quick Start

### Prerequisites

- **Go**: 1.23 or higher
- **Node.js**: 18.0.0 or higher
- **Platform**: Linux, macOS, or Windows

### Automated Demo (Recommended)

The fastest way to see KyberLink in action:

**Linux/macOS:**
```bash
./start-demo.sh
```

**Windows (PowerShell):**
```powershell
.\start-demo.ps1
```

This will automatically:
1. Start the mock backend on port 34890
2. Start the gateway on port 45782
3. Start the frontend dev server on port 5173
4. Open http://localhost:5173 in your browser

### Manual Setup

#### Step 1: Configure the Gateway

```bash
cd gateway
go run main.go
```

Optionally, copy `.env.example` to `.env` to customize configuration:
```bash
cp .env.example .env
```

Available environment variables in `.env`:
```env
PORT=45782
BACKEND_URL=http://localhost:34890
CORS_ORIGIN=http://localhost:5173
```

All environment variables are optional and have sensible defaults.

#### Step 2: Start the Mock Backend

```bash
cd examples/mock-backend
go run main.go
# Backend running on http://localhost:34890
```

#### Step 3: Start the Frontend

```bash
cd examples/frontend
npm install
npm run dev
# Frontend available at http://localhost:5173
```

#### Step 4: Open the Demo

Navigate to [http://localhost:5173](http://localhost:5173) in your browser and test the encrypted protocol.

---

## Protocol Specification

KyberLink implements a four-phase key exchange protocol combining ML-KEM-1024 for key establishment with AES-256-GCM for authenticated encryption.

### Phase 1: Handshake

**Endpoint**: `GET /kempublic`

**Flow**:
1. Client initiates handshake by requesting gateway's public key
2. Gateway generates fresh ML-KEM-1024 keypair: `(pub1, priv1)`
3. Gateway generates a 256-bit cryptographically random session ID (32 bytes from crypto/rand)
4. Gateway stores `priv1` in session map, keyed to the session ID
5. Gateway returns session ID and `pub1` to client

**Response**:
```json
{
  "sessionID": "256-bit-hex-string-64-characters",
  "publicKey": "base64-encoded-ml-kem-1024-public-key"
}
```

**Security Properties**:
- Fresh keypair per handshake
- Session ID is 256-bit random (unguessable, 2^256 possible values)
- No sensitive material leaked

### Phase 2: Client Request

**Endpoint**: `POST /gateway`

**Flow**:
1. Client encapsulates gateway's public key (`pub1`) → ciphertext (`CT1`) + shared secret (`SS1`)
2. Client generates its own ephemeral keypair: `(pub2, priv2)` for receiving the response
3. Client constructs request payload:
   ```json
   {
     "finalApi": "/api/resource",
     "method": "GET|POST|PUT|DELETE",
     "payload": {...},
     "nonce": "unique-per-request-identifier",
     "timestamp": 1708425000
   }
   ```
4. Client derives encryption keys from `SS1` using HKDF-SHA256:
   - `data_key` = HKDF(SS1, "data_encryption", ...)
   - `metadata_key` = HKDF(SS1, "metadata_encryption", ...)
5. Client encrypts request payload with AES-256-GCM using `data_key`
6. Client encrypts salt and IV with AES-256-GCM using `metadata_key`
7. Client sends to gateway:
   ```json
   {
     "sessionID": "256-bit-hex-string-64-characters",
     "ciphertext1": "base64-ct1",
     "encryptedData": "base64-encrypted-payload",
     "encryptedMetadata": "base64-encrypted-salt-and-iv",
     "clientPublicKey": "base64-pub2",
     "nonce": "unique-per-request",
     "timestamp": 1708425000
   }
   ```

**Security Properties**:
- Perfect forward secrecy: fresh `SS1` per request
- Ephemeral client key: `pub2` only used for this response
- Salt & IV encrypted: prevents statistical analysis
- Timestamp + nonce: prevents replay attacks

### Phase 3: Gateway Processing

**Flow**:
1. Gateway receives encrypted request
2. Gateway retrieves `priv1` from session store using the session ID
3. Gateway decapsulates `CT1` with `priv1` → `SS1`
4. Gateway derives same keys using HKDF-SHA256
5. Gateway decrypts payload and metadata
6. Gateway validates timestamp (within 60-second window)
7. Gateway checks nonce against replay cache
8. Gateway **burns** session key: removes `priv1` from store (one-time use, session expires in 5 minutes)
9. Gateway validates request path (only relative paths allowed, SSRF prevention)
10. Gateway forwards decrypted request to backend:
    ```json
    {
      "finalApi": "/api/resource",
      "method": "GET",
      "payload": {...},
      "timestamp": 1708425000,
      "nonce": "..."
    }
    ```
11. Gateway receives backend response
12. Gateway encapsulates client's public key (`pub2`) → ciphertext (`CT2`) + shared secret (`SS2`)
13. Gateway derives keys from `SS2` using same HKDF-SHA256
14. Gateway encrypts response with AES-256-GCM using `SS2` keys
15. Gateway returns to client:
    ```json
    {
      "ciphertext2": "base64-ct2",
      "encryptedResponse": "base64-encrypted-response"
    }
    ```

**Security Properties**:
- Symmetric decryption proves client authenticity
- Single-use sessions prevent key reuse
- Replay detection via timestamp + nonce
- SSRF prevention via path validation

### Phase 4: Client Reception

**Flow**:
1. Client receives gateway response containing `CT2` and encrypted response
2. Client decapsulates `CT2` with its private key (`priv2`) → `SS2`
3. Client derives same keys using HKDF-SHA256
4. Client decrypts and validates response
5. Client returns decrypted response to application

**Security Properties**:
- Only legitimate client (with `priv2`) can decrypt response
- No sensitive data exposed during transport

### Key Derivation (HKDF-SHA256)

All symmetric keys are derived from shared secrets using HKDF-SHA256 with:
- **Hash**: SHA-256
- **Info**: Context-specific string (e.g., "request_encryption", "response_encryption")
- **Salt**: Session-bound random value
- **Output**: 32 bytes for AES-256-GCM

This ensures key independence and prevents cross-protocol attacks.

### Replay Protection

KyberLink prevents replay attacks through:
- **Timestamp Validation**: Requests must have `timestamp` within ±60 seconds of server time
- **Nonce Tracking**: Each nonce is recorded in an in-memory cache with TTL
- **Single-Use Keys**: Session keys are burned after consumption, making replays with same key impossible

---

## API Reference

### GET /kempublic

Initiates a handshake and returns a fresh ML-KEM-1024 public key with a 256-bit random session ID.

**Response** (200 OK):
```json
{
  "sessionID": "string (256-bit random hex, 64 characters)",
  "publicKey": "string (base64-encoded ML-KEM-1024 public key)"
}
```

**Error** (500):
```json
{
  "error": "string (error message)"
}
```

**Rate Limiting**: Handshake endpoint is rate-limited to 100 requests per minute per IP.

---

### POST /gateway

Accepts an encrypted KyberLink request, decrypts it, forwards to backend, encrypts the response.

**Request Body**:
```json
{
  "sessionID": "string (from handshake)",
  "secretCiphertext": "string (base64, CT1 from client encapsulation)",
  "encryptedData": "string (base64, AES-256-GCM encrypted payload)",
  "clientPublicKey": "string (base64, client's ephemeral public key pub2)"
}
```

**Response** (200 OK):
```json
{
  "v": 1,
  "secretCiphertext": "string (base64, CT2 + encrypted salt/IV)",
  "encryptedData": "string (base64, AES-256-GCM encrypted response)"
}
```

**Errors**:
- `400 Bad Request`: Malformed request, missing fields
- `401 Unauthorized`: Invalid session, authentication failure, replay detected
- `408 Request Timeout`: Timestamp outside acceptable window
- `429 Too Many Requests`: Rate limit exceeded
- `500 Internal Server Error`: Backend unavailable or processing error
- `502 Bad Gateway`: Backend unreachable

**Request Validation**:
- Payload size: Limited to 1 MB to prevent memory exhaustion
- Path validation: Only relative paths forwarded to backend (SSRF prevention)
- Timestamp validation: ±60 second window
- Nonce uniqueness: Per-session nonce cache

---

## Configuration

### Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `PORT` | No | `45782` | Port for gateway HTTP server |
| `BACKEND_URL` | No | `http://localhost:34890` | URL of backend service |
| `CORS_ORIGIN` | No | `*` | Allowed CORS origin for frontend requests |

All environment variables are optional and have sensible defaults. Copy `gateway/.env.example` to `gateway/.env` to customize.

### Example .env File

```env
PORT=45782
BACKEND_URL=http://localhost:34890
CORS_ORIGIN=http://localhost:5173
```

---

## Security Features

### 1. Post-Quantum Cryptography

- **ML-KEM-1024**: NIST-standardized key encapsulation mechanism resistant to quantum computing attacks
- **AES-256-GCM**: NIST-standardized authenticated encryption cipher with 256-bit keys

### 2. Perfect Forward Secrecy

Every request uses freshly generated ephemeral keypairs. Compromise of a session key reveals only that single request/response pair, not historical or future traffic.

### 3. Authenticated Encryption

AES-256-GCM provides:
- **Confidentiality**: XOR-based stream cipher
- **Authenticity**: Galois/Counter Mode authentication (128-bit tags)
- **Integrity**: Tamper detection and prevention

### 4. Key Derivation

HKDF-SHA256 ensures:
- **Key Separation**: Different keys for different purposes (payload, metadata, response)
- **Key Stretching**: Shared secrets expanded to multiple derived keys
- **Domain Separation**: Info strings prevent cross-protocol key usage

### 5. Replay Protection

- **Timestamps**: Requests outside ±60 second window rejected
- **Nonces**: Per-request unique identifiers tracked in cache
- **One-Time Keys**: Session keys burned after single use
- **No Replaying with New Keys**: Each request requires handshake

### 6. Session ID Security

Session IDs are 256-bit cryptographically random values (32 bytes from crypto/rand) that are:
- **Unguessable**: 2^256 possible values make brute force infeasible
- **Single-use**: Burned immediately after the first request is processed
- **Time-limited**: Expire after 5 minutes
- **No token validation needed**: Random values inherently prevent enumeration and forgery

### 7. SSRF Prevention

- Only relative paths (starting with `/`) forwarded to backend
- Absolute URLs to other domains rejected
- Backend URL controlled via secure environment variable

### 8. Rate Limiting

Handshake endpoint enforces rate limits (100 req/min per IP) to prevent:
- Brute force attacks
- Denial of service (resource exhaustion)
- Handshake replay flooding

### 9. Request Size Limits

- 1 MB payload limit prevents memory exhaustion attacks
- 10 MB backend response read limit prevents large allocation attacks

### 10. Secure Defaults

- TLS/HTTPS recommended for all connections
- CORS restricted to configured origin
- Timestamp validation enabled by default
- Rate limiting active by default

---

## Project Structure

```
kyberlink/
│
├── gateway/                          # KyberLink Gateway Server (Go)
│   ├── main.go                       # Entry point, HTTP server setup, CORS, body limits
│   ├── go.mod & go.sum              # Go module dependencies
│   ├── .env.example                 # Environment configuration template
│   │
│   ├── crypto/                       # Cryptographic Operations
│   │   ├── kyber.go                  # ML-KEM-1024 key generation, encapsulation, decapsulation
│   │   ├── aes.go                    # AES-256-GCM encryption/decryption + HKDF key derivation
│   │   └── utils.go                  # Base64, hex encoding, random byte generation
│   │
│   └── gateway/                      # HTTP Handlers & Session Management
│       ├── handler.go                # /kempublic and /gateway endpoint handlers
│       ├── security.go               # Replay protection, nonce tracking, rate limiting
│       ├── session.go                # Session store, key lifecycle, session ID generation
│       └── types.go                  # Request/response struct definitions
│
├── examples/
│   │
│   ├── mock-backend/                 # Simple Go Backend (Reference Implementation)
│   │   ├── main.go                   # HTTP server with /test1, /test2, /test3 endpoints
│   │   └── go.mod                    # Go module definition
│   │
│   └── frontend/                     # TypeScript Demo Client (Vite)
│       ├── index.html                # Entry HTML page
│       ├── src/
│       │   ├── main.ts               # Demo UI: send encrypted requests, load testing
│       │   ├── style.css             # Demo styling
│       │   └── lib/kyberlink/        # KyberLink client library
│       │       ├── client.ts         # Main client class (handshake + send)
│       │       ├── encryption.ts     # Client-side AES-256-GCM + ML-KEM-1024
│       │       ├── types.ts          # TypeScript type definitions
│       │       └── index.ts          # Library exports
│       ├── package.json              # npm dependencies (mlkem)
│       └── tsconfig.json             # TypeScript configuration
│
├── start-demo.sh                     # Automated demo launcher (Linux/macOS)
├── start-demo.ps1                    # Automated demo launcher (Windows)
├── LICENSE                           # Apache 2.0 License
├── CONTRIBUTING.md                   # Contribution guidelines
├── SECURITY.md                       # Security policy & vulnerability reporting
└── README.md                         # This file
```

---

## Technology Stack

### Cryptography

| Component | Technology | Purpose |
|-----------|-----------|---------|
| **KEM** | `cloudflare/circl` (Go) | ML-KEM-1024 key encapsulation |
| **KEM** | `mlkem` npm package (TypeScript) | ML-KEM-1024 client-side |
| **AEAD** | Web Crypto API | AES-256-GCM (browser native) |
| **KDF** | SHA-256 (standard library) | HKDF key derivation |

### Backend (Gateway)

| Component | Technology | Purpose |
|-----------|-----------|---------|
| **Language** | Go 1.23+ | High-performance server language |
| **HTTP Framework** | Gin | Fast HTTP router and middleware |
| **HTTP Client** | Go standard library | Backend request forwarding |
| **Crypto** | `cloudflare/circl` | ML-KEM-1024 & HKDF implementations |

### Frontend (Demo Client)

| Component | Technology | Purpose |
|-----------|-----------|---------|
| **Language** | TypeScript | Type-safe client implementation |
| **Bundler** | Vite | Fast development and production builds |
| **Crypto** | `mlkem` npm, Web Crypto API | ML-KEM-1024 & AES-256-GCM |
| **HTTP** | Fetch API | Encrypted request transport |

### Backend (Mock Reference)

| Component | Technology | Purpose |
|-----------|-----------|---------|
| **Language** | Go | Reference backend implementation |
| **HTTP** | Go standard library (`net/http`) | REST API endpoints |

---

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for:
- Code of conduct
- Development setup
- Testing requirements
- Pull request process
- Style guidelines

---

## Security & Vulnerability Reporting

For security vulnerabilities, please **do not** open a public issue. Instead, see [SECURITY.md](SECURITY.md) for responsible disclosure procedures.

---

## License

Copyright 2026 José Antonio Garibay Marcelo. Licensed under the [Apache License 2.0](LICENSE).

**Author**: José Antonio Garibay Marcelo — [GitHub](https://github.com/GaryJAMZ) · [antoniogaribay6@gmail.com](mailto:antoniogaribay6@gmail.com)

---

---

# KyberLink (Español)

> Una Puerta de Enlace de Seguridad Post-Cuántica para Intercambio de Claves Híbrido

> **Nota personal:** KyberLink es mi experimento personal de implementación de un protocolo post-cuántico híbrido en producción. Está funcionando, lo uso, y lo publico para que la comunidad lo revise, critique y mejore.

**Versión**: 1.0.0 | **Licencia**: Apache 2.0 | **Estado**: Experimental | **Go 1.23+** | **Node.js 18+**

---

## ⚠️ Aviso de Seguridad

**Software Experimental** — KyberLink está en producción activa pero **NO** ha sido sometido a una auditoría de seguridad independiente. Se recomienda encarecidamente una auditoría de seguridad independiente antes de utilizarlo en entornos que manejen datos críticos o sensibles. Reporta vulnerabilidades a través de [SECURITY.md](SECURITY.md).

---

## Descripción General

KyberLink es una puerta de enlace de seguridad post-cuántica que implementa un protocolo de intercambio de claves híbrido que combina **ML-KEM-1024** (Kyber1024) y **AES-256-GCM** para cifrado autenticado. Proporciona una capa de cifrado transparente entre clientes y servicios de backend sin requerir cambios criptográficos en la lógica de negocio.

### Capacidades Clave

- **Lista para Post-Cuántica**: Utiliza ML-KEM-1024, estandarizado por NIST como mecanismo de encapsulación de claves resistente a amenazas de computación cuántica
- **Cifrado Híbrido**: Combina la encapsulación de claves de Kyber con AES-256-GCM para cifrado autenticado robusto
- **Confidencialidad Perfecta hacia Adelante (PFS)**: Genera pares de claves efímeras nuevos por solicitud, asegurando que las claves de sesión comprometidas no expongan datos históricos
- **Protección Contra Reproducción**: Validación de marcas de tiempo con ventanas configurables y prevención de ataques basados en nonces
- **Integración Transparente**: La puerta de enlace se sitúa entre clientes y backends, sin requerir conocimiento criptográfico de los desarrolladores de lógica de negocio
- **Sesiones de Uso Único**: Las claves se queman inmediatamente después de su consumo, eliminando vulnerabilidades de reutilización de claves

---

## Inicio Rápido

### Requisitos Previos

- **Go**: Versión 1.23 o superior
- **Node.js**: Versión 18.0.0 o superior
- **Plataforma**: Linux, macOS o Windows

### Demo Automatizada (Recomendada)

```bash
# Linux/macOS
./start-demo.sh

# Windows (PowerShell)
.\start-demo.ps1
```

Esto inicia automáticamente el backend simulado (:34890), la puerta de enlace (:45782) y el frontend (:5173).

### Configuración Manual

```bash
# 1. Puerta de enlace
cd gateway && go run main.go

# 2. Backend simulado
cd examples/mock-backend && go run main.go

# 3. Frontend
cd examples/frontend && npm install && npm run dev
```

Navega a [http://localhost:5173](http://localhost:5173) para probar el protocolo.

Para la documentación completa del protocolo, la referencia de API, las características de seguridad, y la estructura del proyecto, consulta la sección en inglés de este README.

---

## Contribuir

¡Bienvenidas las contribuciones! Consulta [CONTRIBUTING.md](CONTRIBUTING.md) para directrices de contribución.

---

## Licencia

Copyright 2026 José Antonio Garibay Marcelo. Licenciado bajo la [Licencia Apache 2.0](LICENSE).

**Autor**: José Antonio Garibay Marcelo — [GitHub](https://github.com/GaryJAMZ) · [antoniogaribay6@gmail.com](mailto:antoniogaribay6@gmail.com)
