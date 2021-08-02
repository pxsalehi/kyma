package process

func (p *Process) AddSteps() {
	p.Steps = []Step{
		NewCheckIsBebEnabled(p),
		NewScaleDownEventingController(p),
		NewGetSubscriptions(p),
	}
}