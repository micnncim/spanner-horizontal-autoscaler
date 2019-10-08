module github.com/micnncim/spanner-horizontal-autoscaler

go 1.12

require (
	cloud.google.com/go v0.38.0
	github.com/go-logr/logr v0.1.0
	github.com/go-logr/zapr v0.1.0
	github.com/golang/protobuf v1.3.2
	github.com/onsi/ginkgo v1.6.0
	github.com/onsi/gomega v1.4.2
	go.opencensus.io v0.22.1 // indirect
	go.uber.org/zap v1.9.1
	google.golang.org/api v0.10.0
	google.golang.org/genproto v0.0.0-20191007204434-a023cd5227bd
	google.golang.org/grpc v1.24.0 // indirect
	k8s.io/api v0.0.0-20190409021203-6e4e0e4f393b
	k8s.io/apimachinery v0.0.0-20190404173353-6a84e37a896d
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	sigs.k8s.io/controller-runtime v0.2.2
)
