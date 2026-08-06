package notifiers

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/statping-ng/statping-ng/types/core"
	"github.com/statping-ng/statping-ng/types/failures"
	"github.com/statping-ng/statping-ng/types/services"
	"github.com/statping-ng/statping-ng/utils"
)

//go:generate go run generate.go

var log = utils.Log.WithField("type", "notifier")

// replacer contains safe, read-only data for template substitution
type replacer struct {
	Core    coreInfo
	Service serviceInfo
	Failure failureInfo
}

// coreInfo exposes only safe fields from Core
type coreInfo struct {
	Name        string
	Description string
	Domain      string
}

// serviceInfo exposes only safe fields from Service
type serviceInfo struct {
	Id             int64
	Name           string
	Domain         string
	Type           string
	Method         string
	Port           int
	Online         bool
	Latency        int64
	PingTime       int64
	LastStatusCode int
}

// failureInfo exposes only safe fields from Failure
type failureInfo struct {
	Id        int64
	Issue     string
	ErrorCode int
	PingTime  int64
	CreatedAt time.Time
}

// allowedTemplateFields is the whitelist of allowed template field patterns
var allowedTemplatePattern = regexp.MustCompile(`^\{\{\s*\.(?:Core|Service|Failure)\.[A-Za-z]+\s*\}\}$`)

func InitNotifiers() {
	Add(
		slacker,
		mattermoster,
		Command,
		Discorder,
		email,
		LineNotify,
		Telegram,
		Teams,
		Twilio,
		Webhook,
		Pushover,
		Gotify,
		Ntfy,
		AmazonSNS,
	)

	services.UpdateNotifiers()
}

// ValidateTemplate checks that a template only uses allowed field access patterns
func ValidateTemplate(tmpl string) error {
	// Find all template directives
	re := regexp.MustCompile(`\{\{[^}]+\}\}`)
	matches := re.FindAllString(tmpl, -1)

	for _, match := range matches {
		// Allow simple field access only
		if !allowedTemplatePattern.MatchString(match) {
			// Allow literal text like {{if}} etc for backward compat, just block function calls
			if strings.Contains(match, "(") || strings.Contains(match, "call") ||
				strings.Contains(match, "template") || strings.Contains(match, "define") ||
				strings.Contains(match, "block") {
				return fmt.Errorf("template contains disallowed directive: %s", match)
			}
		}
	}
	return nil
}

func ReplaceTemplate(tmpl string, data replacer) string {
	buf := new(bytes.Buffer)
	tmp, err := template.New("replacement").Parse(tmpl)
	if err != nil {
		log.Error(err)
		return err.Error()
	}
	err = tmp.Execute(buf, data)
	if err != nil {
		log.Error(err)
		return err.Error()
	}
	return buf.String()
}

func Add(notifs ...services.ServiceNotifier) {
	for _, n := range notifs {
		notif := n.Select()
		if err := notif.Create(); err != nil {
			log.Error(err)
		}
		services.AddNotifier(n)
	}
}

// makeReplacer creates a safe replacer from service and failure data
func makeReplacer(s *services.Service, f failures.Failure) replacer {
	c := core.GetApp()
	var coreName, coreDesc, coreDomain string
	if c != nil {
		coreName = c.GetName()
		coreDesc = c.Description // read-only after init
		coreDomain = c.GetDomain()
	}
	data := replacer{
		Core: coreInfo{
			Name:        coreName,
			Description: coreDesc,
			Domain:      coreDomain,
		},
		Service: serviceInfo{
			Id:             s.Id,
			Name:           s.Name,
			Domain:         s.Domain,
			Type:           s.Type,
			Method:         s.Method,
			Port:           s.Port,
			Online:         s.Online,
			Latency:        s.Latency,
			PingTime:       s.PingTime,
			LastStatusCode: s.LastStatusCode,
		},
		Failure: failureInfo{
			Id:        f.Id,
			Issue:     f.Issue,
			ErrorCode: f.ErrorCode,
			PingTime:  f.PingTime,
			CreatedAt: f.CreatedAt,
		},
	}
	return data
}

func ReplaceVars(input string, s *services.Service, f failures.Failure) string {
	return ReplaceTemplate(input, makeReplacer(s, f))
}

var exampleFailure = &failures.Failure{
	Id:        1,
	Issue:     "HTTP returned a 500 status code",
	ErrorCode: 500,
	Service:   1,
	PingTime:  43203,
	CreatedAt: utils.Now().Add(-10 * time.Minute),
}
