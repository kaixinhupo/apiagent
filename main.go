package main

import (
	"bufio"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/kaixinhupo/apiagent/config"
	http2 "github.com/kaixinhupo/apiagent/http"
	"github.com/kaixinhupo/apiagent/server"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

var (
	client = http2.DefaultClient()
)

func main() {
	log.Println("代理开始启动")
	log.Println("读取配置文件")
	appConfig, err := config.DefaultConfig()
	if err != nil {
		log.Println(err)
		waitKey()
		return
	}
	count := 0
	for ; count < 5; count++ {
		client.Login()
		if client.IsLogin {
			log.Println("登录成功")
			break
		}
		log.Println("登录失败，5秒后重试")
		time.Sleep(5 * time.Second)
	}
	if client.IsLogin {
		startHttp(appConfig)
	} else {
		log.Println("5次重试后依然失败，请确认登录信息")
		waitKey()
	}
}

func startHttp(config *config.Config) {
	mux := http.NewServeMux()
	mux.Handle("/", &server.HttpHandler{})
	log.Println("启动服务,端口:", config.Port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", config.Port), mux)
	if err != nil {
		log.Println("启动失败：", err)
		return
	}
}

func waitKey() {
	log.Println("按回车键结束...")
	inputReader := bufio.NewReader(os.Stdin)
	_, _ = inputReader.ReadByte()
}

func main2() {
	data, _ := ioutil.ReadFile("C:\\Users\\MSI-PC\\Desktop\\test.html")
	html := string(data)
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
	doc.Find("div.row.card").Each(func(i int, selection *goquery.Selection) {
		fmt.Println("--------------------", i, "-------------------------")
		fmt.Println(selection.Text())
	})
}
