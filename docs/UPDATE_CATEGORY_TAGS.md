# Update Transaction Category & Tags

## Endpoint

```
PUT /api/v2/transactions/:id
```

Requires `Authorization: Bearer <token>` header.

---

## Request Body

All fields are optional. Only fields that are present are updated.

| Field             | Type     | Description                                                  |
|-------------------|----------|--------------------------------------------------------------|
| `category_id`     | `uint`   | Set a new category. Pass `null` to clear the category.       |
| `tag_ids`         | `[]uint` | **Replaces** the full tag list. Pass `[]` to remove all tags. Omit the field entirely to leave tags unchanged. |
| `description`     | `string` | Update the description.                                      |
| `amount`          | `int`    | Update the amount (triggers balance recalculation).          |
| `asset_id`        | `uint64` | Move transaction to a different wallet.                      |
| `transaction_type`| `string` | `"Income"` or `"Expense"`                                    |
| `date`            | `string` | `"YYYY-MM-DD"` or RFC3339                                    |

---

## Examples

### Update category only

```json
PUT /api/v2/transactions/42
{
  "category_id": 3
}
```

### Clear category

```json
PUT /api/v2/transactions/42
{
  "category_id": null
}
```

### Replace all tags

```json
PUT /api/v2/transactions/42
{
  "tag_ids": [1, 5, 9]
}
```

### Remove all tags

```json
PUT /api/v2/transactions/42
{
  "tag_ids": []
}
```

### Update category and tags together

```json
PUT /api/v2/transactions/42
{
  "category_id": 3,
  "tag_ids": [1, 5]
}
```

---

## Response

```json
{
  "success": true,
  "message": "Transaction updated successfully",
  "data": {
    "id": 42,
    "description": "SeaBank - Top Up e-wallet - GOJEK Hxxxxx - 3496 - IDR",
    "amount": 50000,
    "transaction_type": 2,
    "date": "2026-03-25T20:27:00Z",
    "category_name": "Food & Beverage",
    "bank_name": "",
    "asset_id": 1,
    "asset_name": "SeaBank",
    "asset_type": "bank",
    "tags": [
      { "id": 1, "name": "gojek", "color": "#00AA5B" },
      { "id": 5, "name": "top-up", "color": "#0080FF" }
    ]
  }
}
```

---

## Tag Management Endpoints

In addition to bulk replacement via `tag_ids` in the update endpoint, individual tag operations are available:

### Add tags (append, does not remove existing)

```
POST /api/v2/transactions/:id/tags
```

```json
{
  "tag_ids": [2, 7]
}
```

### Remove a single tag

```
DELETE /api/v2/transactions/:id/tags/:tag_id
```

---

## Notes

- `tag_ids` values must belong to the authenticated user — foreign tags are rejected with `400`.
- Omitting `tag_ids` from the body leaves current tags untouched.
- `tag_ids: []` explicitly clears all tags from the transaction.
- `category_id` accepts any category ID returned by `GET /api/my-categories` or `GET /api/categories`.
