package eventmesh

import (
	emsclient "github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/api/events/client"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/api/events/types"
)

type Client struct {
	client *emsclient.Client
}

func NewClient(client *emsclient.Client) Client {
	//authenticator := auth.NewAuthenticator(cfg)
	// client.NewClient(config.GetDefaultConfig(cfg.BebApiUrl), authenticator)
	return Client{client}
}

func (c Client) Delete(bebSubscriptionName string) (*types.DeleteResponse, error) {
	return c.client.Delete(bebSubscriptionName)
}


