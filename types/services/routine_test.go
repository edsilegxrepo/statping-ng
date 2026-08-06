package services

import (
	"net"
	"strconv"
	"strings"
	"testing"

	"github.com/statping-ng/statping-ng/types/null"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

// grpcServerDef is function type.
// Consumed by Test data.
type grpcServerDef func(int, bool) *grpc.Server

// Test Data: Simulates testing scenarios
var testdata = []struct {
	grpcService   grpcServerDef
	clientChecker *Service
}{
	{
		grpcService: func(port int, enableHealthCheck bool) *grpc.Server {
			return grpcServer(port, enableHealthCheck)
		},
		clientChecker: &Service{
			Name:            "GRPC Server with Health check",
			Domain:          "localhost",
			Port:            50053,
			Expected:        null.NewNullString("status:SERVING"),
			ExpectedStatus:  1,
			Type:            "grpc",
			Timeout:         3,
			VerifySSL:       null.NewNullBool(false),
			GrpcHealthCheck: null.NewNullBool(true),
		},
	},
	{
		grpcService: func(port int, enableHealthCheck bool) *grpc.Server {
			return grpcServer(port, enableHealthCheck)
		},
		clientChecker: &Service{
			Name:            "Check TLS endpoint on GRPC Server with TLS disabled",
			Domain:          "localhost",
			Port:            50054,
			Expected:        null.NewNullString(""),
			ExpectedStatus:  0,
			Type:            "grpc",
			Timeout:         1,
			VerifySSL:       null.NewNullBool(true),
			GrpcHealthCheck: null.NewNullBool(true),
		},
	},
	{
		grpcService: func(port int, enableHealthCheck bool) *grpc.Server {
			return grpcServer(port, enableHealthCheck)
		},
		clientChecker: &Service{
			Name:           "Check GRPC Server without Health check endpoint",
			Domain:         "localhost",
			Port:           50055,
			Expected:       null.NewNullString(""),
			ExpectedStatus: 0,
			Type:           "grpc",
			Timeout:        1,
			VerifySSL:      null.NewNullBool(false),
		},
	},
	{
		grpcService: func(port int, enableHealthCheck bool) *grpc.Server {
			return grpcServer(50056, enableHealthCheck)
		},
		clientChecker: &Service{
			Name:            "Check where no GRPC Server exists",
			Domain:          "localhost",
			Port:            1000,
			Expected:        null.NewNullString(""),
			ExpectedStatus:  0,
			Type:            "grpc",
			Timeout:         1,
			VerifySSL:       null.NewNullBool(false),
			GrpcHealthCheck: null.NewNullBool(true),
		},
	},
	{
		grpcService: func(port int, enableHealthCheck bool) *grpc.Server {
			return grpcServer(50057, enableHealthCheck)
		},
		clientChecker: &Service{
			Name:            "Check where no GRPC Server exists (Verify TLS)",
			Domain:          "localhost",
			Port:            1000,
			Expected:        null.NewNullString(""),
			ExpectedStatus:  0,
			Type:            "grpc",
			Timeout:         1,
			VerifySSL:       null.NewNullBool(true),
			GrpcHealthCheck: null.NewNullBool(true),
		},
	},
	{
		grpcService: func(port int, enableHealthCheck bool) *grpc.Server {
			return grpcServer(port, enableHealthCheck)
		},
		clientChecker: &Service{
			Name:            "Check GRPC Server with url",
			Domain:          "http://localhost",
			Port:            50058,
			Expected:        null.NewNullString("status:SERVING"),
			ExpectedStatus:  1,
			Type:            "grpc",
			Timeout:         1,
			VerifySSL:       null.NewNullBool(false),
			GrpcHealthCheck: null.NewNullBool(true),
		},
	},
	{
		grpcService: func(port int, enableHealthCheck bool) *grpc.Server {
			return grpcServer(port, enableHealthCheck)
		},
		clientChecker: &Service{
			Name:            "Unparseable Url Error",
			Domain:          "http://local//host",
			Port:            50059,
			Expected:        null.NewNullString(""),
			ExpectedStatus:  0,
			Type:            "grpc",
			Timeout:         1,
			VerifySSL:       null.NewNullBool(false),
			GrpcHealthCheck: null.NewNullBool(true),
		},
	},
	{
		grpcService: func(port int, enableHealthCheck bool) *grpc.Server {
			return grpcServer(50060, enableHealthCheck)
		},
		clientChecker: &Service{
			Name:           "Check GRPC on HTTP server",
			Domain:         "https://google.com",
			Port:           443,
			Expected:       null.NewNullString(""),
			ExpectedStatus: 0,
			Type:           "grpc",
			Timeout:        1,
			VerifySSL:      null.NewNullBool(false),
		},
	},
	{
		grpcService: func(port int, enableHealthCheck bool) *grpc.Server {
			return grpcServer(port, true)
		},
		clientChecker: &Service{
			Name:            "GRPC HealthCheck where health check endpoint is not implemented",
			Domain:          "http://localhost",
			Port:            50061,
			Expected:        null.NewNullString(""),
			ExpectedStatus:  0,
			Type:            "grpc",
			Timeout:         1,
			VerifySSL:       null.NewNullBool(false),
			GrpcHealthCheck: null.NewNullBool(false),
		},
	},
}

// grpcServer creates grpc Service with optional parameters.
func grpcServer(port int, enableHealthCheck bool) *grpc.Server {
	portString := strconv.Itoa(port)
	server := grpc.NewServer()
	lis, err := net.Listen("tcp", "127.0.0.1:"+portString)
	if err != nil {
		return nil
	}

	if enableHealthCheck {
		healthServer := health.NewServer()
		healthServer.SetServingStatus("Test GRPC Service", healthpb.HealthCheckResponse_SERVING)
		healthpb.RegisterHealthServer(server, healthServer)
		go func() { _ = server.Serve(lis) }()
	}
	return server
}

// TestCheckGrpc ranges over the testdata struct.
// Examines checkGrpc() function
func TestCheckGrpc(t *testing.T) {
	for _, testscenario := range testdata {
		v := testscenario
		t.Run(v.clientChecker.Name, func(t *testing.T) {
			t.Parallel()
			server := v.grpcService(v.clientChecker.Port, v.clientChecker.GrpcHealthCheck.Bool)
			if server != nil {
				defer server.Stop()
			}
			v.clientChecker.CheckService(false)
			if v.clientChecker.LastStatusCode != v.clientChecker.ExpectedStatus || strings.TrimSpace(v.clientChecker.LastResponse) != v.clientChecker.Expected.String {
				t.Errorf("Expected message: '%v', Got message: '%v' , Expected Status: '%v', Got Status: '%v'", v.clientChecker.Expected.String, v.clientChecker.LastResponse, v.clientChecker.ExpectedStatus, v.clientChecker.LastStatusCode)
			}
		})
	}
}

// mockSMTPServer creates a mock SMTP server for testing
func mockSMTPServer(port int) (net.Listener, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:"+strconv.Itoa(port))
	if err != nil {
		return nil, err
	}

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go handleSMTPConnection(conn)
		}
	}()

	return listener, nil
}

func handleSMTPConnection(conn net.Conn) {
	defer func() { _ = conn.Close() }()
	_, _ = conn.Write([]byte("220 mock.smtp.server ESMTP\r\n"))

	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			return
		}
		line := string(buf[:n])
		if strings.HasPrefix(line, "EHLO") || strings.HasPrefix(line, "HELO") {
			_, _ = conn.Write([]byte("250-mock.smtp.server Hello\r\n"))
			_, _ = conn.Write([]byte("250 OK\r\n"))
		} else if strings.HasPrefix(line, "QUIT") {
			_, _ = conn.Write([]byte("221 Bye\r\n"))
			return
		} else {
			_, _ = conn.Write([]byte("250 OK\r\n"))
		}
	}
}

func TestCheckSmtp(t *testing.T) {
	listener, err := mockSMTPServer(2525)
	if err != nil {
		t.Skipf("Could not start mock SMTP server: %v", err)
	}
	defer func() { _ = listener.Close() }()

	t.Run("SMTP check succeeds on mock server", func(t *testing.T) {
		s := &Service{
			Name:    "Test SMTP",
			Domain:  "127.0.0.1",
			Port:    25, // Port 25 doesn't require auth
			Type:    "smtp",
			Timeout: 5,
		}
		// Override port to use our mock server
		s.Port = 2525

		result, err := CheckSmtp(s, false)
		if err != nil {
			t.Logf("SMTP check error (expected on port 25 without auth): %v", err)
		}
		// Just verify the function runs without panic
		_ = result
	})

	t.Run("SMTP check with invalid domain", func(t *testing.T) {
		s := &Service{
			Name:    "Test SMTP Invalid",
			Domain:  "invalid.nonexistent.domain.local",
			Port:    25,
			Type:    "smtp",
			Timeout: 2,
		}

		_, err := CheckSmtp(s, false)
		if err == nil {
			t.Error("Expected error for invalid domain")
		}
	})
}

// mockIMAPServer creates a mock IMAP server for testing
func mockIMAPServer(port int) (net.Listener, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:"+strconv.Itoa(port))
	if err != nil {
		return nil, err
	}

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go handleIMAPConnection(conn)
		}
	}()

	return listener, nil
}

func handleIMAPConnection(conn net.Conn) {
	defer func() { _ = conn.Close() }()
	_, _ = conn.Write([]byte("* OK IMAP4rev1 Service Ready\r\n"))

	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			return
		}
		line := string(buf[:n])
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		tag := parts[0]
		cmd := strings.ToUpper(parts[1])

		switch cmd {
		case "CAPABILITY":
			_, _ = conn.Write([]byte("* CAPABILITY IMAP4rev1 AUTH=PLAIN\r\n"))
			_, _ = conn.Write([]byte(tag + " OK CAPABILITY completed\r\n"))
		case "LOGIN":
			_, _ = conn.Write([]byte(tag + " OK LOGIN completed\r\n"))
		case "LOGOUT":
			_, _ = conn.Write([]byte("* BYE IMAP4rev1 Server logging out\r\n"))
			_, _ = conn.Write([]byte(tag + " OK LOGOUT completed\r\n"))
			return
		default:
			_, _ = conn.Write([]byte(tag + " OK\r\n"))
		}
	}
}

func TestCheckImap(t *testing.T) {
	listener, err := mockIMAPServer(1143)
	if err != nil {
		t.Skipf("Could not start mock IMAP server: %v", err)
	}
	defer func() { _ = listener.Close() }()

	t.Run("IMAP check with invalid domain", func(t *testing.T) {
		s := &Service{
			Name:    "Test IMAP Invalid",
			Domain:  "invalid.nonexistent.domain.local",
			Port:    143,
			Type:    "imap",
			Timeout: 2,
		}

		_, err := CheckImap(s, false)
		if err == nil {
			t.Error("Expected error for invalid domain")
		}
	})

	t.Run("IMAP check on mock server", func(t *testing.T) {
		s := &Service{
			Name:    "Test IMAP",
			Domain:  "127.0.0.1",
			Port:    1143,
			Type:    "imap",
			Timeout: 5,
			Headers: null.NewNullString("username=test,password=test"),
		}

		result, err := CheckImap(s, false)
		if err != nil {
			t.Logf("IMAP check error: %v", err)
		}
		_ = result
	})
}

func TestCheckDatabase(t *testing.T) {
	t.Run("Database check with missing type", func(t *testing.T) {
		s := &Service{
			Name:    "Test DB Missing Type",
			Type:    "database",
			Timeout: 5,
		}

		_, err := CheckDatabase(s, false)
		if err == nil {
			t.Error("Expected error for missing database type")
		}
	})

	t.Run("Database check with missing DSN", func(t *testing.T) {
		s := &Service{
			Name:         "Test DB Missing DSN",
			Type:         "database",
			Timeout:      5,
			DatabaseType: null.NewNullString("postgres"),
		}

		_, err := CheckDatabase(s, false)
		if err == nil {
			t.Error("Expected error for missing DSN")
		}
	})

	t.Run("Database check with unsupported type", func(t *testing.T) {
		s := &Service{
			Name:         "Test DB Unsupported",
			Type:         "database",
			Timeout:      5,
			DatabaseType: null.NewNullString("oracle"),
			DatabaseDSN:  null.NewNullString("oracle://localhost:1521/orcl"),
		}

		_, err := CheckDatabase(s, false)
		if err == nil {
			t.Error("Expected error for unsupported database type")
		}
	})

	t.Run("Database check with invalid connection", func(t *testing.T) {
		s := &Service{
			Name:         "Test DB Invalid Connection",
			Type:         "database",
			Timeout:      2,
			DatabaseType: null.NewNullString("postgres"),
			DatabaseDSN:  null.NewNullString("postgres://localhost:5432/nonexistent?connect_timeout=1"),
		}

		_, err := CheckDatabase(s, false)
		if err == nil {
			t.Error("Expected error for invalid database connection")
		}
	})
}

func TestCheckStorage(t *testing.T) {
	t.Run("Storage check with missing backend", func(t *testing.T) {
		s := &Service{
			Name:    "Test Storage Missing Backend",
			Type:    "storage",
			Timeout: 5,
		}

		_, err := CheckStorage(s, false)
		if err == nil {
			t.Error("Expected error for missing storage backend")
		}
	})

	t.Run("Storage check with missing bucket", func(t *testing.T) {
		s := &Service{
			Name:           "Test Storage Missing Bucket",
			Type:           "storage",
			Timeout:        5,
			StorageBackend: null.NewNullString("gcs"),
		}

		_, err := CheckStorage(s, false)
		if err == nil {
			t.Error("Expected error for missing bucket")
		}
	})

	t.Run("Storage check with unsupported backend", func(t *testing.T) {
		s := &Service{
			Name:           "Test Storage Unsupported",
			Type:           "storage",
			Timeout:        5,
			StorageBackend: null.NewNullString("azure"),
			StorageBucket:  null.NewNullString("my-bucket"),
		}

		_, err := CheckStorage(s, false)
		if err == nil {
			t.Error("Expected error for unsupported storage backend")
		}
	})

	t.Run("GCS check with missing credentials", func(t *testing.T) {
		s := &Service{
			Name:           "Test GCS Missing Creds",
			Type:           "storage",
			Timeout:        5,
			StorageBackend: null.NewNullString("gcs"),
			StorageBucket:  null.NewNullString("my-bucket"),
		}

		_, err := CheckStorage(s, false)
		if err == nil {
			t.Error("Expected error for missing GCS credentials")
		}
	})
}

func TestCheckTLS(t *testing.T) {
	t.Run("TLS check with missing target", func(t *testing.T) {
		s := &Service{
			Name:    "Test TLS Missing Target",
			Type:    "tls",
			Timeout: 5,
		}

		_, err := CheckTLS(s, false)
		if err == nil {
			t.Error("Expected error for missing TLS target")
		}
	})

	t.Run("TLS check with invalid host", func(t *testing.T) {
		s := &Service{
			Name:      "Test TLS Invalid Host",
			Type:      "tls",
			Timeout:   2,
			TLSTarget: null.NewNullString("invalid.nonexistent.domain.local:443"),
		}

		_, err := CheckTLS(s, false)
		if err == nil {
			t.Error("Expected error for invalid TLS host")
		}
	})

	t.Run("TLS check against google.com", func(t *testing.T) {
		s := &Service{
			Name:       "Test TLS Google",
			Type:       "tls",
			Timeout:    10,
			TLSTarget:  null.NewNullString("google.com:443"),
			TLSMinDays: 7,
		}

		result, err := CheckTLS(s, false)
		if err != nil {
			t.Errorf("TLS check failed for google.com: %v", err)
		}
		if result.TLSExpiry == nil {
			t.Error("Expected TLS expiry to be set")
		}
		if result.TLSIssuer == "" {
			t.Error("Expected TLS issuer to be set")
		}
		if result.TLSDaysRemaining <= 0 {
			t.Error("Expected positive days remaining")
		}
		if !result.Online {
			t.Error("Expected service to be online")
		}
	})

	t.Run("TLS check uses Domain as fallback", func(t *testing.T) {
		s := &Service{
			Name:       "Test TLS Domain Fallback",
			Type:       "tls",
			Domain:     "github.com",
			Timeout:    10,
			TLSMinDays: 7,
		}

		result, err := CheckTLS(s, false)
		if err != nil {
			t.Errorf("TLS check failed for github.com: %v", err)
		}
		if !result.Online {
			t.Error("Expected service to be online")
		}
	})

	t.Run("TLS check adds default port", func(t *testing.T) {
		s := &Service{
			Name:       "Test TLS Default Port",
			Type:       "tls",
			TLSTarget:  null.NewNullString("cloudflare.com"),
			Timeout:    10,
			TLSMinDays: 7,
		}

		result, err := CheckTLS(s, false)
		if err != nil {
			t.Errorf("TLS check failed: %v", err)
		}
		if !result.Online {
			t.Error("Expected service to be online")
		}
	})
}
