# Prompt: Implement Tags & Labels Feature on Frontend

## Context

You are working on the **money-manage** frontend application. A new backend feature — **Tags & Labels for Transaction V2** — has been fully implemented and is ready to integrate into the UI.

Before writing any code, **study the existing project thoroughly**:

- Read the existing folder structure, component conventions, naming patterns, and design system
- Understand how the current Transaction V2 form and list pages are built
- Identify the API service/hook patterns already used (e.g., how auth tokens are sent, how responses are handled)
- Follow **all existing patterns** — do not introduce new libraries, state managers, or design styles unless they already exist in the project

---

## Backend API Reference

All endpoints are under `/api/v2` and require a Bearer token in the `Authorization` header.

### Tag CRUD

| Method   | Endpoint               | Notes                                                                    |
| -------- | ---------------------- | ------------------------------------------------------------------------ |
| `GET`    | `/api/v2/tags`         | Supports `?sort=usage`                                                   |
| `GET`    | `/api/v2/tags/:id`     |                                                                          |
| `POST`   | `/api/v2/tags`         | Body: `name`, `color` (hex, default `#6366F1`), `icon` (emoji, optional) |
| `PUT`    | `/api/v2/tags/:id`     |                                                                          |
| `DELETE` | `/api/v2/tags/:id`     | Soft delete                                                              |
| `GET`    | `/api/v2/tags/suggest` | Query: `category_id`, `description`                                      |

### Transaction Tags

| Method   | Endpoint                                | Notes                         |
| -------- | --------------------------------------- | ----------------------------- |
| `POST`   | `/api/v2/transactions/:id/tags`         | Body: `{ "tag_ids": [1, 2] }` |
| `DELETE` | `/api/v2/transactions/:id/tags/:tag_id` | Removes tag from transaction  |

### Analytics

| Method | Endpoint                            | Notes                                        |
| ------ | ----------------------------------- | -------------------------------------------- |
| `GET`  | `/api/v2/analytics/spending-by-tag` | Query: `start_date`, `end_date` (YYYY-MM-DD) |

---

## What to Implement

### 1. Tag Management Page (CRUD)

Create a dedicated page/section where users can manage their tags. It should allow:

- Viewing all their tags in a list or grid
- Creating a new tag (name, color picker using hex, optional emoji icon)
- Editing an existing tag
- Deleting a tag (with a confirmation prompt)
- Sorting by usage count

Match the look and feel of existing list/management pages in the project.

### 2. Tag Selector Component (Reusable)

Create a **reusable component** for selecting tags. It should:

- Be usable on both the **Create Transaction** and **Update Transaction** forms
- Show existing tags as chips/badges using their color and icon
- Allow multi-select
- Show a search/filter input to find tags by name
- Include an inline shortcut to create a new tag without leaving the form
- Show tag suggestions from `GET /api/v2/tags/suggest` when a category and/or description is already filled in

Integrate this component into the existing Transaction V2 create and edit forms.

### 3. Tag Display in Transaction List & Detail

- Show the tags on each transaction row/card in the transaction list view (as small colored chips)
- Show full tag details (with icon) on the transaction detail/show view
- If tags can be edited from the detail view, allow adding/removing individual tags inline

### 4. Tag Analytics Widget/Section

Add a "Spending by Tag" view in the analytics section:

- Allow the user to pick a date range
- Show each tag with its total spending, transaction count, and average amount
- Display tag color and icon alongside its name
- Match the style of existing analytics/chart sections in the project

---

## Behaviour & UX Rules

- **Optimistic UI**: When adding or removing a tag from a transaction, update the UI immediately and roll back on error
- **Usage count**: After creating a tag or attaching it to a transaction, reflect the updated usage_count in the UI without a full page reload
- **Empty states**: Handle the case where a user has no tags yet — show a friendly prompt to create one
- **Error handling**: Show clear, user-friendly messages for errors (e.g., duplicate tag name, tag not found)
- **Authorization**: Never expose another user's tags; the backend enforces this but the frontend must not accidentally pass wrong IDs

---

## Constraints

- Follow the **exact same code structure, file naming, and folder layout** as the rest of the project
- Use the **same HTTP client, auth helper, and error handling utilities** already present in the project
- Match the **existing design system** — same colors, typography, spacing, and component style
- Do not install new dependencies unless absolutely necessary and pre-approved
- Write clean, readable code with the same level of comments and structure as existing files
