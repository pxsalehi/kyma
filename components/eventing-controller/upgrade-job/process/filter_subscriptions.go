package process

import eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"

var _ Step = &FilterSubscriptions{}

type FilterSubscriptions struct {
	name    string
	process *Process
}

func NewFilterSubscriptions(p *Process) FilterSubscriptions {
	return FilterSubscriptions{
		name:    "Filter subscriptions based on migration",
		process: p,
	}
}

func (s FilterSubscriptions) ToString() string {
	return s.name
}

func (s FilterSubscriptions) Do() error {
	//@TODO: Filter subscriptions based on if not migrated acc. to new naming convention
	s.process.State.FilteredSubscriptions = s.process.State.Subscriptions.DeepCopy()

	//1) first generate the new name for webhook
	//2) Check the condition
	//3) if not in condition, then check if

	return nil
}


func MapSubscriptionName(sub *eventingv1alpha1.Subscription) string {
	// Mock function to be replaced
	return sub.Name
}
