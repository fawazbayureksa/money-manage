Got it 👍
I’ll **re-adjust all analysis queries** so they are **100% consistent** with your actual column names:

### ✅ Confirmed schema (based on your query)

**transactions**

* `id`
* `description`
* `category_id`
* `bank_id`
* `amount`
* `transaction_type` → `1 = income`, `2 = expense`
* `date`

**categories**

* `id`
* `category_name`

**banks**

* `id`
* `bank_name`

---

## 1️⃣ Base Transaction List (yours, clean version)

```sql
SELECT 
    t.id,
    t.description AS transaction_name,
    c.category_name AS category,
    b.bank_name AS bank,
    t.amount,
    t.transaction_type,
    t.date
FROM transactions t
LEFT JOIN categories c ON c.id = t.category_id
LEFT JOIN banks b ON b.id = t.bank_id
ORDER BY t.date DESC;
```

---

## 2️⃣ Total Income vs Expense

```sql
SELECT
    SUM(CASE WHEN t.transaction_type = 1 THEN t.amount ELSE 0 END) AS total_income,
    SUM(CASE WHEN t.transaction_type = 2 THEN t.amount ELSE 0 END) AS total_expense,
    SUM(CASE WHEN t.transaction_type = 1 THEN t.amount ELSE -t.amount END) AS net_amount
FROM transactions t;
```

---

## 3️⃣ Income vs Expense per Category

```sql
SELECT
    c.category_name AS category,
    SUM(CASE WHEN t.transaction_type = 1 THEN t.amount ELSE 0 END) AS income,
    SUM(CASE WHEN t.transaction_type = 2 THEN t.amount ELSE 0 END) AS expense
FROM transactions t
LEFT JOIN categories c ON c.id = t.category_id
GROUP BY c.id, c.category_name;
```

---

## 4️⃣ Expense per Category (Top Spending)

```sql
SELECT
    c.category_name AS category,
    SUM(t.amount) AS total_expense
FROM transactions t
LEFT JOIN categories c ON c.id = t.category_id
WHERE t.transaction_type = 2
GROUP BY c.id, c.category_name
ORDER BY total_expense DESC;
```

---

## 5️⃣ Expense per Bank

```sql
SELECT
    b.bank_name AS bank,
    SUM(t.amount) AS total_expense
FROM transactions t
LEFT JOIN banks b ON b.id = t.bank_id
WHERE t.transaction_type = 2
GROUP BY b.id, b.bank_name;
```

---

## 6️⃣ Monthly Cashflow

```sql
SELECT
    DATE_FORMAT(t.date, '%Y-%m') AS month,
    SUM(CASE WHEN t.transaction_type = 1 THEN t.amount ELSE 0 END) AS income,
    SUM(CASE WHEN t.transaction_type = 2 THEN t.amount ELSE 0 END) AS expense,
    SUM(CASE WHEN t.transaction_type = 1 THEN t.amount ELSE -t.amount END) AS net
FROM transactions t
GROUP BY month
ORDER BY month;
```

---

## 7️⃣ Daily Expense Trend

```sql
SELECT
    DATE(t.date) AS day,
    SUM(t.amount) AS total_expense
FROM transactions t
WHERE t.transaction_type = 2
GROUP BY day
ORDER BY day;
```

---

## 8️⃣ Expense Percentage per Category (Pie Chart Ready)

```sql
SELECT
    c.category_name AS category,
    SUM(t.amount) AS total,
    ROUND(
        SUM(t.amount) /
        (SELECT SUM(amount) FROM transactions WHERE transaction_type = 2) * 100,
        2
    ) AS percentage
FROM transactions t
LEFT JOIN categories c ON c.id = t.category_id
WHERE t.transaction_type = 2
GROUP BY c.id, c.category_name;
```

---

## 9️⃣ Top N Transactions (Biggest Expense)

```sql
SELECT
    t.description AS transaction_name,
    c.category_name AS category,
    b.bank_name AS bank,
    t.amount,
    t.date
FROM transactions t
LEFT JOIN categories c ON c.id = t.category_id
LEFT JOIN banks b ON b.id = t.bank_id
WHERE t.transaction_type = 2
ORDER BY t.amount DESC
LIMIT 10;
```

---

## 🔧 Recommended Indexes (MySQL)

Highly recommended for analytics performance:

```sql
CREATE INDEX idx_transactions_date ON transactions(date);
CREATE INDEX idx_transactions_type ON transactions(transaction_type);
CREATE INDEX idx_transactions_category ON transactions(category_id);
CREATE INDEX idx_transactions_bank ON transactions(bank_id);
```

---

## 🚀 What I can do next for you

* Convert these queries into **Laravel Eloquent**
* Prepare **API JSON response format** (chart-friendly)
* Add **budget vs actual** logic
* Create **materialized summary table** for fast dashboard

Just tell me what’s next 👌
