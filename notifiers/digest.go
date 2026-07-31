package notifiers

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"fmt"
	"html/template"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/go-mail/mail"
	"github.com/statping-ng/statping-ng/types/core"
	"github.com/statping-ng/statping-ng/types/failures"
	"github.com/statping-ng/statping-ng/types/notifier"
	"github.com/statping-ng/statping-ng/types/services"
	"github.com/statping-ng/statping-ng/utils"
)

// Email templates are organized by notifier:
//   - Service alerts (online/offline): notifiers/email_templates.go
//   - Daily digest: notifiers/digest.go (this file, below)
//   - Teams/Slack/Discord: JSON payloads inline in their respective notifiers

const digestEmailTemplate = `<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Daily Digest</title>
  <style>
    body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; margin: 0; padding: 20px; background: #f5f5f5; }
    .container { max-width: 600px; margin: 0 auto; background: #fff; border-radius: 8px; overflow: hidden; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
    .header { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: #fff; padding: 30px; text-align: center; }
    .header h1 { margin: 0; font-size: 24px; }
    .header p { margin: 10px 0 0; opacity: 0.9; }
    .content { padding: 30px; }
    .stats { display: flex; justify-content: space-around; margin-bottom: 30px; text-align: center; }
    .stat { padding: 15px; }
    .stat-value { font-size: 32px; font-weight: bold; color: #333; }
    .stat-label { font-size: 12px; color: #666; text-transform: uppercase; }
    .stat-healthy .stat-value { color: #22c55e; }
    .stat-failed .stat-value { color: #ef4444; }
    .section { margin-bottom: 25px; }
    .section h2 { font-size: 16px; color: #333; border-bottom: 2px solid #eee; padding-bottom: 10px; margin-bottom: 15px; }
    table { width: 100%; border-collapse: collapse; }
    th, td { padding: 12px; text-align: left; border-bottom: 1px solid #eee; }
    th { background: #f8f9fa; font-size: 12px; text-transform: uppercase; color: #666; }
    .status-online { color: #22c55e; font-weight: 600; }
    .status-offline { color: #ef4444; font-weight: 600; }
    .error-list { background: #fef2f2; border-left: 4px solid #ef4444; border-radius: 0 4px 4px 0; padding: 15px; }
    .error-item { padding: 8px 0; border-bottom: 1px solid #fecaca; font-size: 13px; color: #991b1b; }
    .error-item:last-child { border-bottom: none; }
    .footer { background: #f8f9fa; padding: 20px; text-align: center; font-size: 12px; color: #666; }
    .footer a { color: #6b7280; }
    .no-issues { text-align: center; padding: 30px; color: #22c55e; }
    @media (max-width: 620px) {
      .stats { flex-direction: column; }
      .stat { padding: 10px 0; }
      th, td { padding: 8px; font-size: 13px; }
    }
  </style>
</head>
<body>
  <div class="container">
    <div class="header">
      <h1>{{.AppName}} Daily Digest</h1>
      <p>{{.Period}}</p>
    </div>
    <div class="content">
      <div class="stats">
        <div class="stat">
          <div class="stat-value">{{.TotalServices}}</div>
          <div class="stat-label">Total Services</div>
        </div>
        <div class="stat stat-healthy">
          <div class="stat-value">{{.HealthyServices}}</div>
          <div class="stat-label">Healthy</div>
        </div>
        <div class="stat stat-failed">
          <div class="stat-value">{{.FailedServices}}</div>
          <div class="stat-label">Currently Down</div>
        </div>
      </div>

      {{if .HasFailures}}
      <div class="section">
        <h2>Service Issues (Last 24h)</h2>
        <table>
          <tr>
            <th>Service</th>
            <th>Status</th>
            <th>Failures</th>
            <th>Last Issue</th>
          </tr>
          {{range .ServiceSummary}}
          <tr>
            <td>{{.Name}}</td>
            <td class="{{if eq .Status "Online"}}status-online{{else}}status-offline{{end}}">{{.Status}}</td>
            <td>{{.FailureCount}}</td>
            <td>{{.LastFailure}}</td>
          </tr>
          {{end}}
        </table>
      </div>
      {{else}}
      <div class="no-issues">
        <div style="font-size: 48px;">&#x2705;</div>
        <p>All services healthy — no issues in the last 24 hours</p>
      </div>
      {{end}}

      {{if .HasAppErrors}}
      <div class="section">
        <h2>Application Errors</h2>
        <div class="error-list">
          {{range .AppErrors}}
          <div class="error-item">{{.Message}}</div>
          {{end}}
        </div>
      </div>
      {{end}}
    </div>
    <div class="footer">
      <p>Generated at {{.GeneratedAt}}</p>
      <p><a href="{{.Domain}}">{{.Domain}}</a></p>
    </div>
  </div>
</body>
</html>`

var (
	digestLog      = utils.Log.WithField("type", "digest")
	digestOnce     sync.Once
	digestStopOnce sync.Once
	stopDigest     chan struct{}
	digestResetMu  sync.Mutex
)

// StartDigestScheduler starts the daily digest email scheduler
func StartDigestScheduler() {
	digestResetMu.Lock()
	defer digestResetMu.Unlock()

	// Reset the stop once so we can stop again after restart
	digestStopOnce = sync.Once{}

	digestOnce.Do(func() {
		stopDigest = make(chan struct{})
		go runDigestScheduler()
	})
}

// StopDigestScheduler stops the daily digest scheduler
func StopDigestScheduler() {
	digestResetMu.Lock()
	defer digestResetMu.Unlock()

	digestStopOnce.Do(func() {
		if stopDigest != nil {
			close(stopDigest)
		}
		// Reset the start once so we can start again
		digestOnce = sync.Once{}
	})
}

func runDigestScheduler() {
	digestLog.Infoln("Daily digest scheduler started")

	for {
		now := time.Now()
		nextRun := calculateNextDigestTime(now)
		waitDuration := nextRun.Sub(now)

		digestLog.Infof("Next daily digest scheduled for %s (in %s)", nextRun.Format("2006-01-02 15:04"), waitDuration.Round(time.Minute))

		select {
		case <-time.After(waitDuration):
			sendDailyDigest()
		case <-stopDigest:
			digestLog.Infoln("Daily digest scheduler stopped")
			return
		}
	}
}

func calculateNextDigestTime(now time.Time) time.Time {
	hour := core.App.DigestHour
	if hour < 0 || hour > 23 {
		hour = 8
	}

	next := time.Date(now.Year(), now.Month(), now.Day(), hour, 0, 0, 0, now.Location())
	if now.After(next) {
		next = next.Add(24 * time.Hour)
	}
	return next
}

// digestData is a local type alias for template rendering
type digestData = notifier.DigestData
type serviceDigest = notifier.ServiceDigest
type appError = notifier.AppError

func sendDailyDigest() {
	c := core.App
	if !c.DigestEnabled.Bool {
		return
	}

	digestLog.Infoln("Generating daily digest...")
	data := generateDigestData()

	// Send to email recipients (legacy behavior)
	emails := strings.TrimSpace(c.DigestEmails)
	if emails != "" && email.Host.String != "" {
		htmlContent := renderDigestEmail(data)
		recipients := parseEmails(emails)
		for _, recipient := range recipients {
			if err := sendDigestEmail(recipient, htmlContent); err != nil {
				digestLog.Errorf("Failed to send digest email to %s: %v", recipient, err)
			} else {
				digestLog.Infof("Daily digest email sent to %s", recipient)
			}
		}
	}

	// Send to all notifiers with receive_digest enabled
	sendDigestToNotifiers(data)
}

// sendDigestToNotifiers sends the digest to all notifiers that have receive_digest enabled
func sendDigestToNotifiers(data digestData) {
	for _, n := range services.AllNotifiers() {
		notif := n.Select()
		if !notif.Enabled.Bool || !notif.ReceiveDigest.Bool {
			continue
		}

		// Check if this notifier implements DigestNotifier
		if dn, ok := n.(notifier.DigestNotifier); ok {
			digestLog.Infof("Sending digest to %s notifier...", notif.Method)
			if _, err := dn.OnDigest(data); err != nil {
				digestLog.Errorf("Failed to send digest via %s: %v", notif.Method, err)
			} else {
				digestLog.Infof("Digest sent via %s", notif.Method)
			}
		}
	}
}

func parseEmails(emails string) []string {
	var result []string
	for _, e := range strings.Split(emails, ",") {
		e = strings.TrimSpace(e)
		if e != "" {
			result = append(result, e)
		}
	}
	return result
}

func generateDigestData() digestData {
	now := utils.Now()
	yesterday := now.Add(-24 * time.Hour)

	allServices := services.AllInOrder()
	var healthyCount, failedCount int
	var serviceSummaries []notifier.ServiceDigest

	for _, s := range allServices {
		fails := failures.Since(yesterday, s).List()
		failCount := len(fails)

		status := "Online"
		if !s.Online {
			status = "Offline"
			failedCount++
		} else {
			healthyCount++
		}

		var totalDowntime time.Duration
		var lastFailureTime string
		for _, f := range fails {
			if f.CreatedAt.After(yesterday) {
				totalDowntime += time.Duration(f.PingTime) * time.Microsecond
				lastFailureTime = f.CreatedAt.Format("15:04")
			}
		}

		if failCount > 0 || !s.Online {
			serviceSummaries = append(serviceSummaries, notifier.ServiceDigest{
				Name:          s.Name,
				Status:        status,
				FailureCount:  failCount,
				TotalDowntime: formatDuration(totalDowntime),
				LastFailure:   lastFailureTime,
			})
		}
	}

	// Get application errors from logs (last 24h)
	appErrors := getRecentAppErrors(yesterday)

	return digestData{
		AppName:         core.App.Name,
		Domain:          core.App.Domain,
		GeneratedAt:     now.Format("2006-01-02 15:04 MST"),
		Period:          fmt.Sprintf("%s to %s", yesterday.Format("2006-01-02 15:04"), now.Format("15:04")),
		TotalServices:   len(allServices),
		HealthyServices: healthyCount,
		FailedServices:  failedCount,
		ServiceSummary:  serviceSummaries,
		AppErrors:       appErrors,
		HasFailures:     len(serviceSummaries) > 0,
		HasAppErrors:    len(appErrors) > 0,
	}
}

func getRecentAppErrors(since time.Time) []notifier.AppError {
	var errors []notifier.AppError
	utils.LockLines.Lock()
	defer utils.LockLines.Unlock()

	for _, line := range utils.LastLines {
		lineStr := line.FormatForHtml()
		if strings.Contains(lineStr, "error") || strings.Contains(lineStr, "Error") || strings.Contains(lineStr, "ERROR") {
			errors = append(errors, notifier.AppError{
				Timestamp: "",
				Message:   truncate(lineStr, 200),
			})
		}
	}

	// Limit to last 50 errors
	if len(errors) > 50 {
		errors = errors[len(errors)-50:]
	}
	return errors
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	return fmt.Sprintf("%dh %dm", int(d.Hours()), int(d.Minutes())%60)
}

func renderDigestEmail(data digestData) string {
	t, err := template.New("digest").Parse(digestEmailTemplate)
	if err != nil {
		digestLog.Errorf("Failed to parse digest template: %v", err)
		return ""
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		digestLog.Errorf("Failed to execute digest template: %v", err)
		return ""
	}

	return buf.String()
}

func sendDigestEmail(to, htmlContent string) error {
	mailer := mail.NewDialer(email.Host.String, int(email.Port.Int64), email.Username.String, email.Password.String)
	mailer.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12}
	mailer.Timeout = 30 * time.Second // Explicit connection timeout

	m := mail.NewMessage()
	m.SetAddressHeader("From", email.Var1.String, fmt.Sprintf("%s Monitoring", core.App.Name))
	m.SetHeader("To", to)
	m.SetHeader("Subject", fmt.Sprintf("[%s] Daily Status Digest - %s", core.App.Name, time.Now().Format("2006-01-02")))
	m.SetBody("text/html", htmlContent)

	return mailer.DialAndSend(m)
}

// SendTestDigest sends a test digest email immediately
func SendTestDigest() error {
	c := core.App
	emails := strings.TrimSpace(c.DigestEmails)
	if emails == "" {
		return fmt.Errorf("no email addresses configured")
	}

	if email.Host.String == "" {
		return fmt.Errorf("SMTP is not configured")
	}

	data := generateDigestData()
	htmlContent := renderDigestEmail(data)

	recipients := parseEmails(emails)
	var lastErr error
	for _, recipient := range recipients {
		if err := sendDigestEmail(recipient, htmlContent); err != nil {
			lastErr = err
			digestLog.Errorf("Failed to send test digest to %s: %v", recipient, err)
		} else {
			digestLog.Infof("Test digest sent to %s", recipient)
		}
	}
	return lastErr
}

// SMTPDiagResult contains SMTP diagnostic results
type SMTPDiagResult struct {
	Host           string `json:"host"`
	Port           int    `json:"port"`
	Connected      bool   `json:"connected"`
	Banner         string `json:"banner,omitempty"`
	TLSSupported   bool   `json:"tls_supported"`
	TLSRequired    bool   `json:"tls_required"`
	AuthSupported  bool   `json:"auth_supported"`
	AuthMethods    string `json:"auth_methods,omitempty"`
	AuthSuccess    bool   `json:"auth_success,omitempty"`
	Error          string `json:"error,omitempty"`
	Recommendations []string `json:"recommendations,omitempty"`
}

// TestSMTPConnection performs SMTP diagnostics
func TestSMTPConnection() *SMTPDiagResult {
	result := &SMTPDiagResult{
		Host: email.Host.String,
		Port: int(email.Port.Int64),
	}

	if result.Host == "" {
		result.Error = "SMTP host not configured"
		result.Recommendations = append(result.Recommendations, "Configure SMTP host in Email notifier settings")
		return result
	}

	if result.Port == 0 {
		result.Port = 25
	}

	address := net.JoinHostPort(result.Host, fmt.Sprintf("%d", result.Port))

	// Test basic TCP connection with timeout
	conn, err := net.DialTimeout("tcp", address, 10*time.Second)
	if err != nil {
		result.Error = fmt.Sprintf("Connection failed: %v", err)
		if strings.Contains(err.Error(), "timeout") {
			result.Recommendations = append(result.Recommendations, "Connection timed out - port may be blocked by firewall")
		} else if strings.Contains(err.Error(), "refused") {
			result.Recommendations = append(result.Recommendations, "Connection refused - check if SMTP service is running")
		}
		return result
	}
	defer conn.Close()

	result.Connected = true

	// Set read deadline
	conn.SetReadDeadline(time.Now().Add(10 * time.Second))

	// Read server banner
	reader := bufio.NewReader(conn)
	banner, err := reader.ReadString('\n')
	if err != nil {
		result.Error = fmt.Sprintf("Failed to read banner: %v", err)
		return result
	}
	result.Banner = strings.TrimSpace(banner)

	// Check if banner indicates a problem
	if strings.HasPrefix(result.Banner, "4") || strings.HasPrefix(result.Banner, "5") {
		result.Error = fmt.Sprintf("Server rejected connection: %s", result.Banner)
		if strings.Contains(result.Banner, "421") {
			result.Recommendations = append(result.Recommendations, "421 error - server temporarily unavailable or rejecting connections from this IP")
		}
		return result
	}

	// Send EHLO to get capabilities
	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	fmt.Fprintf(conn, "EHLO localhost\r\n")

	// Read EHLO response (multi-line)
	var capabilities []string
	for {
		conn.SetReadDeadline(time.Now().Add(10 * time.Second))
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		line = strings.TrimSpace(line)
		capabilities = append(capabilities, line)

		// Check for TLS support
		if strings.Contains(strings.ToUpper(line), "STARTTLS") {
			result.TLSSupported = true
		}

		// Check for AUTH support
		if strings.Contains(strings.ToUpper(line), "AUTH") {
			result.AuthSupported = true
			// Extract auth methods
			parts := strings.SplitN(line, " ", 2)
			if len(parts) > 1 && strings.HasPrefix(strings.ToUpper(parts[0]), "250") {
				authPart := strings.TrimPrefix(strings.ToUpper(line), "250-AUTH ")
				authPart = strings.TrimPrefix(authPart, "250 AUTH ")
				result.AuthMethods = authPart
			}
		}

		// Last line of EHLO response starts with "250 " (space, not dash)
		if strings.HasPrefix(line, "250 ") {
			break
		}
	}

	// Generate recommendations based on findings
	if result.Port == 25 && !result.TLSSupported {
		result.Recommendations = append(result.Recommendations, "Port 25 without TLS - connection is unencrypted")
	}

	if result.Port == 587 && result.TLSSupported {
		result.Recommendations = append(result.Recommendations, "Port 587 with STARTTLS available - recommended configuration")
	}

	if result.Port == 465 {
		result.Recommendations = append(result.Recommendations, "Port 465 uses implicit TLS - make sure 'Disable TLS/SSL' is OFF")
	}

	if result.AuthSupported && email.Username.String == "" {
		result.Recommendations = append(result.Recommendations, "Server supports authentication but no username configured")
	}

	if !result.AuthSupported && email.Username.String != "" {
		result.Recommendations = append(result.Recommendations, "Username configured but server does not advertise AUTH support")
	}

	// Test authentication if credentials provided
	if result.AuthSupported && email.Username.String != "" && email.Password.String != "" {
		// We won't actually test auth here to avoid locking accounts
		// Just note that credentials are configured
		result.Recommendations = append(result.Recommendations, "Credentials configured - authentication will be attempted on send")
	}

	// Send QUIT
	fmt.Fprintf(conn, "QUIT\r\n")

	return result
}
