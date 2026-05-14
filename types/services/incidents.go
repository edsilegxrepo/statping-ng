package services

func (s *Service) DeleteIncidents() error {
	for _, i := range s.Incidents {
		if err := i.Delete(); err != nil {
			return err
		}
	}
	return db.Model(s).Association("Updates").Clear()
}
