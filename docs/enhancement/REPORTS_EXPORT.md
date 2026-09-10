# Reports & Export Implementation Plan

## Problem Statement

Users need to export their financial data for:
- Tax preparation and documentation
- Sharing with financial advisors or accountants
- Personal record keeping and backup
- Expense reports for reimbursement
- Annual financial reviews

---

## Solution Overview

Comprehensive reporting system that:
1. Generates formatted PDF reports
2. Exports raw data as CSV/Excel
3. Provides customizable date ranges and filters
4. Includes visual charts and summaries
5. Supports scheduled automatic reports

---

## Report Types

### 1. Transaction Report
- Complete list of transactions
- Filterable by date, category, asset, tag
- Includes running balance

### 2. Budget Summary Report
- Budget performance overview
- Category-wise budget vs actual
- Alerts and overspending highlights

### 3. Monthly Financial Statement
- Income vs expense summary
- Category breakdown
- Asset balance changes
- Net worth snapshot

### 4. Annual Report
- Year-over-year comparison
- Monthly trends
- Top spending categories
- Savings rate
- Tax-relevant summaries

### 5. Custom Report
- User-defined metrics and filters
- Flexible date ranges
- Selected categories/assets only

---

## API Endpoints

### Generate Report
```http
POST /api/v2/reports/generate
Authorization: Bearer {token}

{
    "report_type": "monthly_statement",
    "format": "pdf",
    "period": {
        "start_date": "2026-01-01",
        "end_date": "2026-01-31"
    },
    "options": {
        "include_charts": true,
        "include_transaction_list": true,
        "currency_format": "IDR"
    }
}
```

### Response
```json
{
    "success": true,
    "message": "Report generated successfully",
    "data": {
        "report_id": "rpt_abc123",
        "status": "ready",
        "format": "pdf",
        "file_size": 245678,
        "download_url": "/api/v2/reports/download/rpt_abc123",
        "expires_at": "2026-02-12T10:30:00Z"
    }
}
```

### Download Report
```http
GET /api/v2/reports/download/{report_id}
Authorization: Bearer {token}
```

### Export Transactions (CSV)
```http
GET /api/v2/transactions/export?format=csv&start_date=2026-01-01&end_date=2026-01-31&category_id=5
Authorization: Bearer {token}
```

### List Generated Reports
```http
GET /api/v2/reports?page=1&limit=10
Authorization: Bearer {token}
```

### Response
```json
{
    "success": true,
    "data": [
        {
            "id": "rpt_abc123",
            "report_type": "monthly_statement",
            "format": "pdf",
            "period": "January 2026",
            "generated_at": "2026-02-11T09:00:00Z",
            "file_size": 245678,
            "status": "ready"
        },
        {
            "id": "rpt_def456",
            "report_type": "transaction_list",
            "format": "csv",
            "period": "Q4 2025",
            "generated_at": "2026-01-15T14:30:00Z",
            "file_size": 89456,
            "status": "ready"
        }
    ]
}
```

### Schedule Automatic Report
```http
POST /api/v2/reports/schedules
Authorization: Bearer {token}

{
    "report_type": "monthly_statement",
    "format": "pdf",
    "frequency": "monthly",
    "day_of_month": 1,
    "delivery_method": "email",
    "email_address": "user@example.com"
}
```

---

## Database Schema

### Table: `reports`

```sql
CREATE TABLE reports (
    id VARCHAR(20) PRIMARY KEY COMMENT 'Format: rpt_xxxxx',
    user_id BIGINT UNSIGNED NOT NULL,
    report_type ENUM('transaction_list', 'budget_summary', 'monthly_statement', 'annual_report', 'custom') NOT NULL,
    format ENUM('pdf', 'csv', 'xlsx') NOT NULL,
    
    -- Period
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    
    -- Options stored as JSON
    options JSON NULL,
    
    -- File info
    file_path VARCHAR(255) NULL,
    file_size INT NULL,
    
    -- Status
    status ENUM('pending', 'processing', 'ready', 'failed', 'expired') DEFAULT 'pending',
    error_message TEXT NULL,
    
    -- Timestamps
    generated_at TIMESTAMP NULL,
    expires_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_user_status (user_id, status),
    INDEX idx_expires (expires_at),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
```

### Table: `report_schedules`

```sql
CREATE TABLE report_schedules (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    report_type ENUM('transaction_list', 'budget_summary', 'monthly_statement', 'annual_report') NOT NULL,
    format ENUM('pdf', 'csv', 'xlsx') DEFAULT 'pdf',
    
    frequency ENUM('weekly', 'monthly', 'quarterly', 'yearly') NOT NULL,
    day_of_week INT NULL COMMENT '0-6 for weekly',
    day_of_month INT NULL COMMENT '1-31 for monthly',
    
    delivery_method ENUM('email', 'in_app') DEFAULT 'in_app',
    email_address VARCHAR(255) NULL,
    
    options JSON NULL,
    is_active BOOLEAN DEFAULT true,
    last_generated_at TIMESTAMP NULL,
    next_run_at TIMESTAMP NULL,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
```

---

## Go Implementation

### Service

```go
// services/report_service.go

package services

import (
    "bytes"
    "encoding/csv"
    "fmt"
    "my-api/models"
    "my-api/repositories"
    "time"
    
    "github.com/jung-kurt/gofpdf"
)

type ReportService interface {
    GenerateReport(userID uint, req *GenerateReportRequest) (*ReportResponse, error)
    GetReport(reportID string, userID uint) (*models.Report, error)
    ListReports(userID uint, page, limit int) ([]models.Report, int64, error)
    DownloadReport(reportID string, userID uint) ([]byte, string, error)
    ExportTransactions(userID uint, req *ExportRequest) ([]byte, string, error)
    CreateSchedule(userID uint, req *ReportScheduleRequest) (*models.ReportSchedule, error)
    ProcessScheduledReports() error
}

type reportService struct {
    repo            repositories.ReportRepository
    transactionRepo repositories.TransactionV2Repository
    budgetRepo      repositories.BudgetRepository
    assetRepo       repositories.AssetRepository
}

func (s *reportService) GenerateReport(userID uint, req *GenerateReportRequest) (*ReportResponse, error) {
    reportID := fmt.Sprintf("rpt_%s", generateRandomString(12))
    
    report := &models.Report{
        ID:         reportID,
        UserID:     userID,
        ReportType: req.ReportType,
        Format:     req.Format,
        StartDate:  req.Period.StartDate,
        EndDate:    req.Period.EndDate,
        Options:    req.Options,
        Status:     "processing",
    }
    
    if err := s.repo.Create(report); err != nil {
        return nil, err
    }
    
    // Generate report asynchronously
    go s.processReport(report)
    
    return &ReportResponse{
        ReportID:  reportID,
        Status:    "processing",
        Format:    req.Format,
        ExpiresAt: time.Now().Add(24 * time.Hour),
    }, nil
}

func (s *reportService) processReport(report *models.Report) {
    var content []byte
    var err error
    
    switch report.ReportType {
    case "transaction_list":
        content, err = s.generateTransactionListReport(report)
    case "monthly_statement":
        content, err = s.generateMonthlyStatementReport(report)
    case "budget_summary":
        content, err = s.generateBudgetSummaryReport(report)
    case "annual_report":
        content, err = s.generateAnnualReport(report)
    }
    
    if err != nil {
        report.Status = "failed"
        report.ErrorMessage = err.Error()
        s.repo.Update(report)
        return
    }
    
    // Save file
    filePath := fmt.Sprintf("reports/%s.%s", report.ID, report.Format)
    // Save to storage (local filesystem or cloud storage)
    
    now := time.Now()
    report.Status = "ready"
    report.FilePath = filePath
    report.FileSize = len(content)
    report.GeneratedAt = &now
    report.ExpiresAt = timePtr(now.Add(7 * 24 * time.Hour))
    
    s.repo.Update(report)
}

func (s *reportService) generateMonthlyStatementReport(report *models.Report) ([]byte, error) {
    userID := report.UserID
    startDate := report.StartDate
    endDate := report.EndDate
    
    // Gather data
    transactions, _ := s.transactionRepo.GetByDateRange(userID, startDate, endDate, nil, nil, nil)
    
    totalIncome := 0
    totalExpense := 0
    categoryBreakdown := make(map[string]int)
    
    for _, tx := range transactions {
        if tx.TransactionType == 1 {
            totalIncome += tx.Amount
        } else {
            totalExpense += tx.Amount
            categoryBreakdown[tx.Category.CategoryName] += tx.Amount
        }
    }
    
    netCashFlow := totalIncome - totalExpense
    
    // Get asset balances
    assets, _ := s.assetRepo.GetByUserID(userID)
    totalBalance := int64(0)
    for _, a := range assets {
        totalBalance += a.Balance
    }
    
    if report.Format == "pdf" {
        return s.generateMonthlyStatementPDF(report, MonthlyStatementData{
            Period:            fmt.Sprintf("%s - %s", startDate.Format("Jan 2, 2006"), endDate.Format("Jan 2, 2006")),
            TotalIncome:       totalIncome,
            TotalExpense:      totalExpense,
            NetCashFlow:       netCashFlow,
            CategoryBreakdown: categoryBreakdown,
            TotalBalance:      int(totalBalance),
            Transactions:      transactions,
        })
    }
    
    return s.generateMonthlyStatementCSV(report, transactions)
}

func (s *reportService) generateMonthlyStatementPDF(report *models.Report, data MonthlyStatementData) ([]byte, error) {
    pdf := gofpdf.New("P", "mm", "A4", "")
    pdf.AddPage()
    
    // Header
    pdf.SetFont("Arial", "B", 20)
    pdf.Cell(0, 10, "Monthly Financial Statement")
    pdf.Ln(12)
    
    pdf.SetFont("Arial", "", 12)
    pdf.Cell(0, 8, data.Period)
    pdf.Ln(15)
    
    // Summary Box
    pdf.SetFillColor(240, 240, 240)
    pdf.SetFont("Arial", "B", 14)
    pdf.Cell(0, 10, "Summary", "0", 1, "L", true)
    pdf.Ln(5)
    
    pdf.SetFont("Arial", "", 12)
    pdf.Cell(60, 8, "Total Income:", "", 0, "L", false)
    pdf.SetTextColor(0, 128, 0)
    pdf.Cell(0, 8, fmt.Sprintf("Rp %s", formatNumber(data.TotalIncome)), "", 1, "L", false)
    
    pdf.SetTextColor(0, 0, 0)
    pdf.Cell(60, 8, "Total Expenses:", "", 0, "L", false)
    pdf.SetTextColor(255, 0, 0)
    pdf.Cell(0, 8, fmt.Sprintf("Rp %s", formatNumber(data.TotalExpense)), "", 1, "L", false)
    
    pdf.SetTextColor(0, 0, 0)
    pdf.Cell(60, 8, "Net Cash Flow:", "", 0, "L", false)
    if data.NetCashFlow >= 0 {
        pdf.SetTextColor(0, 128, 0)
        pdf.Cell(0, 8, fmt.Sprintf("+Rp %s", formatNumber(data.NetCashFlow)), "", 1, "L", false)
    } else {
        pdf.SetTextColor(255, 0, 0)
        pdf.Cell(0, 8, fmt.Sprintf("-Rp %s", formatNumber(-data.NetCashFlow)), "", 1, "L", false)
    }
    
    pdf.SetTextColor(0, 0, 0)
    pdf.Ln(10)
    
    // Category Breakdown
    pdf.SetFont("Arial", "B", 14)
    pdf.Cell(0, 10, "Spending by Category", "", 1, "L", false)
    pdf.Ln(5)
    
    pdf.SetFont("Arial", "", 11)
    for category, amount := range data.CategoryBreakdown {
        percentage := float64(amount) / float64(data.TotalExpense) * 100
        pdf.Cell(80, 7, category, "", 0, "L", false)
        pdf.Cell(50, 7, fmt.Sprintf("Rp %s", formatNumber(amount)), "", 0, "R", false)
        pdf.Cell(0, 7, fmt.Sprintf("(%.1f%%)", percentage), "", 1, "L", false)
    }
    
    pdf.Ln(10)
    
    // Transaction Table (if included)
    if len(data.Transactions) > 0 && len(data.Transactions) <= 50 {
        pdf.SetFont("Arial", "B", 14)
        pdf.Cell(0, 10, "Transactions", "", 1, "L", false)
        pdf.Ln(3)
        
        // Table header
        pdf.SetFont("Arial", "B", 10)
        pdf.SetFillColor(220, 220, 220)
        pdf.Cell(25, 8, "Date", "1", 0, "C", true)
        pdf.Cell(70, 8, "Description", "1", 0, "C", true)
        pdf.Cell(35, 8, "Category", "1", 0, "C", true)
        pdf.Cell(35, 8, "Amount", "1", 1, "C", true)
        
        // Table rows
        pdf.SetFont("Arial", "", 9)
        for _, tx := range data.Transactions {
            pdf.Cell(25, 7, tx.Date.Time.Format("Jan 02"), "1", 0, "C", false)
            pdf.Cell(70, 7, truncateString(tx.Description, 35), "1", 0, "L", false)
            pdf.Cell(35, 7, truncateString(tx.Category.CategoryName, 15), "1", 0, "C", false)
            
            if tx.TransactionType == 1 {
                pdf.SetTextColor(0, 128, 0)
                pdf.Cell(35, 7, fmt.Sprintf("+%s", formatNumber(tx.Amount)), "1", 1, "R", false)
            } else {
                pdf.SetTextColor(255, 0, 0)
                pdf.Cell(35, 7, fmt.Sprintf("-%s", formatNumber(tx.Amount)), "1", 1, "R", false)
            }
            pdf.SetTextColor(0, 0, 0)
        }
    }
    
    // Footer
    pdf.SetY(-20)
    pdf.SetFont("Arial", "I", 8)
    pdf.Cell(0, 10, fmt.Sprintf("Generated on %s | Money Manage", time.Now().Format("Jan 2, 2006 3:04 PM")), "", 0, "C", false)
    
    var buf bytes.Buffer
    err := pdf.Output(&buf)
    return buf.Bytes(), err
}

func (s *reportService) ExportTransactions(userID uint, req *ExportRequest) ([]byte, string, error) {
    transactions, _ := s.transactionRepo.GetByDateRange(
        userID, 
        req.StartDate, 
        req.EndDate,
        req.CategoryID,
        req.AssetID,
        req.TransactionType,
    )
    
    var buf bytes.Buffer
    
    if req.Format == "csv" {
        writer := csv.NewWriter(&buf)
        
        // Header
        writer.Write([]string{"Date", "Description", "Category", "Asset", "Type", "Amount"})
        
        // Rows
        for _, tx := range transactions {
            txType := "Income"
            if tx.TransactionType == 2 {
                txType = "Expense"
            }
            
            writer.Write([]string{
                tx.Date.Time.Format("2006-01-02"),
                tx.Description,
                tx.Category.CategoryName,
                tx.Asset.Name,
                txType,
                fmt.Sprintf("%d", tx.Amount),
            })
        }
        
        writer.Flush()
        return buf.Bytes(), "transactions.csv", nil
    }
    
    // Handle Excel format with excelize library
    // ...
    
    return buf.Bytes(), "transactions.csv", nil
}
```

---

## Frontend Integration Guide

### Reports Screen

```
┌─────────────────────────────────────────┐
│  📊 Reports & Export                    │
├─────────────────────────────────────────┤
│                                         │
│  ┌─────────────────────────────────┐   │
│  │ 📅 January 2026                 │   │
│  │                              ▼  │   │
│  └─────────────────────────────────┘   │
│                                         │
│  GENERATE REPORT                        │
│                                         │
│  ┌─────────────────────────────────┐   │
│  │ 📋 Monthly Statement            │   │
│  │ Income, expenses & balance      │   │
│  │ [ PDF ] [ CSV ]           →     │   │
│  └─────────────────────────────────┘   │
│                                         │
│  ┌─────────────────────────────────┐   │
│  │ 📑 Transaction List             │   │
│  │ All transactions with filters   │   │
│  │ [ PDF ] [ CSV ] [ Excel ] →     │   │
│  └─────────────────────────────────┘   │
│                                         │
│  ┌─────────────────────────────────┐   │
│  │ 💰 Budget Summary               │   │
│  │ Budget vs actual spending       │   │
│  │ [ PDF ]                   →     │   │
│  └─────────────────────────────────┘   │
│                                         │
│  ┌─────────────────────────────────┐   │
│  │ 📈 Annual Report                │   │
│  │ Full year financial overview    │   │
│  │ [ PDF ]                   →     │   │
│  └─────────────────────────────────┘   │
│                                         │
│  ─────────────────────────────────────  │
│                                         │
│  RECENT REPORTS                         │
│                                         │
│  ┌─────────────────────────────────┐   │
│  │ 📄 Monthly Statement            │   │
│  │ January 2026 • PDF • 245 KB     │   │
│  │ Generated Feb 1, 2026           │   │
│  │ Expires Feb 8, 2026     [⬇️]    │   │
│  └─────────────────────────────────┘   │
│                                         │
│  ─────────────────────────────────────  │
│                                         │
│  ⚙️ SCHEDULED REPORTS                   │
│                                         │
│  Monthly Statement on 1st each month    │
│  [ Manage Schedules ]                   │
└─────────────────────────────────────────┘
```

### Report Generation Modal

```
┌─────────────────────────────────────────┐
│  Generate Monthly Statement         ✕   │
├─────────────────────────────────────────┤
│                                         │
│  Period                                 │
│  ┌─────────────────────────────────┐   │
│  │ Custom Range                 ▼  │   │
│  └─────────────────────────────────┘   │
│                                         │
│  ┌────────────┐  ┌────────────┐        │
│  │ Jan 1      │  │ Jan 31     │        │
│  │ 2026    📅 │  │ 2026    📅 │        │
│  └────────────┘  └────────────┘        │
│                                         │
│  Format                                 │
│  ┌────────┐ ┌────────┐ ┌────────┐      │
│  │ PDF ✓ │ │  CSV   │ │ Excel  │      │
│  └────────┘ └────────┘ └────────┘      │
│                                         │
│  Options                                │
│  ☑️ Include charts and graphs           │
│  ☑️ Include transaction list            │
│  ☐ Include asset details                │
│                                         │
│  ─────────────────────────────────────  │
│                                         │
│  ┌─────────────────────────────────┐   │
│  │         Generate Report         │   │
│  └─────────────────────────────────┘   │
└─────────────────────────────────────────┘
```

### Export Success Screen

```
┌─────────────────────────────────────────┐
│  Report Generated!                  ✕   │
├─────────────────────────────────────────┤
│                                         │
│                  ✅                      │
│                                         │
│     Monthly Statement                   │
│     January 2026                        │
│                                         │
│     📄 monthly_statement_jan_2026.pdf   │
│     245 KB                              │
│                                         │
│  ┌─────────────────────────────────┐   │
│  │          ⬇️ Download            │   │
│  └─────────────────────────────────┘   │
│                                         │
│  ┌─────────────────────────────────┐   │
│  │          📧 Email               │   │
│  └─────────────────────────────────┘   │
│                                         │
│  Link expires in 7 days                 │
└─────────────────────────────────────────┘
```

---

## PDF Report Design

### Monthly Statement Layout

```
┌─────────────────────────────────────────────────────────────┐
│                                                             │
│                  MONTHLY FINANCIAL STATEMENT                │
│                     January 1-31, 2026                      │
│                                                             │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  SUMMARY                                                    │
│  ───────────────────────────────────────────────────────    │
│                                                             │
│  Total Income:        Rp 10,000,000                         │
│  Total Expenses:      Rp  8,500,000                         │
│  Net Cash Flow:      +Rp  1,500,000                         │
│                                                             │
│  Total Balance:       Rp 25,000,000                         │
│                                                             │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  SPENDING BY CATEGORY                                       │
│  ───────────────────────────────────────────────────────    │
│                                                             │
│  Food & Dining    ████████████░░░░  Rp 2,500,000  (29.4%)  │
│  Transportation   ████████░░░░░░░░  Rp 1,800,000  (21.2%)  │
│  Utilities        ██████░░░░░░░░░░  Rp 1,200,000  (14.1%)  │
│  Shopping         █████░░░░░░░░░░░  Rp 1,000,000  (11.8%)  │
│  Entertainment    ████░░░░░░░░░░░░  Rp   800,000   (9.4%)  │
│  Other            ████░░░░░░░░░░░░  Rp 1,200,000  (14.1%)  │
│                                                             │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  BUDGET STATUS                                              │
│  ───────────────────────────────────────────────────────    │
│                                                             │
│  Category          Budget      Actual     Status            │
│  ─────────────────────────────────────────────────────      │
│  Food & Dining     2,000,000   2,500,000  ⚠️ Over 25%      │
│  Transportation    2,000,000   1,800,000  ✅ Under          │
│  Entertainment     1,000,000     800,000  ✅ Under          │
│                                                             │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  TRANSACTIONS                                               │
│  ───────────────────────────────────────────────────────    │
│                                                             │
│  Date      Description              Category        Amount  │
│  ─────────────────────────────────────────────────────      │
│  Jan 02    Monthly Salary           Salary      +10,000,000 │
│  Jan 05    Rent Payment             Housing      -3,500,000 │
│  Jan 07    Supermarket Shopping     Groceries      -450,000 │
│  Jan 10    Electric Bill            Utilities      -350,000 │
│  ...                                                        │
│                                                             │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Generated on February 11, 2026 | Money Manage              │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

## Dependencies

```go
// go.mod additions

require (
    github.com/jung-kurt/gofpdf v1.16.2  // PDF generation
    github.com/xuri/excelize/v2 v2.7.0   // Excel generation
)
```

---

## Background Jobs

```go
// jobs/report_jobs.go

// ProcessScheduledReports runs hourly
func (j *ReportJobs) ProcessScheduledReports() {
    schedules, _ := j.repo.FindDueSchedules(time.Now())
    
    for _, schedule := range schedules {
        // Determine period based on frequency
        var startDate, endDate time.Time
        
        switch schedule.Frequency {
        case "monthly":
            lastMonth := time.Now().AddDate(0, -1, 0)
            startDate = time.Date(lastMonth.Year(), lastMonth.Month(), 1, 0, 0, 0, 0, time.Local)
            endDate = startDate.AddDate(0, 1, -1)
        case "weekly":
            endDate = time.Now().AddDate(0, 0, -1)
            startDate = endDate.AddDate(0, 0, -6)
        // ... other frequencies
        }
        
        req := &GenerateReportRequest{
            ReportType: schedule.ReportType,
            Format:     schedule.Format,
            Period: Period{
                StartDate: startDate,
                EndDate:   endDate,
            },
            Options: schedule.Options,
        }
        
        report, err := j.reportService.GenerateReport(schedule.UserID, req)
        if err != nil {
            continue
        }
        
        // Send via configured delivery method
        if schedule.DeliveryMethod == "email" {
            j.emailService.SendReportEmail(schedule.EmailAddress, report)
        }
        
        // Update next run time
        j.repo.UpdateNextRunTime(schedule.ID)
    }
}
```

---

## Security Considerations

1. **Temporary file storage**: Reports expire after 7 days
2. **User isolation**: Users can only access their own reports
3. **Download authentication**: Report downloads require valid auth token
4. **No sensitive data in URLs**: Use report ID, not querystring params
