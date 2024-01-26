// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/nomad/api"
)

func main() {
	logger := hclog.Default()

	// Flags
	pendingDur := 12 * time.Hour
	flag.DurationVar(&pendingDur, "pending", pendingDur, "duration an alloc is allowed to be pending before reported")

	outFn := "nomad-check.json"
	flag.StringVar(&outFn, "out", outFn, "file name for writing results")

	allocsFn := ""
	flag.StringVar(&allocsFn, "allocs", allocsFn, "use a json file instead of http")
	nodesFn := ""
	flag.StringVar(&nodesFn, "nodes", nodesFn, "use a json file instead of http")
	flag.Parse()

	out, err := os.Create(outFn)
	if err != nil {
		logger.Error("error creating output file", "error", err)
		os.Exit(2)
	}
	defer out.Close() // just in case of early exit

	var backend Backend

	if allocsFn == "" && nodesFn == "" {
		logger.Info("Using HTTP API")
		conf := api.DefaultConfig()
		conf.Namespace = "*"
		nomad, err := api.NewClient(conf)
		if err != nil {
			logger.Error("error creating client", "error", err)
			os.Exit(1)
		}
		backend = &HTTPBackend{
			client: nomad,
			logger: logger,
		}
	} else {
		logger.Info("Using files", "allocs", allocsFn, "nodes", nodesFn)
		b, err := NewFileBackend(allocsFn, nodesFn)
		if err != nil {
			logger.Error("error opening files", "error", err)
			os.Exit(1)
		}
		backend = b
	}

	c := &checker{
		backend: backend,
		logger:  logger,
		oldAge:  time.Now().Add(-1 * pendingDur),
	}

	results, err := c.check()
	if err != nil {
		logger.Error("check failed", "error", err)
		os.Exit(97)
	}

	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	if err := enc.Encode(results); err != nil {
		logger.Error("error encoding results", "error", err)
		os.Exit(98)
	}

	if err := out.Close(); err != nil {
		logger.Error("error closing output file", "error", err)
		os.Exit(99)
	}

	logger.Info("Completed!", "results", outFn)
}

type Backend interface {
	ListAllocs() ([]*api.AllocationListStub, error)
	ListNodes() ([]*api.NodeListStub, error)
	GetAlloc(allocID string) (*api.Allocation, error)
	GetNode(nodeID string) (*api.Node, error)
}

type checker struct {
	logger  hclog.Logger
	backend Backend

	oldAge time.Time
}

type Results struct {
	Complete             bool
	NodesTotal           int
	AllocsTotal          int
	AllocsClientTerminal int
	AllocsPendingTooLong []string
	AllocsMissingNode    []string
	AllocsDownNode       []string
	Allocs               map[string]*api.Allocation
	Nodes                map[string]*api.Node
}

func NewResults() *Results {
	return &Results{
		Allocs: map[string]*api.Allocation{},
		Nodes:  map[string]*api.Node{},
	}
}

// TODO pass in a context for signal based cancellation and returning of partial results
func (c checker) check() (*Results, error) {
	r := NewResults()
	c.logger.Info("Fetching all nodes...")
	nodes, err := getNodes(c.backend)
	if err != nil {
		return r, fmt.Errorf("error listing nodes: %w", err)
	}
	r.NodesTotal = len(nodes)

	c.logger.Info("Fetching all allocations...")
	allocs, err := c.backend.ListAllocs()
	if err != nil {
		return r, fmt.Errorf("error listing allocations: %w", err)
	}
	r.AllocsTotal = len(allocs)

	c.logger.Info("Checking allocations...")
	for _, alloc := range allocs {
		switch alloc.ClientStatus {
		case api.AllocClientStatusComplete, api.AllocClientStatusFailed, api.AllocClientStatusLost:
			r.AllocsClientTerminal++
		}

		// Check if non-terminal alloc's been pending too long
		if mod := time.Unix(0, alloc.ModifyTime); alloc.ClientStatus == "pending" && mod.Before(c.oldAge) {
			c.logger.Warn("Allocation has been pending for too long", "alloc", alloc.ID, "modified", mod)
			r.AllocsPendingTooLong = append(r.AllocsPendingTooLong, alloc.ID)
		}

		// Check if non-terminal alloc's node exists
		n, ok := nodes[alloc.NodeID]
		if !ok {
			c.logger.Warn("Non-terminal allocation's node missing", "alloc", alloc.ID, "node", alloc.NodeID)
			r.AllocsMissingNode = append(r.AllocsMissingNode, alloc.ID)
			continue
		}

		// Check if non-terminal alloc's node is up
		if n.Status == "down" {
			c.logger.Warn("Non-terminal allocation's node down", "alloc", alloc.ID, "node", alloc.NodeID)
			r.AllocsDownNode = append(r.AllocsMissingNode, alloc.ID)
		}
	}

	for _, allocList := range [][]string{r.AllocsPendingTooLong, r.AllocsMissingNode, r.AllocsDownNode} {
		for _, allocid := range allocList {
			if _, ok := r.Allocs[allocid]; ok {
				continue
			}

			alloc, err := c.backend.GetAlloc(allocid)
			r.Allocs[allocid] = alloc
			if err != nil {
				c.logger.Error("Error fetching alloc", "error", err, "alloc", allocid)
				continue
			}

			if _, ok := r.Nodes[alloc.NodeID]; ok {
				continue
			}
			node, err := c.backend.GetNode(alloc.NodeID)
			if err != nil {
				c.logger.Error("Error fetching node for alloc", "error", err, "node", alloc.NodeID, "alloc", allocid)
			}
			r.Nodes[alloc.NodeID] = node
		}
	}

	// Mark as Complete as partial results are possible
	r.Complete = true
	return r, nil
}

func getNodes(b Backend) (map[string]*api.NodeListStub, error) {
	s, err := b.ListNodes()
	if err != nil {
		return nil, err
	}

	m := make(map[string]*api.NodeListStub, len(s))
	for _, n := range s {
		m[n.ID] = n
	}

	return m, nil
}
