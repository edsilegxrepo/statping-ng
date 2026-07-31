package notifiers

import (
	"crypto/tls"
	"fmt"

	"github.com/go-mail/mail"
	"github.com/statping-ng/statping-ng/types/core"
	"github.com/statping-ng/statping-ng/types/failures"
	"github.com/statping-ng/statping-ng/types/notifications"
	"github.com/statping-ng/statping-ng/types/notifier"
	"github.com/statping-ng/statping-ng/types/services"
	"github.com/statping-ng/statping-ng/utils"
)

var _ notifier.Notifier = (*emailer)(nil)

var mailer *mail.Dialer

type emailer struct {
	*notifications.Notification
}

func (e *emailer) Select() *notifications.Notification {
	return e.Notification
}

func (e *emailer) Valid(values notifications.Values) error {
	return nil
}

var email = &emailer{
	&notifications.Notification{
		Method:      "email",
		Title:       "SMTP Mail",
		Description: "Send emails via SMTP when services are online or offline.",
		Author:      "Hunter Long",
		AuthorUrl:   "https://github.com/hunterlong",
		Icon:        "far fa-envelope",
		Limits:      30,
		Form: []notifications.NotificationForm{{
			Type:        "text",
			Title:       "SMTP Host",
			Placeholder: "Insert your SMTP Host here.",
			DbField:     "Host",
		}, {
			Type:        "text",
			Title:       "SMTP Username",
			Placeholder: "Insert your SMTP Username here.",
			DbField:     "Username",
		}, {
			Type:        "password",
			Title:       "SMTP Password",
			Placeholder: "Insert your SMTP Password here.",
			DbField:     "Password",
		}, {
			Type:        "number",
			Title:       "SMTP Port",
			Placeholder: "Insert your SMTP Port here.",
			DbField:     "Port",
		}, {
			Type:        "text",
			Title:       "Outgoing Email Address",
			Placeholder: "outgoing@email.com",
			DbField:     "Var1",
		}, {
			Type:        "email",
			Title:       "Send Alerts To",
			Placeholder: "sendto@email.com",
			DbField:     "Var2",
		}, {
			Type:        "switch",
			Title:       "Disable TLS/SSL",
			Placeholder: "",
			SmallText:   "Enabling this will set Insecure Skip Verify to true",
			DbField:     "api_key",
		}},
	},
}

type emailData struct {
	Core    *core.Core
	Service *services.Service
	Failure failures.Failure
	Email   string
}

type emailOutgoing struct {
	To       string
	Subject  string
	Template string
	From     string
}

// OnFailure will trigger failing service
func (e *emailer) OnFailure(s *services.Service, f failures.Failure) (string, error) {
	subject := fmt.Sprintf("🔴 %s is Offline", s.Name)

	data := emailData{
		Core:    core.App,
		Service: s,
		Failure: f,
		Email:   e.Var2.String,
	}

	tmpl, err := RenderEmail(EmailTemplates.ServiceOffline, data)
	if err != nil {
		log.Errorln(err)
		return "", err
	}

	email := &emailOutgoing{
		To:       e.Var2.String,
		Subject:  subject,
		Template: tmpl,
		From:     e.Var1.String,
	}
	return subject, e.dialSend(email)
}

// OnSuccess will trigger successful service
func (e *emailer) OnSuccess(s *services.Service) (string, error) {
	subject := fmt.Sprintf("✅ %s is Back Online", s.Name)

	data := emailData{
		Core:    core.App,
		Service: s,
		Failure: failures.Failure{},
		Email:   e.Var2.String,
	}

	tmpl, err := RenderEmail(EmailTemplates.ServiceOnline, data)
	if err != nil {
		log.Errorln(err)
		return "", err
	}

	email := &emailOutgoing{
		To:       e.Var2.String,
		Subject:  subject,
		Template: tmpl,
		From:     e.Var1.String,
	}
	return subject, e.dialSend(email)
}

// OnTest triggers when this notifier has been saved
func (e *emailer) OnTest() (string, error) {
	service := services.Example(true)
	subject := fmt.Sprintf("🔴 %s is Offline (Test)", service.Name)

	data := emailData{
		Core:    core.App,
		Service: service,
		Failure: failures.Example(),
		Email:   e.Var2.String,
	}

	tmpl, err := RenderEmail(EmailTemplates.ServiceOffline, data)
	if err != nil {
		log.Errorln(err)
		return "", err
	}

	email := &emailOutgoing{
		To:       e.Var2.String,
		Subject:  subject,
		Template: tmpl,
		From:     e.Var1.String,
	}
	return subject, e.dialSend(email)
}

// OnSave will trigger when this notifier is saved
func (e *emailer) OnSave() (string, error) {
	return "", nil
}

func (e *emailer) dialSend(email *emailOutgoing) error {
	mailer = mail.NewDialer(e.Host.String, int(e.Port.Int64), e.Username.String, e.Password.String)
	mailer.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12}
	m := mail.NewMessage()

	m.SetAddressHeader("From", email.From, "Monitoring Service")
	m.SetHeader("To", email.To)
	m.SetHeader("Subject", email.Subject)
	m.SetBody("text/html", email.Template)

	if err := mailer.DialAndSend(m); err != nil {
		utils.Log.Errorln(fmt.Sprintf("email '%v' sent to: %v - error: %v", email.Subject, email.To, err))
		return err
	}

	return nil
}
