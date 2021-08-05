package process

import (
	eventmesh "github.com/kyma-project/kyma/components/eventing-controller/upgrade-job/clients/event-mesh"
	eventingbackend "github.com/kyma-project/kyma/components/eventing-controller/upgrade-job/clients/eventing-backend"
	"github.com/onsi/gomega"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/upgrade-job/clients/deployment"
	"github.com/kyma-project/kyma/components/eventing-controller/upgrade-job/clients/secret"
	"github.com/kyma-project/kyma/components/eventing-controller/upgrade-job/clients/subscription"
	"github.com/kyma-project/kyma/components/eventing-controller/upgrade-job/processtest"

	env "github.com/kyma-project/kyma/components/eventing-controller/upgrade-job/env"
)

type E2ESetup struct {
	secrets     			*corev1.SecretList
	eventingPublishers      *appsv1.DeploymentList
	eventingControllers     *appsv1.DeploymentList
	eventingBackends        *eventingv1alpha1.EventingBackendList
	subscriptions 			*eventingv1alpha1.SubscriptionList
	namespaces       		*corev1.NamespaceList
	config  				env.Config
}

func getProcessClients(e2eSetup E2ESetup, g *gomega.GomegaWithT) Clients {
	fakeSecretClient, err := secret.NewFakeClient(e2eSetup.secrets)
	g.Expect(err).Should(gomega.BeNil())
	fakeDeploymentClient, err := deployment.NewFakeClient(combineEventingControllersAndPublishers(e2eSetup.eventingControllers, e2eSetup.eventingPublishers))
	g.Expect(err).Should(gomega.BeNil())
	fakeSubscriptionClient, err := subscription.NewFakeClient(e2eSetup.subscriptions)
	g.Expect(err).Should(gomega.BeNil())
	fakeEventingBackendClient, err := eventingbackend.NewFakeClient(e2eSetup.eventingBackends)
	g.Expect(err).Should(gomega.BeNil())
	fakeEventMeshClient, err := eventmesh.NewFakeClient()
	g.Expect(err).Should(gomega.BeNil())


	return Clients{
		Deployment: fakeDeploymentClient,
		Subscription: fakeSubscriptionClient,
		EventingBackend: fakeEventingBackendClient,
		Secret: fakeSecretClient,
		EventMesh: fakeEventMeshClient,
	}
}

func newE2ESetup() E2ESetup {
	newEventingControllers := processtest.NewEventingControllers()
	newEventingPublishers := processtest.NewEventingPublishers()
	newSecrets := processtest.NewSecrets()
	newSubscriptions := processtest.NewKymaSubscriptions()
	newEventingBackends := processtest.NewEventingBackends()

	//newNamespaces := processtest.NewNamespaces()

	envConfig := env.Config{
		ReleaseName:            "release",
		KymaNamespace:          "kyma-system",
		EventingControllerName: "eventing-controller",
		EventingPublisherName:  "eventing-publisher-proxy",
		LogFormat: "json",
		LogLevel: "warn",
	}

	e2eSetup := E2ESetup{
		config: envConfig,
		secrets: newSecrets,
		eventingPublishers: newEventingPublishers,
		eventingControllers: newEventingControllers,
		eventingBackends: newEventingBackends,
		subscriptions: newSubscriptions,
		//namespaces:       newNamespaces,
	}
	return e2eSetup
}

func combineEventingControllersAndPublishers(validators, eventServices *appsv1.DeploymentList) *appsv1.DeploymentList {
	result := new(appsv1.DeploymentList)
	for _, v := range validators.Items {
		result.Items = append(result.Items, v)
	}
	for _, es := range eventServices.Items {
		result.Items = append(result.Items, es)
	}
	return result
}
