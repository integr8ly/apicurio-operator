package controller

import (
	"github.com/integr8ly/apicurio-operator/pkg/controller/apicuriodeployment"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, apicuriodeployment.Add)
}
