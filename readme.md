# **Aurion Core Server **

This repository contains the backend server for **Aurion**, written in Go.  
It provides the API, authentication, key management, and integrations with the mail backend.
## **1. Requirements**

- Go **1.22+**
- PostgreSQL **15+**
- Git
- (Optional) Air for hot‑reload

Install Air:

```
go install github.com/air-verse/air@latest
```

## **2. Clone and init env**

```
git clone https://github.com/aurion/core.git
cd core
```
Create a `.env` file:

```
APP_PORT=8080
```

```
go mod tidy
```
## **3. Run the server**

### Normal mode

```
go run ./cmd/app-server
```

### Hot‑reload mode

```
air
```

The server runs on:

```
http://localhost:8080
```