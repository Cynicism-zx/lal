// Copyright 2019, Chef.  All rights reserved.
// https://github.com/q191201771/lal
//
// Use of this source code is governed by a MIT-style license
// that can be found in the License file.
//
// Author: Chef (191201771@qq.com)

package httpflv

import (
	"os"

	"github.com/q191201771/lal/pkg/base"
)

// TODO chef: 结构体重命名为FileWriter，文件名重命名为file_writer.go。所有写流文件的（flv,hls,ts）统一重构

type FlvFileWriter struct {
	fp  *os.File
	Now int
}

// Capability 获取已写入文件的大小
func (ffw *FlvFileWriter) Capability() int64 {
	if ffw.fp != nil {
		fi, _ := ffw.fp.Stat()
		return fi.Size()
	}
	return 0
}

func (ffw *FlvFileWriter) Open(filename string) (err error) {
	ffw.fp, err = os.Create(filename)
	// 一次默认只允许写入单个flv文件50G,写满了再重新创建新的flv文件
	_ = ffw.fp.Truncate(1024 * 1024 * 20)
	return
}

func (ffw *FlvFileWriter) WriteRaw(b []byte) (err error) {
	if ffw.fp == nil {
		return base.ErrFileNotExist
	}
	_, err = ffw.fp.Write(b)
	return
}

func (ffw *FlvFileWriter) WriteFlvHeader() (err error) {
	if ffw.fp == nil {
		return base.ErrFileNotExist
	}
	_, err = ffw.fp.Write(FlvHeader)
	return
}

func (ffw *FlvFileWriter) WriteTag(tag Tag) (err error) {
	if ffw.fp == nil {
		return base.ErrFileNotExist
	}
	_, err = ffw.fp.Write(tag.Raw)
	return
}

func (ffw *FlvFileWriter) Dispose() error {
	if ffw.fp == nil {
		return base.ErrFileNotExist
	}
	return ffw.fp.Close()
}

func (ffw *FlvFileWriter) Name() string {
	if ffw.fp == nil {
		return ""
	}
	return ffw.fp.Name()
}
