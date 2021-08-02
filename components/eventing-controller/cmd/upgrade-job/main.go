package main

import (
	"flag"
	"fmt"
	"github.com/kyma-project/kyma/components/eventing-controller/upgrade-job/clients/deployment"
	eventingbackend "github.com/kyma-project/kyma/components/eventing-controller/upgrade-job/clients/eventing-backend"
	"github.com/kyma-project/kyma/components/eventing-controller/upgrade-job/clients/subscription"
	jobprocess "github.com/kyma-project/kyma/components/eventing-controller/upgrade-job/process"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"time"

	//ctrlClient "sigs.k8s.io/controller-runtime/pkg/client"
)

func main() {
	// @TODO: create logger instance


	fmt.Println("##### Hello world from upgrade-hook main 1.24.xxx")

	// Create k8s client set
	var kubeconfig *string
	//if home := homedir.HomeDir(); home != "" {
	//	kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	//} else {
	//	kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	//}
	kubeconfig = flag.String("kubeconfig", "/Users/faizan/kubeconfigs/kubeconfig--kymatunas--fzn-d1.yml", "absolute path to the kubeconfig file")
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}

	// Create typed client
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	// Create dynamic client
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	// setup clients
	deploymentClient := deployment.NewClient(clientset)
	subscriptionClient := subscription.NewClient(dynamicClient)
	eventingbackendClient := eventingbackend.NewClient(dynamicClient)

	// Create process
	p := jobprocess.Process{
		TimeoutPeriod: 60 * time.Second,
		ReleaseName:  "TEST RS Name 1.24.x",
		BEBNamespace: "TEST BEB NS",
		KymaNamespace: "kyma-system",
		ControllerName: "eventing-controller",
		Clients: jobprocess.Clients{
			Deployment: deploymentClient,
			Subscription: subscriptionClient,
			EventingBackend: eventingbackendClient,
		},
	}

	// Add steps to process
	p.AddSteps()

	// Execute process
	err = p.Execute()
	if err != nil {
		fmt.Println("Upgrade process failed")
		fmt.Println(err)
	}

	fmt.Println("Completed upgrade-hook main 1.24.x")
}
