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

package http

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/fsnotify/fsnotify"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"yuanrong.org/kernel/runtime/faassdk/common/functionlog"
	"yuanrong.org/kernel/runtime/faassdk/config"
)

func TestMain(m *testing.M) {
	m.Run()
}

func TestCollect(t *testing.T) {
	p := gomonkey.NewPatches()
	p.ApplyFunc(functionlog.GetFunctionLogger, func(cfg *config.Configuration) (*functionlog.FunctionLogger, error) {
		return functionlog.NewFunctionLogger(nil, "", ""), nil
	})
	logger := &RuntimeContainerLogger{}
	p.ApplyFunc(NewRuntimeStdLogger, func(logFilePath string, processCB RuntimeContainerLogProcessCB) (*RuntimeContainerLogger, error) {
		logger.processCB = processCB
		return logger, nil
	})
	defer p.Reset()

	l := collectRuntimeContainerLog(&config.Configuration{})
	require.NotNil(t, l)
	l.processCB(RuntimeContainerLog{log: "abc", level: "info", t: time.Now()})
}

func TestRuntimeContainerLogger(t *testing.T) {
	p := gomonkey.NewPatches()
	events := make(chan fsnotify.Event)
	errors := make(chan error)
	p.ApplyFunc(fsnotify.NewWatcher, func() (*fsnotify.Watcher, error) {
		return &fsnotify.Watcher{
			Events: events,
			Errors: errors,
		}, nil
	})

	var watcher *fsnotify.Watcher
	p.ApplyMethod(reflect.TypeOf(watcher), "Add", func(_ *fsnotify.Watcher, name string) error { return nil })
	p.ApplyMethod(reflect.TypeOf(watcher), "Remove", func(_ *fsnotify.Watcher, name string) error { return nil })
	p.ApplyMethod(reflect.TypeOf(watcher), "Close", func(_ *fsnotify.Watcher) error { return nil })

	dir, err := ioutil.TempDir(os.TempDir(), "test-runtime-container-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	fileName := filepath.Join(dir, "log")

	processCB := func(msg RuntimeContainerLog) {
		fmt.Println(msg)
	}
	logger, err := NewRuntimeStdLogger(fileName, processCB)
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		if err := logger.Run(); err != nil {
			fmt.Println(err)
		}
	}()

	before := time.Now()
	content, err := makeDockerLog("abc", "stdin", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	ioutil.WriteFile(fileName, content, 0755)
	events <- fsnotify.Event{Name: fileName, Op: fsnotify.Write}
	logger.syncTo(before)

	os.Remove(fileName)
	events <- fsnotify.Event{Name: fileName, Op: fsnotify.Remove}

	pp := gomonkey.NewPatches()
	pp.ApplyFunc(os.Stat, func(name string) (os.FileInfo, error) { return nil, os.ErrNotExist })
	pp.ApplyFunc(os.Open, func(name string) (*os.File, error) { return nil, os.ErrNotExist })

	content, err = makeDockerLog("abcd", "stdin", time.Now())
	if err != nil {
		pp.Reset()
		t.Fatal(err)
	}

	ioutil.WriteFile(fileName, content, 0755)

	events <- fsnotify.Event{Name: fileName, Op: fsnotify.Create}

	pp.Reset()

	events <- fsnotify.Event{Name: fileName, Op: fsnotify.Write}

	content, err = makeDockerLog("a", "stdin", time.Now())
	if err != nil {
		pp.Reset()
		t.Fatal(err)
	}
	ioutil.WriteFile(fileName, content, 0755)
	events <- fsnotify.Event{Name: fileName, Op: fsnotify.Write}
}

func TestRuntimeContainerLogger2(t *testing.T) {
	p := gomonkey.NewPatches()
	events := make(chan fsnotify.Event)
	errorCh := make(chan error)
	p.ApplyFunc(fsnotify.NewWatcher, func() (*fsnotify.Watcher, error) {
		return &fsnotify.Watcher{
			Events: events,
			Errors: errorCh,
		}, nil
	})

	var watcher *fsnotify.Watcher
	p.ApplyMethod(reflect.TypeOf(watcher), "Add", func(_ *fsnotify.Watcher, name string) error { return nil })
	p.ApplyMethod(reflect.TypeOf(watcher), "Remove", func(_ *fsnotify.Watcher, name string) error { return nil })
	p.ApplyMethod(reflect.TypeOf(watcher), "Close", func(_ *fsnotify.Watcher) error { return nil })

	dir, err := ioutil.TempDir(os.TempDir(), "test-runtime-container-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	fileName := filepath.Join(dir, "log")

	processCB := func(msg RuntimeContainerLog) {
		fmt.Println(msg)
	}
	logger, err := NewRuntimeStdLogger(fileName, processCB)
	if err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := logger.Run(); err != nil {
			fmt.Println(err)
		}
	}()

	before := time.Now()
	content, err := makeDockerLog("abc", "stdin", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	ioutil.WriteFile(fileName, content, 0755)
	events <- fsnotify.Event{Name: fileName, Op: fsnotify.Write}
	logger.syncTo(before)

	os.Remove(fileName)
	errorCh <- errors.New("my error")

	wg.Wait()
}

func TestRuntimeContainerLogger3(t *testing.T) {
	p := gomonkey.NewPatches()
	events := make(chan fsnotify.Event)
	errorCh := make(chan error)
	p.ApplyFunc(fsnotify.NewWatcher, func() (*fsnotify.Watcher, error) {
		return &fsnotify.Watcher{
			Events: events,
			Errors: errorCh,
		}, nil
	})

	var watcher *fsnotify.Watcher
	p.ApplyMethod(reflect.TypeOf(watcher), "Add", func(_ *fsnotify.Watcher, name string) error { return nil })
	p.ApplyMethod(reflect.TypeOf(watcher), "Remove", func(_ *fsnotify.Watcher, name string) error { return nil })
	p.ApplyMethod(reflect.TypeOf(watcher), "Close", func(_ *fsnotify.Watcher) error { return nil })

	dir, err := ioutil.TempDir(os.TempDir(), "test-runtime-container-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	fileName := filepath.Join(dir, "log")

	processCB := func(msg RuntimeContainerLog) {
		fmt.Println(msg)
	}
	logger, err := NewRuntimeStdLogger(fileName, processCB)
	if err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := logger.Run(); err != nil {
			fmt.Println(err)
		}
	}()

	before := time.Now()
	content, err := makeDockerLog("abc", "stdin", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	ioutil.WriteFile(fileName, content, 0755)
	events <- fsnotify.Event{Name: fileName, Op: fsnotify.Write}
	logger.syncTo(before)

	os.Remove(fileName)
	events <- fsnotify.Event{Name: fileName, Op: fsnotify.Remove}

	pp := gomonkey.NewPatches()
	pp.ApplyFunc(os.Stat, func(name string) (os.FileInfo, error) { return nil, os.ErrNotExist })
	pp.ApplyFunc(os.Open, func(name string) (*os.File, error) { return nil, os.ErrNotExist })

	content, err = makeDockerLog("abcd", "stdin", time.Now())
	if err != nil {
		pp.Reset()
		t.Fatal(err)
	}

	ioutil.WriteFile(fileName, content, 0755)

	events <- fsnotify.Event{Name: fileName, Op: fsnotify.Create}

	pp.Reset()

	errorCh <- errors.New("my error")

	content, err = makeDockerLog("a", "stdin", time.Now())
	if err != nil {
		pp.Reset()
		t.Fatal(err)
	}
	ioutil.WriteFile(fileName, content, 0755)

	wg.Wait()
}

func TestLoggerRun(t *testing.T) {
	{
		l := &RuntimeContainerLogger{
			syncPointCh: make(chan *syncPoint),
			t: &tail{
				Lines:  make(chan string),
				Errors: make(chan error),
			},
		}
		close(l.t.Lines)
		err := l.Run()
		assert.NotNil(t, err)
	}

	{
		l := &RuntimeContainerLogger{
			syncPointCh: make(chan *syncPoint),
			t: &tail{
				Lines:  make(chan string),
				Errors: make(chan error),
			},
		}
		close(l.t.Errors)
		err := l.Run()
		assert.NotNil(t, err)
	}

	{
		l := &RuntimeContainerLogger{
			syncPointCh: make(chan *syncPoint),
			t: &tail{
				Lines:  make(chan string),
				Errors: make(chan error),
			},
		}
		close(l.syncPointCh)
		err := l.Run()
		assert.NotNil(t, err)
	}

	{
		finish := make(chan struct{})
		l := &RuntimeContainerLogger{
			syncPointCh: make(chan *syncPoint),
			t: &tail{
				Lines:  make(chan string),
				Errors: make(chan error),
			},
		}
		// close of runtime.ProcessExitSignal should not affect logger
		close(finish)
		go func() { l.Run() }()
		l.syncTo(time.Now())
	}
}

func TestSyncPoint(t *testing.T) {
	before := time.Now()
	time.Sleep(10 * time.Millisecond)
	after := time.Now()
	l := &RuntimeContainerLogger{}
	l.addSyncPoint(&syncPoint{t: before, done: make(chan struct{})})
	assert.Equal(t, 1, len(l.syncPoints))

	l.tick = after
	l.notifySyncPoints()
}

func Test_processLine(t *testing.T) {
	tt := time.Now()
	l := &RuntimeContainerLogger{
		processCB: func(RuntimeContainerLog) {},
	}
	l.processLine("abc")

	content, err := makeDockerLog("abc", "stdout", tt)
	if err != nil {
		t.Fatal(err)
	}
	l.processLine(string(content))

	content, err = makeDockerLog("abc", "stderr", tt)
	if err != nil {
		t.Fatal(err)
	}
	l.processLine(string(content))

	content, err = makeDockerLog("abc", "sync", tt)
	if err != nil {
		t.Fatal(err)
	}
	l.processLine(string(content))

	s := `{"log":"abc","stream":"sync","time":"wrong time"}`
	l.processLine(s)

	assert.True(t, tt.Equal(l.tick))
}

func Test_sync(t *testing.T) {
	p := gomonkey.NewPatches()
	var f *os.File
	p.ApplyMethod(reflect.TypeOf(f), "Close", func(_ *os.File) error { return errors.New("my error") })
	p.ApplyFunc(os.Open, func(name string) (*os.File, error) { return nil, errors.New("my error") })
	defer p.Reset()

	tt := &tail{
		file:   &os.File{},
		Errors: make(chan error, 1),
	}

	logger := &RuntimeContainerLogger{
		syncPointCh: make(chan *syncPoint),
		processCB:   func(RuntimeContainerLog) {},
		logFilePath: "",
	}

	tt.sync(logger)

	e := <-tt.Errors
	assert.NotNil(t, e)
}

func Test_sync2(t *testing.T) {
	p := gomonkey.NewPatches()
	var f *os.File
	p.ApplyMethod(reflect.TypeOf(f), "Close", func(_ *os.File) error { return errors.New("my error") })
	p.ApplyMethod(reflect.TypeOf(f), "Seek", func(_ *os.File, offset int64, whence int) (ret int64, err error) {
		return 0, errors.New("my error")
	})
	p.ApplyFunc(os.Open, func(name string) (*os.File, error) { return &os.File{}, nil })
	defer p.Reset()

	tt := &tail{
		file:   &os.File{},
		Errors: make(chan error, 1),
	}
	logger := &RuntimeContainerLogger{
		syncPointCh: make(chan *syncPoint),
		processCB:   func(RuntimeContainerLog) {},
		logFilePath: "",
	}
	tt.sync(logger)

	e := <-tt.Errors
	assert.NotNil(t, e)
}

func Test_sync3(t *testing.T) {
	p := gomonkey.NewPatches()
	var f *os.File
	p.ApplyMethod(reflect.TypeOf(f), "Close", func(_ *os.File) error { return errors.New("my error") })
	p.ApplyMethod(reflect.TypeOf(f), "Seek", func(_ *os.File, offset int64, whence int) (ret int64, err error) {
		return 0, nil
	})
	p.ApplyFunc(os.Open, func(name string) (*os.File, error) { return &os.File{}, nil })
	var r *bufio.Reader
	p.ApplyMethod(reflect.TypeOf(r), "ReadString", func(_ *bufio.Reader, delim byte) (string, error) {
		return "", errors.New("my error")
	})
	defer p.Reset()

	tt := &tail{
		file:   &os.File{},
		Errors: make(chan error, 1),
	}
	logger := &RuntimeContainerLogger{
		syncPointCh: make(chan *syncPoint),
		processCB:   func(RuntimeContainerLog) {},
		logFilePath: "",
	}

	tt.sync(logger)

	e := <-tt.Errors
	assert.NotNil(t, e)
}

func Test_handleReadLineEOF(t *testing.T) {
	p := gomonkey.NewPatches()
	var f *os.File
	p.ApplyMethod(reflect.TypeOf(f), "Seek", func(_ *os.File, offset int64, whence int) (ret int64, err error) {
		return 0, nil
	})
	var r *bufio.Reader
	p.ApplyMethod(reflect.TypeOf(r), "Reset", func(_ *bufio.Reader, r io.Reader) {})
	defer p.Reset()

	changeNotifier := newChangeNotifier()
	close(changeNotifier.errCh)
	tt := &tail{
		reader:         &bufio.Reader{},
		changeNotifier: changeNotifier,
	}
	logger := &RuntimeContainerLogger{
		syncPointCh: make(chan *syncPoint),
		processCB:   func(RuntimeContainerLog) {},
		logFilePath: "",
	}

	err := tt.handleReadLineEOF(logger, 0, "abc")
	assert.NotNil(t, err)
}

func Test_handleReadLineEOF2(t *testing.T) {
	p := gomonkey.NewPatches()
	var f *os.File
	p.ApplyMethod(reflect.TypeOf(f), "Seek", func(_ *os.File, offset int64, whence int) (ret int64, err error) {
		return 0, nil
	})
	p.ApplyFunc(os.Open, func(name string) (*os.File, error) { return &os.File{}, nil })
	var r *bufio.Reader
	p.ApplyMethod(reflect.TypeOf(r), "Reset", func(_ *bufio.Reader, r io.Reader) {})
	defer p.Reset()

	changeNotifier := newChangeNotifier()
	close(changeNotifier.truncateCh)
	tt := &tail{
		reader:         &bufio.Reader{},
		changeNotifier: changeNotifier,
	}
	logger := &RuntimeContainerLogger{
		syncPointCh: make(chan *syncPoint),
		processCB:   func(RuntimeContainerLog) {},
		logFilePath: "",
	}
	go func(tt *tail) {
		<-tt.changeNotifier.closeCh
		close(tt.changeNotifier.closeDone)
	}(tt)
	err := tt.handleReadLineEOF(logger, 0, "abc")
	assert.Nil(t, err)
}

func Test_handleReadLineEOF3(t *testing.T) {
	p := gomonkey.NewPatches()
	var f *os.File
	p.ApplyMethod(reflect.TypeOf(f), "Seek", func(_ *os.File, offset int64, whence int) (ret int64, err error) {
		return 0, errors.New("my error")
	})
	defer p.Reset()

	tt := &tail{
		reader: &bufio.Reader{},
	}
	logger := &RuntimeContainerLogger{
		syncPointCh: make(chan *syncPoint),
		processCB:   func(RuntimeContainerLog) {},
		logFilePath: "",
	}
	err := tt.handleReadLineEOF(logger, 0, "abc")
	assert.NotNil(t, err)
}

func Test_waitUntilCreate(t *testing.T) {
	p := gomonkey.NewPatches()
	wat := &fsnotify.Watcher{
		Events: make(chan fsnotify.Event),
		Errors: make(chan error),
	}
	p.ApplyMethod(reflect.TypeOf(wat), "Add", func(_ *fsnotify.Watcher, name string) error { return errors.New("my error") })
	defer p.Reset()

	w := &watcher{
		watcher: wat,
	}
	err := w.waitUntilCreate("file")
	assert.NotNil(t, err)
}

func Test_waitUntilCreate2(t *testing.T) {
	p := gomonkey.NewPatches()
	wat := &fsnotify.Watcher{
		Events: make(chan fsnotify.Event),
		Errors: make(chan error),
	}
	p.ApplyMethod(reflect.TypeOf(wat), "Add", func(_ *fsnotify.Watcher, name string) error { return nil })
	p.ApplyMethod(reflect.TypeOf(wat), "Remove", func(_ *fsnotify.Watcher, name string) error { return errors.New("my error") })
	p.ApplyFunc(os.Stat, func(name string) (os.FileInfo, error) { return nil, errors.New("my error") })
	defer p.Reset()

	w := &watcher{
		watcher: wat,
	}
	err := w.waitUntilCreate("file")
	assert.NotNil(t, err)
}

func Test_waitUntilCreate3(t *testing.T) {
	p := gomonkey.NewPatches()
	wat := &fsnotify.Watcher{
		Events: make(chan fsnotify.Event),
		Errors: make(chan error),
	}
	close(wat.Events)
	p.ApplyMethod(reflect.TypeOf(wat), "Add", func(_ *fsnotify.Watcher, name string) error { return nil })
	p.ApplyMethod(reflect.TypeOf(wat), "Remove", func(_ *fsnotify.Watcher, name string) error { return nil })
	p.ApplyFunc(os.Stat, func(name string) (os.FileInfo, error) { return nil, os.ErrNotExist })
	defer p.Reset()

	w := &watcher{
		watcher: wat,
	}
	err := w.waitUntilCreate("file")
	assert.NotNil(t, err)
}

func Test_waitUntilCreate4(t *testing.T) {
	p := gomonkey.NewPatches()
	wat := &fsnotify.Watcher{
		Events: make(chan fsnotify.Event),
		Errors: make(chan error),
	}
	close(wat.Errors)
	p.ApplyMethod(reflect.TypeOf(wat), "Add", func(_ *fsnotify.Watcher, name string) error { return nil })
	p.ApplyMethod(reflect.TypeOf(wat), "Remove", func(_ *fsnotify.Watcher, name string) error { return nil })
	p.ApplyFunc(os.Stat, func(name string) (os.FileInfo, error) { return nil, os.ErrNotExist })
	defer p.Reset()

	w := &watcher{
		watcher: wat,
	}
	err := w.waitUntilCreate("file")
	assert.NotNil(t, err)
}

func Test_waitUntilCreate5(t *testing.T) {
	p := gomonkey.NewPatches()
	wat := &fsnotify.Watcher{
		Events: make(chan fsnotify.Event, 1),
		Errors: make(chan error),
	}
	wat.Events <- fsnotify.Event{Name: "abc"}
	p.ApplyMethod(reflect.TypeOf(wat), "Add", func(_ *fsnotify.Watcher, name string) error { return nil })
	p.ApplyMethod(reflect.TypeOf(wat), "Remove", func(_ *fsnotify.Watcher, name string) error { return nil })
	p.ApplyFunc(os.Stat, func(name string) (os.FileInfo, error) { return nil, os.ErrNotExist })
	p.ApplyFunc(filepath.Abs, func(string) (string, error) { return "", errors.New("my error") })
	defer p.Reset()

	w := &watcher{
		watcher: wat,
	}
	err := w.waitUntilCreate("file")
	assert.NotNil(t, err)
}

func Test_handle(t *testing.T) {
	wat := &fsnotify.Watcher{
		Events: make(chan fsnotify.Event, 1),
		Errors: make(chan error),
	}
	close(wat.Events)

	w := &watcher{
		watcher: wat,
	}
	c := newChangeNotifier()
	size := int64(1)
	go w.handle(c, "abc", &size)
	err := c.errCh
	assert.NotNil(t, err)
}

func Test_handle2(t *testing.T) {
	wat := &fsnotify.Watcher{
		Events: make(chan fsnotify.Event, 1),
		Errors: make(chan error),
	}
	close(wat.Errors)

	w := &watcher{
		watcher: wat,
	}
	c := newChangeNotifier()
	size := int64(1)
	go w.handle(c, "abc", &size)
	err := c.errCh
	assert.NotNil(t, err)
}

func Test_makeDockerLog(t *testing.T) {
	p := gomonkey.NewPatches()
	p.ApplyFunc(json.Marshal, func(v interface{}) ([]byte, error) { return nil, errors.New("my error") })
	defer p.Reset()

	_, err := makeDockerLog("abc", "sync", time.Now())
	assert.NotNil(t, err)
}
