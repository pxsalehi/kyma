package process

type Process struct {
	Steps           []Step
	ReleaseName  string
	BEBNamespace string
	KymaNamespace string
	ControllerName string
	Clients         Clients
	State           State
}
