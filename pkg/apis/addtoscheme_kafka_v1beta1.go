package apis

import (
	"stash.us.cray.com/HMS/hms-trs-operator/pkg/apis/kafka/v1beta1"
)

func init() {
	// Register the types with the Scheme so the components can map objects to GroupVersionKinds and back
	AddToSchemes = append(AddToSchemes, v1beta1.SchemeBuilder.AddToScheme)
}
