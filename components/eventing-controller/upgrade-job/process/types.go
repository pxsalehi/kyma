package process

import (
	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/upgrade-job/clients/deployment"
	eventingbackend "github.com/kyma-project/kyma/components/eventing-controller/upgrade-job/clients/eventing-backend"
	"github.com/kyma-project/kyma/components/eventing-controller/upgrade-job/clients/subscription"
)

type Clients struct {
	Deployment       deployment.Client
	Subscription 	 subscription.Client
	EventingBackend  eventingbackend.Client
}

type State struct {
	Subscriptions          *eventingv1alpha1.SubscriptionList
	isBebEnabled           bool
}

func int32Ptr(i int32) *int32 { return &i }