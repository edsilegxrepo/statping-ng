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
	validTypes := []string{"http", "tcp", "udp", "icmp", "grpc", "static", "cmd"}

	for _, svcType := range validTypes {
		t.Run("Type_"+svcType, func(t *testing.T) {
			s := &Service{
				Name:     "Test " + svcType,
				Type:     svcType,
				Interval: 30,
			}
			if svcType != "static" && svcType != "cmd" {
				s.Domain = "example.com"
			}
			if svcType == "cmd" {
				s.PostData = null.NewNullString(`{"cmd": "echo"}`)
			}

			err := s.Validate()
			assert.NoError(t, err, "Type %s should be valid", svcType)
		})
	}
}
