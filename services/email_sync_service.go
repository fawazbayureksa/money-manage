package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"gorm.io/gorm"
	"my-api/models"
	"my-api/repositories"
	"my-api/utils"
)

// SyncResult summarises a single sync run.
type SyncResult struct {
	Total    int `json:"total"`
	Imported int `json:"imported"`
	Skipped  int `json:"skipped"`
	Failed   int `json:"failed"`
}

// EmailSyncService handles Gmail OAuth2 and transaction importing.
type EmailSyncService interface {
	GetAuthURL(userID uint) (string, error)
	HandleCallback(code, state string) error
	SyncEmails(userID uint) (*SyncResult, error)
	IsConnected(userID uint) bool
	Disconnect(userID uint) error
	GetLogs(userID uint, page, limit int) ([]models.EmailSyncLog, int64, error)
}

type emailSyncService struct {
	emailSyncRepo   repositories.EmailSyncRepository
	assetRepo       *repositories.AssetRepository
	transactionRepo repositories.TransactionV2Repository
	db              *gorm.DB
}

func NewEmailSyncService(
	emailSyncRepo repositories.EmailSyncRepository,
	assetRepo *repositories.AssetRepository,
	transactionRepo repositories.TransactionV2Repository,
	db *gorm.DB,
) EmailSyncService {
	return &emailSyncService{
		emailSyncRepo:   emailSyncRepo,
		assetRepo:       assetRepo,
		transactionRepo: transactionRepo,
		db:              db,
	}
}

// oauthConfig returns the Gmail OAuth2 configuration loaded from env vars.
func oauthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URI"),
		Scopes:       []string{"https://www.googleapis.com/auth/gmail.readonly"},
		Endpoint:     google.Endpoint,
	}
}

// ----- Auth flow -----

func (s *emailSyncService) GetAuthURL(userID uint) (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	state := base64.URLEncoding.EncodeToString(raw)

	// Persist state so we can look it up in the callback
	record := &models.EmailOAuthState{
		UserID:    userID,
		State:     state,
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}
	if err := s.emailSyncRepo.SaveState(record); err != nil {
		return "", err
	}

	// Clean up any stale states while we are here
	_ = s.emailSyncRepo.DeleteExpiredStates()

	url := oauthConfig().AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
	return url, nil
}

func (s *emailSyncService) HandleCallback(code, state string) error {
	record, err := s.emailSyncRepo.FindAndDeleteState(state)
	if err != nil {
		return errors.New("invalid or expired OAuth state")
	}

	token, err := oauthConfig().Exchange(context.Background(), code)
	if err != nil {
		return fmt.Errorf("token exchange failed: %w", err)
	}

	existing, dbErr := s.emailSyncRepo.GetToken(record.UserID)
	if dbErr != nil {
		// No existing token – create a new one
		existing = &models.EmailSyncToken{UserID: record.UserID}
	}

	existing.AccessToken = token.AccessToken
	existing.RefreshToken = token.RefreshToken
	existing.TokenType = token.TokenType
	if !token.Expiry.IsZero() {
		expiry := token.Expiry
		existing.Expiry = &expiry
	}

	return s.emailSyncRepo.SaveToken(existing)
}

func (s *emailSyncService) IsConnected(userID uint) bool {
	_, err := s.emailSyncRepo.GetToken(userID)
	return err == nil
}

func (s *emailSyncService) Disconnect(userID uint) error {
	return s.emailSyncRepo.DeleteToken(userID)
}

func (s *emailSyncService) GetLogs(userID uint, page, limit int) ([]models.EmailSyncLog, int64, error) {
	offset := (page - 1) * limit
	return s.emailSyncRepo.GetLogs(userID, limit, offset)
}

// ----- Sync -----

func (s *emailSyncService) SyncEmails(userID uint) (*SyncResult, error) {
	tokenRecord, err := s.emailSyncRepo.GetToken(userID)
	if err != nil {
		return nil, errors.New("gmail not connected; please connect via /api/v2/email-sync/auth")
	}

	oauthTok := &oauth2.Token{
		AccessToken:  tokenRecord.AccessToken,
		RefreshToken: tokenRecord.RefreshToken,
		TokenType:    tokenRecord.TokenType,
	}
	if tokenRecord.Expiry != nil {
		oauthTok.Expiry = *tokenRecord.Expiry
	}

	cfg := oauthConfig()
	// TokenSource automatically refreshes the access token if expired.
	ts := cfg.TokenSource(context.Background(), oauthTok)
	httpClient := oauth2.NewClient(context.Background(), ts)

	// Persist a refreshed token if it changed
	refreshed, _ := ts.Token()
	if refreshed != nil && refreshed.AccessToken != tokenRecord.AccessToken {
		tokenRecord.AccessToken = refreshed.AccessToken
		if !refreshed.Expiry.IsZero() {
			tokenRecord.Expiry = &refreshed.Expiry
		}
		_ = s.emailSyncRepo.SaveToken(tokenRecord)
	}

	// Fetch user's assets for account-number matching
	assets, err := s.assetRepo.GetAssetsByUser(uint64(userID))
	if err != nil {
		assets = []models.Asset{}
	}

	result := &SyncResult{}

	messageIDs, err := listGmailMessages(httpClient, knownBankQuery, 300)
	if err != nil {
		return nil, fmt.Errorf("failed to list gmail messages: %w", err)
	}

	result.Total = len(messageIDs)

	for _, msgID := range messageIDs {
		// Skip already-synced messages
		if s.emailSyncRepo.IsMessageSynced(userID, msgID) {
			result.Skipped++
			continue
		}

		msg, err := getGmailMessage(httpClient, msgID)
		if err != nil {
			utils.LogErrorf("email-sync: failed to fetch message %s: %v", msgID, err)
			s.saveFailLog(userID, msgID, "", "", err.Error())
			result.Failed++
			continue
		}

		subject := msg.header("Subject")
		fromEmail := msg.header("From")
		emailDate := msg.parseDate()

		body := msg.textBody()

		parsed, parseErr := parseEmailTransaction(subject, fromEmail, body)
		if parseErr != nil {
			s.saveLog(userID, msgID, subject, fromEmail, "", 0, 0, 0, "skipped", parseErr.Error(), emailDate)
			result.Skipped++
			continue
		}

		// Match asset by account suffix
		assetID := matchAsset(assets, parsed.accountSuffix)

		txID, txErr := s.createTransaction(userID, parsed, assetID)
		if txErr != nil {
			utils.LogErrorf("email-sync: failed to create transaction for msg %s: %v", msgID, txErr)
			s.saveLog(userID, msgID, subject, fromEmail, parsed.bankName, parsed.amount, assetID, 0, "failed", txErr.Error(), emailDate)
			result.Failed++
			continue
		}

		s.saveLog(userID, msgID, subject, fromEmail, parsed.bankName, parsed.amount, assetID, txID, "imported", "", emailDate)
		result.Imported++
	}

	return result, nil
}

// createTransaction inserts a new transaction (with or without balance update).
func (s *emailSyncService) createTransaction(userID uint, p *parsedEmail, assetID uint64) (uint, error) {
	tx := &models.TransactionV2{
		UserID:          userID,
		Description:     p.description,
		CategoryID:      nil,
		AssetID:         assetID,
		Amount:          p.amount,
		TransactionType: 2, // expenses by default
		Date:            utils.CustomTime{Time: p.date},
		BankID:          nil,
	}

	if assetID > 0 {
		if err := s.transactionRepo.CreateWithBalanceUpdate(tx); err != nil {
			// Fall back to creating without balance update if the asset has issues
			// (e.g. insufficient balance). Log this so the user is aware.
			utils.LogErrorf("email-sync: balance update failed for asset %d, creating transaction without balance sync: %v", assetID, err)
			tx.AssetID = 0
			if err2 := s.db.Create(tx).Error; err2 != nil {
				return 0, err2
			}
		}
	} else {
		if err := s.db.Create(tx).Error; err != nil {
			return 0, err
		}
	}

	return tx.ID, nil
}

func (s *emailSyncService) saveLog(userID uint, gmailMsgID, subject, fromEmail, bankName string, amount int, assetID uint64, txID uint, status, errMsg string, emailDate *time.Time) {
	log := &models.EmailSyncLog{
		UserID:         userID,
		GmailMessageID: gmailMsgID,
		Subject:        subject,
		FromEmail:      fromEmail,
		BankName:       bankName,
		Amount:         amount,
		AssetID:        assetID,
		TransactionID:  txID,
		Status:         status,
		ErrorMessage:   errMsg,
		EmailDate:      emailDate,
	}
	if err := s.emailSyncRepo.SaveLog(log); err != nil {
		utils.LogErrorf("email-sync: failed to save log for msg %s: %v", gmailMsgID, err)
	}
}

func (s *emailSyncService) saveFailLog(userID uint, gmailMsgID, subject, fromEmail, errMsg string) {
	s.saveLog(userID, gmailMsgID, subject, fromEmail, "", 0, 0, 0, "failed", errMsg, nil)
}

// ----- Gmail REST API helpers -----

type gmailMessageRef struct {
	ID string `json:"id"`
}

type gmailListResponse struct {
	Messages          []gmailMessageRef `json:"messages"`
	NextPageToken     string            `json:"nextPageToken"`
	ResultSizeEstimate int              `json:"resultSizeEstimate"`
}

type gmailMessagePart struct {
	MimeType string             `json:"mimeType"`
	Headers  []gmailHeader      `json:"headers"`
	Body     gmailBody          `json:"body"`
	Parts    []gmailMessagePart `json:"parts"`
}

type gmailHeader struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type gmailBody struct {
	Data string `json:"data"`
	Size int    `json:"size"`
}

type gmailMessage struct {
	ID      string           `json:"id"`
	Payload gmailMessagePart `json:"payload"`
}

func (m *gmailMessage) header(name string) string {
	for _, h := range m.Payload.Headers {
		if strings.EqualFold(h.Name, name) {
			return h.Value
		}
	}
	return ""
}

func (m *gmailMessage) parseDate() *time.Time {
	raw := m.header("Date")
	if raw == "" {
		return nil
	}
	formats := []string{
		time.RFC1123Z,
		"Mon, 2 Jan 2006 15:04:05 -0700",
		"Mon, 02 Jan 2006 15:04:05 -0700 (MST)",
		"2 Jan 2006 15:04:05 -0700",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, raw); err == nil {
			return &t
		}
	}
	return nil
}

// textBody returns the plain-text or HTML body of the message decoded from base64.
func (m *gmailMessage) textBody() string {
	// Prefer text/plain; fall back to text/html
	if data := findPartData(&m.Payload, "text/plain"); data != "" {
		return data
	}
	if data := findPartData(&m.Payload, "text/html"); data != "" {
		return stripHTML(data)
	}
	// Top-level body (single-part message)
	if m.Payload.Body.Data != "" {
		decoded, err := base64Decode(m.Payload.Body.Data)
		if err == nil {
			return stripHTML(decoded)
		}
	}
	return ""
}

func findPartData(part *gmailMessagePart, mimeType string) string {
	if strings.HasPrefix(part.MimeType, mimeType) && part.Body.Data != "" {
		decoded, err := base64Decode(part.Body.Data)
		if err == nil {
			return decoded
		}
	}
	for i := range part.Parts {
		if data := findPartData(&part.Parts[i], mimeType); data != "" {
			return data
		}
	}
	return ""
}

func base64Decode(encoded string) (string, error) {
	// Gmail uses URL-safe base64
	b, err := base64.URLEncoding.DecodeString(encoded)
	if err != nil {
		// Try standard base64 as fallback
		b, err = base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			return "", err
		}
	}
	return string(b), nil
}

func listGmailMessages(client *http.Client, query string, maxResults int) ([]string, error) {
	var ids []string
	pageToken := ""

	for {
		remaining := maxResults - len(ids)
		if remaining <= 0 {
			break
		}
		reqURL := fmt.Sprintf(
			"https://gmail.googleapis.com/gmail/v1/users/me/messages?q=%s&maxResults=%d",
			url.QueryEscape(query), remaining,
		)
		if pageToken != "" {
			reqURL += "&pageToken=" + pageToken
		}

		resp, err := client.Get(reqURL)
		if err != nil {
			return nil, err
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("gmail list API returned %d: %s", resp.StatusCode, string(body))
		}

		var listResp gmailListResponse
		if err := json.Unmarshal(body, &listResp); err != nil {
			return nil, err
		}

		for _, m := range listResp.Messages {
			ids = append(ids, m.ID)
		}

		if listResp.NextPageToken == "" {
			break
		}
		pageToken = listResp.NextPageToken
	}

	return ids, nil
}

func getGmailMessage(client *http.Client, id string) (*gmailMessage, error) {
	url := fmt.Sprintf(
		"https://gmail.googleapis.com/gmail/v1/users/me/messages/%s?format=full", id,
	)
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gmail get API returned %d: %s", resp.StatusCode, string(body))
	}

	var msg gmailMessage
	if err := json.Unmarshal(body, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

// ----- HTML stripping -----

// knownBankQuery is the Gmail search query for known bank transaction emails.
const knownBankQuery = "category:purchases from:(bca@bca.co.id OR contact.center@permatabank.co.id OR alerts@seabank.co.id)"

var (
	reTags     = regexp.MustCompile(`<[^>]+>`)
	reSpaces   = regexp.MustCompile(`[ \t]+`)
	reNBSP     = regexp.MustCompile(`&nbsp;`)
	reAmp      = regexp.MustCompile(`&amp;`)
	reLt       = regexp.MustCompile(`&lt;`)
	reGt       = regexp.MustCompile(`&gt;`)
	reNonDigit = regexp.MustCompile(`[^\d]`)
)

func stripHTML(html string) string {
	s := reNBSP.ReplaceAllString(html, " ")
	s = reLt.ReplaceAllString(s, "<")
	s = reGt.ReplaceAllString(s, ">")
	s = reAmp.ReplaceAllString(s, "&")
	s = reTags.ReplaceAllString(s, " ")
	s = reSpaces.ReplaceAllString(s, " ")
	return s
}

// ----- Email parsing -----

type parsedEmail struct {
	bankName      string
	description   string
	amount        int
	date          time.Time
	accountSuffix string // last visible digits of the source account
}

func parseEmailTransaction(subject, fromEmail, body string) (*parsedEmail, error) {
	from := strings.ToLower(fromEmail)

	switch {
	case strings.Contains(from, "bca@bca.co.id") || strings.Contains(from, "bca.co.id"):
		return parseBCAEmail(subject, body)
	case strings.Contains(from, "permatabank.co.id"):
		return parsePermataEmail(subject, body)
	case strings.Contains(from, "seabank.co.id"):
		return parseSeaBankEmail(subject, body)
	default:
		return nil, fmt.Errorf("unknown sender: %s", fromEmail)
	}
}

// extractField extracts the value after "key : value" or "key: value" in the text.
func extractField(text, key string) string {
	// Build a regex that matches "key\s*:\s*<value up to newline>"
	re := regexp.MustCompile(`(?i)` + regexp.QuoteMeta(key) + `\s*:\s*([^\n\r]+)`)
	m := re.FindStringSubmatch(text)
	if len(m) < 2 {
		return ""
	}
	return strings.TrimSpace(m[1])
}

// extractSeaBankCell handles SeaBank's HTML table format where key and value are in
// separate <td> cells with no colon separator. After HTML stripping the body becomes
// a flat string like "Jumlah IDR 50.000 Biaya Rp1.000 ...". This function tries the
// standard colon-based extractField first, then falls back to a boundary-based match
// using the provided sibling field names to delimit the captured value.
func extractSeaBankCell(text, key string, allKeys []string) string {
	if v := extractField(text, key); v != "" {
		return v
	}
	// Find the key in the text (case-insensitive)
	lowerText := strings.ToLower(text)
	lowerKey := strings.ToLower(key)
	keyIdx := strings.Index(lowerText, lowerKey)
	if keyIdx < 0 {
		return ""
	}
	// Start after the key (and any trailing colon/spaces)
	start := keyIdx + len(key)
	for start < len(text) && (text[start] == ':' || text[start] == ' ' || text[start] == '\t' || text[start] == '\n' || text[start] == '\r') {
		start++
	}
	// Find the earliest next field boundary after start
	end := len(text)
	for _, other := range allKeys {
		if strings.EqualFold(other, key) {
			continue
		}
		idx := strings.Index(strings.ToLower(text[start:]), strings.ToLower(other))
		if idx >= 0 && start+idx < end {
			end = start + idx
		}
	}
	return strings.TrimSpace(text[start:end])
}

// parseAmount converts Indonesian/standard currency strings to integer IDR.
// Examples: "IDR 138,600.00" → 138600, "34,500" → 34500, "IDR 50.000" → 50000
func parseAmount(raw string) (int, error) {
	s := strings.TrimSpace(raw)
	// Remove currency labels
	for _, prefix := range []string{"IDR", "Rp", "idr", "rp"} {
		s = strings.TrimPrefix(s, prefix)
	}
	s = strings.TrimSpace(s)

	// Remove trailing ".00" (standard decimal cents)
	if strings.HasSuffix(s, ".00") {
		s = s[:len(s)-3]
	}

	// Remove all remaining commas and periods (thousand separators)
	s = strings.ReplaceAll(s, ",", "")
	s = strings.ReplaceAll(s, ".", "")
	s = strings.TrimSpace(s)

	// Remove any remaining non-digit characters
	s = reNonDigit.ReplaceAllString(s, "")

	if s == "" {
		return 0, errors.New("empty amount")
	}

	val, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}
	if val <= 0 {
		return 0, errors.New("non-positive amount")
	}
	return val, nil
}

// parseAccountSuffix extracts the last group of digits from a masked account string.
// "TAHAPAN XPRESI - 7055****17" → "17"
// "FAWWAZ BAYUREKSA - 4874 - IDR" → "4874"
// "41XXXXXX01" → "01"
func parseAccountSuffix(raw string) string {
	re := regexp.MustCompile(`\d+`)
	all := re.FindAllString(raw, -1)
	if len(all) == 0 {
		return ""
	}
	// Return the last digit group (most likely account suffix)
	return all[len(all)-1]
}

// matchAsset finds the asset whose AccountNo ends with the given suffix.
func matchAsset(assets []models.Asset, suffix string) uint64 {
	if suffix == "" {
		return 0
	}
	for _, a := range assets {
		if strings.HasSuffix(a.AccountNo, suffix) {
			return a.ID
		}
	}
	return 0
}

// BCA email parser
// Subject: "Internet Transaction Journal" | "Cash Withdrawal Successful"
// From: bca@bca.co.id
func parseBCAEmail(subject, body string) (*parsedEmail, error) {
	// Only process successful transactions
	status := extractField(body, "Status")
	s := strings.ToLower(status)
	if !strings.Contains(s, "successful") &&
		!strings.Contains(s, "sukses") &&
		!strings.Contains(s, "berhasil") {
		return nil, fmt.Errorf("BCA: skipping non-successful transaction (status: %s)", status)
	}

	// Amount
	amountRaw := extractField(body, "Total Payment")
	if amountRaw == "" {
		amountRaw = extractField(body, "Amount")
	}
	amount, err := parseAmount(amountRaw)
	if err != nil {
		return nil, fmt.Errorf("BCA: cannot parse amount %q: %w", amountRaw, err)
	}

	// Date
	dateRaw := extractField(body, "Transaction Date")
	txDate, err := parseBCADate(dateRaw)
	if err != nil {
		txDate = time.Now()
	}

	// Description
	txType := extractField(body, "Transaction Type")
	paymentTo := extractField(body, "Payment to")
	if paymentTo == "" {
		paymentTo = extractField(body, "Beneficiary Account")
	}
	desc := buildDescription("BCA", txType, paymentTo)

	// Account suffix for asset matching
	sourceOfFund := extractField(body, "Source of Fund")
	suffix := parseAccountSuffix(sourceOfFund)

	return &parsedEmail{
		bankName:      "BCA",
		description:   desc,
		amount:        amount,
		date:          txDate,
		accountSuffix: suffix,
	}, nil
}

// wib is the fixed timezone for all Indonesian bank emails (UTC+7).
var wib = time.FixedZone("WIB", 7*3600)

// parseBCADate parses "24 Mar 2026 18:08:10"
func parseBCADate(raw string) (time.Time, error) {
	formats := []string{
		"2 Jan 2006 15:04:05",
		"02 Jan 2006 15:04:05",
		"2006-01-02 15:04:05",
	}
	for _, f := range formats {
		if t, err := time.ParseInLocation(f, strings.TrimSpace(raw), wib); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("cannot parse BCA date: %s", raw)
}

// Permata ME email parser
// Subject: "Permata ME : QR Pay" | "Permata ME : Payment - Virtual Account"
// From: contact.center@permatabank.co.id
func parsePermataEmail(subject, body string) (*parsedEmail, error) {
	// Only successful
	status := extractField(body, "Status Transaksi")
	if status != "" {
		s := strings.ToLower(status)
		if !strings.Contains(s, "sukses") && !strings.Contains(s, "success") && !strings.Contains(s, "berhasil") {
			return nil, fmt.Errorf("Permata: skipping non-successful transaction (status: %s)", status)
		}
	}

	// Amount
	amountRaw := extractField(body, "Total Nominal")
	if amountRaw == "" {
		amountRaw = extractField(body, "Nominal")
	}
	amount, err := parseAmount(amountRaw)
	if err != nil {
		return nil, fmt.Errorf("Permata: cannot parse amount %q: %w", amountRaw, err)
	}

	// Date + time
	dateRaw := extractField(body, "Tanggal")
	timeRaw := extractField(body, "Jam")
	txDate, err := parsePermataDate(dateRaw, timeRaw)
	if err != nil {
		txDate = time.Now()
	}

	// Description
	merchant := extractField(body, "Nama Merchant")
	if merchant == "" {
		merchant = extractField(body, "Rekening Tujuan")
	}
	category := extractField(body, "Kategori")
	desc := buildDescription("Permata ME", category, merchant)

	// Account suffix for asset matching
	rekeningAsal := extractField(body, "Rekening Asal")
	suffix := parseAccountSuffix(rekeningAsal)

	return &parsedEmail{
		bankName:      "Permata",
		description:   desc,
		amount:        amount,
		date:          txDate,
		accountSuffix: suffix,
	}, nil
}

// parsePermataDate parses "23/03/2026" + "17:47:51"
func parsePermataDate(dateRaw, timeRaw string) (time.Time, error) {
	combined := strings.TrimSpace(dateRaw) + " " + strings.TrimSpace(timeRaw)
	formats := []string{
		"02/01/2006 15:04:05",
		"2/1/2006 15:04:05",
		"02/01/2006",
		"2/1/2006",
	}
	for _, f := range formats {
		if t, err := time.ParseInLocation(f, strings.TrimSpace(combined), wib); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("cannot parse Permata date: %s %s", dateRaw, timeRaw)
}

// SeaBank email parser
// Subject: "Notifikasi Top Up e-Wallet di SeaBank" | other SeaBank notifications
// From: alerts@seabank.co.id
//
// SeaBank emails are HTML tables with key and value in separate <td> cells.
// Many rows have no colon, so extractField won't find them. We use extractSeaBankCell
// with explicit sibling field names as value boundaries.
func parseSeaBankEmail(subject, body string) (*parsedEmail, error) {
	// SeaBank known table field names (used as boundaries when extracting table cells)
	seaBankFields := []string{
		"Waktu Transaksi", "Jenis Transaksi", "Transfer Dari",
		"Rekening Tujuan", "Jumlah", "Biaya", "No. Referensi", "Catatan",
	}

	// Amount
	amountRaw := extractSeaBankCell(body, "Jumlah", seaBankFields)
	if amountRaw == "" {
		amountRaw = extractSeaBankCell(body, "Nominal", seaBankFields)
	}
	amount, err := parseAmount(amountRaw)
	if err != nil {
		return nil, fmt.Errorf("SeaBank: cannot parse amount %q: %w", amountRaw, err)
	}

	// Date
	dateRaw := extractSeaBankCell(body, "Waktu Transaksi", seaBankFields)
	txDate, err := parseSeaBankDate(dateRaw)
	if err != nil {
		txDate = time.Now()
	}

	// Description
	txType := extractSeaBankCell(body, "Jenis Transaksi", seaBankFields)
	destination := extractSeaBankCell(body, "Rekening Tujuan", seaBankFields)
	desc := buildDescription("SeaBank", txType, destination)

	// Account suffix for asset matching
	transferDari := extractSeaBankCell(body, "Transfer Dari", seaBankFields)
	suffix := parseAccountSuffix(transferDari)

	return &parsedEmail{
		bankName:      "SeaBank",
		description:   desc,
		amount:        amount,
		date:          txDate,
		accountSuffix: suffix,
	}, nil
}

// parseSeaBankDate parses "25 Mar 2026 20:27"
func parseSeaBankDate(raw string) (time.Time, error) {
	formats := []string{
		"2 Jan 2006 15:04",
		"02 Jan 2006 15:04",
		"2 Jan 2006 15:04:05",
		"02 Jan 2006 15:04:05",
	}
	for _, f := range formats {
		if t, err := time.ParseInLocation(f, strings.TrimSpace(raw), wib); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("cannot parse SeaBank date: %s", raw)
}

func buildDescription(bank, txType, target string) string {
	parts := []string{}
	if bank != "" {
		parts = append(parts, bank)
	}
	if txType != "" {
		parts = append(parts, txType)
	}
	if target != "" {
		parts = append(parts, target)
	}
	desc := strings.Join(parts, " - ")
	if len(desc) > 200 {
		desc = desc[:200]
	}
	return desc
}
