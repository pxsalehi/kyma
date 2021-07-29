package main

import (
	"fmt"
	jobprocess "github.com/kyma-project/kyma/components/eventing-controller/upgrade-job/process"
)

func main() {
	fmt.Println("##### Hello world from upgrade-hook main 1.24.xxx")

	// Create process
	p := jobprocess.Process{
		ReleaseName:  "TEST RS Name 1.24.x",
		BEBNamespace: "TEST BEB NS",
	}

	// Execute process
	err := p.Execute()
	if err != nil {
		fmt.Println("Upgrade process failed")
		fmt.Println(err)
	}

	fmt.Println("Completed upgrade-hook main 1.24.x")
}
