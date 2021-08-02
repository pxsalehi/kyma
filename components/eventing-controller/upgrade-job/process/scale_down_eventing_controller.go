package process

import (
	"fmt"
	"github.com/pkg/errors"
	"time"
)

var _ Step = &ScaleDownEventingController{}

type ScaleDownEventingController struct {
	name    string
	process *Process
}

func NewScaleDownEventingController(p *Process) ScaleDownEventingController {
	return ScaleDownEventingController{
		name:    "Scale down eventing controller to zero replicas",
		process: p,
	}
}

func (s ScaleDownEventingController) ToString() string {
	return s.name
}

func (s ScaleDownEventingController) Do() error {
	if !s.process.State.isBebEnabled {
		fmt.Println("BEB not enabled .. skipping")
		return nil
	}

	// Get eventing-controller deployment object
	oldDeployment, err := s.process.Clients.Deployment.Get(s.process.KymaNamespace, s.process.ControllerName)
	if (err != nil) {
		return err
	}

	// reduce replica count to zero
	desiredContainer := oldDeployment.DeepCopy()
	desiredContainer.Spec.Replicas = int32Ptr(0)

	_, err = s.process.Clients.Deployment.Update(s.process.KymaNamespace, desiredContainer)
	if (err != nil) {
		return err
	}

	// @TODO: check if we need to wait
	// Wait until pod down
	isScaledDownSuccess := false

	start := time.Now()
	for time.Since(start) < s.process.TimeoutPeriod {
		fmt.Println("Checking status")

		time.Sleep(1 * time.Second)

		controllerDeployment, err := s.process.Clients.Deployment.Get(s.process.KymaNamespace, s.process.ControllerName)
		if err != nil {
			// @TODO: print error or someting
			continue
		}

		if controllerDeployment.Status.Replicas == 0 {
			fmt.Println("Controller down success!!!")
			isScaledDownSuccess = true
			break
		}
	}

	if !isScaledDownSuccess {
		return errors.New("Timeout! Failed to scale down eventing controller")
	}

	return nil
}
