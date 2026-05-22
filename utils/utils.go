package utils

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-ping/ping" //nolint:staticcheck
	"github.com/statping-ng/statping-ng/types/metrics"
)

// Directory returns the current path or the STATPING_DIR environment variable
var Directory string

var (
	transports = make(map[string]*http.Transport)
	transMu    sync.RWMutex
	DNSResolver = &net.Resolver{
		PreferGo: true,
	}
)

func getTransport(verifySSL bool, customTLS *tls.Config) *http.Transport {
	key := fmt.Sprintf("ssl-%v", verifySSL)
	if customTLS != nil {
		// Use a more stable key than pointer for custom TLS
		key = fmt.Sprintf("ssl-%v-tls-%p", verifySSL, customTLS)
	}

	// Double-checked locking with RLock for high-concurrency performance
	transMu.RLock()
	if t, ok := transports[key]; ok {
		transMu.RUnlock()
		return t
	}
	transMu.RUnlock()

	transMu.Lock()
	defer transMu.Unlock()

	// Final check inside write lock
	if t, ok := transports[key]; ok {
		return t
	}

	t := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: !verifySSL,
			Renegotiation:      tls.RenegotiateOnceAsClient,
		},
		DisableKeepAlives:     false,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   10,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		Proxy:                 http.ProxyFromEnvironment,
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			dialer := &net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
				Resolver:  DNSResolver,
			}
			return dialer.DialContext(ctx, network, addr)
		},
	}

	if customTLS != nil {
		t.TLSClientConfig.RootCAs = customTLS.RootCAs
		t.TLSClientConfig.Certificates = customTLS.Certificates
	}

	transports[key] = t
	return t
}

func NotNumber(val string) bool {
	_, err := strconv.ParseInt(val, 10, 64)
	return err != nil
}

// ToInt converts a int to a string
func ToInt(s interface{}) int64 {
	switch v := s.(type) {
	case string:
		val, _ := strconv.Atoi(v)
		return int64(val)
	case []byte:
		val, _ := strconv.Atoi(string(v))
		return int64(val)
	case float32:
		return int64(v)
	case float64:
		return int64(v)
	case int:
		return int64(v)
	case int16:
		return int64(v)
	case int32:
		return int64(v)
	case int64:
		return v
	case uint:
		if v > math.MaxInt64 {
			return math.MaxInt64
		}
		return int64(v)
	default:
		return 0
	}
}

// ToString converts a int to a string
func ToString(s interface{}) string {
	switch v := s.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%v", v)
	case float32, float64:
		return fmt.Sprintf("%f", v)
	case []byte:
		return string(v)
	case bool:
		return fmt.Sprintf("%t", v)
	case time.Time:
		return v.Format("Monday January _2, 2006 at 03:04PM")
	case time.Duration:
		return v.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}

// Command will run a terminal command with 'sh -c COMMAND' and return stdout and errOut as strings
//
//	in, out, err := Command("sass assets/scss assets/css/base.css")
func Command(name string, args ...string) (string, string, error) {
	testCmd := exec.Command(name, args...) // #nosec G204
	var stdout, stderr []byte
	var errStdout, errStderr error
	stdoutIn, _ := testCmd.StdoutPipe()
	stderrIn, _ := testCmd.StderrPipe()
	err := testCmd.Start()
	if err != nil {
		return "", "", err
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		stdout, errStdout = copyAndCapture(os.Stdout, stdoutIn)
	}()

	stderr, errStderr = copyAndCapture(os.Stderr, stderrIn)

	// call testCmd.Wait() only after reads from all pipes have completed
	wg.Wait()
	err = testCmd.Wait()
	if err != nil {
		return string(stdout), string(stderr), err
	}

	if errStdout != nil || errStderr != nil {
		return string(stdout), string(stderr), errors.New("failed to capture stdout or stderr")
	}

	outStr, errStr := string(stdout), string(stderr)
	return outStr, errStr, err
}

// copyAndCapture will read a terminal command into bytes
func copyAndCapture(w io.Writer, r io.Reader) ([]byte, error) {
	var out []byte
	buf := make([]byte, 1024)
	for {
		n, err := r.Read(buf[:])
		if n > 0 {
			d := buf[:n]
			out = append(out, d...)
			_, err := w.Write(d)
			if err != nil {
				return out, err
			}
		}
		if err != nil {
			// Read returns io.EOF at the end of file, which is not an error for us
			if err == io.EOF {
				err = nil
			}
			return out, err
		}
	}
}

// DurationReadable will turn a time.Duration into a human readable string
// // t := time.Duration(5 * time.Minute)
// // DurationReadable(t)
// // returns: 5 minutes
func DurationReadable(d time.Duration) string {
	if d.Hours() >= 1 {
		return fmt.Sprintf("%0.0f hours", d.Hours())
	} else if d.Minutes() >= 1 {
		return fmt.Sprintf("%0.0f minutes", d.Minutes())
	} else if d.Seconds() >= 1 {
		return fmt.Sprintf("%0.0f seconds", d.Seconds())
	}
	return d.String()
}

// HumanMicro will turn a microsecond duration value into a human readable string
// // t := 42
// // HumanMicro(t)
// // returns: 42 μs
//
// // t := 45619
// // HumanMicro(t)
// // returns: 45 ms
func HumanMicro(val int64) string {
	if (val > 0 && val < 10000) || (val < 0 && val > -10000) {
		return fmt.Sprintf("%d μs", val)
	}
	return fmt.Sprintf("%0.0f ms", float64(val)*0.001)
}

// HttpRequest is a global function to send a HTTP request
// // url - The URL for HTTP request
// // method - GET, POST, DELETE, PATCH
// // content - The HTTP request content type (text/plain, application/json, or nil)
// // headers - An array of Headers to be sent (KEY=VALUE) []string{"Authentication=12345", ...}
// // body - The body or form data to send with HTTP request
// // timeout - Specific duration to timeout on. time.Duration(30 * time.Seconds)
// // You can use a HTTP Proxy if you HTTP_PROXY environment variable
func HttpRequest(endpoint, method string, contentType interface{}, headers []string, body io.Reader, timeout time.Duration, verifySSL bool, customTLS *tls.Config) ([]byte, *http.Response, error) {
	var err error
	var req *http.Request
	if method == "" {
		method = "GET"
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	t1 := Now()
	if req, err = http.NewRequestWithContext(ctx, method, endpoint, body); err != nil {
		return nil, nil, err
	}

	// Anti-caching headers (CRITICAL for monitoring)
	req.Header.Set("Cache-Control", "no-cache, no-store, must-revalidate")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Expires", "0")

	// set default headers so end user can overwrite them if needed
	req.Header.Set("User-Agent", "Statping-ng")
	req.Header.Set("Statping-Version", Params.GetString("VERSION"))
	req.Header.Set("Accept", "text/html")

	if contentType != nil {
		req.Header.Set("Content-Type", contentType.(string))
	}

	verifyHost := req.URL.Hostname()
	for _, h := range headers {
		keyVal := strings.SplitN(h, "=", 2)
		if len(keyVal) == 2 {
			if keyVal[0] != "" && keyVal[1] != "" {
				if strings.ToLower(keyVal[0]) == "host" {
					req.Host = strings.TrimSpace(keyVal[1])
					verifyHost = req.Host
				} else {
					req.Header.Set(keyVal[0], keyVal[1])
				}
			}
		}
	}

	transport := getTransport(verifySSL, customTLS)
	client := &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}

	// For custom Host headers, we need to handle SNI carefully.
	// If verifyHost (from Host header) differs from URL Host, we use a one-off transport
	// to ensure thread-safety and correct SNI without polluting the shared pool.
	if verifyHost != req.URL.Hostname() {
		tr := transport.Clone()
		tr.TLSClientConfig.ServerName = verifyHost
		client.Transport = tr
	}

	var resp *http.Response
	if req.Header.Get("Redirect") != "true" {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
		req.Header.Del("Redirect")
	}

	if resp, err = client.Do(req); err != nil {
		return nil, resp, err
	}
	defer func() { _ = resp.Body.Close() }()

	// Bounded Reader: Prevent OOM by capping response size to 10MB
	// Most status pages/APIs are < 100KB, 10MB is very safe.
	limitReader := io.LimitReader(resp.Body, 10*1024*1024)
	contents, err := io.ReadAll(limitReader)
	if err != nil {
		return nil, resp, err
	}

	// record HTTP metrics
	metrics.Histo("bytes", float64(len(contents)), endpoint, method)
	metrics.Histo("duration", Now().Sub(t1).Seconds(), endpoint, method)

	return contents, resp, err
}

func Ping(address string, secondsTimeout int) (int64, error) {
	ping, err := ping.NewPinger(address)
	if err != nil {
		return 0, err
	}

	ping.Count = 1
	ping.Timeout = time.Second * time.Duration(secondsTimeout)

	if runtime.GOOS == "windows" {
		ping.SetPrivileged(true)
	}

	err = ping.Run()
	if err != nil {
		return 0, err
	}

	stats := ping.Statistics()

	if stats.PacketsSent-stats.PacketsRecv != 0 {
		return 0, errors.New("destination host unreachable")
	}

	return stats.MinRtt.Microseconds(), nil
}
