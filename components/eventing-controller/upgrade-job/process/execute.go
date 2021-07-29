package process

import (
	"fmt"
)

func (p Process) Execute() error {

	fmt.Println("!! Hello world from upgrade-hook process-main 1.24.xxx")

	return nil
}
