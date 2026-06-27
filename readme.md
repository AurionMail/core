# Aurion Core Server

This repository contains the backend server for **Aurion**, written in Go.

It provides the API, authentication, key management, and integrations with the mail backend.



## **1. Requirements**

* Go **1.22+** (for development/building)
* PostgreSQL **15+**
* Git
* Apache2 & Certbot (for production routing and SSL)
* (Optional) Air for hot‑reload (dev only)



## **2. Database Setup**

Run the following commands in your PostgreSQL instance to initialize the database and user:

```bash
sudo -u postgres psql

```

```sql
CREATE USER aurionuser WITH PASSWORD 'coolpassword';
CREATE DATABASE auriondb OWNER aurionuser;
\c auriondb
GRANT ALL ON SCHEMA public TO aurionuser;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO aurionuser;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO aurionuser;

```

> **Note:** Remember to apply the database migrations. If you have `goose` installed:
> ```bash
> goose -dir internal/db/migrations postgres "host=localhost user=aurionuser password=coolpassword dbname=auriondb sslmode=disable" up
> 
> ```
> 
> 



## **3. Environment Variables (`.env`)**

Clone the repository and create your configuration file:

```bash
git clone https://github.com/aurion/core.git
cd core
cp .env.example .env

```

Edit the `.env` file with your specific configuration:

```ini
APP_ENV=production # change to 'dev' for local development
APP_PORT=8080

DB_HOST=localhost
DB_PORT=5432
DB_USER=aurionuser
DB_PASS=coolpassword
DB_NAME=auriondb

MAIL_BACKEND=jmap  
JMAP_URL=hello 
# IMAP_URL=hello 
STALWART_API_KEY=hello

# Generate a secure key using: openssl rand -base64 32
AUTH_FAKE_SALT_SECRET="e4jQQdVPo4+71ceUgE2K+6XHjbNJtj3pCP94BXjMIiY="

```



## **4. Development Installation & Run**

Install dependencies:

```bash
go mod tidy

```

### Normal mode

```bash
go run ./cmd/app-server

```

### Hot‑reload mode (requires Air)

Install Air first: `go install github.com/air-verse/air@latest`

```bash
air

```

The server will be available at: `http://localhost:8080`



## **5. Production Installation (Standard Usage)**

For standard production usage, it is recommended to compile the Go binary rather than using `go run`.

### Step 1: Build the Binary

Compile the application into a single executable:

```bash
go mod tidy
go build -ldflags="-w -s" -o aurion-core ./cmd/app-server

```

*(The `-ldflags="-w -s"` flags strip debugging information to reduce the binary size).*

### Step 2: Deployment

To run the server, you only need the compiled `aurion-core` binary and the `.env` file in the same directory.

```bash
./aurion-core

```

### Step 3: Production Service (Systemd)

To ensure the server runs continuously and restarts automatically on crash or reboot, create a systemd service.

Create the file `/etc/systemd/system/aurion.service`:

```ini
[Unit]
Description=Aurion Core Server
After=network.target postgresql.service

[Service]
Type=simple
User=www-data
WorkingDirectory=/var/www/aurion
ExecStart=/var/www/aurion/aurion-core
Restart=always
RestartSec=5
EnvironmentFile=/var/www/aurion/.env

[Install]
WantedBy=multi-user.target

```

Enable and start the service:

```bash
sudo systemctl daemon-reload
sudo systemctl enable aurion
sudo systemctl start aurion

```



## **6. Production Routing: Apache Reverse Proxy & Certbot**

In production, you should expose the server via a reverse proxy like Apache to handle SSL/TLS termination, custom domains, and headers.

### Step 1: Install Apache and Enable Modules

Install Apache2 and ensure the mandatory proxy modules are enabled:

```bash
sudo apt update
sudo apt install apache2 -y

# Enable proxy modules
sudo a2enmod proxy
sudo a2enmod proxy_http
sudo a2enmod headers
sudo a2enmod rewrite

```

### Step 2: Configure the Apache VirtualHost

Create a new configuration file for your domain (e.g., `api.aurion.example.com`):

```bash
sudo nano /etc/apache2/sites-available/aurion.conf

```

Paste the following configuration (replace `api.aurion.example.com` with your real domain):

```apache
<VirtualHost *:80>
    ServerName api.aurion.example.com

    # Forward requests to the Go application running on port 8080
    ProxyPreserveHost On
    ProxyPass / http://localhost:8080/
    ProxyPassReverse / http://localhost:8080/

    # Security Headers
    Header always set X-Content-Type-Options "nosniff"
    Header always set X-Frame-Options "DENY"
    Header always set X-XSS-Protection "1; mode=block"

    ErrorLog ${APACHE_LOG_DIR}/aurion-error.log
    CustomLog ${APACHE_LOG_DIR}/aurion-access.log combined
</VirtualHost>

```

Enable the new site and restart Apache:

```bash
sudo a2ensite aurion.conf
sudo systemctl restart apache2

```

### Step 3: Install Certbot and Configure SSL

Install Certbot along with the Apache plugin to automatically fetch and configure your Let's Encrypt certificate:

```bash
sudo apt install certbot python3-certbot-apache -y

```

Run Certbot to generate the SSL certificate. Certbot will automatically read your Apache configuration, look for the `ServerName`, and prompt you to handle HTTPS redirection:

```bash
sudo certbot --apache -d api.aurion.example.com

```