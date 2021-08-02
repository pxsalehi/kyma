package process

import "github.com/mfaizanse/kyma/components/eventing-controller/upgrade-job/clients/deployment"

type Clients struct {
	Deployment       deployment.Client
}
