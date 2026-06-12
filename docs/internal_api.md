# Internal Routing API (v1)
Purpose: 
This API is used exclusively by the internal SMTP proxy to resolve any incoming or outgoing email address into:

- the **final identity** that owns the address  
- the **public key** used for encryption  
- the **list of recipient user emails** associated with that identity  

This enables the proxy to correctly route, encrypt, and deliver messages inside the system.

## Endpoint

### POST `/internal/routing/resolve`

Resolves an email address into its internal identity and recipients.

```json
{
  "email": "team@example.com"
}
```

### Rules:
- The email must be lowercase and valid.
- The address may be:
  - a **primary identity**
  - a **shared identity**
  - an **alias identity** (implemented as a shared identity with one member)
  - a **catch‑all match**

## Resolution Logic

Resolution follows this deterministic order:

### 1. Direct identity match
If `email` exists in `identities.email`, return that identity.

### 2. Catch‑all match
If no identity matches, check whether the domain part matches an entry in `routing_catchall.domain`.

If yes → return the identity associated with that catch‑all.

### 3. Otherwise
Return `404 unknown address`.

## Successful Response

```json
{
  "identity_email": "team@example.com",
  "public_key": "--BEGIN PGP PUBLIC KEY BLOCK-- ...",
  "recipients": [
    "alice@example.com",
    "bob@example.com",
    "charlie@example.com"
  ]
}
```

### **Fields**

| Field | Description |
|-|-|
| **identity_email** | The resolved identity address (primary/shared/alias). |
| **public_key** | The active OpenPGP public key for this identity. |
| **recipients** | List of user email addresses that belong to this identity. |

### Notes:
- `recipients` always contains **at least one email**.  
- For a simple alias, it contains exactly **one** email.  
- For shared identities, it contains **multiple** emails.  
- The proxy uses `public_key` to encrypt outgoing messages when needed.

## Error Responses

### Invalid request
```json
{
  "error": "invalid request"
}
```

### Invalid email format
```json
{
  "error": "invalid email"
}
```

### Unknown address
```json
{
  "error": "unknown address"
}
```

### Identity has no active public key
```json
{
  "error": "identity has no active public key"
}
```

## Security
- This API is **internal only**.  
- It must be accessible exclusively by the SMTP proxy.

## Example Workflows

### Simple alias
`support@example.com → paul@example.com`

Response:

```json
{
  "identity_email": "support@example.com",
  "public_key": "...",
  "recipients": ["paul@example.com"]
}
```

### Shared identity
`team@example.com → {alice, bob, charlie}`

Response:

```json
{
  "identity_email": "team@example.com",
  "public_key": "...",
  "recipients": [
    "alice@example.com",
    "bob@example.com",
    "charlie@example.com"
  ]
}
```

### Catch‑all
`anything@company.com → admin@company.com`

Response:

```json
{
  "identity_email": "admin@company.com",
  "public_key": "...",
  "recipients": ["admin@company.com"]
}
```