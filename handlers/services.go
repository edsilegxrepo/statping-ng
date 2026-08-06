package handlers

import (
	"fmt"
	"math"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/statping-ng/statping-ng/database"
	"github.com/statping-ng/statping-ng/types/checkins"
	"github.com/statping-ng/statping-ng/types/errors"
	"github.com/statping-ng/statping-ng/types/failures"
	"github.com/statping-ng/statping-ng/types/hits"
	"github.com/statping-ng/statping-ng/types/services"
	"github.com/statping-ng/statping-ng/utils"
)

type serviceOrder struct {
	Id    int64 `json:"service"`
	Order int   `json:"order"`
}

func findService(r *http.Request) (*services.Service, error) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	if utils.NotNumber(idStr) {
		return nil, errors.NotNumber
	}
	id := utils.ToInt(idStr)
	if id <= 0 {
		return nil, errors.NotNumber
	}
	servicer, err := services.Find(id)
	if err != nil {
		return nil, err
	}
	if !servicer.Public.Bool && !IsReadAuthenticated(r) {
		return nil, errors.NotAuthenticated
	}
	return servicer, nil
}

func reorderServiceHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		sendErrorJson(err, w, r)
		return
	}
	var newOrder []*serviceOrder
	if err := DecodeJSON(r, &newOrder); err != nil {
		sendErrorJson(err, w, r)
		return
	}

	for _, s := range newOrder {
		service, err := services.Find(s.Id)
		if err != nil {
			sendErrorJson(err, w, r)
			return
		}
		service.Order = s.Order
		if err := service.Update(); err != nil {
			sendErrorJson(err, w, r)
			return
		}
	}
	returnJson(newOrder, w, r)
}

func apiServiceHandler(r *http.Request) interface{} {
	srv, err := findService(r)
	if err != nil {
		return err
	}
	srv = srv.UpdateStats()
	// ExpectedStatus 0 is stored in database as MinInt32,
	// to circumvent the problem of gorm not updating zero value.
	if srv.Type == "cmd" && srv.ExpectedStatus == math.MinInt32 {
		srv.ExpectedStatus = 0
	}
	// Mask sensitive credentials before returning to API
	srv.MaskSecrets()
	return srv
}

func apiCreateServiceHandler(w http.ResponseWriter, r *http.Request) {
	var service *services.Service
	if err := DecodeJSON(r, &service); err != nil {
		sendErrorJson(err, w, r)
		return
	}
	if err := service.Validate(); err != nil {
		sendErrorJson(err, w, r)
		return
	}
	if err := service.Create(); err != nil {
		sendErrorJson(err, w, r)
		return
	}
	go services.ServiceCheckQueue(service, true)

	sendJsonAction(service, "create", w, r)
}

type servicePatchReq struct {
	Online  bool   `json:"online"`
	Issue   string `json:"issue,omitempty"`
	Latency int64  `json:"latency,omitempty"`
}

func apiServicePatchHandler(w http.ResponseWriter, r *http.Request) {
	service, err := findService(r)
	if err != nil {
		sendErrorJson(err, w, r)
		return
	}
	var req servicePatchReq
	if err := DecodeJSON(r, &req); err != nil {
		sendErrorJson(err, w, r)
		return
	}

	service.Online = req.Online
	service.Latency = req.Latency

	issueDefault := "Service was triggered to be offline"
	if req.Issue != "" {
		issueDefault = req.Issue
	}

	if !req.Online {
		services.RecordFailure(service, issueDefault, "trigger")
	} else {
		services.RecordSuccess(service)
	}

	if err := service.Update(); err != nil {
		sendErrorJson(err, w, r)
		return
	}

	sendJsonAction(service, "update", w, r)
}

func apiServiceUpdateHandler(w http.ResponseWriter, r *http.Request) {
	service, err := findService(r)
	if err != nil {
		sendErrorJson(err, w, r)
		return
	}
	if err := DecodeJSON(r, &service); err != nil {
		sendErrorJson(err, w, r)
		return
	}
	if err := service.Validate(); err != nil {
		sendErrorJson(err, w, r)
		return
	}
	if err := service.Update(); err != nil {
		sendErrorJson(err, w, r)
		return
	}
	go service.CheckService(true)
	sendJsonAction(service, "update", w, r)
}

func apiServiceDataHandler(w http.ResponseWriter, r *http.Request) {
	service, err := findService(r)
	if err != nil {
		sendErrorJson(err, w, r)
		return
	}

	groupQuery, err := database.ParseQueriesForTable(r, service.AllHits().Db().Select("latency, created_at"), "hits")
	if err != nil {
		sendErrorJson(err, w, r)
		return
	}

	objs, err := groupQuery.GraphData(database.ByAverage("latency", 1000))
	if err != nil {
		sendErrorJson(err, w, r)
		return
	}
	returnJson(objs, w, r)
}

func apiServiceFailureDataHandler(w http.ResponseWriter, r *http.Request) {
	service, err := findService(r)
	if err != nil {
		sendErrorJson(err, w, r)
		return
	}

	groupQuery, err := database.ParseQueriesForTable(r, service.AllFailures().Db().Select("created_at"), "failures")
	if err != nil {
		sendErrorJson(err, w, r)
		return
	}

	objs, err := groupQuery.GraphData(database.ByCount)
	if err != nil {
		sendErrorJson(err, w, r)
		return
	}

	returnJson(objs, w, r)
}

func apiServicePingDataHandler(w http.ResponseWriter, r *http.Request) {
	service, err := findService(r)
	if err != nil {
		sendErrorJson(err, w, r)
		return
	}

	groupQuery, err := database.ParseQueriesForTable(r, service.AllHits().Db().Select("ping_time, created_at"), "hits")
	if err != nil {
		sendErrorJson(err, w, r)
		return
	}

	objs, err := groupQuery.GraphData(database.ByAverage("ping_time", 1000))
	if err != nil {
		sendErrorJson(err, w, r)
		return
	}
	returnJson(objs, w, r)
}

func apiServiceTimeDataHandler(w http.ResponseWriter, r *http.Request) {
	service, err := findService(r)
	if err != nil {
		sendErrorJson(err, w, r)
		return
	}

	groupHits, err := database.ParseQueriesForTable(r, service.AllHits().Db().Select("created_at"), "hits")
	if err != nil {
		sendErrorJson(err, w, r)
		return
	}

	groupCheckinHits, err := database.ParseQueriesForTable(r, service.AllCheckinHits().Db().Select("checkin_hits.created_at"), "checkin_hits")
	if err != nil {
		sendErrorJson(err, w, r)
		return
	}

	groupFailures, err := database.ParseQueriesForTable(r, service.AllFailures().Db().Select("created_at"), "failures")
	if err != nil {
		sendErrorJson(err, w, r)
		return
	}

	var allFailures []*failures.Failure
	var allHits []*hits.Hit
	var allCheckinHits []*checkins.CheckinHit

	if err := groupHits.Find(&allHits); err != nil {
		sendErrorJson(err, w, r)
		return
	}

	if err := groupCheckinHits.Find(&allCheckinHits); err != nil {
		sendErrorJson(err, w, r)
		return
	}

	if err := groupFailures.Find(&allFailures); err != nil {
		sendErrorJson(err, w, r)
		return
	}

	uptimeData, err := service.UptimeData(allHits, allCheckinHits, allFailures)
	if err != nil {
		sendErrorJson(err, w, r)
		return
	}

	returnJson(uptimeData, w, r)
}

func apiServiceHitsDeleteHandler(w http.ResponseWriter, r *http.Request) {
	service, err := findService(r)
	if err != nil {
		sendErrorJson(err, w, r)
		return
	}
	if err := service.AllHits().DeleteAll(); err != nil {
		sendErrorJson(err, w, r)
		return
	}
	sendJsonAction(service, "delete", w, r)
}

func apiServiceDeleteHandler(w http.ResponseWriter, r *http.Request) {
	service, err := findService(r)
	if err != nil {
		sendErrorJson(err, w, r)
		return
	}
	err = service.Delete()
	if err != nil {
		sendErrorJson(err, w, r)
		return
	}

	sendJsonAction(service, "delete", w, r)
}

func apiAllServicesHandler(r *http.Request) interface{} {
	var srvs []*services.Service
	for _, v := range services.AllInOrder() {
		if !v.Public.Bool && !IsUser(r) {
			continue
		}
		// ExpectedStatus 0 is stored in database as MinInt32,
		// to circumvent the problem of gorm not updating zero value.
		if v.Type == "cmd" && v.ExpectedStatus == math.MinInt32 {
			v.ExpectedStatus = 0
		}
		// Mask sensitive credentials before returning to API
		v.MaskSecrets()
		srvs = append(srvs, v)
	}
	return srvs
}

func servicesDeleteFailuresHandler(w http.ResponseWriter, r *http.Request) {
	service, err := findService(r)
	if err != nil {
		sendErrorJson(err, w, r)
		return
	}
	if err := service.AllFailures().DeleteAll(); err != nil {
		sendErrorJson(err, w, r)
		return
	}
	sendJsonAction(service, "delete_failures", w, r)
}

func apiServiceFailuresHandler(r *http.Request) interface{} {
	service, err := findService(r)
	if err != nil {
		return err
	}
	var fails []*failures.Failure
	query, err := database.ParseQueriesForTable(r, service.AllFailures(), "failures")
	if err != nil {
		return err
	}
	if err := query.Find(&fails); err != nil {
		return err
	}
	return fails
}

func apiServiceHitsHandler(r *http.Request) interface{} {
	service, err := findService(r)
	if err != nil {
		return err
	}
	var hts []*hits.Hit
	query, err := database.ParseQueriesForTable(r, service.AllHits(), "hits")
	if err != nil {
		return err
	}
	if err := query.Find(&hts); err != nil {
		return err
	}
	return hts
}

// ServiceTestResponse contains the result of a service connectivity test
type ServiceTestResponse struct {
	Success bool              `json:"success"`
	Message string            `json:"message"`
	Latency int64             `json:"latency,omitempty"`
	Details string            `json:"details,omitempty"`
	Info    map[string]string `json:"info,omitempty"`
}

// apiServiceTestHandler tests connectivity for a service without saving
func apiServiceTestHandler(w http.ResponseWriter, r *http.Request) {
	var service services.Service
	if err := DecodeJSON(r, &service); err != nil {
		sendErrorJson(err, w, r)
		return
	}

	// Set default timeout if not specified
	if service.Timeout == 0 {
		service.Timeout = 10
	}

	var testErr error
	switch service.Type {
	case "http":
		_, testErr = services.CheckHttp(&service, false)
	case "tcp", "udp":
		_, testErr = services.CheckTcp(&service, false)
	case "icmp":
		_, testErr = services.CheckIcmp(&service, false)
	case "grpc":
		_, testErr = services.CheckGrpc(&service, false)
	case "smtp":
		_, testErr = services.CheckSmtp(&service, false)
	case "imap":
		_, testErr = services.CheckImap(&service, false)
	case "database":
		_, testErr = services.CheckDatabase(&service, false)
	case "storage":
		_, testErr = services.CheckStorage(&service, false)
	case "tls":
		_, testErr = services.CheckTLS(&service, false)
	case "cmd":
		_, testErr = services.CheckCmd(&service, false)
	default:
		sendErrorJson(fmt.Errorf("unsupported service type: %s", service.Type), w, r)
		return
	}

	resp := ServiceTestResponse{
		Success: testErr == nil,
		Latency: service.Latency,
		Info:    make(map[string]string),
	}

	if testErr != nil {
		resp.Message = "Connection test failed"
		resp.Details = testErr.Error()
	} else {
		resp.Message = "Connection test successful"
		// Add type-specific technical details
		switch service.Type {
		case "http":
			resp.Info["status_code"] = fmt.Sprintf("%d", service.LastStatusCode)
			if len(service.LastResponse) > 0 {
				resp.Info["response_size"] = fmt.Sprintf("%d bytes", len(service.LastResponse))
				resp.Details = service.LastResponse
			}
		case "tcp", "udp":
			resp.Info["endpoint"] = fmt.Sprintf("%s:%d", service.Domain, service.Port)
		case "icmp":
			resp.Info["host"] = service.Domain
		case "grpc":
			resp.Info["endpoint"] = fmt.Sprintf("%s:%d", service.Domain, service.Port)
			if service.LastResponse != "" {
				resp.Info["response"] = service.LastResponse
			}
		case "database":
			resp.Info["type"] = service.DatabaseType.String
		case "storage":
			resp.Info["backend"] = service.StorageBackend.String
			resp.Info["bucket"] = service.StorageBucket.String
		case "tls":
			if service.TLSExpiry != nil {
				resp.Info["expiry"] = service.TLSExpiry.Format("2006-01-02")
				resp.Info["days_remaining"] = fmt.Sprintf("%d", service.TLSDaysRemaining)
			}
			if service.TLSIssuer != "" {
				resp.Info["issuer"] = service.TLSIssuer
			}
		}
	}

	returnJson(resp, w, r)
}
