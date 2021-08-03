package process

import (
	"errors"
	"fmt"
	"github.com/kyma-project/kyma/components/eventing-controller/reconciler/backend"
	corev1 "k8s.io/api/core/v1"
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
	if !s.process.State.IsBebEnabled {
		fmt.Println("BEB not enabled .. skipping")
		return nil
	}

	// Get BEB k8s secret
	secretLabel := backend.BEBBackendSecretLabelKey + "=" + backend.BEBBackendSecretLabelValue
	secretList, err := s.process.Clients.Secret.ListByMatchingLabels(corev1.NamespaceAll, secretLabel)
	if err != nil {
		return err
	}
	if len(secretList.Items) == 0 {
		return errors.New("no BEB secrets found")
	}
	if len(secretList.Items) > 1 {
		return errors.New("more than 1 BEB secrets found")
	}

	// Initialize BEB client with this secret
	err = s.process.Clients.EventMesh.Init(&secretList.Items[0])
	if err != nil {
		return err
	}

	// Traverse through the subscriptions and migrate
	subscriptionListItems := s.process.State.FilteredSubscriptions.Items

	for _, subscription := range subscriptionListItems {
		fmt.Println("Deleting: ", subscription.Name)
		result, err := s.process.Clients.EventMesh.Delete(subscription.Name)
		if err != nil {
			fmt.Println(err)
			continue
		}

		fmt.Println(result.StatusCode, result.Message)

		// @TODO: should we check for response
	}

	return nil
}
