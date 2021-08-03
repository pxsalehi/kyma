package process

import eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"

var _ Step = &CheckIsBebEnabled{}

type CheckIsBebEnabled struct {
	name    string
	process *Process
}

func NewCheckIsBebEnabled(p *Process) CheckIsBebEnabled {
	return CheckIsBebEnabled{
		name:    "Check if BEB enabled",
		process: p,
	}
}

func (s CheckIsBebEnabled) ToString() string {
	return s.name
}

func (s CheckIsBebEnabled) Do() error {
	eventingbackendObj, err := s.process.Clients.EventingBackend.Get(s.process.KymaNamespace, "eventing-backend")
	if (err != nil) {
		return err
	}

	s.process.State.IsBebEnabled = false
	if eventingbackendObj.Status.Backend == eventingv1alpha1.BebBackendType {
		s.process.State.IsBebEnabled = true
	}

	return nil
}
