// Copyright 2022, Chef.  All rights reserved.
// https://github.com/q191201771/lal
//
// Use of this source code is governed by a MIT-style license
// that can be found in the License file.
//
// Author: Chef (191201771@qq.com)

package logic

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/q191201771/lal/pkg/httpflv"
)

// startRecordFlvIfNeeded 必要时开启flv录制
//
func (group *Group) startRecordFlvIfNeeded(nowUnix string) {
	if !group.config.RecordConfig.EnableFlv {
		return
	}
	streamPath := filepath.Join(group.config.RecordConfig.FlvOutPath, group.streamName)
	group.ensureDir(streamPath)
	// 构造文件名
	filename := fmt.Sprintf("%s.flv", nowUnix)
	filenameWithPath := filepath.Join(streamPath, filename)

	// 初始化录制
	group.recordFlv = &httpflv.FlvFileWriter{}
	if err := group.recordFlv.Open(filenameWithPath); err != nil {
		Log.Errorf("[%s] record flv open file failed. filename=%s, err=%+v",
			group.UniqueKey, filenameWithPath, err)
		group.recordFlv = nil
	}
	Log.Infof("[%s] record flv open file succeeded. filename=%s", group.UniqueKey, filenameWithPath)
	if err := group.recordFlv.WriteFlvHeader(); err != nil {
		Log.Errorf("[%s] record flv write flv header failed. filename=%s, err=%+v",
			group.UniqueKey, filenameWithPath, err)
		group.recordFlv = nil
	}
}

func (group *Group) stopRecordFlvIfNeeded() {
	if !group.config.RecordConfig.EnableFlv {
		return
	}

	if group.recordFlv != nil {
		_ = group.recordFlv.Dispose()
		group.recordFlv = nil
	}
}

func (group *Group) ensureDir(dir string) {
	// 注意，如果路径已经存在，则啥也不干
	err := os.MkdirAll(dir, 0777)
	Log.Assert(nil, err)
}
