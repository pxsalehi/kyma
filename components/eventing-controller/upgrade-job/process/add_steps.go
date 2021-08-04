package process

func (p *Process) AddSteps() {
	p.Steps = []Step{
		NewScaleDownEventingController(p),
		NewDeletePublisherDeployment(p),
		NewGetSubscriptions(p),
		NewFilterSubscriptions(p),
		NewDeleteBebSubscriptions(p),
	}
}