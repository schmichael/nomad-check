package main

import "github.com/hashicorp/nomad/api"

type Backend interface {
	GetAlloc(allocID string) (*api.Allocation, error)
	GetNode(nodeID string) (*api.Node, error)
	ListAllocs() ([]*api.AllocationListStub, error)
	ListNamespaces() ([]*api.Namespace, error)
	ListNodes() ([]*api.NodeListStub, error)
}
