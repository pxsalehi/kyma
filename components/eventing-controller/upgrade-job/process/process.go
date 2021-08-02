package process

import "time"

type Process struct {
	Steps           []Step
	ReleaseName  string
	BEBNamespace string
	KymaNamespace string
	ControllerName string
	Clients         Clients
	State           State
	TimeoutPeriod time.Duration
}
