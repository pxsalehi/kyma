package main

import (
	"fmt"
	"github.com/kyma-project/kyma/components/eventing-controller/upgrade-job/clients/deployment"
	eventmesh "github.com/kyma-project/kyma/components/eventing-controller/upgrade-job/clients/event-mesh"
	eventingbackend "github.com/kyma-project/kyma/components/eventing-controller/upgrade-job/clients/eventing-backend"
	"github.com/kyma-project/kyma/components/eventing-controller/upgrade-job/clients/secret"
	"github.com/kyma-project/kyma/components/eventing-controller/upgrade-job/clients/subscription"
	jobprocess "github.com/kyma-project/kyma/components/eventing-controller/upgrade-job/process"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"time"
)

func main() {
	// @TODO: create logger instance

	// Generate dynamic clients
	k8sConfig := config.GetConfigOrDie()

	// Create dynamic client
	dynamicClient := dynamic.NewForConfigOrDie(k8sConfig)

	// setup clients
	deploymentClient := deployment.NewClient(dynamicClient)
	subscriptionClient := subscription.NewClient(dynamicClient)
	eventingBackendClient := eventingbackend.NewClient(dynamicClient)
	secretClient := secret.NewClient(dynamicClient)
	eventMeshClient := eventmesh.NewClient()

	// Create process
	p := jobprocess.Process{
		TimeoutPeriod: 60 * time.Second,
		ReleaseName:  "TEST RS Name 1.24.x",
		KymaNamespace: "kyma-system",
		ControllerName: "eventing-controller",
		PublisherName: "eventing-publisher-proxy",
		Clients: jobprocess.Clients{
			Deployment: deploymentClient,
			Subscription: subscriptionClient,
			EventingBackend: eventingBackendClient,
			Secret: secretClient,
			EventMesh: eventMeshClient,
		},
	}

	// Add steps to process
	p.AddSteps()

	// Execute process
	err := p.Execute()
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Completed upgrade-hook main 1.24.x")
}
