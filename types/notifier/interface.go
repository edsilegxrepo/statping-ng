package notifier

import (
	"github.com/statping-ng/statping-ng/types/failures"
	"github.com/statping-ng/statping-ng/types/services"
)

// Notifier interface is required to create a new Notifier
type Notifier interface {
	OnSuccess(*services.Service) (string, error)                   // OnSuccess is triggered when a service is successful
	OnFailure(*services.Service, failures.Failure) (string, error) // OnFailure is triggered when a service is failing
	OnTest() (string, error)                                       // OnTest is triggered for testing
	OnSave() (string, error)                                       // OnSave is triggered for when saved
}

// DigestData contains the data for a daily digest
type DigestData struct {
	AppName         string
	Domain          string
	GeneratedAt     string
	Period          string
	TotalServices   int
	HealthyServices int
	FailedServices  int
	ServiceSummary  []ServiceDigest
	AppErrors       []AppError
	HasFailures     bool
	HasAppErrors    bool
}

type ServiceDigest struct {
	Name          string
	Status        string
	FailureCount  int
	TotalDowntime string
	LastFailure   string
}

type AppError struct {
	Timestamp string
	Message   string
}

// DigestNotifier is an optional interface for notifiers that support daily digest
type DigestNotifier interface {
	OnDigest(DigestData) (string, error)
}
