package services

import (
	"sync"

	"github.com/statping-ng/statping-ng/types/failures"
	"github.com/statping-ng/statping-ng/types/notifications"
)

var (
	allNotifiers  = make(map[string]ServiceNotifier)
	notifiersLock sync.RWMutex
)

func AllNotifiers() map[string]ServiceNotifier {
	notifiersLock.RLock()
	defer notifiersLock.RUnlock()
	list := make(map[string]ServiceNotifier, len(allNotifiers))
	for k, v := range allNotifiers {
		list[k] = v
	}
	return list
}

func ReturnNotifier(method string) ServiceNotifier {
	notifiersLock.RLock()
	defer notifiersLock.RUnlock()
	return allNotifiers[method]
}

func FindNotifier(method string) *notifications.Notification {
	notifiersLock.RLock()
	n := allNotifiers[method]
	notifiersLock.RUnlock()
	if n != nil {
		notif := n.Select()
		no, err := notifications.Find(notif.Method)
		if err != nil {
			log.Error(err)
			return nil
		}
		return notif.UpdateFields(no)
	}
	return nil
}

type ServiceNotifier interface {
	OnSuccess(Service) (string, error)                   // OnSuccess is triggered when a service is successful
	OnFailure(Service, failures.Failure) (string, error) // OnFailure is triggered when a service is failing
	OnTest() (string, error)                             // OnTest is triggered for testing
	OnSave() (string, error)                             // OnSave is triggered for testing
	Select() *notifications.Notification                 // OnTest is triggered for testing
	Valid(notifications.Values) error                    // Valid checks your form values
}
