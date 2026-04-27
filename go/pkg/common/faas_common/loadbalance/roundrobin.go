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

// Previous - returns the previous scheduled node of a function
func (w *WNGINX) Previous(name string, move bool) interface{} {
	return nil
}

// DeleteBalancer -
func (w *WNGINX) DeleteBalancer(name string) {
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
