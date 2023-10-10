package main

import (
	"context"
	"flag"
	"path/filepath"

	appv1alpha1 "github.com/nickpoorman/hoper/api/app.nickpoorman.com/v1alpha1"
	"github.com/nickpoorman/hoper/client/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func main() {
	var kubeconfig string

	if home := homedir.HomeDir(); home != "" {
		flag.StringVar(&kubeconfig, "kubeconfig", filepath.Join(home, ".kube", "config"), "absolute path to the kubeconfig file")
	} else {
		flag.StringVar(&kubeconfig, "kubeconfig", "", "absolute path to the kubeconfig file")
	}

	var tenantName string
	flag.StringVar(&tenantName, "tenant-name", "", "name of the tenant")

	flag.Parse()

	if tenantName == "" {
		flag.PrintDefaults()
		panic("tenant-name is required")
	}

	// Build config from the kubeconfig file
	println("Using kubeconfig: " + kubeconfig)
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// Create the clientset
	clientset, err := versioned.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	tenant := &appv1alpha1.Tenant{
		ObjectMeta: metav1.ObjectMeta{
			Name: tenantName,
		},
		Spec: appv1alpha1.TenantSpec{
			Name: tenantName,
		},
		Status: appv1alpha1.TenantStatus{
			InstanceCreated: false,
			Conditions:      []metav1.Condition{},
		},
	}

	// Use the clientset to create the Tenant CR
	_, err = clientset.AppV1alpha1().Tenants("default").Create(context.TODO(), tenant, metav1.CreateOptions{})
	if err != nil {
		panic(err.Error())
	}

	println("Tenant CR created!")
}
