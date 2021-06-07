module github.com/tamalsaha/kubebuilder-multi-apigroup

go 1.15

require (
	github.com/go-logr/logr v0.4.0
	github.com/nats-io/nats-server/v2 v2.2.6 // indirect
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.13.0
	go.bytebuilders.dev/audit v0.0.7
	go.bytebuilders.dev/license-verifier v0.9.2
	go.bytebuilders.dev/license-verifier/kubernetes v0.9.2
	k8s.io/apimachinery v0.21.1
	k8s.io/client-go v0.21.1
	kmodules.xyz/client-go v0.0.0-20210617233340-13d22e91512b
	sigs.k8s.io/controller-runtime v0.9.0
)

replace k8s.io/apimachinery => github.com/kmodules/apimachinery v0.21.1-rc.0.0.20210405112358-ad4c2289ba4c
