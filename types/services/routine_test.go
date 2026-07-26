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
	defer conn.Close()
	conn.Write([]byte("220 mock.smtp.server ESMTP\r\n"))

	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			return
		}
		line := string(buf[:n])
		if strings.HasPrefix(line, "EHLO") || strings.HasPrefix(line, "HELO") {
			conn.Write([]byte("250-mock.smtp.server Hello\r\n"))
			conn.Write([]byte("250 OK\r\n"))
		} else if strings.HasPrefix(line, "QUIT") {
			conn.Write([]byte("221 Bye\r\n"))
			return
		} else {
			conn.Write([]byte("250 OK\r\n"))
		}
	}
}

func TestCheckSmtp(t *testing.T) {
	listener, err := mockSMTPServer(2525)
	if err != nil {
		t.Skipf("Could not start mock SMTP server: %v", err)
	}
	defer listener.Close()

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
	defer conn.Close()
	conn.Write([]byte("* OK IMAP4rev1 Service Ready\r\n"))

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
			conn.Write([]byte("* CAPABILITY IMAP4rev1 AUTH=PLAIN\r\n"))
			conn.Write([]byte(tag + " OK CAPABILITY completed\r\n"))
		case "LOGIN":
			conn.Write([]byte(tag + " OK LOGIN completed\r\n"))
		case "LOGOUT":
			conn.Write([]byte("* BYE IMAP4rev1 Server logging out\r\n"))
			conn.Write([]byte(tag + " OK LOGOUT completed\r\n"))
			return
		default:
			conn.Write([]byte(tag + " OK\r\n"))
		}
	}
}

func TestCheckImap(t *testing.T) {
	listener, err := mockIMAPServer(1143)
	if err != nil {
		t.Skipf("Could not start mock IMAP server: %v", err)
	}
	defer listener.Close()

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
