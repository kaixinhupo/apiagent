package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"sort"
	"sync"
)

var (
	defaultConfig *Config = nil
	clientMutex           = new(sync.Mutex)
)

const (
	TYPE_CSS   = "css"
	TYPE_JSON  = "json"
	TYPE_REGEX = "reg"
)

type KeyValuePair struct {
	Key   string
	Value string
}

// 参数
type Param struct {
	KeyValuePair
	IsConst bool //是否常量
}

// 表示HTTP请求
type Req struct {
	Url          string            //地址
	UrlParams    map[string]string //地址参数
	Method       string            //Http方法
	Headers      map[string]string //请求头
	Params       []*Param          //请求参数
	TemplatePath string            // 模板路径
	Encoding     string            //编码方式
}

// 代表响应数据中单项的解析规则
type ItemRule struct {
	Type  string //XPATH,JSON,CSS,REGEX
	Expr  string //表达式
	Key   string // 键
	Regex string // 对结果惊醒处理的正则表达式
}

// 表示响应数据中集合的解析规则
type CollectionRule struct {
	Type      string      //XPATH,JSON
	Expr      string      //表达式,返回一个集合
	ItemRules []*ItemRule //子项规则
	Key       string      // 键
}

// 表示HTTP响应
type Res struct {
	Extract         bool              //是否对结果经行处理
	Scope           string            //通过正则表达式限制结果范围
	Encoding        string            //编码方式
	ItemRules       []*ItemRule       //单条规则
	CollectionRules []*CollectionRule //集合规则
	Check           *Param            //校验规则
}

// 步骤
type Step struct {
	Sort   int  //序号
	Input  *Req //输入
	Output *Res //输出
}

// 任务
type Task struct {
	Name  string  //名称
	Steps []*Step //步骤
}

// 配置文件的GO表示
type Config struct {
	Port      int             //端口号
	AesKey    string          //加密Key
	CheckData string          //校验数据
	Arguments []*KeyValuePair //预设参数
	Tasks     []*Task         //任务列表
}

// 获取指定名称的Task
func (c *Config) GetTaskByName(name string) *Task {
	if c.Tasks == nil || len(c.Tasks) == 0 {
		return nil
	}
	for _, t := range c.Tasks {
		if t.Name == name {
			return t
		}
	}
	return nil
}

// 对Step进行排序
func (t *Task) SortSteps() {
	if t.Steps == nil || len(t.Steps) == 0 {
		return
	}
	sort.Slice(t.Steps, func(l, r int) bool {
		left := t.Steps[l]
		right := t.Steps[r]
		if left == nil {
			return true
		}
		if right == nil {
			return false
		}
		return left.Sort < right.Sort
	})
}

// 获取配置 单例
func DefaultConfig() (*Config, error) {
	if defaultConfig == nil {
		clientMutex.Lock()
		if defaultConfig == nil {
			var err error
			defaultConfig, err = readApplicationConfig()
			if err != nil {
				return nil, err
			}
		}
		clientMutex.Unlock()
	}
	return defaultConfig, nil
}

func readApplicationConfig() (*Config, error) {
	config, err1 := configPath()
	if err1 != nil {
		return nil, err1
	}
	configJson := path.Join(config, "config.json")
	if !pathExists(configJson) {
		return nil, errors.New("错误的应用程序路径")
	}
	configFile, err2 := os.Open(configJson)
	if err2 != nil {
		log.Println(err2)
		return nil, errors.New("读取配置发生错误")
	}
	data, err3 := ioutil.ReadAll(configFile)
	_ = configFile.Close()
	if err3 != nil {
		log.Println(err3)
		return nil, errors.New("读取配置发生错误")
	}
	appConfig := new(Config)
	err4 := json.Unmarshal(data, appConfig)
	if err4 != nil {
		log.Println(err4)
		return nil, errors.New("读取配置发生错误")
	}
	if appConfig.Tasks != nil {
		for _, t := range appConfig.Tasks {
			t.SortSteps()
		}
	}

	return appConfig, nil
}

func AppPath() (string, error) {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return "", err
	}
	return dir, nil
}

func configPath() (string, error) {
	current, err := AppPath()
	if err != nil {
		fmt.Println("获取应用程序路径发生错误")
		return "", err
	}
	config := path.Join(current, "config")

	if !pathExists(config) {
		return "", errors.New("错误的应用程序路径")
	}
	return config, nil
}

func pathExists(path string) bool {
	_, err := os.Stat(path) //os.Stat获取文件信息
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}
