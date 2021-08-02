package main

import (
	"flag"
	"fmt"
	"github.com/kyma-project/kyma/components/eventing-controller/upgrade-job/clients/deployment"
	jobprocess "github.com/kyma-project/kyma/components/eventing-controller/upgrade-job/process"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
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
	kubeconfig = flag.String("kubeconfig", "$HOME/kubeconfigs/kubeconfig--kymatunas--fzn-d1.yml", "absolute path to the kubeconfig file")
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}


	// setup clients
	deploymentClient := deployment.NewClient(clientset)

	// Create process
	p := jobprocess.Process{
		ReleaseName:  "TEST RS Name 1.24.x",
		BEBNamespace: "TEST BEB NS",
		Clients: jobprocess.Clients{
			Deployment: deploymentClient,
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
