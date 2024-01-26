// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/nomad/api"
)

type HTTPBackend struct {
	client *api.Client

	logger hclog.Logger
}

func (b *HTTPBackend) GetAlloc(allocID string) (*api.Allocation, error) {
	alloc, _, err := b.client.Allocations().Info(allocID, nil)
	return alloc, err
}

func (b *HTTPBackend) GetNode(nodeID string) (*api.Node, error) {
	node, _, err := b.client.Nodes().Info(nodeID, nil)
	return node, err
}

func (b *HTTPBackend) ListAllocs() ([]*api.AllocationListStub, error) {
	allocs, _, err := b.client.Allocations().List(nil)
	return allocs, err
}

func (b *HTTPBackend) ListNamespaces() ([]*api.Namespace, error) {
	nses, _, err := b.client.Namespaces().List(nil)
	return nses, err
}

func (b *HTTPBackend) ListNodes() ([]*api.NodeListStub, error) {
	nodes, _, err := b.client.Nodes().List(nil)
	return nodes, err
}
