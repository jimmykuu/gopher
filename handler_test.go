package gopher

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// 获取路由.
func TestParam(t *testing.T) {
	res := httptest.NewRecorder()
	testId := "24aaaaaaaaaaaaaaaaaaaaaa"
	req, err := http.NewRequest("GET", Config.Host+"/t/"+testId, nil)
	if err != nil {
		panic(err)
	}
	t.Log(res.Code)
	r := getRoute()
	testParam = func() {
		handler := &Handler{
			ResponseWriter: res,
			Request:        req,
		}
		id := handler.param("topicId")
		if id != testId {
			t.Fatal("Handler.param failed,want 'id' but get '" + id + "'")
		}
	}
	r.ServeHTTP(res, req)
}
