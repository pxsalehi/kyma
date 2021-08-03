package process

var _ Step = &FilterSubscriptions{}

type FilterSubscriptions struct {
	name    string
	process *Process
}

func NewFilterSubscriptions(p *Process) FilterSubscriptions {
	return FilterSubscriptions{
		name:    "Get list of subscriptions",
		process: p,
	}
}

func (s FilterSubscriptions) ToString() string {
	return s.name
}

func (s FilterSubscriptions) Do() error {
	if !s.process.State.IsBebEnabled {
		s.process.Logger.WithContext().Info("BEB not enabled .. skipping")
		return nil
	}

	//@TODO: Filter subscriptions based on if not migrated acc. to new naming convention
	s.process.State.FilteredSubscriptions = s.process.State.Subscriptions.DeepCopy()

	return nil
}
