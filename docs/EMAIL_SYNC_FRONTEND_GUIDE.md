# Email Sync — Frontend Integration Guide

Auto-imports bank transaction emails (BCA, Permata, SeaBank) from the user's Gmail inbox.

---

## How it works

```
User clicks "Connect Gmail"
  → GET /api/v2/email-sync/auth          (your backend)
  → Open returned auth_url in browser
  → User logs in & grants access on Google
  → Google redirects to /api/v2/email-sync/callback (your backend)
  → Backend saves token, redirects to EMAIL_SYNC_REDIRECT_URL?email_sync=success
  → Frontend detects ?email_sync=success and updates UI
```

The backend env var `EMAIL_SYNC_REDIRECT_URL` must point to your frontend URL (e.g. `http://localhost:3000` or your production domain).

---

## API Reference

All endpoints require `Authorization: Bearer <jwt_token>` except the callback.

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/api/v2/email-sync/auth` | ✅ | Get Google OAuth URL |
| GET | `/api/v2/email-sync/callback` | ❌ | Google redirect (do not call manually) |
| POST | `/api/v2/email-sync/sync` | ✅ | Trigger email sync |
| GET | `/api/v2/email-sync/status?page=1&limit=20` | ✅ | Connection status + logs |
| DELETE | `/api/v2/email-sync/disconnect` | ✅ | Remove Gmail connection |

### GET `/api/v2/email-sync/auth`
```json
{
  "success": true,
  "data": {
    "auth_url": "https://accounts.google.com/o/oauth2/auth?client_id=..."
  }
}
```

### POST `/api/v2/email-sync/sync`
```json
{
  "success": true,
  "data": {
    "total": 15,
    "imported": 10,
    "skipped": 4,
    "failed": 1
  }
}
```

### GET `/api/v2/email-sync/status`
```json
{
  "success": true,
  "data": {
    "connected": true,
    "logs": [
      {
        "id": 1,
        "gmail_message_id": "18efa2c...",
        "subject": "Internet Transaction Journal",
        "from_email": "bca@bca.co.id",
        "bank_name": "BCA",
        "amount": 138600,
        "asset_id": 2,
        "transaction_id": 45,
        "status": "imported",
        "error_message": "",
        "email_date": "2026-03-24T18:08:10Z",
        "created_at": "2026-03-28T10:00:00Z"
      }
    ],
    "total": 42,
    "page": 1,
    "limit": 20
  }
}
```

`status` values: `imported` | `skipped` | `failed`

---

## React JS

### 1. Install nothing extra — uses native `fetch` and `window.open`

### 2. Email Sync Hook (`useEmailSync.js`)

```js
import { useState, useEffect, useCallback } from 'react';

const BASE_URL = process.env.REACT_APP_API_URL || 'http://localhost:8080';

export function useEmailSync() {
  const [connected, setConnected]   = useState(false);
  const [logs, setLogs]             = useState([]);
  const [total, setTotal]           = useState(0);
  const [syncing, setSyncing]       = useState(false);
  const [connecting, setConnecting] = useState(false);
  const [syncResult, setSyncResult] = useState(null);
  const [error, setError]           = useState(null);

  const token = localStorage.getItem('auth_token');
  const headers = { Authorization: `Bearer ${token}` };

  const fetchStatus = useCallback(async (page = 1, limit = 20) => {
    try {
      const res = await fetch(
        `${BASE_URL}/api/v2/email-sync/status?page=${page}&limit=${limit}`,
        { headers }
      );
      const json = await res.json();
      if (json.success) {
        setConnected(json.data.connected);
        setLogs(json.data.logs || []);
        setTotal(json.data.total || 0);
      }
    } catch (e) {
      setError(e.message);
    }
  }, [token]);

  // Check status on mount
  useEffect(() => {
    fetchStatus();
  }, [fetchStatus]);

  // Handle ?email_sync=success / error after OAuth redirect
  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    const result = params.get('email_sync');
    if (result === 'success') {
      fetchStatus();
      // Clean URL
      window.history.replaceState({}, '', window.location.pathname);
    } else if (result === 'error') {
      const reason = params.get('reason') || 'unknown error';
      setError(`Gmail connection failed: ${reason}`);
      window.history.replaceState({}, '', window.location.pathname);
    }
  }, [fetchStatus]);

  const connect = async () => {
    setConnecting(true);
    setError(null);
    try {
      const res = await fetch(`${BASE_URL}/api/v2/email-sync/auth`, { headers });
      const json = await res.json();
      if (json.success) {
        // Open Google login in a new tab
        window.open(json.data.auth_url, '_blank');
      } else {
        setError(json.message);
      }
    } catch (e) {
      setError(e.message);
    } finally {
      setConnecting(false);
    }
  };

  const sync = async () => {
    setSyncing(true);
    setError(null);
    setSyncResult(null);
    try {
      const res = await fetch(`${BASE_URL}/api/v2/email-sync/sync`, {
        method: 'POST',
        headers,
      });
      const json = await res.json();
      if (json.success) {
        setSyncResult(json.data);
        await fetchStatus(); // refresh logs
      } else {
        setError(json.message);
      }
    } catch (e) {
      setError(e.message);
    } finally {
      setSyncing(false);
    }
  };

  const disconnect = async () => {
    try {
      await fetch(`${BASE_URL}/api/v2/email-sync/disconnect`, {
        method: 'DELETE',
        headers,
      });
      setConnected(false);
      setLogs([]);
    } catch (e) {
      setError(e.message);
    }
  };

  return {
    connected, logs, total, syncing, connecting, syncResult, error,
    connect, sync, disconnect, fetchStatus,
  };
}
```

### 3. Email Sync Component (`EmailSyncCard.jsx`)

```jsx
import React from 'react';
import { useEmailSync } from './useEmailSync';

export default function EmailSyncCard() {
  const {
    connected, logs, total, syncing, connecting, syncResult, error,
    connect, sync, disconnect, fetchStatus,
  } = useEmailSync();

  return (
    <div style={{ padding: 24, maxWidth: 640 }}>
      <h2>Gmail Bank Sync</h2>

      {error && (
        <div style={{ color: 'red', marginBottom: 12 }}>{error}</div>
      )}

      {/* Connection status */}
      <div style={{ marginBottom: 16 }}>
        <span style={{ marginRight: 8 }}>
          Status: <strong>{connected ? '🟢 Connected' : '🔴 Not connected'}</strong>
        </span>
        {!connected ? (
          <button onClick={connect} disabled={connecting}>
            {connecting ? 'Opening Google...' : 'Connect Gmail'}
          </button>
        ) : (
          <button onClick={disconnect} style={{ marginLeft: 8 }}>
            Disconnect
          </button>
        )}
      </div>

      {/* Sync button */}
      {connected && (
        <div style={{ marginBottom: 16 }}>
          <button onClick={sync} disabled={syncing}>
            {syncing ? 'Syncing...' : 'Sync Now'}
          </button>
          {syncResult && (
            <span style={{ marginLeft: 12, fontSize: 14, color: '#555' }}>
              ✅ {syncResult.imported} imported · ⏭ {syncResult.skipped} skipped · ❌ {syncResult.failed} failed
            </span>
          )}
        </div>
      )}

      {/* Logs */}
      {logs.length > 0 && (
        <>
          <h3>Sync Logs ({total})</h3>
          <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: 13 }}>
            <thead>
              <tr>
                {['Date', 'Bank', 'Amount', 'Status', 'Description'].map(h => (
                  <th key={h} style={{ textAlign: 'left', borderBottom: '1px solid #ddd', padding: '6px 8px' }}>{h}</th>
                ))}
              </tr>
            </thead>
            <tbody>
              {logs.map(log => (
                <tr key={log.id}>
                  <td style={{ padding: '6px 8px' }}>
                    {log.email_date ? new Date(log.email_date).toLocaleDateString('id-ID') : '—'}
                  </td>
                  <td style={{ padding: '6px 8px' }}>{log.bank_name || '—'}</td>
                  <td style={{ padding: '6px 8px' }}>
                    {log.amount > 0 ? `Rp ${log.amount.toLocaleString('id-ID')}` : '—'}
                  </td>
                  <td style={{ padding: '6px 8px' }}>
                    <StatusBadge status={log.status} />
                  </td>
                  <td style={{ padding: '6px 8px', color: '#888', fontSize: 12 }}>
                    {log.error_message || log.subject}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </>
      )}
    </div>
  );
}

function StatusBadge({ status }) {
  const colors = {
    imported: { bg: '#d1fae5', text: '#065f46' },
    skipped:  { bg: '#fef9c3', text: '#854d0e' },
    failed:   { bg: '#fee2e2', text: '#991b1b' },
  };
  const c = colors[status] || { bg: '#f3f4f6', text: '#374151' };
  return (
    <span style={{
      backgroundColor: c.bg, color: c.text,
      padding: '2px 8px', borderRadius: 4, fontSize: 12, fontWeight: 600,
    }}>
      {status}
    </span>
  );
}
```

### 4. Handle the OAuth redirect

After Google redirects back, the page at `EMAIL_SYNC_REDIRECT_URL` (e.g. `http://localhost:3000/settings`) will receive:
- `?email_sync=success` — connection established
- `?email_sync=error&reason=...` — something went wrong

The hook above already handles this in the `useEffect`. Make sure `EmailSyncCard` is rendered on the page that `EMAIL_SYNC_REDIRECT_URL` points to.

> **`window.open` vs same tab**: The example uses `window.open` (new tab) so the user returns to your app automatically after OAuth. Alternatively you can do `window.location.href = json.data.auth_url` to redirect in the same tab.

---

## Swift (iOS)

### 1. Email Sync Service (`EmailSyncService.swift`)

```swift
import Foundation

struct SyncResult: Decodable {
    let total: Int
    let imported: Int
    let skipped: Int
    let failed: Int
}

struct EmailSyncStatus: Decodable {
    let connected: Bool
    let logs: [EmailSyncLog]
    let total: Int
    let page: Int
    let limit: Int
}

struct EmailSyncLog: Decodable, Identifiable {
    let id: Int
    let gmailMessageId: String
    let subject: String
    let fromEmail: String
    let bankName: String
    let amount: Int
    let assetId: Int
    let transactionId: Int
    let status: String          // "imported" | "skipped" | "failed"
    let errorMessage: String
    let emailDate: String?
    let createdAt: String

    enum CodingKeys: String, CodingKey {
        case id
        case gmailMessageId  = "gmail_message_id"
        case subject, fromEmail = "from_email"
        case bankName        = "bank_name"
        case amount
        case assetId         = "asset_id"
        case transactionId   = "transaction_id"
        case status, errorMessage = "error_message"
        case emailDate       = "email_date"
        case createdAt       = "created_at"
    }
}

struct APIResponse<T: Decodable>: Decodable {
    let success: Bool
    let message: String?
    let data: T?
}

class EmailSyncService {
    static let shared = EmailSyncService()
    private let baseURL = "http://localhost:8080"   // change for production

    private var authToken: String {
        UserDefaults.standard.string(forKey: "auth_token") ?? ""
    }

    private var defaultHeaders: [String: String] {
        ["Authorization": "Bearer \(authToken)", "Content-Type": "application/json"]
    }

    // MARK: - Get OAuth URL
    func getAuthURL(completion: @escaping (Result<String, Error>) -> Void) {
        guard let url = URL(string: "\(baseURL)/api/v2/email-sync/auth") else { return }
        var req = URLRequest(url: url)
        defaultHeaders.forEach { req.setValue($1, forHTTPHeaderField: $0) }

        URLSession.shared.dataTask(with: req) { data, _, error in
            if let error { completion(.failure(error)); return }
            guard let data else { return }
            if let json = try? JSONDecoder().decode(APIResponse<[String: String]>.self, from: data),
               let authURL = json.data?["auth_url"] {
                completion(.success(authURL))
            }
        }.resume()
    }

    // MARK: - Sync Emails
    func syncEmails(completion: @escaping (Result<SyncResult, Error>) -> Void) {
        guard let url = URL(string: "\(baseURL)/api/v2/email-sync/sync") else { return }
        var req = URLRequest(url: url)
        req.httpMethod = "POST"
        defaultHeaders.forEach { req.setValue($1, forHTTPHeaderField: $0) }

        URLSession.shared.dataTask(with: req) { data, _, error in
            if let error { completion(.failure(error)); return }
            guard let data else { return }
            if let json = try? JSONDecoder().decode(APIResponse<SyncResult>.self, from: data),
               let result = json.data {
                completion(.success(result))
            }
        }.resume()
    }

    // MARK: - Get Status
    func getStatus(page: Int = 1, limit: Int = 20,
                   completion: @escaping (Result<EmailSyncStatus, Error>) -> Void) {
        guard let url = URL(string: "\(baseURL)/api/v2/email-sync/status?page=\(page)&limit=\(limit)") else { return }
        var req = URLRequest(url: url)
        defaultHeaders.forEach { req.setValue($1, forHTTPHeaderField: $0) }

        URLSession.shared.dataTask(with: req) { data, _, error in
            if let error { completion(.failure(error)); return }
            guard let data else { return }
            if let json = try? JSONDecoder().decode(APIResponse<EmailSyncStatus>.self, from: data),
               let status = json.data {
                completion(.success(status))
            }
        }.resume()
    }

    // MARK: - Disconnect
    func disconnect(completion: @escaping (Bool) -> Void) {
        guard let url = URL(string: "\(baseURL)/api/v2/email-sync/disconnect") else { return }
        var req = URLRequest(url: url)
        req.httpMethod = "DELETE"
        defaultHeaders.forEach { req.setValue($1, forHTTPHeaderField: $0) }

        URLSession.shared.dataTask(with: req) { _, response, _ in
            let ok = (response as? HTTPURLResponse)?.statusCode == 200
            completion(ok)
        }.resume()
    }
}
```

### 2. Email Sync ViewModel (`EmailSyncViewModel.swift`)

```swift
import Foundation
import UIKit   // for UIApplication.open

@MainActor
class EmailSyncViewModel: ObservableObject {
    @Published var connected   = false
    @Published var logs        = [EmailSyncLog]()
    @Published var total       = 0
    @Published var syncing     = false
    @Published var syncResult: SyncResult? = nil
    @Published var errorMessage: String?   = nil

    private let service = EmailSyncService.shared

    func loadStatus() {
        service.getStatus { [weak self] result in
            DispatchQueue.main.async {
                switch result {
                case .success(let status):
                    self?.connected = status.connected
                    self?.logs      = status.logs
                    self?.total     = status.total
                case .failure(let error):
                    self?.errorMessage = error.localizedDescription
                }
            }
        }
    }

    func connect() {
        service.getAuthURL { [weak self] result in
            DispatchQueue.main.async {
                switch result {
                case .success(let urlString):
                    guard let url = URL(string: urlString) else { return }
                    // Opens Safari (or in-app browser) for Google OAuth
                    UIApplication.shared.open(url)
                case .failure(let error):
                    self?.errorMessage = error.localizedDescription
                }
            }
        }
    }

    func sync() {
        syncing = true
        syncResult = nil
        service.syncEmails { [weak self] result in
            DispatchQueue.main.async {
                self?.syncing = false
                switch result {
                case .success(let res):
                    self?.syncResult = res
                    self?.loadStatus()
                case .failure(let error):
                    self?.errorMessage = error.localizedDescription
                }
            }
        }
    }

    func disconnect() {
        service.disconnect { [weak self] _ in
            DispatchQueue.main.async {
                self?.connected = false
                self?.logs      = []
            }
        }
    }
}
```

### 3. SwiftUI View (`EmailSyncView.swift`)

```swift
import SwiftUI

struct EmailSyncView: View {
    @StateObject private var vm = EmailSyncViewModel()

    var body: some View {
        NavigationView {
            List {
                // Status section
                Section {
                    HStack {
                        Circle()
                            .fill(vm.connected ? Color.green : Color.red)
                            .frame(width: 10, height: 10)
                        Text(vm.connected ? "Gmail Connected" : "Not Connected")
                            .font(.headline)
                        Spacer()
                        if vm.connected {
                            Button("Disconnect", role: .destructive) { vm.disconnect() }
                                .font(.caption)
                        } else {
                            Button("Connect") { vm.connect() }
                                .font(.caption)
                        }
                    }
                }

                // Sync button + result
                if vm.connected {
                    Section {
                        Button {
                            vm.sync()
                        } label: {
                            HStack {
                                if vm.syncing {
                                    ProgressView().padding(.trailing, 4)
                                }
                                Text(vm.syncing ? "Syncing…" : "Sync Now")
                            }
                        }
                        .disabled(vm.syncing)

                        if let r = vm.syncResult {
                            HStack(spacing: 16) {
                                Label("\(r.imported) imported", systemImage: "checkmark.circle.fill")
                                    .foregroundColor(.green).font(.caption)
                                Label("\(r.skipped) skipped", systemImage: "forward.fill")
                                    .foregroundColor(.orange).font(.caption)
                                Label("\(r.failed) failed", systemImage: "xmark.circle.fill")
                                    .foregroundColor(.red).font(.caption)
                            }
                        }
                    }
                }

                // Logs section
                if !vm.logs.isEmpty {
                    Section(header: Text("Sync Logs (\(vm.total))")) {
                        ForEach(vm.logs) { log in
                            EmailSyncLogRow(log: log)
                        }
                    }
                }

                // Error
                if let err = vm.errorMessage {
                    Section {
                        Text(err).foregroundColor(.red).font(.caption)
                    }
                }
            }
            .navigationTitle("Email Sync")
            .onAppear { vm.loadStatus() }
        }
    }
}

struct EmailSyncLogRow: View {
    let log: EmailSyncLog

    var body: some View {
        VStack(alignment: .leading, spacing: 4) {
            HStack {
                Text(log.bankName.isEmpty ? "Unknown" : log.bankName)
                    .font(.subheadline).bold()
                Spacer()
                StatusBadge(status: log.status)
            }
            if log.amount > 0 {
                Text("Rp \(log.amount.formatted())")
                    .font(.subheadline).foregroundColor(.secondary)
            }
            Text(log.subject).font(.caption).foregroundColor(.secondary).lineLimit(1)
            if !log.errorMessage.isEmpty {
                Text(log.errorMessage).font(.caption).foregroundColor(.red).lineLimit(2)
            }
        }
        .padding(.vertical, 4)
    }
}

struct StatusBadge: View {
    let status: String

    var color: Color {
        switch status {
        case "imported": return .green
        case "skipped":  return .orange
        case "failed":   return .red
        default:         return .gray
        }
    }

    var body: some View {
        Text(status)
            .font(.caption2).bold()
            .padding(.horizontal, 8).padding(.vertical, 3)
            .background(color.opacity(0.15))
            .foregroundColor(color)
            .cornerRadius(6)
    }
}
```

### 4. Handle the OAuth redirect (Deep Link)

After Google completes the OAuth flow, the backend redirects to `EMAIL_SYNC_REDIRECT_URL?email_sync=success`. For iOS you have two options:

**Option A — Universal Link / Custom Scheme** (recommended for production)

Set `EMAIL_SYNC_REDIRECT_URL` to your app's deep link, e.g. `myapp://email-sync`.

In `AppDelegate` or `SceneDelegate`:
```swift
func scene(_ scene: UIScene, openURLContexts contexts: Set<UIOpenURLContext>) {
    guard let url = contexts.first?.url else { return }
    if url.scheme == "myapp", url.host == "email-sync" {
        let params = URLComponents(url: url, resolvingAgainstBaseURL: false)?.queryItems
        let result = params?.first(where: { $0.name == "email_sync" })?.value
        if result == "success" {
            NotificationCenter.default.post(name: .emailSyncConnected, object: nil)
        }
    }
}

extension Notification.Name {
    static let emailSyncConnected = Notification.Name("emailSyncConnected")
}
```

Then observe in your ViewModel:
```swift
NotificationCenter.default.addObserver(forName: .emailSyncConnected, object: nil, queue: .main) { _ in
    self.loadStatus()
}
```

**Option B — Safari opens your web frontend URL**

Set `EMAIL_SYNC_REDIRECT_URL` to your React web app URL. The web page shows the result, and the user returns to the app manually. Simpler, but less seamless.

---

## Environment Variables Checklist

| Variable | Example |
|----------|---------|
| `GOOGLE_CLIENT_ID` | `123456.apps.googleusercontent.com` |
| `GOOGLE_CLIENT_SECRET` | `GOCSPX-...` |
| `GOOGLE_REDIRECT_URI` | `http://localhost:8080/api/v2/email-sync/callback` |
| `EMAIL_SYNC_REDIRECT_URL` | `http://localhost:3000` (React) or `myapp://email-sync` (iOS) |

Add `GOOGLE_REDIRECT_URI` to your **Google Cloud Console → OAuth 2.0 Credentials → Authorized redirect URIs**.
