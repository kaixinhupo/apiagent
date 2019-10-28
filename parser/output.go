package parser

import (
	config2 "github.com/kaixinhupo/apiagent/config"
	errors2 "github.com/kaixinhupo/apiagent/errors"
	"github.com/kaixinhupo/apiagent/util"
	"log"
)

func ParseStepResult(step *config2.Step, body string) (map[string]interface{}, error) {
	if body == "" {
		return make(map[string]interface{}), nil
	}
	if !step.Output.Extract {
		return map[string]interface{}{"body": body}, nil
	}
	var _body string
	if step.Output.Scope != "" {
		parser, err := NewOutputParser(body).ValueByRegex(body)
		if err != nil {
			return nil, err
		}
		_body = parser
	} else {
		_body = body
	}
	result := make(map[string]interface{})
	context := NewOutputParser(_body)
	log.Println("提取单个元素")
	rst := extractByRules(step.Output.ItemRules, context)
	if rst != nil {
		util.CopyMap(rst, result)
	}
	collection := step.Output.CollectionRules
	if collection != nil {
		log.Println("提取集合元素")
		for _, v := range collection {
			if v.Type == config2.TYPE_REGEX {
				continue
			}
			group, err := collectionByRule(v, context)
			if err != nil {
				break
			}
			if group != nil {
				arr := make([]map[string]interface{}, len(group))
				for i, g := range group {
					rst = extractByRules(v.ItemRules, NewOutputParser(g))
					arr[i] = rst
				}
				result[v.Key] = arr
			}
		}
	}
	log.Println("result len:", len(result))
	return result, nil
}

func extractByRules(rules []*config2.ItemRule, context *OutputParser) map[string]interface{} {
	result := make(map[string]interface{})
	if rules != nil {
	rulesloop:
		for _, rule := range rules {
			val, err := valueByRule(rule, context)
			if err != nil {
				break rulesloop
			}
			result[rule.Key] = val
		}
	}
	return result
}

func valueByRule(rule *config2.ItemRule, context *OutputParser) (string, error) {
	switch rule.Type {
	case config2.TYPE_CSS:
		return extractCssValue(context, rule)
	case config2.TYPE_JSON:
		return extractJsonValue(context, rule)
	case config2.TYPE_REGEX:
		return extractRegexValue(context, rule), nil
	}
	return "", nil
}

func collectionByRule(rule *config2.CollectionRule, context *OutputParser) ([]string, error) {
	var val []string
	var err error
	switch rule.Type {
	case config2.TYPE_CSS:
		val, err = context.SliceByCss(rule.Expr)
		break
	case config2.TYPE_JSON:
		val, err = context.SliceByJson(rule.Expr)
		break
	}
	if err != nil {
		if _, ok := err.(errors2.ContextError); ok {
			return nil, err
		} else {
			return nil, nil
		}
	}
	return val, nil
}

func extractJsonValue(parser *OutputParser, rule *config2.ItemRule) (string, error) {
	var val string
	var err error
	if rule.Regex != "" {
		val, err = parser.ValueByJsonWithRegex(rule.Expr, rule.Regex)
	} else {
		val, err = parser.ValueByJson(rule.Expr)
	}
	if err != nil {
		if _, ok := err.(errors2.ContextError); ok {
			return "", err
		} else {
			return "", nil
		}
	}
	return val, nil
}

func extractRegexValue(parser *OutputParser, rule *config2.ItemRule) string {
	val, err := parser.ValueByRegex(rule.Expr)
	if err != nil {
		return ""
	}
	return val
}

func extractCssValue(parser *OutputParser, rule *config2.ItemRule) (string, error) {
	var val string
	var err error
	if rule.Regex != "" {
		val, err = parser.ValueByCssWithRegex(rule.Expr, rule.Regex)
	} else {
		val, err = parser.ValueByCss(rule.Expr)
	}
	if err != nil {
		if _, ok := err.(errors2.ContextError); ok {
			return "", err
		} else {
			return "", nil
		}
	}
	return val, nil
}
