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
	fp    *os.File
	CutAt uint32 // flv切割时间点
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
	defer func() {
		if e := recover(); e != nil {
			Log.Error(e)
		}
	}()
	ffw.fp, err = os.Create(filename)
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
