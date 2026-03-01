# Nginx Guide — Why and How to Use It With This API

This guide explains **why nginx is used**, **when you need it**, and **how to configure it** for this project.

---

## 1. Why Is Nginx Needed?

The Go API runs on HTTP (port 8080) and does **not** handle HTTPS, SSL certificates, or load balancing itself. In production, you typically put a **reverse proxy** in front of it.

| Need | Why nginx helps |
|------|-----------------|
| **HTTPS** | Browsers and APIs require TLS in production. Nginx terminates SSL (handles certificates) and forwards plain HTTP to your API. |
| **Single entry point** | One domain (e.g. `api.example.com`) can serve multiple backends. Nginx routes `/api` → your Go app, `/` → static site, etc. |
| **Rate limiting / DDoS** | Extra layer before your app. Nginx can limit connections, block bad IPs. |
| **Static files** | Serve docs, Swagger UI, or a frontend from nginx; API stays focused on JSON. |
| **Load balancing** | If you run multiple API instances, nginx can distribute traffic. |

**Summary:** Your Go API is the "app server." Nginx is the "front door" — it handles HTTPS, routing, and optional protection. The API stays simple and doesn't need to know about certificates.

---

## 2. When Do You Need Nginx?

| Scenario | Need nginx? |
|----------|-------------|
| **Local development** | No. Use `http://localhost:8080` directly. |
| **Production (VPS, EC2, etc.)** | Yes. You need HTTPS. Use nginx (or Caddy, or a cloud load balancer). |
| **PaaS (Railway, Render, Fly.io)** | Usually no. The platform provides HTTPS and a reverse proxy. |
| **Docker-only local** | Optional. Only if you want to test HTTPS locally. |

---

## 3. Use Cases for This Project

### Use case A: Production on a VPS

You deploy the API on a server (DigitalOcean, Linode, etc.). You want:
- `https://api.yourdomain.com` → your Go API
- Valid SSL certificate (Let's Encrypt)

**Setup:** Install nginx on the server, get a certificate (e.g. certbot), configure nginx to proxy to `localhost:8080`.

### Use case B: API + frontend on same domain

- `https://yourdomain.com` → React/Vue frontend (static files)
- `https://yourdomain.com/api` → your Go API

**Setup:** Nginx serves `/` from a folder and proxies `/api` to the Go app.

### Use case C: Local HTTPS testing

You want to test CORS or cookies with `https://localhost` locally.

**Setup:** Generate a self-signed cert, configure nginx to use it and proxy to `localhost:8080`.

---

## 4. Example Nginx Config (This API)

### Minimal: Proxy to Go API (HTTP only, for testing)

```nginx
server {
    listen 80;
    server_name api.example.com;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### Production: HTTPS with Let's Encrypt

```nginx
server {
    listen 80;
    server_name api.example.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl;
    server_name api.example.com;

    ssl_certificate     /etc/letsencrypt/live/api.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/api.example.com/privkey.pem;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

**Get certificates:** `certbot --nginx -d api.example.com` (installs certbot, gets cert, can auto-configure nginx).

---

## 5. How to Use in Your Project

### Step 1: Install nginx

- **macOS:** `brew install nginx`
- **Ubuntu/Debian:** `sudo apt install nginx`
- **CentOS/RHEL:** `sudo yum install nginx`

### Step 2: Create config file

```bash
sudo nano /etc/nginx/sites-available/ecom-api
# (paste one of the configs above; change server_name and paths)
```

### Step 3: Enable and reload

```bash
sudo ln -s /etc/nginx/sites-available/ecom-api /etc/nginx/sites-enabled/
sudo nginx -t    # test config
sudo systemctl reload nginx
```

### Step 4: Ensure API is running

Your Go API must be running on `127.0.0.1:8080` (or whatever port you use). Use systemd, Docker, or a process manager to keep it running.

---

## 6. CORS Note

If your frontend is on `https://app.example.com` and API on `https://api.example.com`, set in `.env`:

```
CORS_ALLOWED_ORIGINS=https://app.example.com
```

The API handles CORS. Nginx just proxies; it doesn't need to add CORS headers for this setup.

---

## 7. Alternatives to Nginx

| Tool | Notes |
|------|-------|
| **Caddy** | Simpler config, auto-HTTPS. Good alternative. |
| **Cloud load balancer** | AWS ALB, GCP LB, etc. Handle HTTPS at the edge. |
| **PaaS** | Railway, Render, Fly.io — they provide HTTPS; no nginx needed. |

---

## 8. Related Docs

- [DEPLOYMENT.md](../DEPLOYMENT.md) — Production checklist
- [.env.example](../.env.example) — `CORS_ALLOWED_ORIGINS`, `HTTP_ADDR`
