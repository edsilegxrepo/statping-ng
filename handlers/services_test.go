package handlers

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/statping-ng/statping-ng/types"
	"github.com/statping-ng/statping-ng/types/core"
	"github.com/statping-ng/statping-ng/types/services"
	"github.com/statping-ng/statping-ng/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnAuthenticatedServicesRoutes(t *testing.T) {
	tests := []HTTPTest{
		{
			Name:           "No Authentication - New Service",
			URL:            "/api/services",
			Method:         "POST",
			ExpectedStatus: 403, // CSRF rejects before auth check
			NoAuth:         true,
		},
		{
			Name:           "No Authentication - Update Service",
			URL:            "/api/services/1",
			Method:         "POST",
			ExpectedStatus: 403, // CSRF rejects before auth check
			NoAuth:         true,
		},
		{
			Name:           "No Authentication - Delete Service",
			URL:            "/api/services/1",
			Method:         "DELETE",
			ExpectedStatus: 403, // CSRF rejects before auth check
			NoAuth:         true,
		},
		{
			Name:           "No Authentication - Delete Service Hits",
			URL:            "/api/services/1/hits",
			Method:         "DELETE",
			ExpectedStatus: 403, // CSRF rejects before auth check
			NoAuth:         true,
		},
		{
			Name:           "No Authentication - Delete Service Failures",
			URL:            "/api/services/1/failures",
			Method:         "DELETE",
			ExpectedStatus: 403, // CSRF rejects before auth check
			NoAuth:         true,
		},
		{
			Name:           "No Authentication - Reorder Services",
			URL:            "/api/reorder/services",
			Method:         "POST",
			ExpectedStatus: 403, // CSRF rejects before auth check
			NoAuth:         true,
		},
	}

	for _, v := range tests {
		t.Run(v.Name, func(t *testing.T) {
			str, t, err := RunHTTPTest(v, t)
			t.Logf("Test %s: \n %v\n", v.Name, str)
			assert.Nil(t, err)
		})
	}
}

func TestServicesRoutes(t *testing.T) {
	ensureHandlerSetup(t)
	since := utils.Now().Add(-30 * types.Day)
	end := utils.Now().Add(-30 * time.Minute)
	startEndQuery := fmt.Sprintf("?start=%d&end=%d", since.Unix(), end.Unix()+15)

	tests := []HTTPTest{
		{
			Name:             "Statping All Public and Private Services",
			URL:              "/api/services",
			Method:           "GET",
			HttpHeaders:      []string{"Authorization=" + core.App.ApiSecret},
			ExpectedContains: []string{`"name":"Google"`},
			ExpectedStatus:   200,
			ResponseLen:      7,
			BeforeTest:       SetTestENV,
			FuncTest: func(t *testing.T) error {
				count := len(services.Services())
				if count != 7 {
					return errors.Errorf("incorrect services count: %d", count)
				}
				return nil
			},
		},
		{
			Name:             "Statping All Public Services",
			URL:              "/api/services",
			Method:           "GET",
			ExpectedContains: []string{`"name":"Google"`},
			ExpectedStatus:   200,
			GreaterThan:      5, // At least 6 public services from sample data
			FuncTest: func(t *testing.T) error {
				count := len(services.Services())
				if count < 7 {
					return errors.Errorf("expected at least 7 services, got: %d", count)
				}
				return nil
			},
		},
		{
			Name:             "Statping Public Service 1",
			URL:              "/api/services/1",
			Method:           "GET",
			ExpectedContains: []string{`"name":"Google"`},
			ExpectedStatus:   200,
			BeforeTest:       UnsetTestENV,
		},
		{
			Name:             "Statping Private Service 6",
			URL:              "/api/services/6",
			Method:           "GET",
			ExpectedContains: []string{`"error":"user not authenticated"`},
			ExpectedStatus:   401, // GET doesn't need CSRF, so auth check runs
			NoAuth:           true,
		},
		{
			Name:           "Statping Authenticated Private Service 6",
			URL:            "/api/services/6",
			Method:         "GET",
			HttpHeaders:    []string{"Authorization=" + core.App.ApiSecret},
			ExpectedStatus: 200,
			BeforeTest:     SetTestENV,
		},
		{
			Name:           "Statping Private Service with API Key",
			URL:            "/api/services/6?api=" + core.App.ApiSecret,
			Method:         "GET",
			ExpectedStatus: 200,
			BeforeTest:     UnsetTestENV,
		},
		{
			Name:           "Statping Private Service with API Header",
			URL:            "/api/services/6?api=" + core.App.ApiSecret,
			Method:         "GET",
			HttpHeaders:    []string{"Authorization=" + core.App.ApiSecret},
			ExpectedStatus: 200,
			BeforeTest:     UnsetTestENV,
		},
		{
			Name:             "Statping Service 1 with Private responses",
			URL:              "/api/services/1",
			Method:           "GET",
			ExpectedContains: []string{`"name":"Google"`},
			ExpectedStatus:   200,
			BeforeTest:       SetTestENV,
		},
		{
			Name:           "Statping Service Failures",
			URL:            "/api/services/3/failures" + startEndQuery,
			Method:         "GET",
			GreaterThan:    0,
			ExpectedStatus: 200,
		},
		{
			Name:           "Statping Service 1 Hits",
			URL:            "/api/services/1/hits" + startEndQuery,
			Method:         "GET",
			GreaterThan:    8580,
			ExpectedStatus: 200,
		},
		{
			Name:           "Statping Service 2 Hits",
			URL:            "/api/services/2/hits" + startEndQuery,
			Method:         "GET",
			GreaterThan:    8580,
			ExpectedStatus: 200,
		},
		{
			Name:           "Statping Service Failures Limited",
			URL:            "/api/services/3/failures?limit=1",
			Method:         "GET",
			ExpectedStatus: 200,
		},
		{
			Name:           "Statping Service 1 Hits Data",
			URL:            "/api/services/1/hits_data" + startEndQuery,
			Method:         "GET",
			GreaterThan:    70,
			ExpectedStatus: 200,
		},
		{
			Name:           "Statping Service 1 Ping Data",
			URL:            "/api/services/1/ping_data" + startEndQuery,
			Method:         "GET",
			ExpectedStatus: 200,
			GreaterThan:    70,
		},
		{
			Name:           "Statping Service 3 Failure Data - 24 Hour",
			URL:            "/api/services/3/failure_data" + startEndQuery + "&group=24h",
			Method:         "GET",
			ExpectedStatus: 200,
			GreaterThan:    0,
		},
		{
			Name:           "Statping Service 3 Failure Data - 12 Hour",
			URL:            "/api/services/3/failure_data" + startEndQuery + "&group=12h",
			Method:         "GET",
			ExpectedStatus: 200,
			GreaterThan:    0,
		},
		{
			Name:           "Statping Service 3 Failure Data - 1 Hour",
			URL:            "/api/services/3/failure_data" + startEndQuery + "&group=1h",
			Method:         "GET",
			ExpectedStatus: 200,
			GreaterThan:    0,
		},
		{
			Name:           "Statping Service 3 Failure Data - 15 Minute",
			URL:            "/api/services/3/failure_data" + startEndQuery + "&group=15m",
			Method:         "GET",
			GreaterThan:    0,
			ExpectedStatus: 200,
		},
		{
			Name:           "Statping Service 1 Hits",
			URL:            "/api/services/1/hits_data" + startEndQuery,
			Method:         "GET",
			GreaterThan:    70,
			ExpectedStatus: 200,
		},
		{
			Name:           "Statping Service 1 Uptime",
			URL:            "/api/services/1/uptime_data" + startEndQuery,
			Method:         "GET",
			ExpectedStatus: 200,
			ResponseFunc: func(req *httptest.ResponseRecorder, t *testing.T, resp []byte) error {
				var uptime *services.UptimeSeries
				if err := json.Unmarshal(resp, &uptime); err != nil {
					return err
				}
				assert.GreaterOrEqual(t, uptime.Uptime, int64(0))
				return nil
			},
		},
		{
			Name:           "Statping Service 3 Failure Data",
			URL:            "/api/services/3/failure_data" + startEndQuery,
			Method:         "GET",
			GreaterThan:    0,
			ExpectedStatus: 200,
		},
		{
			Name:           "Statping Reorder Services",
			URL:            "/api/reorder/services",
			Method:         "POST",
			Body:           `[{"service":1,"order":1},{"service":4,"order":2},{"service":2,"order":3},{"service":3,"order":4}]`,
			ExpectedStatus: 200,
			HttpHeaders:    []string{"Content-Type=application/json"},
			SecureRoute:    true,
		},
		{
			Name:        "Statping Create Service",
			URL:         "/api/services",
			HttpHeaders: []string{"Content-Type=application/json"},
			Method:      "POST",
			Body: `{
					"name": "New Private Service",
					"domain": "https://statping.com",
					"expected": "",
					"expected_status": 200,
					"check_interval": 30,
					"type": "http",
					"public": false,
					"group_id": 1,
					"method": "GET",
					"post_data": "",
					"port": 0,
					"timeout": 30,
					"order_id": 0
				}`,
			ExpectedStatus:   200,
			ExpectedContains: []string{Success, MethodCreate, `"type":"service",`, `"public":false`, `"group_id":1`},
			FuncTest: func(t *testing.T) error {
				count := len(services.Services())
				if count != 8 {
					return errors.Errorf("incorrect services count: %d", count)
				}
				return nil
			},
			SecureRoute: true,
		},
		{
			Name:        "Statping Update Service",
			URL:         "/api/services/1",
			HttpHeaders: []string{"Content-Type=application/json"},
			Method:      "POST",
			Body: `{
					"name": "Updated New Service",
					"domain": "https://google.com",
					"expected": "",
					"expected_status": 200,
					"check_interval": 60,
					"type": "http",
					"method": "GET",
					"post_data": "",
					"port": 0,
					"timeout": 10,
					"order_id": 0
				}`,
			ExpectedStatus:   200,
			ExpectedContains: []string{Success, `"name":"Updated New Service"`, MethodUpdate},
			FuncTest: func(t *testing.T) error {
				item, err := services.Find(1)
				require.Nil(t, err)
				if item.Interval != 60 {
					return errors.Errorf("incorrect service check interval: %d", item.Interval)
				}
				return nil
			},
			SecureRoute: true,
		},
		{
			Name:             "Statping Delete Hits",
			URL:              "/api/services/1/hits",
			Method:           "DELETE",
			ExpectedStatus:   200,
			ExpectedContains: []string{Success, `"method":"delete"`},
			SecureRoute:      true,
		},
		{
			Name:             "Statping Delete Failures",
			URL:              "/api/services/1/failures",
			Method:           "DELETE",
			ExpectedStatus:   200,
			ExpectedContains: []string{Success, `"method":"delete_failures"`},
			FuncTest: func(t *testing.T) error {
				item, err := services.Find(1)
				require.Nil(t, err)
				fails := item.AllFailures().Count()
				if fails != 0 {
					return errors.Errorf("incorrect service failures count: %d", fails)
				}
				return nil
			},
			SecureRoute: true,
		},
		{
			Name:             "Statping Delete Service",
			URL:              "/api/services/1",
			Method:           "DELETE",
			ExpectedStatus:   200,
			ExpectedContains: []string{Success, MethodDelete},
			FuncTest: func(t *testing.T) error {
				count := len(services.Services())
				if count != 7 {
					return errors.Errorf("incorrect services count: %d", count)
				}
				return nil
			},
			SecureRoute: true,
		},
		{
			Name:           "Statping Patch Static Service",
			URL:            "/api/services/7",
			Method:         "PATCH",
			ExpectedStatus: 200,
			Body: `{
						"online": false,
						"latency": 30500,
						"issue": "This is a failure string you can create"
					},`,
			ExpectedContains: []string{Success, MethodUpdate},
			FuncTest: func(t *testing.T) error {
				count := len(services.Services())
				if count != 7 {
					return errors.Errorf("incorrect services count: %d", count)
				}
				return nil
			},
			SecureRoute: true,
		},
		{
			Name:             "Incorrect JSON POST",
			URL:              "/api/services",
			Body:             BadJSON,
			ExpectedContains: []string{BadJSONResponse},
			BeforeTest:       SetTestENV,
			Method:           "POST",
			ExpectedStatus:   422,
		},
	}

	for _, v := range tests {
		t.Run(v.Name, func(t *testing.T) {
			_, t, err := RunHTTPTest(v, t)
			assert.Nil(t, err)
		})
	}
}

func TestServiceUpdateHandler(t *testing.T) {
	ensureHandlerSetup(t)

	tests := []HTTPTest{
		{
			Name:        "Update Service with Valid Data",
			URL:         "/api/services/2",
			HttpHeaders: []string{"Content-Type=application/json"},
			Method:      "POST",
			Body: `{
				"name": "Updated Service Name",
				"domain": "https://example.com",
				"expected": "",
				"expected_status": 200,
				"check_interval": 45,
				"type": "http",
				"method": "GET",
				"post_data": "",
				"port": 0,
				"timeout": 15,
				"order_id": 0
			}`,
			ExpectedStatus:   200,
			ExpectedContains: []string{Success, `"name":"Updated Service Name"`, MethodUpdate},
			FuncTest: func(t *testing.T) error {
				item, err := services.Find(2)
				require.Nil(t, err)
				if item.Interval != 45 {
					return errors.Errorf("incorrect service check interval: %d, expected 45", item.Interval)
				}
				if item.Name != "Updated Service Name" {
					return errors.Errorf("incorrect service name: %s", item.Name)
				}
				return nil
			},
			SecureRoute: true,
		},
		{
			Name:             "Update Service with Invalid JSON",
			URL:              "/api/services/2",
			HttpHeaders:      []string{"Content-Type=application/json"},
			Method:           "POST",
			Body:             BadJSON,
			ExpectedStatus:   422,
			ExpectedContains: []string{BadJSONResponse},
			SecureRoute:      true,
		},
		{
			Name:        "Update Service with Missing Required Fields",
			URL:         "/api/services/2",
			HttpHeaders: []string{"Content-Type=application/json"},
			Method:      "POST",
			Body: `{
				"name": "",
				"domain": ""
			}`,
			ExpectedStatus:   422,
			ExpectedContains: []string{`"error"`},
			SecureRoute:      true,
		},
		{
			Name:        "Update Non-Existent Service",
			URL:         "/api/services/99999",
			HttpHeaders: []string{"Content-Type=application/json"},
			Method:      "POST",
			Body: `{
				"name": "Test Service",
				"domain": "https://test.com",
				"expected_status": 200,
				"check_interval": 30,
				"type": "http",
				"method": "GET",
				"timeout": 30
			}`,
			ExpectedStatus:   404,
			ExpectedContains: []string{`"error"`},
			SecureRoute:      true,
		},
		{
			Name:             "Update Service with Invalid ID (non-numeric)",
			URL:              "/api/services/invalid",
			HttpHeaders:      []string{"Content-Type=application/json"},
			Method:           "POST",
			Body:             `{"name":"Test"}`,
			ExpectedStatus:   422,
			ExpectedContains: []string{`"error"`},
			SecureRoute:      true,
		},
	}

	for _, v := range tests {
		t.Run(v.Name, func(t *testing.T) {
			_, t, err := RunHTTPTest(v, t)
			assert.Nil(t, err)
		})
	}
}

func TestServiceDataEndpoints(t *testing.T) {
	ensureHandlerSetup(t)
	since := utils.Now().Add(-30 * types.Day)
	end := utils.Now().Add(-30 * time.Minute)
	startEndQuery := fmt.Sprintf("?start=%d&end=%d", since.Unix(), end.Unix()+15)

	tests := []HTTPTest{
		{
			Name:           "Service Hits Data with Valid Service",
			URL:            "/api/services/2/hits_data" + startEndQuery,
			Method:         "GET",
			ExpectedStatus: 200,
		},
		{
			Name:           "Service Hits Data with Non-Existent Service",
			URL:            "/api/services/99999/hits_data" + startEndQuery,
			Method:         "GET",
			ExpectedStatus: 404,
		},
		{
			Name:           "Service Hits Data with Invalid ID",
			URL:            "/api/services/invalid/hits_data" + startEndQuery,
			Method:         "GET",
			ExpectedStatus: 422,
		},
		{
			Name:           "Service Failure Data with Valid Service",
			URL:            "/api/services/3/failure_data" + startEndQuery,
			Method:         "GET",
			ExpectedStatus: 200,
		},
		{
			Name:           "Service Failure Data with Non-Existent Service",
			URL:            "/api/services/99999/failure_data" + startEndQuery,
			Method:         "GET",
			ExpectedStatus: 404,
		},
		{
			Name:           "Service Failure Data with Invalid ID",
			URL:            "/api/services/invalid/failure_data" + startEndQuery,
			Method:         "GET",
			ExpectedStatus: 422,
		},
		{
			Name:           "Service Ping Data with Valid Service",
			URL:            "/api/services/2/ping_data" + startEndQuery,
			Method:         "GET",
			ExpectedStatus: 200,
		},
		{
			Name:           "Service Ping Data with Non-Existent Service",
			URL:            "/api/services/99999/ping_data" + startEndQuery,
			Method:         "GET",
			ExpectedStatus: 404,
		},
		{
			Name:           "Service Ping Data with Invalid ID",
			URL:            "/api/services/invalid/ping_data" + startEndQuery,
			Method:         "GET",
			ExpectedStatus: 422,
		},
		{
			Name:           "Service Uptime Data with Valid Service",
			URL:            "/api/services/2/uptime_data" + startEndQuery,
			Method:         "GET",
			ExpectedStatus: 200,
		},
		{
			Name:           "Service Uptime Data with Non-Existent Service",
			URL:            "/api/services/99999/uptime_data" + startEndQuery,
			Method:         "GET",
			ExpectedStatus: 404,
		},
		{
			Name:           "Service Uptime Data with Invalid ID",
			URL:            "/api/services/invalid/uptime_data" + startEndQuery,
			Method:         "GET",
			ExpectedStatus: 422,
		},
		{
			Name:           "Service Failure Data with 6h Group",
			URL:            "/api/services/3/failure_data" + startEndQuery + "&group=6h",
			Method:         "GET",
			ExpectedStatus: 200,
		},
		{
			Name:           "Service Failure Data with 30m Group",
			URL:            "/api/services/3/failure_data" + startEndQuery + "&group=30m",
			Method:         "GET",
			ExpectedStatus: 200,
		},
	}

	for _, v := range tests {
		t.Run(v.Name, func(t *testing.T) {
			_, t, err := RunHTTPTest(v, t)
			assert.Nil(t, err)
		})
	}
}

func TestServiceDeleteOperations(t *testing.T) {
	ensureHandlerSetup(t)

	tests := []HTTPTest{
		{
			Name:             "Delete Service Hits for Valid Service",
			URL:              "/api/services/3/hits",
			Method:           "DELETE",
			ExpectedStatus:   200,
			ExpectedContains: []string{Success, `"method":"delete"`},
			SecureRoute:      true,
		},
		{
			Name:           "Delete Service Hits for Non-Existent Service",
			URL:            "/api/services/99999/hits",
			Method:         "DELETE",
			ExpectedStatus: 404,
			SecureRoute:    true,
		},
		{
			Name:           "Delete Service Hits with Invalid ID",
			URL:            "/api/services/invalid/hits",
			Method:         "DELETE",
			ExpectedStatus: 422,
			SecureRoute:    true,
		},
		{
			Name:             "Delete Service Failures for Valid Service",
			URL:              "/api/services/3/failures",
			Method:           "DELETE",
			ExpectedStatus:   200,
			ExpectedContains: []string{Success, `"method":"delete_failures"`},
			FuncTest: func(t *testing.T) error {
				item, err := services.Find(3)
				require.Nil(t, err)
				fails := item.AllFailures().Count()
				if fails != 0 {
					return errors.Errorf("incorrect service failures count: %d, expected 0", fails)
				}
				return nil
			},
			SecureRoute: true,
		},
		{
			Name:           "Delete Service Failures for Non-Existent Service",
			URL:            "/api/services/99999/failures",
			Method:         "DELETE",
			ExpectedStatus: 404,
			SecureRoute:    true,
		},
		{
			Name:           "Delete Service Failures with Invalid ID",
			URL:            "/api/services/invalid/failures",
			Method:         "DELETE",
			ExpectedStatus: 422,
			SecureRoute:    true,
		},
		{
			Name:             "Delete Service - Valid Service",
			URL:              "/api/services/4",
			Method:           "DELETE",
			ExpectedStatus:   200,
			ExpectedContains: []string{Success, MethodDelete},
			SecureRoute:      true,
		},
		{
			Name:           "Delete Service - Non-Existent Service",
			URL:            "/api/services/99999",
			Method:         "DELETE",
			ExpectedStatus: 404,
			SecureRoute:    true,
		},
		{
			Name:           "Delete Service - Invalid ID",
			URL:            "/api/services/invalid",
			Method:         "DELETE",
			ExpectedStatus: 422,
			SecureRoute:    true,
		},
	}

	for _, v := range tests {
		t.Run(v.Name, func(t *testing.T) {
			_, t, err := RunHTTPTest(v, t)
			assert.Nil(t, err)
		})
	}
}

func TestReorderServiceHandler(t *testing.T) {
	ensureHandlerSetup(t)

	tests := []HTTPTest{
		{
			Name:           "Reorder Services - Valid Order",
			URL:            "/api/reorder/services",
			Method:         "POST",
			Body:           `[{"service":2,"order":1},{"service":3,"order":2},{"service":5,"order":3}]`,
			ExpectedStatus: 200,
			HttpHeaders:    []string{"Content-Type=application/json"},
			FuncTest: func(t *testing.T) error {
				srv, err := services.Find(2)
				require.Nil(t, err)
				if srv.Order != 1 {
					return errors.Errorf("service 2 order incorrect: %d, expected 1", srv.Order)
				}
				srv, err = services.Find(3)
				require.Nil(t, err)
				if srv.Order != 2 {
					return errors.Errorf("service 3 order incorrect: %d, expected 2", srv.Order)
				}
				return nil
			},
			SecureRoute: true,
		},
		{
			Name:             "Reorder Services - Invalid JSON",
			URL:              "/api/reorder/services",
			Method:           "POST",
			Body:             BadJSON,
			ExpectedStatus:   422,
			HttpHeaders:      []string{"Content-Type=application/json"},
			ExpectedContains: []string{BadJSONResponse},
			SecureRoute:      true,
		},
		{
			Name:           "Reorder Services - Non-Existent Service ID",
			URL:            "/api/reorder/services",
			Method:         "POST",
			Body:           `[{"service":99999,"order":1}]`,
			ExpectedStatus: 404,
			HttpHeaders:    []string{"Content-Type=application/json"},
			SecureRoute:    true,
		},
		{
			Name:           "Reorder Services - Empty Array",
			URL:            "/api/reorder/services",
			Method:         "POST",
			Body:           `[]`,
			ExpectedStatus: 200,
			HttpHeaders:    []string{"Content-Type=application/json"},
			SecureRoute:    true,
		},
		{
			Name:           "Reorder Services - Single Service",
			URL:            "/api/reorder/services",
			Method:         "POST",
			Body:           `[{"service":2,"order":5}]`,
			ExpectedStatus: 200,
			HttpHeaders:    []string{"Content-Type=application/json"},
			FuncTest: func(t *testing.T) error {
				srv, err := services.Find(2)
				require.Nil(t, err)
				if srv.Order != 5 {
					return errors.Errorf("service 2 order incorrect: %d, expected 5", srv.Order)
				}
				return nil
			},
			SecureRoute: true,
		},
	}

	for _, v := range tests {
		t.Run(v.Name, func(t *testing.T) {
			_, t, err := RunHTTPTest(v, t)
			assert.Nil(t, err)
		})
	}
}

func TestServicePatchHandler(t *testing.T) {
	ensureHandlerSetup(t)

	tests := []HTTPTest{
		{
			Name:           "Patch Service - Set Offline with Custom Issue",
			URL:            "/api/services/5",
			Method:         "PATCH",
			ExpectedStatus: 200,
			Body: `{
				"online": false,
				"latency": 25000,
				"issue": "Manual maintenance mode"
			}`,
			HttpHeaders:      []string{"Content-Type=application/json"},
			ExpectedContains: []string{Success, MethodUpdate},
			SecureRoute:      true,
		},
		{
			Name:           "Patch Service - Set Online",
			URL:            "/api/services/5",
			Method:         "PATCH",
			ExpectedStatus: 200,
			Body: `{
				"online": true,
				"latency": 100
			}`,
			HttpHeaders:      []string{"Content-Type=application/json"},
			ExpectedContains: []string{Success, MethodUpdate},
			SecureRoute:      true,
		},
		{
			Name:           "Patch Service - Set Offline with Default Issue",
			URL:            "/api/services/5",
			Method:         "PATCH",
			ExpectedStatus: 200,
			Body: `{
				"online": false,
				"latency": 50000
			}`,
			HttpHeaders:      []string{"Content-Type=application/json"},
			ExpectedContains: []string{Success, MethodUpdate},
			SecureRoute:      true,
		},
		{
			Name:             "Patch Service - Invalid JSON",
			URL:              "/api/services/5",
			Method:           "PATCH",
			ExpectedStatus:   422,
			Body:             BadJSON,
			HttpHeaders:      []string{"Content-Type=application/json"},
			ExpectedContains: []string{BadJSONResponse},
			SecureRoute:      true,
		},
		{
			Name:           "Patch Service - Non-Existent Service",
			URL:            "/api/services/99999",
			Method:         "PATCH",
			ExpectedStatus: 404,
			Body:           `{"online": false}`,
			HttpHeaders:    []string{"Content-Type=application/json"},
			SecureRoute:    true,
		},
		{
			Name:           "Patch Service - Invalid ID",
			URL:            "/api/services/invalid",
			Method:         "PATCH",
			ExpectedStatus: 422,
			Body:           `{"online": false}`,
			HttpHeaders:    []string{"Content-Type=application/json"},
			SecureRoute:    true,
		},
	}

	for _, v := range tests {
		t.Run(v.Name, func(t *testing.T) {
			_, t, err := RunHTTPTest(v, t)
			assert.Nil(t, err)
		})
	}
}

func TestServicePrivateAccess(t *testing.T) {
	ensureHandlerSetup(t)

	tests := []HTTPTest{
		{
			Name:             "Private Service Hits Data - Unauthenticated",
			URL:              "/api/services/6/hits_data",
			Method:           "GET",
			ExpectedStatus:   401,
			ExpectedContains: []string{`"error":"user not authenticated"`},
			NoAuth:           true,
		},
		{
			Name:             "Private Service Failure Data - Unauthenticated",
			URL:              "/api/services/6/failure_data",
			Method:           "GET",
			ExpectedStatus:   401,
			ExpectedContains: []string{`"error":"user not authenticated"`},
			NoAuth:           true,
		},
		{
			Name:             "Private Service Ping Data - Unauthenticated",
			URL:              "/api/services/6/ping_data",
			Method:           "GET",
			ExpectedStatus:   401,
			ExpectedContains: []string{`"error":"user not authenticated"`},
			NoAuth:           true,
		},
		{
			Name:             "Private Service Uptime Data - Unauthenticated",
			URL:              "/api/services/6/uptime_data",
			Method:           "GET",
			ExpectedStatus:   401,
			ExpectedContains: []string{`"error":"user not authenticated"`},
			NoAuth:           true,
		},
		{
			Name:           "Private Service Hits Data - Authenticated with API Key",
			URL:            "/api/services/6/hits_data",
			Method:         "GET",
			HttpHeaders:    []string{"Authorization=" + core.App.ApiSecret},
			ExpectedStatus: 200,
			BeforeTest:     SetTestENV,
		},
		{
			Name:           "Private Service Failure Data - Authenticated with API Key",
			URL:            "/api/services/6/failure_data",
			Method:         "GET",
			HttpHeaders:    []string{"Authorization=" + core.App.ApiSecret},
			ExpectedStatus: 200,
			BeforeTest:     SetTestENV,
		},
	}

	for _, v := range tests {
		t.Run(v.Name, func(t *testing.T) {
			_, t, err := RunHTTPTest(v, t)
			assert.Nil(t, err)
		})
	}
}
