package process

import (
	"fmt"
)

var _ Step = &DeleteBebSubscriptions{}

type DeleteBebSubscriptions struct {
	name    string
	process *Process
}

func NewDeleteBebSubscriptions(p *Process) DeleteBebSubscriptions {
	return DeleteBebSubscriptions{
		name:    "Get list of subscriptions",
		process: p,
	}
}

func (s DeleteBebSubscriptions) ToString() string {
	return s.name
}

func (s DeleteBebSubscriptions) Do() error {
	if !s.process.State.isBebEnabled {
		fmt.Println("BEB not enabled .. skipping")
		return nil
	}

	// s.process.State.Subscriptions.Items[0]

	//s.process.State.Subscriptions = subscriptionList

	return nil
}
