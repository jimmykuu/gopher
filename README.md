#Gopher

Golang中国([www.golangtc.com](http://www.golangtc.com))源代码.

##Install

    $ go get -v github.com/jimmykuu/gopher/server

修改文件 *etc/config.json.default* 为 *etc/config.json* 作为配置文件

- superusers: 内容为用户名,如果没有管理员,内容为"",如果有多个,用英文逗号隔开
- analytics_file: 内容为统计分析代码的文件名
- time_zone_offset: 时差，跟UTC的时间差，单位小时
- github_login_redirect: 第三方登录失败无法获取cookie跳转地址
- github_login_success_redirect:第三放登录成功后跳转地址
- cookie_secure:第三方登录需要使用HTTPS,当设置为false供本地测试使用

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
        "4ublic_salt": "nXweu8Jq44FgEfgM1Pv4xH51"
		"github_auth_client_id":"example",
		"github_auth_client_secret":"example",
		"github_login_redirect":"/",
		"github_login_success_redirect":"/auth/signup"
    }

先启动MongoDB

    $GOPATH/bin/server

##Contributors

- [Contributors](https://github.com/jimmykuu/gopher/graphs/contributors)


##License

Copyright (c) 2012-2013

Released under the MIT license:

- [www.opensource.org/licenses/MIT](http://www.opensource.org/licenses/MIT)
