package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	Logger "github.com/sirupsen/logrus"
)

// LogShipper sends logs to external systems (Loki, Elasticsearch, Splunk, Cribl, webhook)
type LogShipper struct {
	enabled    bool
	shipType   string // "loki", "elasticsearch", "splunk", "cribl", "webhook"
	endpoint   string
	authToken  string
	splunkHEC  string // Splunk HEC token (separate from authToken for clarity)
	labels     map[string]string
	batchSize  int
	buffer     []logEntry
	mu         sync.Mutex
	client     *http.Client
	stopCh     chan struct{}
	sourceType string // For Splunk/Cribl
	index      string // For Splunk
}

type logEntry struct {
	Timestamp time.Time
	Level     string
	Message   string
	Fields    map[string]interface{}
}

var (
	shipper     *LogShipper
	shipperOnce sync.Once
)

// LogShipConfig holds log shipping configuration from either env vars or database
type LogShipConfig struct {
	Enabled    bool
	Type       string
	Endpoint   string
	Token      string
	Index      string
	Sourcetype string
	Labels     string
}

// InitLogShipper initializes log shipping based on environment variables
// LOG_SHIP_TYPE: loki, elasticsearch, splunk, cribl, webhook
// LOG_SHIP_ENDPOINT: URL to send logs to
// LOG_SHIP_TOKEN: Bearer token for authentication (or Splunk HEC token)
// LOG_SHIP_LABELS: comma-separated key=value pairs for labels/metadata
// LOG_SHIP_INDEX: Splunk index name (default: "main")
// LOG_SHIP_SOURCETYPE: Splunk/Cribl sourcetype (default: "statping")
func InitLogShipper() {
	InitLogShipperWithConfig(nil)
}

// InitLogShipperWithConfig initializes log shipping with optional config from database
// Pass nil to use environment variables only
// Pass config to use database settings as fallback (env vars take precedence)
func InitLogShipperWithConfig(dbConfig *LogShipConfig) {
	shipperOnce.Do(func() {
		// Environment variables take precedence
		shipType := Params.GetString("LOG_SHIP_TYPE")
		endpoint := Params.GetString("LOG_SHIP_ENDPOINT")
		token := Params.GetString("LOG_SHIP_TOKEN")
		labelStr := Params.GetString("LOG_SHIP_LABELS")
		sourceType := Params.GetString("LOG_SHIP_SOURCETYPE")
		index := Params.GetString("LOG_SHIP_INDEX")

		// Fall back to database config if env vars not set
		if dbConfig != nil && shipType == "" && endpoint == "" {
			if dbConfig.Enabled && dbConfig.Type != "" && dbConfig.Endpoint != "" {
				shipType = dbConfig.Type
				endpoint = dbConfig.Endpoint
				token = dbConfig.Token
				labelStr = dbConfig.Labels
				sourceType = dbConfig.Sourcetype
				index = dbConfig.Index
			}
		}

		if shipType == "" || endpoint == "" {
			return
		}

		labels := make(map[string]string)
		labels["app"] = "statping"
		labels["instance"] = Params.GetString("NAME")

		// Parse additional labels
		if labelStr != "" {
			for _, pair := range strings.Split(labelStr, ",") {
				kv := strings.SplitN(pair, "=", 2)
				if len(kv) == 2 {
					labels[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
				}
			}
		}

		if sourceType == "" {
			sourceType = "statping"
		}

		if index == "" {
			index = "main"
		}

		shipper = &LogShipper{
			enabled:    true,
			shipType:   strings.ToLower(shipType),
			endpoint:   endpoint,
			authToken:  token,
			labels:     labels,
			batchSize:  100,
			buffer:     make([]logEntry, 0, 100),
			client:     &http.Client{Timeout: 10 * time.Second},
			stopCh:     make(chan struct{}),
			sourceType: sourceType,
			index:      index,
		}

		// Start background flusher
		go shipper.flushLoop()

		Log.Infof("Log shipping enabled: %s -> %s", shipType, endpoint)
	})
}

// ReloadLogShipper stops and restarts log shipper with new config
func ReloadLogShipper(config *LogShipConfig) {
	// Stop existing shipper
	if shipper != nil && shipper.enabled {
		close(shipper.stopCh)
		shipper = nil
	}
	// Reset once so we can reinitialize
	shipperOnce = sync.Once{}

	// Reinitialize with new config
	InitLogShipperWithConfig(config)
}

// ShipLog adds a log entry to the shipping buffer
func ShipLog(entry *Logger.Entry) {
	if shipper == nil || !shipper.enabled {
		return
	}

	shipper.mu.Lock()
	defer shipper.mu.Unlock()

	shipper.buffer = append(shipper.buffer, logEntry{
		Timestamp: entry.Time,
		Level:     entry.Level.String(),
		Message:   entry.Message,
		Fields:    entry.Data,
	})

	// Flush if buffer is full
	if len(shipper.buffer) >= shipper.batchSize {
		go shipper.flush()
	}
}

func (s *LogShipper) flushLoop() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.flush()
		case <-s.stopCh:
			s.flush() // Final flush
			return
		}
	}
}

func (s *LogShipper) flush() {
	s.mu.Lock()
	if len(s.buffer) == 0 {
		s.mu.Unlock()
		return
	}
	entries := s.buffer
	s.buffer = make([]logEntry, 0, s.batchSize)
	s.mu.Unlock()

	var err error
	switch s.shipType {
	case "loki":
		err = s.sendToLoki(entries)
	case "elasticsearch":
		err = s.sendToElasticsearch(entries)
	case "splunk":
		err = s.sendToSplunk(entries)
	case "cribl":
		err = s.sendToCribl(entries)
	case "webhook":
		err = s.sendToWebhook(entries)
	default:
		err = fmt.Errorf("unknown log ship type: %s", s.shipType)
	}

	if err != nil {
		// Don't use Log here to avoid infinite loop
		fmt.Printf("Log shipping error: %v\n", err)
	}
}

// Loki push format
type lokiPush struct {
	Streams []lokiStream `json:"streams"`
}

type lokiStream struct {
	Stream map[string]string `json:"stream"`
	Values [][]string        `json:"values"`
}

func (s *LogShipper) sendToLoki(entries []logEntry) error {
	// Build Loki push payload
	values := make([][]string, len(entries))
	for i, e := range entries {
		// Loki expects [timestamp_ns, log_line]
		ts := fmt.Sprintf("%d", e.Timestamp.UnixNano())
		line := e.Message
		if len(e.Fields) > 0 {
			fieldsJSON, _ := json.Marshal(e.Fields)
			line = fmt.Sprintf("[%s] %s %s", e.Level, e.Message, string(fieldsJSON))
		} else {
			line = fmt.Sprintf("[%s] %s", e.Level, e.Message)
		}
		values[i] = []string{ts, line}
	}

	push := lokiPush{
		Streams: []lokiStream{
			{
				Stream: s.labels,
				Values: values,
			},
		},
	}

	body, err := json.Marshal(push)
	if err != nil {
		return err
	}

	endpoint := strings.TrimSuffix(s.endpoint, "/") + "/loki/api/v1/push"
	return s.doRequest("POST", endpoint, body)
}

func (s *LogShipper) sendToElasticsearch(entries []logEntry) error {
	// Build bulk request
	var buf bytes.Buffer
	for _, e := range entries {
		// Index action
		action := map[string]interface{}{
			"index": map[string]interface{}{
				"_index": fmt.Sprintf("statping-logs-%s", e.Timestamp.Format("2006.01.02")),
			},
		}
		actionJSON, _ := json.Marshal(action)
		buf.Write(actionJSON)
		buf.WriteByte('\n')

		// Document
		doc := map[string]interface{}{
			"@timestamp": e.Timestamp.Format(time.RFC3339Nano),
			"level":      e.Level,
			"message":    e.Message,
			"app":        "statping",
		}
		for k, v := range e.Fields {
			doc[k] = v
		}
		for k, v := range s.labels {
			doc[k] = v
		}
		docJSON, _ := json.Marshal(doc)
		buf.Write(docJSON)
		buf.WriteByte('\n')
	}

	endpoint := strings.TrimSuffix(s.endpoint, "/") + "/_bulk"
	return s.doRequest("POST", endpoint, buf.Bytes())
}

// Splunk HEC (HTTP Event Collector) format
type splunkEvent struct {
	Time       float64                `json:"time"`
	Host       string                 `json:"host,omitempty"`
	Source     string                 `json:"source,omitempty"`
	Sourcetype string                 `json:"sourcetype,omitempty"`
	Index      string                 `json:"index,omitempty"`
	Event      map[string]interface{} `json:"event"`
}

func (s *LogShipper) sendToSplunk(entries []logEntry) error {
	var buf bytes.Buffer

	hostname := s.labels["instance"]
	if hostname == "" {
		hostname = "statping"
	}

	for _, e := range entries {
		event := splunkEvent{
			Time:       float64(e.Timestamp.UnixNano()) / 1e9,
			Host:       hostname,
			Source:     "statping",
			Sourcetype: s.sourceType,
			Index:      s.index,
			Event: map[string]interface{}{
				"level":   e.Level,
				"message": e.Message,
			},
		}

		// Add fields to event
		for k, v := range e.Fields {
			event.Event[k] = v
		}
		for k, v := range s.labels {
			event.Event[k] = v
		}

		eventJSON, err := json.Marshal(event)
		if err != nil {
			continue
		}
		buf.Write(eventJSON)
	}

	// Splunk HEC endpoint
	endpoint := strings.TrimSuffix(s.endpoint, "/")
	if !strings.HasSuffix(endpoint, "/services/collector/event") {
		endpoint += "/services/collector/event"
	}

	return s.doSplunkRequest(endpoint, buf.Bytes())
}

func (s *LogShipper) doSplunkRequest(url string, body []byte) error {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	// Splunk HEC uses "Splunk <token>" format
	if s.authToken != "" {
		req.Header.Set("Authorization", "Splunk "+s.authToken)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("splunk HEC failed: HTTP %d", resp.StatusCode)
	}

	return nil
}

func (s *LogShipper) sendToCribl(entries []logEntry) error {
	// Cribl HTTP source accepts JSON array or newline-delimited JSON
	// Using NDJSON format for better streaming support
	var buf bytes.Buffer

	hostname := s.labels["instance"]
	if hostname == "" {
		hostname = "statping"
	}

	for _, e := range entries {
		event := map[string]interface{}{
			"_time":      float64(e.Timestamp.UnixNano()) / 1e9,
			"_raw":       e.Message,
			"host":       hostname,
			"source":     "statping",
			"sourcetype": s.sourceType,
			"level":      e.Level,
		}

		// Add fields
		for k, v := range e.Fields {
			event[k] = v
		}
		for k, v := range s.labels {
			event[k] = v
		}

		eventJSON, err := json.Marshal(event)
		if err != nil {
			continue
		}
		buf.Write(eventJSON)
		buf.WriteByte('\n')
	}

	return s.doRequest("POST", s.endpoint, buf.Bytes())
}

func (s *LogShipper) sendToWebhook(entries []logEntry) error {
	// Simple JSON array of log entries
	payload := make([]map[string]interface{}, len(entries))
	for i, e := range entries {
		entry := map[string]interface{}{
			"timestamp": e.Timestamp.Format(time.RFC3339Nano),
			"level":     e.Level,
			"message":   e.Message,
			"app":       "statping",
		}
		for k, v := range e.Fields {
			entry[k] = v
		}
		for k, v := range s.labels {
			entry[k] = v
		}
		payload[i] = entry
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return s.doRequest("POST", s.endpoint, body)
}

func (s *LogShipper) doRequest(method, url string, body []byte) error {
	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	if s.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+s.authToken)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("log shipping failed: HTTP %d", resp.StatusCode)
	}

	return nil
}

// StopLogShipper gracefully stops the log shipper
func StopLogShipper() {
	if shipper != nil && shipper.enabled {
		close(shipper.stopCh)
	}
}
