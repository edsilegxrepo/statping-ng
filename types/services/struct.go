package services

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/statping-ng/statping-ng/types/checkins"
	"github.com/statping-ng/statping-ng/types/failures"
	"github.com/statping-ng/statping-ng/types/incidents"
	"github.com/statping-ng/statping-ng/types/messages"
	"github.com/statping-ng/statping-ng/types/null"
)

// Service is the main struct for Services
type Service struct {
	Id              int64           `gorm:"primary_key;column:id" json:"id" yaml:"id"`
	Name            string          `gorm:"column:name" json:"name" yaml:"name"`
	Domain          string          `gorm:"column:domain" json:"domain" yaml:"domain" private:"true" scope:"user,admin"`
	Expected        null.NullString `gorm:"column:expected" json:"expected" yaml:"expected" scope:"user,admin"`
	ExpectedStatus  int             `gorm:"default:200;column:expected_status" json:"expected_status" yaml:"expected_status" scope:"user,admin"`
	Interval        int             `gorm:"default:30;column:check_interval" json:"check_interval" yaml:"check_interval"`
	Type            string          `gorm:"column:check_type" json:"type" scope:"user,admin" yaml:"type"`
	Method          string          `gorm:"column:method" json:"method" scope:"user,admin" yaml:"method"`
	PostData        null.NullString `gorm:"column:post_data" json:"post_data" scope:"user,admin" yaml:"post_data"`
	Port            int             `gorm:"not null;column:port" json:"port" scope:"user,admin" yaml:"port"`
	Timeout         int             `gorm:"default:30;column:timeout" json:"timeout" scope:"user,admin" yaml:"timeout"`
	Order           int             `gorm:"default:0;column:order_id" json:"order_id" yaml:"order_id"`
	VerifySSL       null.NullBool   `gorm:"default:false;column:verify_ssl" json:"verify_ssl" scope:"user,admin" yaml:"verify_ssl"`
	GrpcHealthCheck null.NullBool   `gorm:"default:false;column:grpc_health_check" json:"grpc_health_check" scope:"user,admin" yaml:"grpc_health_check"`
	Public          null.NullBool   `gorm:"default:true;column:public" json:"public" yaml:"public"`
	GroupId         int             `gorm:"index;default:0;column:group_id" json:"group_id" yaml:"group_id"`
	TLSCert         null.NullString `gorm:"column:tls_cert" json:"tls_cert" scope:"user,admin" yaml:"tls_cert"`
	TLSCertKey      null.NullString `gorm:"column:tls_cert_key" json:"tls_cert_key" scope:"user,admin" yaml:"tls_cert_key"`
	TLSCertRoot     null.NullString `gorm:"column:tls_cert_root" json:"tls_cert_root" scope:"user,admin" yaml:"tls_cert_root"`
	// Database service fields
	DatabaseType  null.NullString `gorm:"column:database_type" json:"database_type" scope:"user,admin" yaml:"database_type"`        // postgres, mysql, sqlite, sqlserver, mongodb
	DatabaseDSN   null.NullString `gorm:"column:database_dsn" json:"database_dsn" scope:"admin" yaml:"database_dsn" private:"true"` // connection string (encrypted)
	DatabaseQuery null.NullString `gorm:"column:database_query" json:"database_query" scope:"user,admin" yaml:"database_query"`     // optional query to execute
	// Storage service fields
	StorageBackend     null.NullString `gorm:"column:storage_backend" json:"storage_backend" scope:"user,admin" yaml:"storage_backend"` // gcs, s3, azure
	StorageBucket      null.NullString `gorm:"column:storage_bucket" json:"storage_bucket" scope:"user,admin" yaml:"storage_bucket"`
	StorageProjectID   null.NullString `gorm:"column:storage_project_id" json:"storage_project_id" scope:"user,admin" yaml:"storage_project_id"` // GCS project ID
	StorageCredentials null.NullString `gorm:"column:storage_credentials" json:"storage_credentials" scope:"admin" yaml:"storage_credentials" private:"true"` // service account JSON (encrypted)
	// TLS certificate monitoring fields
	TLSTarget        null.NullString `gorm:"column:tls_target" json:"tls_target" scope:"user,admin" yaml:"tls_target"`                  // host:port to check
	TLSMinDays       int             `gorm:"column:tls_min_days;default:30" json:"tls_min_days" scope:"user,admin" yaml:"tls_min_days"` // alert if cert expires within N days
	TLSExpectedSAN   null.NullString `gorm:"column:tls_expected_san" json:"tls_expected_san" scope:"user,admin" yaml:"tls_expected_san"`
	TLSCheckOCSP     null.NullBool   `gorm:"default:false;column:tls_check_ocsp" json:"tls_check_ocsp" scope:"user,admin" yaml:"tls_check_ocsp"`
	TLSRequireSCT    null.NullBool   `gorm:"default:false;column:tls_require_sct" json:"tls_require_sct" scope:"user,admin" yaml:"tls_require_sct"`
	TLSExpiry        *time.Time      `gorm:"-" json:"tls_expiry,omitempty" yaml:"-"`         // certificate expiry date (runtime)
	TLSIssuer        string          `gorm:"-" json:"tls_issuer,omitempty" yaml:"-"`         // certificate issuer (runtime)
	TLSDaysRemaining int             `gorm:"-" json:"tls_days_remaining,omitempty" yaml:"-"` // days until expiry (runtime)
	Headers             null.NullString       `gorm:"column:headers" json:"headers" scope:"user,admin" yaml:"headers"`
	Permalink           null.NullString       `gorm:"index;column:permalink" json:"permalink" yaml:"permalink"`
	Redirect            null.NullBool         `gorm:"default:false;column:redirect" json:"redirect" scope:"user,admin" yaml:"redirect"`
	AllowInternal       null.NullBool         `gorm:"default:false;column:allow_internal" json:"allow_internal" scope:"admin" yaml:"allow_internal"`
	Priority            int                   `gorm:"default:3;column:priority" json:"priority" yaml:"priority"` // 1=Critical, 2=High, 3=Normal, 4=Low
	CreatedAt           time.Time             `gorm:"column:created_at" json:"created_at" yaml:"-"`
	UpdatedAt           time.Time             `gorm:"column:updated_at" json:"updated_at" yaml:"-"`
	Online              bool                  `gorm:"-" json:"online" yaml:"-"`
	Latency             int64                 `gorm:"-" json:"latency" yaml:"-"`
	PingTime            int64                 `gorm:"-" json:"ping_time" yaml:"-"`
	Online24Hours       float32               `gorm:"-" json:"online_24_hours" yaml:"-"`
	Online7Days         float32               `gorm:"-" json:"online_7_days" yaml:"-"`
	Online1Year         float32               `gorm:"-" json:"online_1_year" yaml:"-"`
	AvgResponse         int64                 `gorm:"-" json:"avg_response" yaml:"-"`
	FailuresLast24Hours int                   `gorm:"-" json:"failures_24_hours" yaml:"-"`
	Running             chan bool             `gorm:"-" json:"-" yaml:"-"`
	runningMu           sync.Mutex            `gorm:"-" json:"-" yaml:"-"`
	fieldsMu            sync.RWMutex          `gorm:"-" json:"-" yaml:"-"` // protects runtime fields during JSON serialization
	checkpoint          time.Time             `gorm:"-" json:"-" yaml:"-"`
	sleepDuration       time.Duration         `gorm:"-" json:"-" yaml:"-"`
	LastResponse        string                `gorm:"-" json:"-" yaml:"-"`
	NotifyAfter         int64                 `gorm:"column:notify_after" json:"notify_after" yaml:"notify_after" scope:"user,admin"`
	AllowNotifications  null.NullBool         `gorm:"default:true;column:allow_notifications" json:"allow_notifications" yaml:"allow_notifications" scope:"user,admin"`
	UpdateNotify        null.NullBool         `gorm:"default:true;column:notify_all_changes" json:"notify_all_changes" yaml:"notify_all_changes" scope:"user,admin"` // This Variable is a simple copy of `core.CoreApp.UpdateNotify.Bool`
	DownText            string                `gorm:"-" json:"-" yaml:"-"`                                                                                           // Contains the current generated Downtime Text 	// Is 'true' if the user has already be informed that the Services now again available // Is 'true' if the user has already be informed that the Services now again available
	LastStatusCode      int                   `gorm:"-" json:"status_code" yaml:"-"`
	LastLookupTime      int64                 `gorm:"-" json:"-" yaml:"-"`
	LastLatency         int64                 `gorm:"-" json:"-" yaml:"-"`
	LastCheck           time.Time             `gorm:"-" json:"-" yaml:"-"`
	LastOnline          time.Time             `gorm:"-" json:"last_success" yaml:"-"`
	LastOffline         time.Time             `gorm:"-" json:"last_error" yaml:"-"`
	Stats               *Stats                `gorm:"-" json:"stats,omitempty" yaml:"-"`
	Messages            []*messages.Message   `gorm:"foreignKey:Service;references:Id" json:"messages,omitempty" yaml:"messages"`
	Incidents           []*incidents.Incident `gorm:"foreignKey:Service;references:Id" json:"incidents,omitempty" yaml:"incidents"`
	Checkins            []*checkins.Checkin   `gorm:"foreignKey:Service;references:Id" json:"checkins,omitempty" yaml:"-" scope:"user,admin"`
	Failures            []*failures.Failure   `gorm:"-" json:"failures,omitempty" yaml:"-" scope:"user,admin"`

	notifyAfterCount int64 `gorm:"-" json:"-" yaml:"-"`
	prevOnline       bool  `gorm:"-" json:"-" yaml:"-"`
}

// RLock acquires a read lock on the service's runtime fields.
// Use this before reading fields that workers may be updating.
func (s *Service) RLock() {
	s.fieldsMu.RLock()
}

// RUnlock releases the read lock on the service's runtime fields.
func (s *Service) RUnlock() {
	s.fieldsMu.RUnlock()
}

// Lock acquires a write lock on the service's runtime fields.
// Use this before updating fields that may be read concurrently.
func (s *Service) Lock() {
	s.fieldsMu.Lock()
}

// Unlock releases the write lock on the service's runtime fields.
func (s *Service) Unlock() {
	s.fieldsMu.Unlock()
}

// serviceAlias is used by MarshalJSON to avoid infinite recursion
type serviceAlias Service

// MarshalJSON implements json.Marshaler with read lock protection.
// This prevents races between JSON serialization and worker updates.
func (s *Service) MarshalJSON() ([]byte, error) {
	s.fieldsMu.RLock()
	defer s.fieldsMu.RUnlock()
	return json.Marshal((*serviceAlias)(s))
}

// ServicePtrOrder sorts service pointers by Order field without copying mutex
type ServicePtrOrder []*Service

func (c ServicePtrOrder) Len() int           { return len(c) }
func (c ServicePtrOrder) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }
func (c ServicePtrOrder) Less(i, j int) bool { return c[i].Order < c[j].Order }

type Stats struct {
	Failures int       `gorm:"-" json:"failures"`
	Hits     int       `gorm:"-" json:"hits"`
	FirstHit time.Time `gorm:"-" json:"first_hit"`
}

type ser struct {
	Time   time.Time
	Online bool
}

type UptimeSeries struct {
	Start    time.Time `json:"start"`
	End      time.Time `json:"end"`
	Uptime   int64     `json:"uptime"`
	Downtime int64     `json:"downtime"`
	Series   []series  `json:"series"`
}

type ByTime []ser

func (a ByTime) Len() int           { return len(a) }
func (a ByTime) Less(i, j int) bool { return a[i].Time.Before(a[j].Time) }
func (a ByTime) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

type series struct {
	Start    time.Time `json:"start"`
	End      time.Time `json:"end"`
	Duration int64     `json:"duration"`
	Online   bool      `json:"online"`
}
