package process

import "fmt"

var _ Step = &ScaleDownEventingController{}

type ScaleDownEventingController struct {
	name    string
	process *Process
}

func NewScaleDownEventingController(p *Process) ScaleDownEventingController {
	return ScaleDownEventingController{
		name:    "Scale down eventing controller to zero replicas",
		process: p,
	}
}

func (s ScaleDownEventingController) ToString() string {
	return s.name
}

func (s ScaleDownEventingController) Do() error {

	deployment, err := s.process.Clients.Deployment.Get("kyma-system", "eventing-controller")
	if (err != nil) {
		return err
	}

	fmt.Println(deployment.Name)

	return nil
}