package actions

import (
	"errors"
	"math"

	"gopkg.in/mgo.v2"

	"github.com/jimmykuu/gopher/models"
)

// Pagination 分页
type Pagination struct {
	query   interface{} // 查询体，可能是 mgo.Query 也可能是 slice
	current int         // 当前页码
	count   int         // 内容总数
	perPage int         // 每页多少
}

// Page 返回第几页的查询
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

		p.current = number

		if number > 1 {
			query = query.Skip(p.perPage * (number - 1))
		}
		return query.Limit(p.perPage), nil
	case []models.CollectTopic:
		cts := p.query.([]models.CollectTopic)
		if count := len(cts); count == 0 {
			return cts, nil
		}
		if !(number > 0 && number <= pageCount) {
			return nil, errors.New("页码不在范围内")
		}

		p.current = number

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

// Pages 页码列表, 0 表示 ...
// 连续显示10个页码
func (p *Pagination) Pages() []int {
	last := p.count / p.perPage

	var result = []int{}

	if p.count == 0 {
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

	// 后补补省略值或补最后值
	if result[len(result)-1] == last-1 {
		result = append(result, last)
	} else if result[len(result)-1] < last-1 {
		result = append(result, 0, last)
	}

	return result
}

// NewPagination 创建一个分页结构体
func NewPagination(query interface{}, perPage int) *Pagination {
	p := Pagination{}
	p.query = query
	switch query.(type) {
	case *mgo.Query:
		p.count, _ = query.(*mgo.Query).Count()
	case []models.CollectTopic:
		p.count = len(query.([]models.CollectTopic))
	}
	p.perPage = perPage

	return &p
}
