/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2025. All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package loadbalance provides roundrobin algorithm
package loadbalance

import (
	"math/rand"
	"time"
)

// WeightNginx weight nginx
type WeightNginx struct {
	Node            interface{}
	Weight          int
	CurrentWeight   int
	EffectiveWeight int
}

// WNGINX w nginx
type WNGINX struct {
	nodes []*WeightNginx
}

// Add add node
func (w *WNGINX) Add(node interface{}, weight int) {
	weightNginx := &WeightNginx{
		Node:            node,
		Weight:          weight,
		EffectiveWeight: weight}
	w.nodes = append(w.nodes, weightNginx)
}

// Remove removes a node
func (w *WNGINX) Remove(node interface{}) {
	for i, weighted := range w.nodes {
		if weighted.Node == node {
			w.nodes = append(w.nodes[:i], w.nodes[i+1:]...)
			break
		}
	}
}

// RemoveAll remove all nodes
func (w *WNGINX) RemoveAll() {
	w.nodes = w.nodes[:0]
}

// Next get next node
func (w *WNGINX) Next(_ string, _ bool) interface{} {
	if len(w.nodes) == 0 {
		return nil
	}
	if len(w.nodes) == 1 {
		return w.nodes[0].Node
	}
	return nextWeightedNode(w.nodes).Node
}

// nextWeightedNode get best next node info
func nextWeightedNode(nodes []*WeightNginx) *WeightNginx {
	total := 0
	if len(nodes) == 0 {
		return nil
	}
	best := nodes[0]
	for _, w := range nodes {
		w.CurrentWeight += w.EffectiveWeight
		total += w.EffectiveWeight
		if w.CurrentWeight > best.CurrentWeight {
			best = w
		}
	}
	best.CurrentWeight -= total
	return best
}

// Reset reset all nodes
func (w *WNGINX) Reset() {
	for _, s := range w.nodes {
		s.EffectiveWeight = s.Weight
		s.CurrentWeight = 0
	}
}

// Done -
func (w *WNGINX) Done(node interface{}) {}

// NextWithRequest -
func (w *WNGINX) NextWithRequest(req *Request, move bool) interface{} {
	return w.Next(req.Name, move)
}

// SetConcurrency -
func (w *WNGINX) SetConcurrency(concurrency int) {}

// Start -
func (w *WNGINX) Start() {}

// Stop -
func (w *WNGINX) Stop() {}

// NoLock -
func (w *WNGINX) NoLock() bool {
	return false
}

// WeightLvs weight lv5
type WeightLvs struct {
	Node   interface{}
	Weight int
}

// WLVS w lv5
type WLVS struct {
	nodes []*WeightLvs
	gcd   int // general weight divisor
	maxW  int // maximum weight
	i     int // number of times selected
	cw    int // current weight
}

// Next get next node
func (w *WLVS) Next(_ string, _ bool) interface{} {
	if len(w.nodes) == 0 {
		return nil
	}
	if len(w.nodes) == 1 {
		return w.nodes[0].Node
	}
	for {
		if w.updateCwAnsIsReturn() {
			return nil
		}

		if w.i < len(w.nodes) && w.nodes[w.i].Weight >= w.cw {
			return w.nodes[w.i].Node
		}
	}
}

// updateCwAnsIsReturn update current weight and return the value whether to return
func (w *WLVS) updateCwAnsIsReturn() bool {
	w.i = (w.i + 1) % len(w.nodes)
	if w.i == 0 {
		if w.cw = w.cw - w.gcd; w.cw <= 0 {
			if w.cw = w.maxW; w.cw == 0 {
				return true
			}
		}
	}

	return false
}

// Add add a node
func (w *WLVS) Add(node interface{}, weight int) {
	weighted := &WeightLvs{Node: node, Weight: weight}
	if weight > 0 {
		w.gcd = gcd(w.gcd, weight)
		if w.maxW < weight {
			w.maxW = weight
		}
	}
	w.nodes = append(w.nodes, weighted)
}

// Returns the maximum divisor
func gcd(x, y int) int {
	for {
		if y == 0 {
			return x
		}
		x, y = y, x%y
	}
}

// Remove removes a node
func (w *WLVS) Remove(node interface{}) {
	for i, weighted := range w.nodes {
		if weighted.Node == node {
			w.nodes = append(w.nodes[:i], w.nodes[i+1:]...)
			break
		}
	}
}

// RemoveAll remove all nodes
func (w *WLVS) RemoveAll() {
	w.nodes = w.nodes[:0]
	w.gcd = 0
	w.maxW = 0
	w.i = -1
	w.cw = 0
}

// Reset reset all nodes
func (w *WLVS) Reset() {
	w.i = -1
	w.cw = 0
}

// Done -
func (w *WLVS) Done(node interface{}) {}

// NextWithRequest -
func (w *WLVS) NextWithRequest(req *Request, move bool) interface{} {
	return w.Next(req.Name, move)
}

// SetConcurrency -
func (w *WLVS) SetConcurrency(concurrency int) {}

// Start -
func (w *WLVS) Start() {}

// Stop -
func (w *WLVS) Stop() {}

// NoLock -
func (w *WLVS) NoLock() bool {
	return false
}

// NewRandomRR init RandomRR
func NewRandomRR() *RandomRR {
	return &RandomRR{
		rnd: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// RandomRR is random version of WLVS
// it will shuffle all nodes randomly when reset
type RandomRR struct {
	rnd *rand.Rand
	WLVS
}

// Reset reset RR and shuffle all nodes
func (r *RandomRR) Reset() {
	r.WLVS.Reset()
	r.rnd.Shuffle(len(r.WLVS.nodes), func(i, j int) {
		r.WLVS.nodes[i], r.WLVS.nodes[j] = r.WLVS.nodes[j], r.WLVS.nodes[i]
	})
}
