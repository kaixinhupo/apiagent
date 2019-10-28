package server

import (
	"encoding/json"
	"github.com/kaixinhupo/apiagent/config"
	http2 "github.com/kaixinhupo/apiagent/http"
	"github.com/kaixinhupo/apiagent/util"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

type HttpHandler struct {
}

type Message struct {
	Task string
	Data map[string]string
}

func (h *HttpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	log.Println("<<< ", r.Method, path)
	if path == "" || path == "/" {
		writeResponse(w, http.StatusOK, "Welcome")
	} else if path == "/favicon.ico" {
		writeResponse(w, http.StatusNotFound, "")
	} else if path == "call" && strings.ToLower(r.Method) == "post" {
		cfg, _ := config.DefaultConfig()
		if verify(r, cfg) {
			data, err := ioutil.ReadAll(r.Body)
			_ = r.Body.Close()
			if err != nil {
				log.Println("读取请求数据发生错误")
				writeResponse(w, http.StatusInternalServerError, "")
				return
			}
			msg := &Message{}
			err = json.Unmarshal(data, msg)
			if err != nil {
				log.Println("解析请求数据发生错误")
				writeResponse(w, http.StatusBadRequest, "")
				return
			}
			task := cfg.GetTaskByName(msg.Task)
			if task == nil {
				writeResponse(w, http.StatusBadRequest, "任务未定义")
				return
			}
			rst, err := http2.DefaultClient().RunTask(task)
			if err != nil {
				log.Println("执行任务时发生错误：", err)
				writeResponse(w, http.StatusInternalServerError, err.Error())
			} else {
				rstJson, err := json.Marshal(rst)
				if err == nil {
					writeResponse(w, http.StatusOK, string(rstJson))
				} else {
					log.Println("序列化结果数据时发生错误：", err)
					writeResponse(w, http.StatusInternalServerError, err.Error())
				}
			}
		} else {
			writeResponse(w, http.StatusForbidden, "")
		}
	} else {
		writeResponse(w, http.StatusNotFound, "")
	}
}

func verify(request *http.Request, cfg *config.Config) bool {
	token := request.Header.Get("x-token")
	if token == "" {
		return false
	}
	data := util.AesDecrypt([]byte(token), []byte(cfg.AesKey))
	return string(data) == cfg.CheckData
}

func writeResponse(w http.ResponseWriter, status int, body string) {
	w.WriteHeader(status)
	w.Header().Add("Content-Type", "plain/text")
	if len(body) > 0 {
		_, _ = w.Write([]byte(body))
	}
	log.Printf(">>> %d %s", status, body)
}
