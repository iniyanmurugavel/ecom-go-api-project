# Make Mac Postgres Use Port 15432

You can configure Mac Postgres (Postgres.app or Homebrew) to listen on port **15432** instead of 5432. Then the project uses the same port whether you run Docker or Mac Postgres — no conflicts, consistent config.

---

## Steps

### 1. Find postgresql.conf

**Homebrew (Apple Silicon):**
```bash
ls /opt/homebrew/var/postgresql*/postgresql.conf
```

**Homebrew (Intel):**
```bash
ls /usr/local/var/postgres/postgresql.conf
```

**Postgres.app:**
```
~/Library/Application Support/Postgres/var/postgresql.conf
```

**Or query Postgres:**
```bash
psql -U $USER -d postgres -c 'SHOW config_file;'
```

### 2. Edit the file

Open `postgresql.conf` and change:
```
port = 5432
```
to:
```
port = 15432
```

Save the file.

### 3. Restart Postgres

**Homebrew:**
```bash
brew services restart postgresql
```

**Postgres.app:** Quit the app and reopen it.

### 4. Create postgres user (if needed)

Mac Postgres often uses your Mac username by default. To create a `postgres` user:
```bash
createuser -s postgres
```

Set a password:
```bash
psql -U postgres -d postgres -c "ALTER USER postgres PASSWORD 'postgres';"
```

### 5. Create database

```bash
createdb -U postgres ecom
```

### 6. Update .env

```
GOOSE_DBSTRING="host=localhost port=15432 user=postgres password=postgres dbname=ecom sslmode=disable"
```

### 7. Run migrations

```bash
source .env
goose up
```

---

## Result

- **Mac Postgres:** port 15432
- **Docker Postgres:** port 15432 (from docker-compose)
- **Project .env:** port 15432
- **DB client:** port 15432

Everything uses **15432** consistently.
