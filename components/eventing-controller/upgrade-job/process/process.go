package process

import "time"

type Process struct {
	Steps           []Step
	ReleaseName  string
	KymaNamespace string
	ControllerName string
	PublisherName string
	Clients         Clients
	State           State
	TimeoutPeriod time.Duration
}
