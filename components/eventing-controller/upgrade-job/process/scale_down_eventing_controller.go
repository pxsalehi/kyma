package process

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
	// Get eventing-controller deployment object
	oldDeployment, err := s.process.Clients.Deployment.Get(s.process.KymaNamespace, s.process.ControllerName)
	if (err != nil) {
		return err
	}

	// reduce replica count to zero
	desiredContainer := oldDeployment.DeepCopy()
	desiredContainer.Spec.Replicas = int32Ptr(1)

	_, err = s.process.Clients.Deployment.Update(s.process.KymaNamespace, desiredContainer)
	if (err != nil) {
		return err
	}

	return nil
}
