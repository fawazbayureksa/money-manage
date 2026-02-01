# Pay Cycle Feature - Frontend Implementation Plan

## 📋 Table of Contents
1. [Feature Overview](#feature-overview)
2. [What's New vs What's Enhanced](#whats-new-vs-whats-enhanced)
3. [API Endpoints Summary](#api-endpoints-summary)
4. [CRUD Implementation - Pay Cycle Settings](#crud-implementation---pay-cycle-settings)
5. [Analytics Enhancement](#analytics-enhancement)
6. [UI/UX Design Guidelines](#uiux-design-guidelines)
7. [Mobile & Web Implementation](#mobile--web-implementation)
8. [Type Definitions](#type-definitions)
9. [Implementation Checklist](#implementation-checklist)

---

## Feature Overview

The Pay Cycle feature allows users to configure custom financial periods that align with their actual income schedule, rather than standard calendar months.

### 🎯 Problem Solved
| Before | After |
|--------|-------|
| User gets paid Jan 30, but Jan 31 expense counts as "January" | User can set pay cycle, Jan 31 expense now counts as "February" period |
| Analytics always use calendar month | Analytics can use user's actual financial period |
| No personalization for financial tracking | Fully customizable pay cycle with 4 types |

---

## What's New vs What's Enhanced

### 🆕 NEW Features (Requires New UI/Pages)

| Feature | Description | Priority |
|---------|-------------|----------|
| **Pay Cycle Settings Page** | CRUD interface for managing pay cycle configuration | High |
| **Settings Form** | Form to create/edit pay cycle type, pay day, offset | High |
| **Period Preview** | Visual preview of calculated financial periods | Medium |

### 🔄 ENHANCED Features (Update Existing UI)

| Feature | Change Required | Priority |
|---------|-----------------|----------|
| **Analytics Trend Chart** | Add toggle/switch for "Use Pay Cycle" | High |
| **Analytics Period Display** | Show `period_start` and `period_end` dates | Medium |
| **Dashboard Summary** | Optional: Show current financial period info | Low |

---

## API Endpoints Summary

### User Settings (Pay Cycle) - NEW CRUD

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/user/settings` | Get user's pay cycle settings |
| `POST` | `/api/user/settings` | Create pay cycle settings |
| `PUT` | `/api/user/settings` | Update pay cycle settings |
| `DELETE` | `/api/user/settings` | Delete/reset to calendar |

### Analytics - ENHANCED

| Method | Endpoint | New Parameter |
|--------|----------|---------------|
| `GET` | `/api/analytics/trend` | `use_pay_cycle=true` |

---

## CRUD Implementation - Pay Cycle Settings

### Type Definitions (TypeScript)

```typescript
// types/userSettings.ts

export type PayCycleType = 'calendar' | 'last_weekday' | 'custom_day' | 'bi_weekly';

export interface UserSettings {
  id: number;
  user_id: number;
  pay_cycle_type: PayCycleType;
  pay_day: number | null;        // Day of month (1-31) or day of week (0-6)
  cycle_start_offset: number;    // Days after payday to start counting
  created_at: string;
  updated_at: string | null;
}

export interface CreateUserSettingsRequest {
  pay_cycle_type: PayCycleType;
  pay_day?: number | null;
  cycle_start_offset?: number;   // Default: 1
}

export interface UpdateUserSettingsRequest {
  pay_cycle_type?: PayCycleType;
  pay_day?: number | null;
  cycle_start_offset?: number;
}

// Pay Cycle Type Options for UI
export const PAY_CYCLE_OPTIONS = [
  {
    value: 'calendar',
    label: 'Calendar Month',
    description: 'Standard calendar months (Jan 1-31, Feb 1-28, etc.)',
    requiresPayDay: false,
    icon: '📅'
  },
  {
    value: 'last_weekday',
    label: 'Last Weekday of Month',
    description: 'Last working day (Mon-Fri) of each month',
    requiresPayDay: false,
    icon: '💼'
  },
  {
    value: 'custom_day',
    label: 'Specific Day of Month',
    description: 'Set a specific day (e.g., 25th of every month)',
    requiresPayDay: true,
    payDayRange: { min: 1, max: 31 },
    payDayLabel: 'Pay Day (1-31)',
    icon: '📆'
  },
  {
    value: 'bi_weekly',
    label: 'Bi-Weekly',
    description: 'Every 2 weeks on a specific day',
    requiresPayDay: true,
    payDayRange: { min: 0, max: 6 },
    payDayLabel: 'Pay Day (Day of Week)',
    payDayOptions: [
      { value: 0, label: 'Sunday' },
      { value: 1, label: 'Monday' },
      { value: 2, label: 'Tuesday' },
      { value: 3, label: 'Wednesday' },
      { value: 4, label: 'Thursday' },
      { value: 5, label: 'Friday' },
      { value: 6, label: 'Saturday' }
    ],
    icon: '🔄'
  }
];

export const OFFSET_OPTIONS = [
  { value: 0, label: 'Same day as payday', description: 'Period starts on payday' },
  { value: 1, label: '1 day after payday', description: 'Period starts next day (Recommended)' },
  { value: 2, label: '2 days after payday', description: 'Period starts 2 days after' }
];

Implementation Checklist
Phase 1: Pay Cycle Settings CRUD (Priority: High)
 Types & Interfaces

 Create UserSettings type definition
 Create request/response interfaces
 Define pay cycle constants
 API Service

 Implement getSettings()
 Implement createSettings()
 Implement updateSettings()
 Implement resetSettings()
 UI Components

 Create Pay Cycle Settings page/screen
 Create radio card component for pay cycle type
 Create conditional pay day input
 Create offset dropdown
 Create period preview component
 State Management

 Add user settings query
 Add mutation for create/update/delete
 Handle loading and error states
Phase 2: Analytics Enhancement (Priority: High)
 API Updates

 Add use_pay_cycle parameter to trend API call
 UI Updates

 Add toggle switch to analytics page
 Show period start/end dates when pay cycle enabled
 Update chart labels for pay cycle periods
Phase 3: Polish & UX (Priority: Medium)
 Mobile Optimization

 Responsive layout for settings page
 Bottom sheet for quick toggle
 Haptic feedback
 Web Optimization

 Keyboard navigation
 Tooltips and help text
 Settings in user profile section
 Testing

 Unit tests for API service
 Integration tests for settings CRUD
 E2E tests for analytics toggle