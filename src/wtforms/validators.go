package wtforms

import (
    "regexp"
)

type IValidator interface {
    CleanData(value string) (bool, string)
}

type Required struct {
}

func (v Required) CleanData(value string) (bool, string) {
    if value == "" {
        return false, "该字段不能为空"
    }

    return true, ""
}

type Regexp struct {
    Expr    string
    Message string
}

func (v Regexp) CleanData(value string) (bool, string) {
    reg, err := regexp.Compile(v.Expr)
    if err != nil {
        panic(err)
    }

    if reg.MatchString(value) {
        return true, ""
    }

    return false, v.Message
}

type Email struct {
}

func (v Email) CleanData(value string) (bool, string) {
    tmp := Regexp{Expr: `^.+@[^.].*\.[a-z]{2,10}$`, Message: "无效的电子邮件地址"}

    return tmp.CleanData(value)
}

type URL struct {
}

func (v URL) CleanData(value string) (bool, string) {
    tmp := Regexp{Expr: `^(http|https)?://([^/:]+|([0-9]{1,3}\.){3}[0-9]{1,3})(:[0-9]+)?(\/.*)?$`, Message: "无效的URL"}

    return tmp.CleanData(value)
}
