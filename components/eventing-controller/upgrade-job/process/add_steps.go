package process

func (p *Process) AddSteps() {
	p.Steps = []Step{
		NewCheckIsBebEnabled(p),
		NewScaleDownEventingController(p),
		NewDeletePublisherDeployment(p),
		NewGetSubscriptions(p),
		NewFilterSubscriptions(p),
		NewDeleteBebSubscriptions(p),
	}
}