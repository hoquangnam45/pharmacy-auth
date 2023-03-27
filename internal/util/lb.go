package util

import (
	"fmt"

	"github.com/hoquangnam45/pharmacy-common-go/microservice/consul"
)

type LoadBalancer struct {
	clusterPrefix string
	consulClient  *consul.Client
}

func NewClusterLoadBalancer(clusterPrefix string, consulClient *consul.Client) *LoadBalancer {
	return &LoadBalancer{clusterPrefix, consulClient}
}

func (a *LoadBalancer) Cluster(serviceName string) string {
	if a.clusterPrefix == "" {
		return serviceName
	}
	return fmt.Sprintf("%s-%s", a.clusterPrefix, serviceName)
}

func (a *LoadBalancer) LoadBalancing(serviceName string) (string, error) {
	svc := a.Cluster(serviceName)
	return a.consulClient.LoadBalancing(svc)
}
