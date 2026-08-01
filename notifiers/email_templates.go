package notifiers

import (
	"bytes"
	"html/template"
)

// Email templates for service alerts
// Clean, enterprise-grade HTML emails compatible with Gmail, Outlook, Apple Mail
//
// Note: Digest emails have their own template in digest.go

const baseEmailTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <meta http-equiv="X-UA-Compatible" content="IE=edge">
  <title>{{.Title}}</title>
  <style>
    /* Reset */
    body, table, td, p, a { -webkit-text-size-adjust: 100%; -ms-text-size-adjust: 100%; }
    table, td { mso-table-lspace: 0pt; mso-table-rspace: 0pt; }
    img { -ms-interpolation-mode: bicubic; border: 0; height: auto; line-height: 100%; outline: none; text-decoration: none; }
    body { margin: 0; padding: 0; width: 100%; height: 100%; }

    /* Base */
    body { background-color: #f4f5f7; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; }

    /* Container */
    .wrapper { width: 100%; table-layout: fixed; background-color: #f4f5f7; padding: 40px 0; }
    .container { max-width: 600px; margin: 0 auto; background-color: #ffffff; border-radius: 8px; overflow: hidden; box-shadow: 0 1px 3px rgba(0,0,0,0.1); }

    /* Header */
    .header { padding: 32px 40px; text-align: center; }
    .header-success { background: linear-gradient(135deg, #059669 0%, #10b981 100%); }
    .header-failure { background: linear-gradient(135deg, #dc2626 0%, #ef4444 100%); }
    .header h1 { margin: 0; color: #ffffff; font-size: 24px; font-weight: 600; line-height: 1.3; }
    .header p { margin: 8px 0 0; color: rgba(255,255,255,0.9); font-size: 14px; }

    /* Status badge */
    .status-badge { display: inline-block; padding: 6px 16px; border-radius: 20px; font-size: 13px; font-weight: 600; text-transform: uppercase; letter-spacing: 0.5px; margin-bottom: 12px; background-color: rgba(255,255,255,0.2); color: #ffffff; }

    /* Body */
    .body { padding: 32px 40px; }
    .body p { margin: 0 0 16px; color: #374151; font-size: 15px; line-height: 1.6; }
    .body p:last-child { margin-bottom: 0; }

    /* Info cards */
    .info-card { background-color: #f9fafb; border-radius: 6px; padding: 20px; margin: 20px 0; }
    .info-row { display: table; width: 100%; margin-bottom: 12px; }
    .info-row:last-child { margin-bottom: 0; }
    .info-label { display: table-cell; width: 120px; color: #6b7280; font-size: 13px; font-weight: 500; text-transform: uppercase; letter-spacing: 0.3px; vertical-align: top; padding-right: 12px; }
    .info-value { display: table-cell; color: #111827; font-size: 14px; word-break: break-word; }

    /* Issue box */
    .issue-box { background-color: #fef2f2; border-left: 4px solid #ef4444; border-radius: 0 6px 6px 0; padding: 16px 20px; margin: 20px 0; }
    .issue-box p { margin: 0; color: #991b1b; font-size: 14px; }
    .issue-label { font-weight: 600; margin-bottom: 4px; }

    /* Downtime indicator */
    .downtime { text-align: center; padding: 24px 0; }
    .downtime-value { font-size: 36px; font-weight: 700; color: #111827; line-height: 1; }
    .downtime-label { font-size: 13px; color: #6b7280; margin-top: 4px; text-transform: uppercase; letter-spacing: 0.5px; }

    /* Button */
    .btn-container { text-align: center; padding: 24px 0 8px; }
    .btn { display: inline-block; padding: 14px 32px; background-color: #2563eb; color: #ffffff !important; text-decoration: none; border-radius: 6px; font-size: 14px; font-weight: 600; }

    /* Footer */
    .footer { padding: 24px 40px; background-color: #f9fafb; border-top: 1px solid #e5e7eb; text-align: center; }
    .footer p { margin: 0; color: #9ca3af; font-size: 12px; line-height: 1.5; }
    .footer a { color: #6b7280; text-decoration: none; }

    /* Responsive */
    @media only screen and (max-width: 620px) {
      .container { margin: 0 16px; }
      .header, .body, .footer { padding-left: 24px; padding-right: 24px; }
      .header h1 { font-size: 20px; }
      .info-row { display: block; margin-bottom: 16px; }
      .info-label { display: block; margin-bottom: 4px; }
      .info-value { display: block; }
    }
  </style>
</head>
<body>
  <div class="wrapper">
    <table class="container" cellpadding="0" cellspacing="0" border="0" width="100%">
      {{.Content}}
    </table>
  </div>
</body>
</html>`

const serviceOnlineContent = `
      <tr>
        <td class="header header-success">
          <div class="status-badge">● Back Online</div>
          <h1>{{.Service.Name}}</h1>
          <p>Service has recovered and is now operational</p>
        </td>
      </tr>
      <tr>
        <td class="body">
          <div class="downtime">
            <div class="downtime-value">{{.Service.Downtime.Human}}</div>
            <div class="downtime-label">Total Downtime</div>
          </div>

          <div class="info-card">
            <div class="info-row">
              <span class="info-label">Service</span>
              <span class="info-value">{{.Service.Name}}</span>
            </div>
            <div class="info-row">
              <span class="info-label">Domain</span>
              <span class="info-value">{{.Service.Domain}}</span>
            </div>
            <div class="info-row">
              <span class="info-label">Status</span>
              <span class="info-value" style="color: #059669; font-weight: 600;">Operational</span>
            </div>
          </div>

          <div class="btn-container">
            <a href="{{.Core.Domain}}/service/{{.Service.Id}}" class="btn">View Service Dashboard</a>
          </div>
        </td>
      </tr>
      <tr>
        <td class="footer">
          <p>This alert was sent by <a href="{{.Core.Domain}}">{{.Core.Name}}</a></p>
        </td>
      </tr>`

const serviceOfflineContent = `
      <tr>
        <td class="header header-failure">
          <div class="status-badge">● Offline</div>
          <h1>{{.Service.Name}}</h1>
          <p>Service is currently unreachable</p>
        </td>
      </tr>
      <tr>
        <td class="body">
          <div class="downtime">
            <div class="downtime-value">{{.Service.Downtime.Human}}</div>
            <div class="downtime-label">Downtime Duration</div>
          </div>

          {{if .Failure.Issue}}
          <div class="issue-box">
            <p class="issue-label">Error Details</p>
            <p>{{.Failure.Issue}}</p>
          </div>
          {{end}}

          <div class="info-card">
            <div class="info-row">
              <span class="info-label">Service</span>
              <span class="info-value">{{.Service.Name}}</span>
            </div>
            <div class="info-row">
              <span class="info-label">Domain</span>
              <span class="info-value">{{.Service.Domain}}</span>
            </div>
            <div class="info-row">
              <span class="info-label">Status</span>
              <span class="info-value" style="color: #dc2626; font-weight: 600;">Unreachable</span>
            </div>
          </div>

          <div class="btn-container">
            <a href="{{.Core.Domain}}/service/{{.Service.Id}}" class="btn">View Service Dashboard</a>
          </div>
        </td>
      </tr>
      <tr>
        <td class="footer">
          <p>This alert was sent by <a href="{{.Core.Domain}}">{{.Core.Name}}</a></p>
        </td>
      </tr>`

// EmailTemplates provides access to service alert templates
var EmailTemplates = struct {
	ServiceOnline  string
	ServiceOffline string
}{
	ServiceOnline:  serviceOnlineContent,
	ServiceOffline: serviceOfflineContent,
}

// RenderEmail renders an email template with the provided data
func RenderEmail(content string, data interface{}) (string, error) {
	contentTmpl, err := template.New("content").Parse(content)
	if err != nil {
		return "", err
	}

	var contentBuf bytes.Buffer
	if err := contentTmpl.Execute(&contentBuf, data); err != nil {
		return "", err
	}

	type baseData struct {
		Title   string
		Content template.HTML
	}

	baseTmpl, err := template.New("base").Parse(baseEmailTemplate)
	if err != nil {
		return "", err
	}

	title := "Service Notification"

	var finalBuf bytes.Buffer
	if err := baseTmpl.Execute(&finalBuf, baseData{
		Title:   title,
		Content: template.HTML(contentBuf.String()), // #nosec G203 - content is from pre-defined templates, already escaped
	}); err != nil {
		return "", err
	}

	return finalBuf.String(), nil
}
