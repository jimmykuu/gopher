/*
分页
*/

package gopher

import (
	"errors"
	"fmt"
	"html/template"
	"math"
	"strings"

	"labix.org/v2/mgo"
)

// 分页结构体
type Pagination struct {
	query   interface{}
	count   int
	perPage int
	url     string
}

// 在页面显示分页信息, 内容为 上一页 当前页/下一页 下一页
func (p *Pagination) Html(number int) template.HTML {
	pageCount := int(math.Ceil(float64(p.count) / float64(p.perPage)))

	if pageCount <= 1 {
		return template.HTML("")
	}

	linkFlag := "?"

	if strings.Index(p.url, "?") > -1 {
		linkFlag = "&"
	}

	html := `<ul class="pager">`
	if number > 1 {
		html += fmt.Sprintf(`<li class="previous"><a href="%s%sp=%d">&larr; 上一页</a></li>`, p.url, linkFlag, number-1)
	}

	html += fmt.Sprintf(`<li class="number">%d/%d</li>`, number, pageCount)

	if number < pageCount {
		html += fmt.Sprintf(`<li class="next"><a href="%s%sp=%d">下一页 &rarr;</a></li>`, p.url, linkFlag, number+1)
	}

	return template.HTML(html)
}

// 返回第几页的查询
func (p *Pagination) Page(number int) (interface{}, error) {
	pageCount := int(math.Ceil(float64(p.count) / float64(p.perPage)))
	switch p.query.(type) {
	case *mgo.Query:

		query := p.query.(*mgo.Query)

		if count, _ := query.Count(); count == 0 {
			return query, nil
		}

		if !(number > 0 && number <= pageCount) {
			return nil, errors.New("页码不在范围内")
		}

		if number > 1 {
			query = query.Skip(p.perPage * (number - 1))
		}
		return query.Limit(p.perPage), nil
	case []CollectTopic:
		cts := p.query.([]CollectTopic)
		if count := len(cts); count == 0 {
			return cts, nil
		}
		if !(number > 0 && number <= pageCount) {
			return nil, errors.New("页码不在范围内")
		}
		var end int
		if number*p.perPage > p.count {
			end = p.count
		} else {
			end = number * p.perPage
		}
		return cts[p.perPage*(number-1) : end], nil
	}
	return nil, errors.New("Query type is not *mgo.Query or slice")
}

// 内容总数
func (p *Pagination) Count() int {
	return p.count
}

// 创建一个分页结构体
func NewPagination(query interface{}, url string, perPage int) *Pagination {
	p := Pagination{}
	p.query = query
	switch query.(type) {
	case *mgo.Query:
		p.count, _ = query.(*mgo.Query).Count()
	case []CollectTopic:
		p.count = len(query.([]CollectTopic))
	}
	p.perPage = perPage
	p.url = url

	return &p
}
