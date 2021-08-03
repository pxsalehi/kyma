package process

import (
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
)

var _ Step = &DeletePublisherDeployment{}

type DeletePublisherDeployment struct {
	name    string
	process *Process
}

func NewDeletePublisherDeployment(p *Process) DeletePublisherDeployment {
	return DeletePublisherDeployment{
		name:    "Delete eventing publisher deployment",
		process: p,
	}
}

func (s DeletePublisherDeployment) ToString() string {
	return s.name
}

func (s DeletePublisherDeployment) Do() error {
	if !s.process.State.IsBebEnabled {
		s.process.Logger.WithContext().Info("BEB not enabled .. skipping")
		return nil
	}

	// Get eventing-controller deployment object
	err := s.process.Clients.Deployment.Delete(s.process.KymaNamespace, s.process.PublisherName)
	// Ignore the error if its 404 error
	if err != nil && !k8serrors.IsNotFound(err){
		return err
	}

	return nil
}
