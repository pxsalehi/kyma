module github.com/kyma-project/kyma/components/eventing-controller

go 1.14

require (
	github.com/avast/retry-go/v3 v3.1.1
	github.com/cloudevents/sdk-go/protocol/nats/v2 v2.5.0
	github.com/cloudevents/sdk-go/v2 v2.5.0
	github.com/go-logr/logr v0.4.0
	github.com/go-logr/zapr v0.4.0
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/kyma-incubator/api-gateway v0.0.0-20211012135230-f748d862a3ba
	github.com/kyma-project/kyma/common/logging v0.0.0-20210601142757-445a3b6021fe
	github.com/kyma-project/kyma/components/application-operator v0.0.0-20210204131215-a368a90f2525
	github.com/mitchellh/hashstructure/v2 v2.0.2
	github.com/nats-io/nats-server/v2 v2.3.4
	github.com/nats-io/nats.go v1.11.1-0.20210623165838-4b75fc59ae30
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.16.0
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.7.0
	go.uber.org/multierr v1.6.0 // indirect
	go.uber.org/zap v1.16.0
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
	golang.org/x/tools v0.1.0 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1
	k8s.io/api v0.20.7
	k8s.io/apiextensions-apiserver v0.20.7 // indirect
	k8s.io/apimachinery v0.22.2
	k8s.io/client-go v0.20.7
	sigs.k8s.io/controller-runtime v0.8.3
)

replace github.com/nats-io/nats.go => github.com/nats-io/nats.go v1.11.0

replace github.com/gogo/protobuf v1.3.1 => github.com/gogo/protobuf v1.3.2

replace github.com/go-logr/zapr => github.com/go-logr/zapr v0.3.0
