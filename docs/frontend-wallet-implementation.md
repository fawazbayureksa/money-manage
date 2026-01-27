# Wallet / Asset (Wallet) - Frontend Implementation Plan

This document outlines how to implement the Wallet/Asset feature on the frontend to complement the Go backend MVP you added. It assumes a REST API surface mounted on /api/wallets as described in the backend plan.

## 1. Objective
- Build a user-facing wallet dashboard that lists per-user assets, allows creating/updating/deleting assets, and provides a per-currency balance summary.
- Provide an initial, production-friendly UX while keeping the code maintainable and extensible for future enhancements (e.g., currency conversion, market value).

## 2. API surface and data contracts
- Endpoints (REST, protected):
  - GET /api/wallets -> list all assets for the current user
  - GET /api/wallets/:id -> get a specific asset
  - POST /api/wallets -> create asset
  - PUT /api/wallets/:id -> update asset
  - DELETE /api/wallets/:id -> delete asset
  - GET /api/wallets/summary -> per-currency totals

- Asset (server response shape):
  - id: number
  - user_id: number
  - name: string
  - type: string
  - balance: number
  - currency: string
  - bank_name: string|null
  - account_no: string|null
  - created_at: string (ISO 8601)
  - updated_at: string (ISO 8601)

- Create/Update payloads (client to server):
  - Create:
    { name, type, balance, currency, bank_name?, account_no? }
  - Update:
    { name?, type?, balance?, currency?, bank_name?, account_no? }

## 3. Client-side data models (TypeScript hints)
- Asset interface
```
interface Asset {
  id: number;
  user_id: number;
  name: string;
  type: string;
  balance: number;
  currency: string;
  bank_name?: string | null;
  account_no?: string | null;
  created_at: string;
  updated_at: string;
}
```
- DTOs for create/update can mirror the API payloads above.

## 4. Tech and architecture choices
- Frontend framework: React with TypeScript (recommended). Alternatively Vue or Svelte can be used with a similar approach.
- Data fetching: React Query (preferred) or SWR for caching and stale-while-revalidate behavior.
- State management: Local component state for forms; a small Auth/User context to supply the current user token. React Query handles server data caching and invalidation.
- Styling: CSS variables with a deliberate design system. Avoid purple-on-white defaults; ensure responsive layout.
- Charts: Chart.js or Recharts for the per-currency summary visualization.

## 5. UI components and composition
- WalletDashboard: the main page showing an overview, summary chart, and a list/grid of assets.
- WalletList: list of asset items with search/sort and quick actions (edit/delete).
- WalletCard: a compact card showing asset name, type, balance, and currency.
- WalletFormDrawer/Modal: used for Create and Update flows; reuse fields and validation.
- WalletSummaryChart: per-currency breakdown (bar or doughnut chart).
- Helpers: currency formatter, number formatting, and date display.

## 6. UX guidelines
- Quick add: a prominent “New Asset” button to open the drawer.
- Consistency: enforce the same input patterns across create and update forms.
- Accessibility: proper aria-labels, keyboard navigation, and readable contrast.
- Internationalization: format balances with Intl.NumberFormat, respecting currency code when possible.

## 7. Data flow and state management
- On login, load assets via GET /api/wallets and populate the WalletList and SummaryChart.
- Mutations (create/update/delete) should optimistically reflect changes where feasible, then reconcile with server response.
- After any mutation, invalidate/fetch /api/wallets and /api/wallets/summary to keep data in sync.
- Authentication: attach Authorization header with token; reuse existing auth utilities from your app.

## 8. Data fetching hooks (suggested)
- useWallets(): fetch list of assets for current user
- useWallet(id): fetch single asset
- useCreateWallet(dto): create a new asset
- useUpdateWallet(id, dto): update asset
- useDeleteWallet(id): delete asset
- useWalletSummary(): fetch /api/wallets/summary
- These can be implemented with React Query or your preferred data layer.

## 9. Validation and error handling
- Client-side: require name and currency; balance >= 0.
- Server responses: surface error messages in toasts or inline errors.
- Ownership: server ensures assets belong to the current user; the frontend should handle 403/401 cleanly.

## 10. Accessibility and localization
- Keyboard friendly forms and dialogs; aria-live for errors.
- Currency formatting and locale-aware date strings.

## 11. Testing plan
- Unit tests for formatting helpers and small utility functions.
- Integration tests for API data fetching and mutation flows (mock API or test API).
- End-to-end tests (Cypress/Playwright) for core wallet flows: list -> create -> update -> delete -> summary.

## 12. API error handling mapping (frontend)
- Map 400/422 to user-friendly validation messages.
- Map 401/403 to authentication prompts (redirect to login or token refresh).
- 5xx to generic error toast with retry suggestion.

## 13. Migration and compatibility considerations
- Ensure the backend asset endpoints are ready; if not, implement a stub/mock layer for local prototyping.
- If the shape changes (e.g., new fields), update the frontend type definitions accordingly.

## 14. Future enhancements (optional)
- Multi-currency support with real-time currency conversion.
- Asset value history and market value integration.
- Sorting, filtering by type/currency, and pagination for asset lists.

## 15. Milestones and phased plan
- Phase 1 — Read-only wallet view (list + summary):
  - UI scaffolding, fetch assets, render list and summary chart.
- Phase 2 — CRUD operations:
  - Create, edit, delete assets via modals/drawers; ensure optimistic updates and re-fetch strategy.
- Phase 3 — UX polish and charts:
  - Improve chart visuals, add currency formatting, add empty states and skeleton loaders.
- Phase 4 — Testing:
  - Add unit/integration tests and E2E tests.

## 16. Quickstart usage examples
- Hook usage pattern (pseudo-code):
```
const { data: wallets, isLoading } = useWallets(); // GET /api/wallets
const { mutate: createWallet } = useCreateWallet(); // POST /api/wallets
```
- Example payloads:
```
POST /api/wallets
{ "name": "PayPal", "type": "Online Wallet", "balance": 250.00, "currency": "USD" }
```

## 17. Risks and mitigations
- Risk: mismatch between frontend data shapes and backend models. Mitigation: align TypeScript interfaces with Go structs; add runtime validation and error handling.
- Risk: exposing sensitive bank account numbers. Mitigation: mask or omit account_no in list views; only show when editing if necessary.

## 18. Developer notes
- Reuse existing auth flows to obtain user context; implement a stable getCurrentUser hook to feed wallet endpoints.
- Use a single source of truth for currency formatting to reduce bugs.

## 19. Appendix: sample data
- Sample asset object:
```
{
  "id": 1,
  "user_id": 42,
  "name": "Main Checking",
  "type": "Bank",
  "balance": 1250.75,
  "currency": "USD",
  "bank_name": "Example Bank",
  "account_no": "123456789",
  "created_at": "2026-01-21T12:00:00Z",
  "updated_at": "2026-01-21T12:00:00Z"
}
```

This plan is designed to be actionable and integrate smoothly with the backend work you’ve started. If you want, I can convert this into a repo patch that scaffolds React components and hooks, or tailor it to your chosen frontend framework (React, Vue, Svelte). 
