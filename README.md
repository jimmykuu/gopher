#Gopher

Golang中国([www.golangtc.com](http://www.golangtc.com))源代码.

##Requirements

- Go1.2+
- MongoDB
- github.com/gorilla/mux
- github.com/gorilla/sessions
- github.com/qiniu/bytes
- github.com/qiniu/rpc
- github.com/qiniu/api.v6
- labix.org/v2/mgo
- code.google.com/p/go-uuid/uuid
- github.com/jimmykuu/webhelpers
- github.com/jimmykuu/wtforms
- github.com/deferpanic/deferclient/deferclient

##Install

    $ go get github.com/jimmykuu/gopher/server


复制文件 *etc/config.json.default* 并改名为 *etc/config.json* 作为配置文件

- superusers: 内容为用户名,如果没有管理员,内容为"",如果有多个,用英文逗号隔开
- analytics_file: 内容为统计分析代码的文件名
- time_zone_offset: 时差，跟UTC的时间差，单位小时
- github_login_redirect: 第三方登录失败无法获取cookie跳转地址
- github_login_success_redirect: 第三放登录成功后跳转地址
- cookie_secure: 第三方登录需要使用HTTPS，当设置为false供本地测试使用
- deferpanic_api_key: deferpanic.com的Api Key，用于监控panic


内容如下:

    {
        "host": "http://localhost:8888",
        "port": 8888,
        "db": "localhost:27017",
        "cookie_secret": "05e0ba2eca9411e18155109add4b8aac",
        "smtp_username": "username@example.com",
        "smtp_password": "password",
        "smtp_host": "smtp.example.com",
        "smtp_addr": "smtp.example.com:25",
        "from_email": "who@example.com",
        "superusers": "jimmykuu,another",
        "analytics_file": "",
        "time_zone_offset": 8,
        "static_file_version": 1,
        "go_get_path": "/tmp/download",
        "packages_download_path": "/var/go/gopher/static/download/packages",
        "public_salt": "",
		"github_auth_client_id": "example",
		"github_auth_client_secret": "example",
		"github_login_redirect": "/",
		"github_login_success_redirect": "/auth/signup",
		"deferpanic_api_key": ""
    }

需要先启动MongoDB

生成私钥和证书,会在当前目录生成`key.pem`和`cert.pem`

	go run $GOROOT/src/crypto/tls/generate_cert.go --host 域名

Linux/Unix/OS X:

    $ $GOPATH/bin/server

Windows:

    > $GOPATH\bin\server.exe

或者:

	$ go build -o binary github.com/jimmykuu/gopher/server
	$ ./binary

##Contributors

- [Contributors](https://github.com/jimmykuu/gopher/graphs/contributors)


##License

Copyright (c) 2012-2015

Released under the MIT license:

- [www.opensource.org/licenses/MIT](http://www.opensource.org/licenses/MIT)
