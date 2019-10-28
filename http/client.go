package http

import (
	"errors"
	"github.com/Jeffail/gabs/v2"
	"github.com/axgle/mahonia"
	"github.com/hoisie/mustache"
	cfg "github.com/kaixinhupo/apiagent/config"
	errors2 "github.com/kaixinhupo/apiagent/errors"
	"github.com/kaixinhupo/apiagent/parser"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type HttpClient struct {
	http.Client
	IsLogin bool
}

var defaultClient *HttpClient = nil

var clientMutex = new(sync.Mutex)

func DefaultClient() *HttpClient {
	if defaultClient == nil {
		clientMutex.Lock()
		if defaultClient == nil {
			defaultClient = &HttpClient{IsLogin: true}
		}
		clientMutex.Unlock()
	}
	return defaultClient
}

func (h *HttpClient) Login() {
	if h.IsLogin {
		return
	}
	config, _ := cfg.DefaultConfig()
	task := config.GetTaskByName("login")
	_, err := h.RunTask(task)
	if err != nil {
		h.IsLogin = false
	}
}

func (h *HttpClient) RunTask(task *cfg.Task) (map[string]interface{}, error) {
	if task == nil {
		return nil, errors.New("Task does not exist")
	}
	if task.Steps == nil || len(task.Steps) == 0 {
		return nil, errors.New("Task does not contain any steps")
	}
	result := make(map[string]interface{})
	context := make(map[string]string)

	config, _ := cfg.DefaultConfig()
	if config.Arguments != nil {
		for _, p := range config.Arguments {
			context[p.Key] = p.Value
		}
	}

	for _, s := range task.Steps {
		body, err := h.RunStep(s, &context)
		if err != nil {
			return nil, err
		}

		if body != nil {
			check := s.Output.Check
			if check != nil {
				var chkVal string
				if !check.IsConst {
					if _v, ok := context[check.Key]; ok {
						chkVal = _v
					}
				} else {
					chkVal = check.Value
				}

				if v, ok := body[check.Key]; ok {
					if v != chkVal {
						return nil, errors2.NewCheckError("校验不通过")
					}
				} else {
					return nil, errors2.NewCheckError("校验不通过")
				}
			}

			for k, v := range body {
				if str, ok := v.(string); ok {
					context[k] = str
				}
				result[k] = v
			}
		}
	}
	return result, nil
}

func (h *HttpClient) RunStep(step *cfg.Step, context *map[string]string) (map[string]interface{}, error) {
	method := strings.ToLower(step.Input.Method)
	if method == "" {
		method = "get"
	}
	var req *http.Request
	if method == "get" {
		req = parseGet(step, context)
	} else {
		req = parsePost(step, context)
	}

	if step.Input.Headers != nil {
		for k, v := range step.Input.Headers {
			req.Header.Set(k, v)
		}
	}

	resp, err := h.Do(req)
	if err != nil {
		return nil, err
	}
	data, _ := ioutil.ReadAll(resp.Body)
	_ = resp.Body.Close()

	body, err := parseBody(data, step.Output.Encoding)
	if err != nil {
		return nil, err
	}
	return parser.ParseStepResult(step, body)
}

func parseBody(data []byte, encoding string) (string, error) {
	if encoding == "" || strings.ToLower(encoding) == "utf-8" {
		return string(data), nil
	}
	decoder := mahonia.NewDecoder(encoding)
	_, bodyBytes, err := decoder.Translate(data, true)
	if err != nil {
		return "", err
	}
	body := string(bodyBytes)

	return body, nil
}

func parsePost(step *cfg.Step, context *map[string]string) *http.Request {
	urlStr := buildUrl(step, context)
	log.Println("url:", urlStr)
	formBody := buildBody(step, context)
	request, _ := http.NewRequest(http.MethodPost, urlStr, strings.NewReader(formBody))
	return request
}

func buildBody(step *cfg.Step, context *map[string]string) string {
	templatePath := step.Input.TemplatePath
	var mimeType string
	if ct, ok := step.Input.Headers["Content-Type"]; ok {
		mimeType = ct
	} else if ct, ok := step.Input.Headers["content-type"]; ok {
		mimeType = ct
	} else if templatePath != "" {
		mimeType = "application/json"
	} else {
		mimeType = "application/x-www-form-urlencoded"
	}
	if step.Input.TemplatePath != "" {
		return renderTemplate(step.Input.TemplatePath, step.Input.Params, context)
	} else {
		params := step.Input.Params
		if params != nil {
			if strings.Contains(mimeType, "json") {
				json := gabs.New()
				for _, v := range params {
					var val string
					if v.IsConst {
						val = v.Value
					} else {
						if value, ok := (*context)[v.Value]; ok {
							val = value
						}
					}
					_, _ = json.Set(val, v.Key)
				}
				return json.String()
			} else {
				return buildFormStr(step.Input.Encoding, params, context)
			}
		}
	}
	return ""
}

func EncodeStr(input string, encoding string) string {
	return input
}

func renderTemplate(path string, params []*cfg.Param, context *map[string]string) string {

	appPath, _ := cfg.AppPath()
	templatePath := filepath.Join(appPath, "config", "templates", path)
	templateFile, err := os.Open(templatePath)
	if err != nil {
		log.Println("读取模板发生错误", err)
		return ""
	}
	data, err := ioutil.ReadAll(templateFile)
	_ = templateFile.Close()
	_context := mergeContext(params, context)
	return mustache.Render(string(data), _context)
}

func mergeContext(params []*cfg.Param, ctx *map[string]string) map[string]string {
	context := *ctx
	result := make(map[string]string, len(params)+len(context))
	for k, v := range context {
		result[k] = v
	}
	for _, v := range params {
		var val string
		if v.IsConst {
			val = v.Value
		} else {
			if value, ok := context[v.Value]; ok {
				val = value
			}
		}
		result[v.Key] = val
	}
	return result
}

func parseGet(step *cfg.Step, context *map[string]string) *http.Request {
	urlStr := buildUrl(step, context)
	log.Println(urlStr)
	params := step.Input.Params
	if params != nil {
		query := buildFormStr(step.Input.Encoding, params, context)
		if !strings.Contains(urlStr, "?") {
			urlStr += "?"
		}
		if !strings.HasSuffix(urlStr, "&") && !strings.HasSuffix(urlStr, "?") {
			urlStr += "&"
		}
		urlStr = urlStr + query
	}
	request, _ := http.NewRequest(http.MethodGet, urlStr, nil)
	return request
}

func buildFormStr(encoding string, params []*cfg.Param, context *map[string]string) string {
	builder := strings.Builder{}
	for _, v := range params {
		var val string
		if v.IsConst {
			val = v.Value
		} else {
			if value, ok := (*context)[v.Value]; ok {
				val = value
			}
		}
		builder.WriteString(EncodeStr(v.Key, encoding))
		builder.WriteByte('=')
		builder.WriteString(EncodeStr(val, encoding))
	}
	return builder.String()
}

func buildUrl(step *cfg.Step, context *map[string]string) string {
	urlStr := step.Input.Url
	if step.Input.UrlParams != nil {
		for k, v := range step.Input.UrlParams {
			if val, ok := (*context)[v]; ok {
				urlStr = strings.ReplaceAll(urlStr, k, val)
			}
		}
	}
	return urlStr
}
