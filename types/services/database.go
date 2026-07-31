package services

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"sync"

	"github.com/statping-ng/statping-ng/database"
	"github.com/statping-ng/statping-ng/types/errors"
	"github.com/statping-ng/statping-ng/types/metrics"
	"github.com/statping-ng/statping-ng/utils"
	"gorm.io/gorm"
)

var (
	db           database.Database
	log          = utils.Log.WithField("type", "service")
	allServices  map[int64]*Service
	servicesLock sync.RWMutex
)

func (s *Service) Validate() error {
	if s.Name == "" {
		return errors.ServiceNameMissing
	} else if s.Domain == "" && s.Type != "cmd" && s.Type != "static" {
		return errors.DomainNameMissing
	} else if s.Type == "" {
		return errors.ServiceTypeMissing
	} else if s.Interval == 0 && s.Type != "static" {
		return errors.CheckIntervalMissing
	}

	if s.Type == "cmd" {
		var cmdConfig CmdConfig
		err := json.Unmarshal([]byte(s.PostData.String), &cmdConfig)
		if err != nil {
			return errors.CommandConfigNotJson
		}
		if cmdConfig.Cmd == "" {
			return errors.CommandConfigFieldCmdMissing
		}

		// ExpectedStatus 0 is stored in database as MinInt32,
		// to circumvent the problem of gorm not updating zero value.
		if s.ExpectedStatus == 0 {
			s.ExpectedStatus = math.MinInt32
		}
	}

	return nil
}

func (s *Service) BeforeCreate(tx *gorm.DB) (err error) {
	return s.Validate()
}

func (s *Service) BeforeUpdate(tx *gorm.DB) (err error) {
	return s.Validate()
}

// AfterFind is called after finding a service
// Note: Related data (Incidents, Messages, Checkins) should be loaded using
// Preload() in the query, not in AfterFind, to avoid N+1 query problems.
// Use FindWithRelations() for loading services with all related data.
func (s *Service) AfterFind(tx *gorm.DB) (err error) {
	metrics.Query("service", "find")
	return nil
}

func (s *Service) AfterCreate(tx *gorm.DB) (err error) {
	s.prevOnline = true
	servicesLock.Lock()
	allServices[s.Id] = s
	servicesLock.Unlock()
	metrics.Query("service", "create")
	return nil
}

func (s *Service) AfterUpdate(tx *gorm.DB) (err error) {
	metrics.Query("service", "update")
	return nil
}

func (s *Service) AfterDelete(tx *gorm.DB) (err error) {
	metrics.Query("service", "delete")
	return nil
}

func init() {
	allServices = make(map[int64]*Service)
}

// ClearCache clears the in-memory services cache.
// Used for testing to ensure fresh state.
func ClearCache() {
	servicesLock.Lock()
	defer servicesLock.Unlock()
	allServices = make(map[int64]*Service)
}

// StopAll stops all running service goroutines.
// Must be called before ClearCache to avoid orphan goroutines.
func StopAll() {
	servicesLock.RLock()
	servicesCopy := make([]*Service, 0, len(allServices))
	for _, s := range allServices {
		servicesCopy = append(servicesCopy, s)
	}
	servicesLock.RUnlock()

	for _, s := range servicesCopy {
		s.Close()
	}
}

func Services() map[int64]*Service {
	servicesLock.RLock()
	defer servicesLock.RUnlock()
	res := make(map[int64]*Service, len(allServices))
	for k, v := range allServices {
		res[k] = v
	}
	return res
}

func SetDB(database database.Database) {
	db = database
}

func Find(id int64) (*Service, error) {
	servicesLock.RLock()
	srv := allServices[id]
	servicesLock.RUnlock()
	if srv != nil {
		db.First(&srv, id)
		return srv, nil
	}
	var service Service
	if err := db.First(&service, id).Error(); err != nil {
		return nil, errors.Missing(&Service{}, id)
	}
	servicesLock.Lock()
	allServices[service.Id] = &service
	servicesLock.Unlock()
	return &service, nil
}

func all() []*Service {
	var services []*Service
	db.Find(&services)
	return services
}

// allWithRelations loads all services with their related data using Preload
// This avoids N+1 query problems by batching related data queries
func allWithRelations() []*Service {
	var services []*Service
	db.Preload("Incidents").Preload("Messages").Preload("Checkins").Find(&services)
	return services
}

// FindWithRelations finds a service by ID with all related data preloaded
func FindWithRelations(id int64) (*Service, error) {
	var service Service
	if err := db.Preload("Incidents").Preload("Messages").Preload("Checkins").First(&service, id).Error(); err != nil {
		return nil, errors.Missing(&Service{}, id)
	}
	servicesLock.Lock()
	allServices[service.Id] = &service
	servicesLock.Unlock()
	return &service, nil
}

func All() map[int64]*Service {
	return Services()
}

func AllInOrder() []*Service {
	servicesLock.RLock()
	var services []*Service
	for _, service := range allServices {
		service.UpdateStats()
		services = append(services, service)
	}
	servicesLock.RUnlock()
	sort.Sort(ServicePtrOrder(services))
	return services
}

func (s *Service) Create() error {
	err := db.Create(s)
	if err.Error() != nil {
		log.Errorln(fmt.Sprintf("Failed to create service %v #%v: %v", s.Name, s.Id, err))
		return err.Error()
	}
	return nil
}

func (s *Service) Update() error {
	q := db.Update(s)
	s.Close()
	servicesLock.Lock()
	allServices[s.Id] = s
	servicesLock.Unlock()
	s.SetSleepDuration(s.Duration())
	go ServiceCheckQueue(s, true)
	return q.Error()
}

func (s *Service) Delete() error {
	s.Close()
	if err := s.AllFailures().DeleteAll(); err != nil {
		return err
	}
	if err := s.AllHits().DeleteAll(); err != nil {
		return err
	}
	if err := s.DeleteCheckins(); err != nil {
		return err
	}
	if err := s.DeleteIncidents(); err != nil {
		return err
	}
	if err := s.DeleteMessages(); err != nil {
		return err
	}
	s.Checkins = nil
	s.Incidents = nil
	s.Messages = nil

	servicesLock.Lock()
	delete(allServices, s.Id)
	servicesLock.Unlock()
	q := db.Model(&Service{}).Delete(s)
	return q.Error()
}

func (s *Service) DeleteMessages() error {
	for _, m := range s.Messages {
		if err := m.Delete(); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) DeleteCheckins() error {
	for _, c := range s.Checkins {
		if err := c.Delete(); err != nil {
			return err
		}
	}
	return nil
}
