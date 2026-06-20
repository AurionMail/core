# Public API Documentation — v1

This document describes the public HTTP API of the Aurion Core service.

All endpoints are JSON‑based and follow standard HTTP semantics.

# 1. Health Check

### GET `/health`

Returns service status.

#### Response

```json
{
  "status": "ok"
}

```

# 2. Authentication API

Authentication uses session tokens returned at login or signup.

Authenticated endpoints require the header:

```
Authorization: Bearer <token>

```

## 2.1 Get Encryption Salts

### POST `/auth/salt`

Retrieves the client and server salts required to process the password before authentication or signup. Returns deterministic fake salts if the user does not exist to prevent user enumeration.

#### Request

```json
{
  "email": "user@example.com"
}

```

#### Response

```json
{
  "id": "UUID-or-null",
  "salt_server": "base64-salt",
  "salt_client": "base64-salt"
}

```

## 2.2 Sign Up

### POST `/auth/signup`

Creates a new user account after validating the external mail server password.

#### Request

```json
{
  "email": "user@example.com",
  "password": "hashed-client-password",
  "server_password": "external-mail-password",
  "salt_client": "base64-salt",
  "salt_server": "base64-salt"
}

```

#### Response

```json
{
  "user_id": "UUID",
  "email": "user@example.com",
  "token": "..."
}

```

#### Errors

* `400` — Invalid request or user creation failed
* `401` — External mail server authentication failed

## 2.3 Verify Mail Server Credentials

### POST `/auth/verify`

Pre-validates IMAP/JMAP mail server credentials before starting the registration workflow.

#### Request

```json
{
  "email": "user@example.com",
  "server_password": "external-mail-password"
}

```

#### Response

```json
{
  "status": "valid"
}

```

#### Errors

* `401` — External mail server authentication failed
* `501` — Mail backend integration not configured on server

## 2.4 Login

### POST `/auth/login`

Authenticates a user and returns a session token.

#### Request

```json
{
  "email": "user@example.com",
  "password": "hashed-client-password"
}

```

#### Response

```json
{
  "user_id": "UUID",
  "email": "user@example.com",
  "token": "..."
}

```

## 2.5 Session Validation

### GET `/auth/session`

Returns the authenticated user’s session information.

#### Response

```json
{
  "user_id": "UUID",
  "email": "user@example.com"
}

```

# 3. Public Key Discovery API

These endpoints are public and require no authentication.

## 3.1 Get Public Key for an Email

### GET `/keys/public/:email`

Returns the active OpenPGP public key for the given identity.

#### Response

```json
{
  "email": "user@example.com",
  "armored_key": "--BEGIN PGP PUBLIC KEY BLOCK-- ..."
}

```

#### Errors

* `404` — identity not found or no active key

# 4. WKD (Web Key Directory)

These endpoints implement the OpenPGP WKD standard.

## 4.1 WKD Policy

### GET `/.well-known/openpgpkey/policy`

Returns the WKD policy file.

## 4.2 WKD Key Lookup

### GET `/.well-known/openpgpkey/hu/:hash`

Returns the public key associated with the hashed local‑part.

# 5. Key Management & Sync API (Authenticated)

Requires `Authorization: Bearer <token>`.

## 5.1 Synchronize Routing and Cryptographic States

### GET `/sync/routing`

Returns all aliases and groups associated with the connected user, along with instructions for key generation or fetching.

#### Response

```json
{
  "identities": [
    {
      "identity_id": "UUID",
      "email": "team@example.com",
      "type": "shared",
      "needs_key_gen": true,
      "needs_key_fetch": false,
      "members": [
        {
          "user_id": "UUID-member-1",
          "public_key": "--BEGIN PGP PUBLIC KEY BLOCK-- ..."
        },
        {
          "user_id": "UUID-member-2",
          "public_key": "--BEGIN PGP PUBLIC KEY BLOCK-- ..."
        }
      ]
    },
    {
      "identity_id": "UUID",
      "email": "alias@example.com",
      "type": "primary",
      "needs_key_gen": false,
      "needs_key_fetch": false,
      "wkd_hash": "zbase32-hash",
      "encrypted_private_key": "base64-blob"
    }
  ]
}

```

## 5.2 Upload Synchronized Keys (E2EE Shared Envelopes)

### POST `/keys/sync`

Pushes newly generated public keys and their encrypted private counterparts distributed per member (shares) to the server.

#### Request

```json
{
  "identity_id": "UUID",
  "armored_public_key": "--BEGIN PGP PUBLIC KEY BLOCK-- ...",
  "wkd_hash": "zbase32-hash",
  "shares": [
    {
      "user_id": "UUID-member-1",
      "encrypted_private_key": "base64-blob-encrypted-with-member-1-pubkey"
    },
    {
      "user_id": "UUID-member-2",
      "encrypted_private_key": "base64-blob-encrypted-with-member-2-pubkey"
    }
  ]
}

```

#### Response

```json
{
  "status": "keys_synchronized"
}

```

## 5.3 Upload Public Key

### POST `/keys/public`

Registers a public key for an identity.

#### Request

```json
{
  "email": "user@example.com",
  "armored_key": "--BEGIN PGP PUBLIC KEY BLOCK-- ...",
  "wkd_hash": "zbase32-hash"
}

```

#### Response

```json
{
  "id": "UUID"
}

```

## 5.4 Upload Private Key

### POST `/keys/private`

Stores the user’s encrypted private key envelope for a specific identity.

#### Request

```json
{
  "identity_email": "user@example.com",
  "encrypted_private_key": "base64-blob"
}

```

#### Response

```json
{
  "id": "UUID"
}

```

## 5.5 Get All Private Keys for Current User

### GET `/keys/private/me`

Returns the list of all available encrypted private keys across all identities (primary, aliases, and groups) to which the authenticated user belongs.

#### Response

```json
{
  "keys": [
    {
      "identity_email": "user@example.com",
      "encrypted_private_key": "base64-blob"
    },
    {
      "identity_email": "team@example.com",
      "encrypted_private_key": "base64-blob"
    }
  ]
}

```

#### Errors

* `404` — No identity found or no private keys available for this user