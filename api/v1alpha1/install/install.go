package install

import (
	v1alpha1 "github.com/nickpoorman/hoper/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
)

// Install adds the v1alpha1 group to a given scheme.
func Install(scheme *runtime.Scheme) {
	v1alpha1.AddToScheme(scheme)
}
