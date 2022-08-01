// Copyright 2020, Chef.  All rights reserved.
// https://github.com/q191201771/lal
//
// Use of this source code is governed by a MIT-style license
// that can be found in the License file.
//
// Author: Chef (191201771@qq.com)

package logic

import (
	"encoding/json"
	"github.com/q191201771/naza/pkg/nazajson"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"

	"github.com/q191201771/naza/pkg/nazahttp"

	"github.com/q191201771/lal/pkg/base"
)

type HttpApiServer struct {
	addr string
	sm   *ServerManager

	ln net.Listener
}

func NewHttpApiServer(addr string, sm *ServerManager) *HttpApiServer {
	return &HttpApiServer{
		addr: addr,
		sm:   sm,
	}
}

func (h *HttpApiServer) Listen() (err error) {
	if h.ln, err = net.Listen("tcp", h.addr); err != nil {
		return
	}
	Log.Infof("start http-api server listen. addr=%s", h.addr)
	return
}

func (h *HttpApiServer) RunLoop() error {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/stat/group", h.statGroupHandler)
	mux.HandleFunc("/api/stat/all_group", h.statAllGroupHandler)
	mux.HandleFunc("/api/stat/lal_info", h.statLalInfoHandler)

	mux.HandleFunc("/api/ctrl/start_relay_pull", h.ctrlStartRelayPullHandler)
	mux.HandleFunc("/api/ctrl/stop_relay_pull", h.ctrlStopRelayPullHandler)
	mux.HandleFunc("/api/ctrl/kick_session", h.ctrlKickSessionHandler)
	mux.HandleFunc("/api/video/get", h.getVideo)
	mux.HandleFunc("/api/video/del", h.delVideo)
	mux.HandleFunc("/api/ctrl/start_rtp_pub", h.ctrlStartRtpPubHandler)
	mux.HandleFunc("/", h.notFoundHandler)

	var srv http.Server
	srv.Handler = mux
	return srv.Serve(h.ln)
}

// TODO chef: dispose

// ---------------------------------------------------------------------------------------------------------------------

func (h *HttpApiServer) statLalInfoHandler(w http.ResponseWriter, req *http.Request) {
	var v base.ApiStatLalInfo
	v.ErrorCode = base.ErrorCodeSucc
	v.Desp = base.DespSucc
	v.Data = h.sm.StatLalInfo()
	feedback(v, w)
}

func (h *HttpApiServer) statAllGroupHandler(w http.ResponseWriter, req *http.Request) {
	var v base.ApiStatAllGroup
	v.ErrorCode = base.ErrorCodeSucc
	v.Desp = base.DespSucc
	v.Data.Groups = h.sm.StatAllGroup()
	feedback(v, w)
}

func (h *HttpApiServer) statGroupHandler(w http.ResponseWriter, req *http.Request) {
	var v base.ApiStatGroup

	q := req.URL.Query()
	streamName := q.Get("stream_name")
	if streamName == "" {
		v.ErrorCode = base.ErrorCodeParamMissing
		v.Desp = base.DespParamMissing
		feedback(v, w)
		return
	}

	v.Data = h.sm.StatGroup(streamName)
	if v.Data == nil {
		v.ErrorCode = base.ErrorCodeGroupNotFound
		v.Desp = base.DespGroupNotFound
		feedback(v, w)
		return
	}

	v.ErrorCode = base.ErrorCodeSucc
	v.Desp = base.DespSucc
	feedback(v, w)
	return
}

// ---------------------------------------------------------------------------------------------------------------------

func (h *HttpApiServer) ctrlStartRelayPullHandler(w http.ResponseWriter, req *http.Request) {
	var v base.ApiCtrlStartRelayPull
	var info base.ApiCtrlStartRelayPullReq

	j, err := unmarshalRequestJsonBody(req, &info, "url")
	if err != nil {
		Log.Warnf("http api start pull error. err=%+v", err)
		v.ErrorCode = base.ErrorCodeParamMissing
		v.Desp = base.DespParamMissing
		feedback(v, w)
		return
	}

	if !j.Exist("pull_timeout_ms") {
		info.PullTimeoutMs = 5000
	}
	if !j.Exist("pull_retry_num") {
		info.PullRetryNum = base.PullRetryNumNever
	}
	if !j.Exist("auto_stop_pull_after_no_out_ms") {
		info.AutoStopPullAfterNoOutMs = base.AutoStopPullAfterNoOutMsNever
	}
	if !j.Exist("rtsp_mode") {
		info.RtspMode = base.RtspModeTcp
	}

	Log.Infof("http api start pull. req info=%+v", info)

	resp := h.sm.CtrlStartRelayPull(info)
	feedback(resp, w)
	return
}

func (h *HttpApiServer) ctrlStopRelayPullHandler(w http.ResponseWriter, req *http.Request) {
	var v base.ApiCtrlStopRelayPull

	q := req.URL.Query()
	streamName := q.Get("stream_name")
	if streamName == "" {
		v.ErrorCode = base.ErrorCodeParamMissing
		v.Desp = base.DespParamMissing
		feedback(v, w)
		return
	}

	Log.Infof("http api stop pull. stream_name=%s", streamName)

	resp := h.sm.CtrlStopRelayPull(streamName)
	feedback(resp, w)
	return
}

func (h *HttpApiServer) ctrlKickSessionHandler(w http.ResponseWriter, req *http.Request) {
	var v base.HttpResponseBasic
	var info base.ApiCtrlKickSessionReq

	_, err := unmarshalRequestJsonBody(req, &info, "stream_name", "session_id")
	if err != nil {
		Log.Warnf("http api kick session error. err=%+v", err)
		v.ErrorCode = base.ErrorCodeParamMissing
		v.Desp = base.DespParamMissing
		feedback(v, w)
		return
	}

	Log.Infof("http api kick out session. req info=%+v", info)

	resp := h.sm.CtrlKickSession(info)
	feedback(resp, w)
	return
}

func (h *HttpApiServer) ctrlStartRtpPubHandler(w http.ResponseWriter, req *http.Request) {
	var v base.ApiCtrlStartRtpPub
	var info base.ApiCtrlStartRtpPubReq

	_, err := unmarshalRequestJsonBody(req, &info, "stream_name")
	if err != nil {
		Log.Warnf("http api start rtp pub error. err=%+v", err)
		v.ErrorCode = base.ErrorCodeParamMissing
		v.Desp = base.DespParamMissing
		feedback(v, w)
		return
	}

	Log.Infof("http api start rtp pub. req info=%+v", info)

	resp := h.sm.CtrlStartRtpPub(info)
	feedback(resp, w)
	return
}

// ---------------------------------------------------------------------------------------------------------------------

func (h *HttpApiServer) notFoundHandler(w http.ResponseWriter, req *http.Request) {
	Log.Warnf("invalid http-api request. uri=%s, raddr=%s", req.RequestURI, req.RemoteAddr)
}

func feedback(v interface{}, w http.ResponseWriter) {
	resp, _ := json.Marshal(v)
	w.Header().Add("Server", base.LalHttpApiServer)
	_, _ = w.Write(resp)
}

// unmarshalRequestJsonBody
//
// TODO(chef): [refactor] 搬到naza中 202205
//
func unmarshalRequestJsonBody(r *http.Request, info interface{}, keyFieldList ...string) (nazajson.Json, error) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nazajson.Json{}, err
	}

	j, err := nazajson.New(body)
	if err != nil {
		return j, err
	}
	for _, kf := range keyFieldList {
		if !j.Exist(kf) {
			return j, nazahttp.ErrParamMissing
		}
	}

	return j, json.Unmarshal(body, info)
}

// getVideo 获取历史视频资源
func (h *HttpApiServer) getVideo(w http.ResponseWriter, req *http.Request) {
	video := make([]string, 100)
	// 先暂时不鉴权
	q := req.URL.Query()
	stream := q.Get("stream")
	dir := filepath.Join(h.sm.config.RecordConfig.FlvOutPath, stream)
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		feedback(base.ApiVideo{
			HttpResponseBasic: base.HttpResponseBasic{
				ErrorCode: 400,
				Desp:      "读取视频文件目录失败",
			},
		}, w)
		return
	}
	for _, f := range files {
		video = append(video, f.Name())
	}
	feedback(base.ApiVideo{
		HttpResponseBasic: base.HttpResponseBasic{
			ErrorCode: 200,
			Desp:      "获取成功",
		},
		Data: video,
	}, w)
	return
}

// delVideo 删除历史视频资源
func (h *HttpApiServer) delVideo(w http.ResponseWriter, req *http.Request) {
	// 先暂时不鉴权
	q := req.URL.Query()
	name := q.Get("video")
	// FIXME: 需要知道是哪个流下边的视频资源
	file := filepath.Join(h.sm.config.RecordConfig.FlvOutPath, name)
	if err := os.Remove(file); err != nil {
		feedback(base.ApiVideo{
			HttpResponseBasic: base.HttpResponseBasic{
				ErrorCode: 400,
				Desp:      "删除视频失败",
			},
		}, w)
		return
	}
	feedback(base.ApiVideo{
		HttpResponseBasic: base.HttpResponseBasic{
			ErrorCode: 200,
			Desp:      "删除成功",
		},
	}, w)
	return
}
