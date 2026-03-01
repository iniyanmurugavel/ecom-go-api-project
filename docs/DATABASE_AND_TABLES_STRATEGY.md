# Database and Tables Strategy — Learning Guide

This document explains **how the database is structured**, **why tables are designed this way**, and **the strategy** behind it. For learning.

---

## 1. Overview: What Tables Exist

| Table | Purpose |
|-------|---------|
| **products** | Catalog: name, price, stock |
| **customers** | Users: email, password hash, name |
| **orders** | Order header: which customer, when |
| **order_items** | Line items: which product, quantity, price at time of order |

---

## 2. Table-by-Table Strategy

### 2.1 `products`

```
id (BIGSERIAL PK)
name (TEXT NOT NULL)
price_in_centers (INTEGER, >= 0)   -- "centers" = typo for "cents"
quantity (INTEGER, default 0)      -- stock
created_at (TIMESTAMPTZ)
```

**Why this design:**
- **id:** Auto-increment primary key. Simple, works for small/medium scale.
- **price_in_centers:** Store in cents to avoid floating-point (e.g. $19.99 → 1999).
- **quantity:** Stock. Decremented when an order is placed.
- **created_at:** Audit trail; useful for "new arrivals" or reporting.

**Strategy:** One row per product. No variants (size, color) — that would need a `product_variants` table.

---

### 2.2 `customers`

```
id (BIGSERIAL PK)
email (TEXT UNIQUE NOT NULL)
password_hash (TEXT NOT NULL)
name (TEXT NOT NULL)
created_at (TIMESTAMPTZ)
```

**Why this design:**
- **email UNIQUE:** One account per email. Login uses email.
- **password_hash:** Never store plain passwords. We use bcrypt.
- **name:** Display name (e.g. "John Doe").

**Strategy:** One row per customer. Auth (register/login) creates/validates against this table.

---

### 2.3 `orders`

```
id (BIGSERIAL PK)
customer_id (BIGINT NOT NULL)
created_at (TIMESTAMPTZ)
```

**Why this design:**
- **customer_id:** Who placed the order. Links to `customers.id`.
- **created_at:** When the order was placed.

**Strategy:** One row per order (the "header"). Line items go in `order_items`. This is a classic **header-detail** pattern: one order, many items.

**Note:** There is no FK from `orders.customer_id` to `customers.id` in the schema (to allow flexibility with existing data). In a fresh project you'd add it.

---

### 2.4 `order_items`

```
id (BIGSERIAL PK)
order_id (BIGINT NOT NULL) → orders.id
product_id (BIGINT NOT NULL) → products.id
quantity (INTEGER NOT NULL)
price_cents (INTEGER NOT NULL)
```

**Why this design:**
- **order_id:** Which order this line belongs to.
- **product_id:** Which product. FK to `products` ensures it exists.
- **quantity:** How many of this product in this order.
- **price_cents:** Price **at time of order**. Products can change price later; we store snapshot.

**Strategy:** One row per product in an order. If someone orders 2x Product A and 1x Product B, that's 2 rows in `order_items`.

**Why store price_cents?** If you only stored `product_id` and `quantity`, you'd look up current price from `products`. But prices change. We need historical accuracy: "They paid $19.99 when they ordered" — so we copy the price into the order.

---

## 3. Relationships (ER Diagram, Text)

```
customers (1) ──────< orders (many)
    │                      │
    │                      └──< order_items (many)
    │                               │
    └───────────────────────────────┴──> products (many)
```

- One customer → many orders
- One order → many order_items
- One product → many order_items (across different orders)

---

## 4. Why Separate `orders` and `order_items`?

**Alternative:** One table `orders` with columns like `product_1_id`, `product_2_id`, … (repeated columns).

**Problems:**
- Fixed number of items per order
- Hard to query "all orders containing product X"
- Wastes space

**Our approach:** Normalized. `orders` = header, `order_items` = detail. Flexible, easy to query, standard pattern.

---

## 5. Migration Order (Why This Sequence)

1. **products** — No dependencies
2. **orders + order_items** — order_items needs orders
3. **customers** — Auth; orders reference customer_id
4. **FK order_items → products** — Ensures we can't order a non-existent product

Order matters: you can't create `order_items` before `orders` exists.

---

## 6. Indexes (Current and Future)

**Current:** Primary keys (`id`) are indexed by default.

**Useful to add later:**
- `products(name)` — if you search by name
- `orders(customer_id)` — list orders by customer
- `order_items(order_id)` — list items for an order (often covered by FK)
- `customers(email)` — login lookup (UNIQUE implies index)

---

## 7. Summary

| Decision | Reason |
|----------|--------|
| Header-detail (orders + order_items) | Flexible, normalized, standard |
| price_cents in order_items | Historical accuracy |
| customers separate from orders | Auth vs orders; reuse customer |
| BIGSERIAL for IDs | Room to grow; consistent with Go int64 |

---

## 8. Related Docs

- [MIGRATIONS_README.md](../MIGRATIONS_README.md) — How migrations work
- [docs/MIGRATION_EXAMPLE.md](MIGRATION_EXAMPLE.md) — Adding a column step-by-step
