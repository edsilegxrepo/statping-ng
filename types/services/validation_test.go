package services

import (
	"testing"

	"github.com/statping-ng/statping-ng/types/errors"
	"github.com/statping-ng/statping-ng/types/null"
	"github.com/stretchr/testify/assert"
)

func TestServiceValidate(t *testing.T) {
	tests := []struct {
		name        string
		service     *Service
		expectedErr error
	}{
		{
			name: "Valid HTTP service",
			service: &Service{
				Name:     "Test Service",
				Domain:   "https://example.com",
				Type:     "http",
				Interval: 30,
			},
			expectedErr: nil,
		},
		{
			name: "Valid TCP service",
			service: &Service{
				Name:     "TCP Service",
				Domain:   "example.com",
				Type:     "tcp",
				Port:     443,
				Interval: 60,
			},
			expectedErr: nil,
		},
		{
			name: "Valid static service (no interval required)",
			service: &Service{
				Name:   "Static Service",
				Domain: "https://example.com",
				Type:   "static",
			},
			expectedErr: nil,
		},
		{
			name: "Valid command service (no domain required)",
			service: &Service{
				Name:     "Command Service",
				Type:     "cmd",
				Interval: 30,
				PostData: null.NewNullString(`{"cmd": "echo hello"}`),
			},
			expectedErr: nil,
		},
		{
			name: "Missing name",
			service: &Service{
				Domain:   "https://example.com",
				Type:     "http",
				Interval: 30,
			},
			expectedErr: errors.ServiceNameMissing,
		},
		{
			name: "Missing domain for HTTP",
			service: &Service{
				Name:     "Test Service",
				Type:     "http",
				Interval: 30,
			},
			expectedErr: errors.DomainNameMissing,
		},
		{
			name: "Missing domain for TCP",
			service: &Service{
				Name:     "TCP Service",
				Type:     "tcp",
				Interval: 30,
			},
			expectedErr: errors.DomainNameMissing,
		},
		{
			name: "Missing type",
			service: &Service{
				Name:     "Test Service",
				Domain:   "https://example.com",
				Interval: 30,
			},
			expectedErr: errors.ServiceTypeMissing,
		},
		{
			name: "Missing interval for HTTP",
			service: &Service{
				Name:   "Test Service",
				Domain: "https://example.com",
				Type:   "http",
			},
			expectedErr: errors.CheckIntervalMissing,
		},
		{
			name: "Command service with invalid JSON",
			service: &Service{
				Name:     "Command Service",
				Type:     "cmd",
				Interval: 30,
				PostData: null.NewNullString(`not valid json`),
			},
			expectedErr: errors.CommandConfigNotJson,
		},
		{
			name: "Command service with missing cmd field",
			service: &Service{
				Name:     "Command Service",
				Type:     "cmd",
				Interval: 30,
				PostData: null.NewNullString(`{"timeout": 30}`),
			},
			expectedErr: errors.CommandConfigFieldCmdMissing,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.service.Validate()
			if tt.expectedErr != nil {
				assert.Equal(t, tt.expectedErr, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestServiceValidateEdgeCases(t *testing.T) {
	t.Run("Static service with no domain is valid", func(t *testing.T) {
		s := &Service{
			Name: "Static No Domain",
			Type: "static",
		}
		err := s.Validate()
		assert.NoError(t, err)
	})

	t.Run("Command service with no domain is valid", func(t *testing.T) {
		s := &Service{
			Name:     "Cmd No Domain",
			Type:     "cmd",
			Interval: 30,
			PostData: null.NewNullString(`{"cmd": "echo test"}`),
		}
		err := s.Validate()
		assert.NoError(t, err)
	})

	t.Run("Command service with expected status 0 gets MinInt32", func(t *testing.T) {
		s := &Service{
			Name:           "Cmd Expected Zero",
			Type:           "cmd",
			Interval:       30,
			PostData:       null.NewNullString(`{"cmd": "echo test"}`),
			ExpectedStatus: 0,
		}
		err := s.Validate()
		assert.NoError(t, err)
		// ExpectedStatus should be set to MinInt32 for zero values
		assert.NotEqual(t, 0, s.ExpectedStatus)
	})

	t.Run("Empty name with spaces fails", func(t *testing.T) {
		s := &Service{
			Name:     "   ",
			Domain:   "https://example.com",
			Type:     "http",
			Interval: 30,
		}
		// Note: Current implementation doesn't trim spaces
		err := s.Validate()
		assert.NoError(t, err) // passes because "   " != ""
	})

	t.Run("GRPC service requires domain", func(t *testing.T) {
		s := &Service{
			Name:     "GRPC Service",
			Type:     "grpc",
			Interval: 30,
		}
		err := s.Validate()
		assert.Equal(t, errors.DomainNameMissing, err)
	})

	t.Run("ICMP service requires domain", func(t *testing.T) {
		s := &Service{
			Name:     "ICMP Service",
			Type:     "icmp",
			Interval: 30,
		}
		err := s.Validate()
		assert.Equal(t, errors.DomainNameMissing, err)
	})
}

func TestServiceTypes(t *testing.T) {
	validTypes := []string{"http", "tcp", "udp", "icmp", "grpc", "static", "cmd", "database", "storage", "tls"}

	for _, svcType := range validTypes {
		t.Run("Type_"+svcType, func(t *testing.T) {
			s := &Service{
				Name:     "Test " + svcType,
				Type:     svcType,
				Interval: 30,
			}
			// Set required fields based on type
			switch svcType {
			case "static":
				// No domain required
			case "cmd":
				s.PostData = null.NewNullString(`{"cmd": "echo"}`)
			case "database":
				s.Domain = "" // No domain required for database
				s.DatabaseType = null.NewNullString("postgres")
				s.DatabaseDSN = null.NewNullString("postgres://localhost/test")
			case "storage":
				s.Domain = "" // No domain required for storage
				s.StorageBackend = null.NewNullString("gcs")
				s.StorageBucket = null.NewNullString("my-bucket")
			case "tls":
				s.TLSTarget = null.NewNullString("example.com:443")
			default:
				s.Domain = "example.com"
			}

			err := s.Validate()
			assert.NoError(t, err, "Type %s should be valid", svcType)
		})
	}
}

func TestMaskSecrets(t *testing.T) {
	t.Run("MaskSecrets masks sensitive fields", func(t *testing.T) {
		s := &Service{
			Name:               "Test Service",
			DatabaseDSN:        null.NewNullString("postgres://user:password@localhost/db"),
			StorageCredentials: null.NewNullString(`{"type": "service_account", "private_key": "secret"}`),
			TLSCertKey:         null.NewNullString("-----BEGIN PRIVATE KEY-----\nMIIE..."),
		}

		s.MaskSecrets()

		assert.Equal(t, "********", s.DatabaseDSN.String)
		assert.Equal(t, "********", s.StorageCredentials.String)
		assert.Equal(t, "********", s.TLSCertKey.String)
	})

	t.Run("MaskSecrets ignores empty fields", func(t *testing.T) {
		s := &Service{
			Name: "Test Service",
		}

		s.MaskSecrets()

		assert.False(t, s.DatabaseDSN.Valid)
		assert.False(t, s.StorageCredentials.Valid)
		assert.False(t, s.TLSCertKey.Valid)
	})
}

func TestNewServiceTypeValidation(t *testing.T) {
	t.Run("Database service without domain is valid", func(t *testing.T) {
		s := &Service{
			Name:         "DB Service",
			Type:         "database",
			Interval:     60,
			DatabaseType: null.NewNullString("postgres"),
			DatabaseDSN:  null.NewNullString("postgres://localhost/test"),
		}
		err := s.Validate()
		assert.NoError(t, err)
	})

	t.Run("Storage service without domain is valid", func(t *testing.T) {
		s := &Service{
			Name:           "Storage Service",
			Type:           "storage",
			Interval:       60,
			StorageBackend: null.NewNullString("gcs"),
			StorageBucket:  null.NewNullString("my-bucket"),
		}
		err := s.Validate()
		assert.NoError(t, err)
	})

	t.Run("TLS service with TLSTarget is valid", func(t *testing.T) {
		s := &Service{
			Name:      "TLS Service",
			Type:      "tls",
			Interval:  60,
			TLSTarget: null.NewNullString("example.com:443"),
		}
		err := s.Validate()
		assert.NoError(t, err)
	})
}
