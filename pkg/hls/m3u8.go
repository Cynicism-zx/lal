// Copyright 2020, Chef.  All rights reserved.
// https://github.com/q191201771/lal
//
// Use of this source code is governed by a MIT-style license
// that can be found in the License file.
//
// Author: Chef (191201771@qq.com)

package hls

import (
	"bytes"
	"fmt"
	"github.com/q191201771/lal/pkg/base"
	"github.com/q191201771/naza/pkg/nazaerrors"
	"strconv"
)

// writeM3u8File
//
// @param content     需写入文件的内容
// @param filename    m3u8文件名
// @param filenameBak m3u8临时文件名
//
func writeM3u8File(content []byte, filename string, filenameBak string) error {
	if err := fslCtx.WriteFile(filenameBak, content, 0666); err != nil {
		return err
	}

	return fslCtx.Rename(filenameBak, filename)
}

// updateTargetDurationInM3u8 如果当前duration比原m3u8文件的`EXT-X-TARGETDURATION`大，则更新`EXT-X-TARGETDURATION`的值
//
// @param content      原m3u8文件的内容
// @param currDuration 当前duration
//
// @return 处理后的m3u8文件内容
//
func updateTargetDurationInM3u8(content []byte, currDuration int) ([]byte, error) {
	l := bytes.Index(content, []byte("#EXT-X-TARGETDURATION:"))
	if l == -1 {
		return content, nazaerrors.Wrap(base.ErrHls)
	}
	r := bytes.Index(content[l:], []byte{'\n'})
	if r == -1 {
		return content, nazaerrors.Wrap(base.ErrHls)
	}
	oldDurationStr := bytes.TrimPrefix(content[l:l+r], []byte("#EXT-X-TARGETDURATION:"))
	oldDuration, err := strconv.Atoi(string(oldDurationStr))
	if err != nil {
		return content, err
	}
	if err != nil {
		return content, err
	}
	if currDuration > oldDuration {
		tmpContent := make([]byte, l)
		copy(tmpContent, content[:l])
		tmpContent = append(tmpContent, []byte(fmt.Sprintf("#EXT-X-TARGETDURATION:%d", currDuration))...)
		tmpContent = append(tmpContent, content[l+r:]...)
		content = tmpContent
	}
	return content, nil
}

// deleteTsInM3u8 删除record.m3u8文件中过期的ts文件
// @param content      原m3u8文件的内容
// @param ts           过期ts文件
//
// @return 处理后的m3u8文件内容
//
func delTsInM3u8(content []byte, frag *fragmentInfo) ([]byte, error) {
	// 更新SEQUENCE
	s := bytes.Index(content, []byte("#EXT-X-MEDIA-SEQUENCE:"))
	rr := bytes.Index(content[s:], []byte{'\n'})
	tmp := make([]byte, s)
	copy(tmp, content[:s])
	tmp = append(tmp, []byte(fmt.Sprintf("#EXT-X-MEDIA-SEQUENCE:%d", frag.id+1))...)
	tmp = append(tmp, content[s+rr:]...)
	content = tmp
	// 删除ts文件记录
	delLines := fmt.Sprintf("#EXTINF:%.3f,\n%s\n", frag.duration, frag.filename)
	l := bytes.Index(content, []byte(delLines))
	if l == -1 {
		return nil, nazaerrors.Wrap(base.ErrHls)
	}
	delLines = fmt.Sprintf("%s\n", frag.filename)
	ll := bytes.Index(content, []byte(delLines))
	r := bytes.Index(content[ll:], []byte{'\n'})
	if r == -1 {
		return nil, nazaerrors.Wrap(base.ErrHls)
	}
	tmpContent := make([]byte, l)
	copy(tmpContent, content[:l])
	tmpContent = append(tmpContent, content[ll+r+1:]...)
	content = tmpContent
	return content, nil
}

// CalcM3u8Duration
//
// @param content 传入m3u8文件内容
//
// @return durationSec m3u8中所有ts的时间总和。注意，使用的是m3u8文件中描述的ts时间，而不是读取ts文件中实际音视频数据的时间。
//
func CalcM3u8Duration(content []byte) (durationSec float64, err error) {
	lines := bytes.Split(content, []byte{'\n'})
	for _, line := range lines {
		if bytes.HasPrefix(line, []byte("#EXTINF:")) {
			line = bytes.TrimSpace(line)
			v := bytes.TrimSuffix(bytes.TrimPrefix(line, []byte("#EXTINF:")), []byte{','})
			v = bytes.TrimSpace(v)
			vv, err := strconv.ParseFloat(string(v), 64)
			if err != nil {
				return durationSec, err
			}
			durationSec += vv
		}
	}
	return
}
