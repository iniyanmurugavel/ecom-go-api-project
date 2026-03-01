# ACID Properties — Explained for Learning

This document explains **ACID** (Atomicity, Consistency, Isolation, Durability) — the four properties that make database transactions reliable. With examples from this project.

---

## 1. What Is a Transaction?

A **transaction** is a group of operations that the database treats as **one unit**. Either **all** succeed, or **none** do.

**Example (PlaceOrder):**
1. Create order row
2. Create order_item for product A
3. Create order_item for product B
4. Decrement stock for A
5. Decrement stock for B

If step 4 fails (e.g. out of stock), we **roll back** steps 1–3. The database is as if we never started.

---

## 2. ACID — The Four Properties

### A — Atomicity

**Definition:** All operations in a transaction succeed together, or **none** of them do. No partial updates.

**In this project:** `PlaceOrder` runs in one transaction. If `DecrementProductStock` fails (e.g. another request took the last unit), we `Rollback`. The order and order_items we created are **undone**. The customer sees "product has not enough stock" — no orphan order, no incorrect stock.

```go
tx, _ := pool.Begin(ctx)
defer tx.Rollback(ctx)  // runs if we return before Commit

// ... create order, order_items, decrement stock ...

tx.Commit(ctx)  // only now is everything permanent
```

---

### C — Consistency

**Definition:** The database moves from one **valid state** to another. Constraints (NOT NULL, FK, CHECK) are never violated.

**In this project:**
- `order_items.product_id` has FK to `products.id` — we can't order a non-existent product
- `products.quantity >= 0` (CHECK) — we never go negative
- `DecrementProductStock` uses `WHERE quantity >= $1` — if stock is 2 and we want 5, the UPDATE affects 0 rows and we roll back

**Consistency** means: before and after the transaction, the DB satisfies all rules. We never commit a state where stock is negative or an order references a missing product.

---

### I — Isolation

**Definition:** Concurrent transactions don't see each other's **uncommitted** changes. One transaction's work is invisible to others until it commits.

**In this project:** Two users order the last unit of the same product at the same time.

- Transaction A: reads stock = 1, creates order, tries to decrement
- Transaction B: reads stock = 1, creates order, tries to decrement

With **isolation**, one of them will "win" (UPDATE affects 1 row). The other will get 0 rows (stock already 0) and roll back. We avoid overselling.

PostgreSQL uses **MVCC** (Multi-Version Concurrency Control): each transaction sees a snapshot. Writes are serialized; you don't get dirty reads (seeing uncommitted data).

---

### D — Durability

**Definition:** Once a transaction is **committed**, the changes are **permanent**. A crash or power loss won't lose them.

**In this project:** After `tx.Commit(ctx)`, the order and stock updates are written to disk (via WAL — Write-Ahead Log). Even if the server crashes right after, PostgreSQL replays the log on restart and the data is there.

---

## 3. How This Project Uses ACID

| Step in PlaceOrder | ACID in action |
|--------------------|----------------|
| `Begin` | Start transaction |
| `CreateOrder` | Part of atomic unit |
| `CreateOrderItem` (each) | Part of atomic unit |
| `DecrementProductStock` | Fails if stock insufficient → rollback (Atomicity) |
| `Commit` | All changes durable (Durability) |
| FK, CHECK constraints | Consistency |
| Concurrent requests | Isolation (no dirty reads, no oversell) |

---

## 4. What Happens Without a Transaction?

If we did **not** use a transaction:

1. Create order ✓
2. Create order_item ✓
3. Decrement stock ✗ (fails — out of stock)

Result: We'd have an order and order_item in the DB, but stock wouldn't be decremented. **Inconsistent** — we "sold" something we don't have. Atomicity and consistency would be broken.

---

## 5. Summary Table

| Property | Meaning | Example in this project |
|----------|---------|-------------------------|
| **Atomicity** | All or nothing | Rollback order + items if stock decrement fails |
| **Consistency** | Valid state before/after | FK, CHECK; no negative stock |
| **Isolation** | Concurrent tx don't interfere | Two users, last item — one wins, one rolls back |
| **Durability** | Committed = permanent | WAL; survives crash |

---

## 6. Related Code

- `internal/orders/service.go` — `PlaceOrder` uses `Begin`, `Commit`, `Rollback`
- `DecrementProductStock` — `UPDATE ... WHERE quantity >= $1` ensures we never oversell

---

## 7. Further Reading

- PostgreSQL: [Transaction Isolation](https://www.postgresql.org/docs/current/transaction-iso.html)
- MVCC: How PostgreSQL keeps isolation without locking everything
