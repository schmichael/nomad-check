// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/hashicorp/nomad/api"
)

type FileBackend struct {
	allocs []*api.AllocationListStub
	nodes  []*api.NodeListStub
}

func NewFileBackend(allocsFn, nodesFn string) (*FileBackend, error) {
	allocsFd, err := os.Open(allocsFn)
	if err != nil {
		return nil, fmt.Errorf("error opening allocs: %w", err)
	}
	defer allocsFd.Close()

	nodesFd, err := os.Open(nodesFn)
	if err != nil {
		return nil, fmt.Errorf("error opening nodes: %w", err)
	}
	defer nodesFd.Close()

	b := FileBackend{
		allocs: []*api.AllocationListStub{},
		nodes:  []*api.NodeListStub{},
	}

	dec := json.NewDecoder(allocsFd)
	for dec.More() {
		a := &api.AllocationListStub{}
		if err := dec.Decode(a); err != nil {
			return nil, fmt.Errorf("error decoding allocs: %w", err)
		}
		b.allocs = append(b.allocs, a)
	}

	if len(b.allocs) == 0 {
		return nil, fmt.Errorf("no allocs found")
	}

	dec = json.NewDecoder(nodesFd)
	for dec.More() {
		n := &api.NodeListStub{}
		if err := dec.Decode(n); err != nil {
			return nil, fmt.Errorf("error decoding nodes: %w", err)
		}
		b.nodes = append(b.nodes, n)
	}

	return &b, nil
}

func (f *FileBackend) ListAllocs() ([]*api.AllocationListStub, error) {
	return f.allocs, nil
}

func (f *FileBackend) ListNodes() ([]*api.NodeListStub, error) {
	return f.nodes, nil
}

func (f *FileBackend) GetAlloc(allocID string) (*api.Allocation, error) {
	var stub *api.AllocationListStub
	for _, a := range f.allocs {
		if a.ID == allocID {
			stub = a
			break
		}
	}

	if stub == nil {
		return nil, fmt.Errorf("alloc id %s not found", allocID)
	}

	return &api.Allocation{
		ID:                    stub.ID,
		EvalID:                stub.EvalID,
		Name:                  stub.Name,
		Namespace:             stub.Namespace,
		NodeID:                stub.NodeID,
		NodeName:              stub.NodeName,
		JobID:                 stub.JobID,
		TaskGroup:             stub.TaskGroup,
		AllocatedResources:    stub.AllocatedResources,
		DesiredStatus:         stub.DesiredStatus,
		DesiredDescription:    stub.DesiredDescription,
		ClientStatus:          stub.ClientStatus,
		ClientDescription:     stub.ClientDescription,
		TaskStates:            stub.TaskStates,
		DeploymentStatus:      stub.DeploymentStatus,
		FollowupEvalID:        stub.FollowupEvalID,
		NextAllocation:        stub.NextAllocation,
		RescheduleTracker:     stub.RescheduleTracker,
		PreemptedAllocations:  stub.PreemptedAllocations,
		PreemptedByAllocation: stub.PreemptedByAllocation,
		CreateIndex:           stub.CreateIndex,
		ModifyIndex:           stub.ModifyIndex,
		CreateTime:            stub.CreateTime,
		ModifyTime:            stub.ModifyTime,
	}, nil
}

func (f *FileBackend) GetNode(nodeID string) (*api.Node, error) {
	var stub *api.NodeListStub
	for _, n := range f.nodes {
		if n.ID == nodeID {
			stub = n
			break
		}
	}

	if stub == nil {
		return nil, fmt.Errorf("node id %s not found", nodeID)
	}

	return &api.Node{
		ID:                    stub.ID,
		Attributes:            stub.Attributes,
		Datacenter:            stub.Datacenter,
		Name:                  stub.Name,
		NodeClass:             stub.NodeClass,
		NodePool:              stub.NodePool,
		Drain:                 stub.Drain,
		SchedulingEligibility: stub.SchedulingEligibility,
		Status:                stub.Status,
		StatusDescription:     stub.StatusDescription,
		Drivers:               stub.Drivers,
		NodeResources:         stub.NodeResources,
		ReservedResources:     stub.ReservedResources,
		LastDrain:             stub.LastDrain,
		CreateIndex:           stub.CreateIndex,
		ModifyIndex:           stub.ModifyIndex,
	}, nil
}
