# Migration Example — Adding a Column Without Data Loss

This guide walks through **adding a new column** to an existing table **without losing data**. Step-by-step, for learning.

---

## 1. Scenario: Add `sku` to `products`

**Goal:** Add a `sku` (Stock Keeping Unit) column to `products`. Existing rows should keep their data; new column can be NULL for old rows until we backfill.

**Result:** No data loss. Existing products stay as-is.

---

## 2. Step-by-Step Workflow

### Step 1: Create the migration file

Goose expects files in order: `00001_`, `00002_`, … `00005_`, etc.

```bash
# From project root
goose -dir ./internal/adapters/postgresql/migrations create add_sku_to_products sql
```

This creates `00005_add_sku_to_products.sql` (or next number). Edit it:

```sql
-- +goose Up
-- +goose StatementBegin
ALTER TABLE products ADD COLUMN sku TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE products DROP COLUMN sku;
-- +goose StatementEnd
```

**What this does:**
- **Up:** Adds `sku` column. Existing rows get `NULL` for `sku`. No rows deleted.
- **Down:** Removes `sku`. Data in that column is lost (but we're rolling back, so that's expected).

---

### Step 2: Run the migration

```bash
source .env   # or: export GOOSE_DRIVER=postgres GOOSE_DBSTRING="..."
goose up
```

Output: `goose: successfully migrated database to version 5`

**What happens:** PostgreSQL runs `ALTER TABLE products ADD COLUMN sku TEXT`. Existing data is untouched.

---

### Step 3: Update SQLC queries (if needed)

If you have a query that does `SELECT * FROM products`, SQLC will now include `sku` in the generated struct. No change needed to `queries.sql` for that.

If you add a new query that uses `sku`:

```sql
-- In queries.sql
-- name: FindProductBySKU :one
SELECT * FROM products WHERE sku = $1;
```

---

### Step 4: Regenerate SQLC code

```bash
sqlc generate
```

This updates `internal/adapters/postgresql/sqlc/` with the new column in Go structs.

---

### Step 5: Update your Go code

Use the new field in handlers/services:

```go
// Example: product now has .Sku (nullable *string)
if product.Sku != nil {
    // use *product.Sku
}
```

---

### Step 6: Restart the API

```bash
go run ./cmd
```

---

## 3. What If You Need NOT NULL?

If `sku` must be NOT NULL, you can't add it as `NULL` for existing rows. Two approaches:

### Option A: Add with DEFAULT

```sql
ALTER TABLE products ADD COLUMN sku TEXT NOT NULL DEFAULT 'UNKNOWN';
```

- Existing rows get `sku = 'UNKNOWN'`
- No data loss
- You can later `UPDATE` to real values, then change default if needed

### Option B: Add nullable, backfill, then add NOT NULL

```sql
-- Step 1: Add nullable
ALTER TABLE products ADD COLUMN sku TEXT;

-- Step 2: Backfill (e.g. generate from id)
UPDATE products SET sku = 'SKU-' || id WHERE sku IS NULL;

-- Step 3: Add NOT NULL
ALTER TABLE products ALTER COLUMN sku SET NOT NULL;
```

---

## 4. What You Need to Know for Learning

| Concept | Explanation |
|---------|-------------|
| **Migrations are forward-only in prod** | You rarely run `goose down` in production. It can drop data. |
| **Up = apply, Down = rollback** | Down undoes Up. Use Down for local dev rollback only. |
| **ALTER ADD COLUMN** | Safe. Existing rows get NULL or default. No data loss. |
| **ALTER DROP COLUMN** | Destructive. All data in that column is gone. |
| **Order matters** | Migration 5 runs after 4. Dependencies (e.g. FK) must exist first. |
| **SQLC follows schema** | After migration, run `sqlc generate` so Go code matches DB. |

---

## 5. Checklist: Add Column Without Data Loss

- [ ] Create migration file with `ALTER TABLE ... ADD COLUMN`
- [ ] Use `TEXT` or nullable type (or `DEFAULT`) so existing rows are valid
- [ ] Run `goose up`
- [ ] Run `sqlc generate`
- [ ] Update handlers/services to use new field
- [ ] Restart API
- [ ] (Optional) Backfill existing rows if needed

---

## 6. What NOT to Do

| Don't | Why |
|-------|-----|
| Add NOT NULL without DEFAULT or backfill | Migration fails if existing rows would have NULL |
| Drop column in Up without understanding | Data is lost |
| Run `goose down` in production | Can drop tables/columns and lose data |
| Forget `sqlc generate` | Go structs won't match DB; code may break |

---

## 7. Quick Use Cases (Add / Delete Columns)

### Use case A: Add one column

```sql
-- +goose Up
ALTER TABLE products ADD COLUMN sku TEXT;

-- +goose Down
ALTER TABLE products DROP COLUMN sku;
```

---

### Use case B: Add two columns

```sql
-- +goose Up
ALTER TABLE products ADD COLUMN sku TEXT;
ALTER TABLE products ADD COLUMN is_active BOOLEAN NOT NULL DEFAULT true;

-- +goose Down
ALTER TABLE products DROP COLUMN is_active;
ALTER TABLE products DROP COLUMN sku;
```

**Note:** Drop in reverse order (last added, first dropped). Or drop both in one Down — order doesn't matter for DROP.

---

### Use case C: Delete one column

```sql
-- +goose Up
ALTER TABLE products DROP COLUMN sku;

-- +goose Down
ALTER TABLE products ADD COLUMN sku TEXT;
```

**Warning:** All data in `sku` is **lost**. Down adds an empty column (NULL for existing rows). You cannot restore the old values.

---

### Use case D: Add one, delete one (rename / replace)

Example: Replace `price_in_centers` with `price_in_cents` (fix typo, keep data).

```sql
-- +goose Up
ALTER TABLE products ADD COLUMN price_in_cents INTEGER;
UPDATE products SET price_in_cents = price_in_centers;
ALTER TABLE products DROP COLUMN price_in_centers;
ALTER TABLE products ALTER COLUMN price_in_cents SET NOT NULL;

-- +goose Down
ALTER TABLE products ADD COLUMN price_in_centers INTEGER;
UPDATE products SET price_in_centers = price_in_cents;
ALTER TABLE products ALTER COLUMN price_in_centers SET NOT NULL;
ALTER TABLE products DROP COLUMN price_in_cents;
```

---

## 8. Related Docs

- [MIGRATIONS_README.md](../MIGRATIONS_README.md) — Full migration guide
- [docs/DATABASE_AND_TABLES_STRATEGY.md](DATABASE_AND_TABLES_STRATEGY.md) — Table design
