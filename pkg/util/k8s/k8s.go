package k8s

import "io/ioutil"

func CurrentNamespace() (string, error) {
	bytes, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	return string(bytes), err
}
