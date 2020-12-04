package controller

import (
	"stash.us.cray.com/HMS/hms-trs-operator/pkg/controller/trsworker"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, trsworker.Add)
}
