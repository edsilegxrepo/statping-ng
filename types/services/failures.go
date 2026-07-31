package services

import (
	"strings"
	"time"

	"github.com/statping-ng/statping-ng/types/failures"
)

const limitFailures = 32

func (s *Service) FailuresColumnID() (string, int64) {
	return "service", s.Id
}

func (s *Service) AllFailures() failures.Failurer {
	return failures.AllFailures(s)
}

func (s *Service) FailuresSince(t time.Time) failures.Failurer {
	return failures.Since(t, s)
}

func (s *Service) DowntimeText() string {
	last := s.AllFailures().Last()
	if last == nil {
		return ""
	}
	return parseError(last)
}

// ParseError returns a human readable error for a Failure
func parseError(f *failures.Failure) string {
	if f.Method == "checkin" {
		return "Checkin is Offline"
	}
	err := strings.Contains(f.Issue, "connection reset by peer")
	if err {
		return "Connection Reset"
	}
	err = strings.Contains(f.Issue, "operation timed out")
	if err {
		return "HTTP Request Timed Out"
	}
	err = strings.Contains(f.Issue, "x509: certificate is valid")
	if err {
		return "SSL Certificate invalid"
	}
	err = strings.Contains(f.Issue, "Client.Timeout exceeded while awaiting headers")
	if err {
		return "Connection Timed Out"
	}
	err = strings.Contains(f.Issue, "no such host")
	if err {
		return "Domain is offline or not found"
	}
	err = strings.Contains(f.Issue, "HTTP Status Code")
	if err {
		return "Incorrect HTTP Status Code"
	}
	err = strings.Contains(f.Issue, "connection refused")
	if err {
		return "Connection Failed"
	}
	err = strings.Contains(f.Issue, "can't assign requested address")
	if err {
		return "Unable to Request Address"
	}
	err = strings.Contains(f.Issue, "no route to host")
	if err {
		return "Domain is offline or not found"
	}
	err = strings.Contains(f.Issue, "i/o timeout")
	if err {
		return "Connection Timed Out"
	}
	err = strings.Contains(f.Issue, "Client.Timeout exceeded while reading body")
	if err {
		return "Timed Out on Response Body"
	}
	return f.Issue
}
