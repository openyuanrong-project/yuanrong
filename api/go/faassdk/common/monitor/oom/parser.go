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

// Package oom
package oom

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"strconv"

	"yuanrong.org/kernel/runtime/faassdk/utils"
)

var (
	rssPrefix = []byte("rss ")
)

const (
	decimalBase = 10
	bitSize     = 64
)

// NewCGroupMemoryParserWithPath creates parser with cgroupMemoryParserFunc
func NewCGroupMemoryParserWithPath(path string) (*Parser, error) {
	return NewParser(path, cgroupMemoryParserFunc)
}

var cgroupMemoryParserFunc = func(reader *bufio.Reader) (uint64, error) {
	for {
		lineBytes, _, err := reader.ReadLine()
		if err != nil {
			return 0, err
		}

		if bytes.HasPrefix(lineBytes, rssPrefix) {
			lineBytes = bytes.TrimSpace(lineBytes[len(rssPrefix):])
			return strconv.ParseUint(utils.BytesToString(lineBytes), decimalBase, bitSize)
		}
	}
}

// ParserFunc func that parser content of reader to uint64
type ParserFunc func(reader *bufio.Reader) (uint64, error)

// NewParser creates new Parser with file path and ParserFunc
func NewParser(path string, parserFunc ParserFunc) (*Parser, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return &Parser{
		f:      f,
		reader: bufio.NewReader(nil),
		parser: parserFunc,
	}, nil
}

// Parser aims to parse file content that updated frequently (such as cgroup file) with high performance.
// It opens file only once and seek to start every time before read.
// NOTICE: Parser is not thread safe
type Parser struct {
	reader *bufio.Reader
	f      io.ReadSeekCloser
	parser ParserFunc
}

// Close closes file to parse
func (p *Parser) Close() error {
	p.reader.Reset(nil)
	return p.f.Close()
}

// Read resets reader to the start of the file and parses it.
// This method is not thread safe
func (p *Parser) Read() (uint64, error) {
	_, err := p.f.Seek(0, io.SeekStart)
	if err != nil {
		return 0, err
	}
	p.reader.Reset(p.f)
	return p.parser(p.reader)
}
