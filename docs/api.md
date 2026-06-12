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

## 2.1 Sign Up

### POST `/auth/signup`

Creates a new user account.

#### Request
```json
{
  "email": "user@example.com",
  "password": "plaintext-password"
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

## 2.2 Login

### POST `/auth/login`

Authenticates a user and returns a session token.

#### Request
```json
{
  "email": "user@example.com",
  "password": "plaintext-password"
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

## 2.3 Session Validation

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
- `404` — identity not found or no active key
# 4. WKD (Web Key Directory)

These endpoints implement the OpenPGP WKD standard.
## 4.1 WKD Policy

### GET `/.well-known/openpgpkey/policy`

Returns the WKD policy file.

## 4.2 WKD Key Lookup

### GET `/.well-known/openpgpkey/hu/:hash`

Returns the public key associated with the hashed local‑part.

# 5. Key Management API (Authenticated)

Requires `Authorization: Bearer <token>`.

## 5.1 Upload Public Key

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
## 5.2 Upload Private Key

### POST `/keys/private`

Stores the user’s encrypted private key for an identity.

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
## 5.3 Get Private Key for Current User

### GET `/keys/private/me`

Returns the encrypted private key for the authenticated user’s primary identity.

#### Response
```json
{
  "identity_email": "user@example.com",
  "encrypted_private_key": "base64-blob"
}
```

# 6. Mail API (Authenticated)

All mail operations use the internal mail service (JMAP backend).
## 6.1 Send Mail

### POST `/mail/send`

Sends an encrypted or plaintext message.

#### Request
```json
{
  "to": "recipient@example.com",
  "subject": "Hello",
  "ciphertext_for_sender": "base64-blob",
  "ciphertext_for_receiver": "base64-blob",
  "attachments": [
    {
      "Filename": "file.txt",
      "MimeType": "text/plain",
      "Data": "base64-blob"
    }
  ]
}
```

#### Response
```json
{
  "status": "sent"
}
```
## 6.2 List Messages

### GET `/mail/messages`

Query parameters:

- `folder` (optional)
- `limit`
- `offset`

#### Response
```json
[
  {
    "id": "UUID",
    "from": "alice@example.com",
    "subject": "Hello",
    "seen": false,
    "tags": []
  }
]
```
## 6.3 Get Message

### GET `/mail/message/:id`

Returns full message content.

## 6.4 Delete Message

### DELETE `/mail/message/:id`

Deletes a message.
## 6.5 Mark as Seen

### POST `/mail/message/:id/seen`

#### Request
```json
{
  "seen": true
}
```
## 6.6 Update Tags

### POST `/mail/message/:id/tags`

#### Request
```json
{
  "tags": ["important"]
}
```
# 7. Mailbox API (Authenticated)
## 7.1 List Mailboxes
### GET `/mail/mailboxes`
## 7.2 Create Mailbox
### POST `/mail/mailbox/create`
#### Request
```json
{
  "name": "Projects"
}
```
## 7.3 Rename Mailbox

### POST `/mail/mailbox/rename`

#### Request
```json
{
  "id": "UUID",
  "name": "NewName"
}
```
## 7.4 Delete Mailbox

### POST `/mail/mailbox/delete`

#### Request
```json
{
  "id": "UUID"
}
```
# 8. Drafts API (Authenticated)

## 8.1 Create Draft

### POST `/mail/draft`

#### Request
```json
{
  "to": ["recipient@example.com"],
  "subject": "Draft",
  "payload": "base64-blob"
}
```
## 8.2 Update Draft

### PUT `/mail/draft/:id`

## 8.3 Delete Draft

### DELETE `/mail/draft/:id`