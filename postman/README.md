# Postman Collection for Ecom API

Pre-built collection. No setup needed — import and run.

---

## Step 1: Import into Postman

### Import Collection
1. Open [Postman](https://www.postman.com/downloads/)
2. Click **Import** (top left)
3. Drag `Ecom-API.postman_collection.json` or click **Upload Files** and select it
4. Click **Import**

### Import Environment
1. Click **Import** again
2. Select `Local.postman_environment.json`
3. Click **Import**
4. Select **Local** from the environment dropdown (top right)

---

## Step 2: Start the API

From project root:
```bash
./scripts/setup-and-run.sh
```

---

## Step 3: Run the Requests (in order)

| # | Request       | What it does                          |
|---|---------------|----------------------------------------|
| 1 | Health Check  | Verifies API is running               |
| 2 | Register      | Creates user (skip if already done)   |
| 3 | Login         | Gets token — **saves it automatically** |
| 4 | Get Products  | Lists products (uses saved token)     |
| 5 | Place Order   | Creates order (uses saved token)       |

**Important:** Run **Login** before Products and Orders. The token is saved automatically.

**Place Order body:** `{"items":[{"productId":1,"quantity":2}]}` — customer ID comes from the token.

---

## Run from Terminal (Newman)

If you have Node.js:

```bash
npm install -g newman
newman run postman/Ecom-API.postman_collection.json -e postman/Local.postman_environment.json
```

---

## Different Port?

If API runs on 8081: Edit **Local** environment → change `base_url` to `http://localhost:8081`
