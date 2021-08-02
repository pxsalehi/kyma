package process

type Process struct {
	Steps           []Step
	ReleaseName  string
	BEBNamespace string
	Clients         Clients
}
