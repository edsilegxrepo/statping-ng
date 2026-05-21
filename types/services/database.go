package services

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"

	"github.com/statping-ng/statping-ng/database"
	"github.com/statping-ng/statping-ng/types/errors"
	"github.com/statping-ng/statping-ng/types/metrics"
	"github.com/statping-ng/statping-ng/utils"
	"gorm.io/gorm"
)

var (
	db          database.Database
	log         = utils.Log.WithField("type", "service")
	allServices map[int64]*Service
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

func (s *Service) AfterFind(tx *gorm.DB) (err error) {
	tx.Where("service = ?", s.Id).Find(&s.Incidents)
	tx.Where("service = ?", s.Id).Find(&s.Messages)
	tx.Where("service = ?", s.Id).Find(&s.Checkins)
	metrics.Query("service", "find")
	return nil
}

func (s *Service) AfterCreate(tx *gorm.DB) (err error) {
	s.prevOnline = true
	allServices[s.Id] = s
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

func Services() map[int64]*Service {
	return allServices
}

func SetDB(database database.Database) {
	db = database
}

func Find(id int64) (*Service, error) {
	srv := allServices[id]
	if srv == nil {
		return nil, errors.Missing(&Service{}, id)
	}
	db.First(&srv, id)
	return srv, nil
}

func all() []*Service {
	var services []*Service
	db.Find(&services)
	return services
}

func All() map[int64]*Service {
	return allServices
}

func AllInOrder() []Service {
	var services []Service
	for _, service := range allServices {
		service.UpdateStats()
		services = append(services, *service)
	}
	sort.Sort(ServiceOrder(services))
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
	allServices[s.Id] = s
	s.SleepDuration = s.Duration()
	go ServiceCheckQueue(allServices[s.Id], true)
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
	if err := db.Model(s).Association("Checkins").Clear(); err != nil {
		return err
	}
	if err := s.DeleteIncidents(); err != nil {
		return err
	}
	if err := db.Model(s).Association("Incidents").Clear(); err != nil {
		return err
	}
	if err := s.DeleteMessages(); err != nil {
		return err
	}
	if err := db.Model(s).Association("Messages").Clear(); err != nil {
		return err
	}

	delete(allServices, s.Id)
	q := db.Model(&Service{}).Delete(s)
	return q.Error()
}

func (s *Service) DeleteMessages() error {
	for _, m := range s.Messages {
		if err := m.Delete(); err != nil {
			return err
		}
	}
	return db.Model(s).Association("messages").Clear()
}

func (s *Service) DeleteCheckins() error {
	for _, c := range s.Checkins {
		if err := c.Delete(); err != nil {
			return err
		}
	}
	return db.Model(s).Association("checkins").Clear()
}
