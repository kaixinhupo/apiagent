package parser

import (
	"github.com/Jeffail/gabs/v2"
	"github.com/PuerkitoBio/goquery"
	errors2 "github.com/kaixinhupo/apiagent/errors"
	"log"
	"regexp"
	"strings"
)

type OutputParser struct {
	Body        string
	jsonContext *gabs.Container
	htmlContext *goquery.Document
}

func NewOutputParser(body string) *OutputParser {
	return &OutputParser{Body: body}
}

func (p *OutputParser) ValueByJson(path string) (string, error) {
	ctx := p.getJsonContext()
	if ctx == nil {
		return "", errors2.NewContextError("创建上下文时发生异常")
	}
	if val, ok := ctx.Path(path).Data().(string); ok {
		return val, nil
	}
	return "", nil
}

func (p *OutputParser) ValueByJsonWithRegex(path string, regex string) (string, error) {
	val, err := p.ValueByJson(path)
	if err != nil {
		return "", err
	}
	return p.filterByRegex(val, regex)
}

func (p *OutputParser) filterByRegex(val string, regex string) (string, error) {
	reg, err := regexp.Compile(regex)
	if err != nil {
		log.Println("正则表达式错误：", err)
		return val, err
	}
	group := reg.FindAllStringSubmatch(val, -1)
	if group == nil || len(group) == 0 {
		return val, nil
	}
	return group[0][1], nil
}

func (p *OutputParser) ValueByCss(selector string) (string, error) {
	ctx := p.getHtmlContext()
	if ctx == nil {
		return "", errors2.NewContextError("创建上下文时发生异常")
	}
	return ctx.Find(selector).Text(), nil
}

func (p *OutputParser) ValueByCssWithRegex(selector string, regex string) (string, error) {
	val, err := p.ValueByCss(selector)
	if err != nil {
		return "", err
	}
	return p.filterByRegex(val, regex)
}

func (p *OutputParser) ValueByRegex(regex string) (string, error) {
	return p.filterByRegex(p.Body, regex)
}

func (p *OutputParser) SliceByJson(path string) ([]string, error) {

	ctx := p.getJsonContext()
	if ctx == nil {
		return nil, errors2.NewContextError("创建上下文时发生异常")
	}
	containers := ctx.Path(path).Children()

	length := len(containers)
	if length == 0 {
		return nil, nil
	}
	rst := make([]string, length)

	var i int
	for i = 0; i < length; i++ {
		node := containers[i]
		if val, ok := node.Data().(string); ok {
			rst[i] = val
		}
	}
	return rst, nil
}

func (p *OutputParser) SliceByCss(selector string) ([]string, error) {
	ctx := p.getHtmlContext()
	if ctx == nil {
		return nil, errors2.NewContextError("创建上下文时发生异常")
	}
	selection := ctx.Find(selector)
	length := selection.Size()
	log.Println("selector:", selector, " length:", length)
	if length == 0 {
		return nil, nil
	}
	rst := make([]string, length)
	selection.Each(func(i int, sel *goquery.Selection) {
		html, _ := sel.Html()
		rst[i] = html
	})
	return rst, nil
}

func (p *OutputParser) getHtmlContext() *goquery.Document {
	if p.Body == "" {
		return nil
	}
	if p.htmlContext == nil {
		ctx, err := goquery.NewDocumentFromReader(strings.NewReader(p.Body))
		if err != nil {
			log.Println("创建HtmlContext时发生错误:", err)
			return nil
		}
		p.htmlContext = ctx
	}
	return p.htmlContext
}

func (p *OutputParser) getJsonContext() *gabs.Container {
	if p.Body == "" {
		return nil
	}
	if p.jsonContext == nil {
		ctx, err := gabs.ParseJSON([]byte(p.Body))
		if err != nil {
			log.Println("创建JsonContext时发生错误:", err)
			return nil
		}
		p.jsonContext = ctx
	}
	return p.jsonContext
}
