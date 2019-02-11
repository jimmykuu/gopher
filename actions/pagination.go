package actions

import (
	"errors"
	"fmt"
	"math"
	"strings"

	"gopkg.in/mgo.v2"
)

// Pagination 分页
type Pagination struct {
	query   *mgo.Query // 查询体
	current int        // 当前页码
	count   int        // 内容总数
	perPage int        // 每页多少
}

// Page 返回第几页的查询
func (p *Pagination) Page(number int) (*mgo.Query, error) {
	pageCount := int(math.Ceil(float64(p.count) / float64(p.perPage)))
	query := p.query
	if count, _ := query.Count(); count == 0 {
		return query, nil
	}

	if !(number > 0 && number <= pageCount) {
		return nil, errors.New("页码不在范围内")
	}

	p.current = number

	if number > 1 {
		query = query.Skip(p.perPage * (number - 1))
	}

	return query.Limit(p.perPage), nil
}

// Count 内容总数
func (p *Pagination) Count() int {
	return p.count
}

// Current 当前页码
func (p *Pagination) Current() int {
	return p.current
}

// Prev 上一页，如果没有上一页，返回 0
func (p *Pagination) Prev() int {
	if p.current == 1 {
		return 0
	}

	return p.current - 1
}

// Next 下一页，如果没下一页，返回 0
func (p *Pagination) Next() int {
	last := p.count / p.perPage

	if p.current == last {
		return 0
	}

	return p.current + 1
}

// Url 生成链接
func (p *Pagination) Url(url string, page int) string {
	if strings.Contains(url, "?") {
		return fmt.Sprintf("%s&p=%d", url, page)
	} else {
		return fmt.Sprintf("%s?p=%d", url, page)
	}
}

// Pages 页码列表, 0 表示 ...
// 连续显示10个页码
func (p *Pagination) Pages() []int {
	last := p.count / p.perPage

	var result = []int{}

	if p.count == 0 || last == 0 {
		return result
	}

	// 当前页码位于第5个
	i := p.current

	result = append(result, i)

	// 计算当前值前面的4个值
	count := 0
	for i > 1 {
		i--
		result = append([]int{i}, result...)

		if count++; count == 4 {
			break
		}
	}

	i = p.current
	// 计算当前值后面的5个值
	count = 0
	for i < last {
		i++
		result = append(result, i)

		if count++; count == 5 {
			break
		}
	}

	// 如果当前页 < 5，补全10个
	if p.current < 5 {
		for len(result) < 10 {
			if result[len(result)-1] == last {
				break
			}

			result = append(result, result[len(result)-1]+1)
		}
	}

	// 如果当前页 > last - 5，补全10个
	if p.current > last-5 {
		for len(result) < 10 {
			if result[0] == 1 {
				break
			}

			result = append([]int{result[0] - 1}, result...)
		}
	}

	// 前部补1或补省略值
	if result[0] > 2 {
		result = append([]int{1, 0}, result...)
	} else if result[0] == 2 {
		result = append([]int{1}, result...)
	}

	// 后部补省略值或补最后值
	if result[len(result)-1] == last-1 {
		result = append(result, last)
	} else if result[len(result)-1] < last-1 {
		result = append(result, 0, last)
	}

	return result
}

// NewPagination 创建一个分页结构体
func NewPagination(query *mgo.Query, perPage int) *Pagination {
	p := Pagination{}
	p.query = query
	p.count, _ = query.Count()
	p.perPage = perPage

	return &p
}
