package test

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Cluster represents a testing cluster
type Cluster interface {
	// Name is a unique string name by which this cluster can be identified, generally a UUID
	Name() string

	// Client produces a client for communicating with the Kubernetes cluster
	Client() kubernetes.Interface

	// Config produces a rest.Config from which the caller could generate their own clients
	Config() *rest.Config

	// ConfigPath produces a file path to the rest.Config for use with tools like kubectl
	ConfigPath() string

	// Cleanup performs all necessary cleanup for the cluster when the caller is done using it
	Cleanup() error
}
