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

// Package timewheel -
package timewheel

import (
	"sync"
	"testing"
	"time"
)

func TestNewSimpleTimeWheel(t *testing.T) {
	// Test normal creation
	tw := NewSimpleTimeWheel(10*time.Millisecond, 10)
	if tw == nil {
		t.Error("Expected SimpleTimeWheel to be created")
	}

	// Test with pace below minimum
	tw = NewSimpleTimeWheel(1*time.Millisecond, 10)
	// Should use minPace instead
	if tw == nil {
		t.Error("Expected SimpleTimeWheel to be created")
	}
	// Test with slotNum below minimum
	tw = NewSimpleTimeWheel(10*time.Millisecond, 0)
	// Should use minSlotNum instead
	if tw == nil {
		t.Error("Expected SimpleTimeWheel to be created")
	}
}

func TestSimpleTimeWheel_AddTask(t *testing.T) {
	tw := NewSimpleTimeWheel(10*time.Millisecond, 10)
	defer tw.Stop()

	// Test adding a valid task
	err := tw.AddTask("task1", 100*time.Millisecond, 1)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Test adding duplicate task
	err = tw.AddTask("task1", 100*time.Millisecond, 1)
	if err == nil {
		t.Error("Expected error for duplicate task")
	}

	// Test adding task with interval smaller than perimeter
	err = tw.AddTask("task2", 50*time.Millisecond, 1) // 50ms < 100ms (10*10ms)
	if err == nil {
		t.Error("Expected error for invalid interval")
	}
}

func TestSimpleTimeWheel_DelTask(t *testing.T) {
	tw := NewSimpleTimeWheel(10*time.Millisecond, 10)
	defer tw.Stop()

	// Add a task
	tw.AddTask("task1", 100*time.Millisecond, 1)

	// Delete existing task
	err := tw.DelTask("task1")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Delete non-existing task
	err = tw.DelTask("nonexistent")
	if err != nil {
		t.Errorf("Expected no error for non-existent task, got %v", err)
	}
}

func TestSimpleTimeWheel_TaskTrigger(t *testing.T) {
	tw := NewSimpleTimeWheel(10*time.Millisecond, 10)
	defer tw.Stop()

	var wg sync.WaitGroup
	wg.Add(1)

	// Add a task that should trigger once
	err := tw.AddTask("task1", 100*time.Millisecond, 1)
	if err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}

	go func() {
		defer wg.Done()
		tasks, err := tw.Wait()
		if err != nil {
			t.Errorf("Error waiting for tasks: %v", err)
			return
		}
		if len(tasks) != 1 || tasks[0] != "task1" {
			t.Errorf("Expected task1, got %v", tasks)
		}
	}()

	// Wait for task to be triggered
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Success
	case <-time.After(200 * time.Millisecond):
		t.Error("Task was not triggered within expected time")
	}
}

func TestSimpleTimeWheel_RepeatingTask(t *testing.T) {
	tw := NewSimpleTimeWheel(10*time.Millisecond, 10)
	defer tw.Stop()

	// Add a task that should trigger 3 times
	err := tw.AddTask("task1", 100*time.Millisecond, 3)
	if err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}

	// Wait for all triggers
	triggerCount := 0
	timeout := time.After(500 * time.Millisecond)
	for triggerCount < 3 {
		select {
		case tasks := <-tw.(*SimpleTimeWheel).readyCh:
			if len(*tasks) == 1 && (*tasks)[0] == "task1" {
				triggerCount++
			}
		case <-timeout:
			t.Errorf("Expected 3 triggers, got %d within timeout", triggerCount)
			return
		}
	}

	if triggerCount != 3 {
		t.Errorf("Expected 3 triggers, got %d", triggerCount)
	}
}

func TestSimpleTimeWheel_EndlessTask(t *testing.T) {
	tw := NewSimpleTimeWheel(10*time.Millisecond, 10)
	defer tw.Stop()

	// Add a task that runs endlessly
	err := tw.AddTask("task1", 100*time.Millisecond, -1)
	if err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}

	// Wait for at least 2 triggers
	triggerCount := 0
	timeout := time.After(300 * time.Millisecond)
	for triggerCount < 2 {
		select {
		case tasks := <-tw.(*SimpleTimeWheel).readyCh:
			if len(*tasks) == 1 && (*tasks)[0] == "task1" {
				triggerCount++
			}
		case <-timeout:
			if triggerCount < 2 {
				t.Errorf("Expected at least 2 triggers, got %d within timeout", triggerCount)
			}
			return
		}
	}
}

func TestSimpleTimeWheel_UpdateTask(t *testing.T) {
	tw := NewSimpleTimeWheel(10*time.Millisecond, 10)
	defer tw.Stop()

	// Add a task
	err := tw.AddTask("task1", 100*time.Millisecond, 1)
	if err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}

	// Update the task
	err = tw.UpdateTask("task1", 200*time.Millisecond, 2)
	if err != nil {
		t.Errorf("Failed to update task: %v", err)
	}

	// Update non-existing task (should add it)
	err = tw.UpdateTask("task2", 100*time.Millisecond, 1)
	if err != nil {
		t.Errorf("Failed to update/add non-existing task: %v", err)
	}
}

func TestSimpleTimeWheel_Stop(t *testing.T) {
	tw := NewSimpleTimeWheel(10*time.Millisecond, 10)

	// Stop the time wheel
	tw.Stop()

	// Try to add a task after stopping
	err := tw.AddTask("task1", 100*time.Millisecond, 1)
	// This should not panic, but the behavior might vary based on implementation
	_ = err
}

func TestSimpleTimeWheel_WaitAfterStop(t *testing.T) {
	tw := NewSimpleTimeWheel(10*time.Millisecond, 10)
	tw.Stop()

	// Wait should return nil, nil after stop
	tasks, err := tw.Wait()
	if tasks != nil || err != nil {
		t.Errorf("Expected nil, nil after stop, got %v, %v", tasks, err)
	}
}

func TestSimpleTimeWheel_InvalidTaskInterval(t *testing.T) {
	tw := NewSimpleTimeWheel(10*time.Millisecond, 10)
	defer tw.Stop()

	// Try to add a task with invalid interval
	err := tw.AddTask("task1", 50*time.Millisecond, 1) // 50ms < 100ms perimeter
	if err == nil {
		t.Error("Expected error for invalid task interval")
	}
}
