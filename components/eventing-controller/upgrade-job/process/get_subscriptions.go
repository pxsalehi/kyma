package process

import (
	"fmt"
	corev1 "k8s.io/api/core/v1"
)

var _ Step = &GetSubscriptions{}

type GetSubscriptions struct {
	name    string
	process *Process
}

func NewGetSubscriptions(p *Process) GetSubscriptions {
	return GetSubscriptions{
		name:    "Get list of subscriptions",
		process: p,
	}
}

func (s GetSubscriptions) ToString() string {
	return s.name
}

func (s GetSubscriptions) Do() error {
	if !s.process.State.IsBebEnabled {
		fmt.Println("BEB not enabled .. skipping")
		return nil
	}

	namespace := corev1.NamespaceAll

	subscriptionList, err := s.process.Clients.Subscription.List(namespace)
	if (err != nil) {
		return err
	}

	fmt.Println(len(subscriptionList.Items))

	s.process.State.Subscriptions = subscriptionList

	return nil
}
