package process

func (p *Process) AddSteps() {
	p.Steps = []Step{
		NewScaleDownEventingController(p),
	}
}