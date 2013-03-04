#Gopher

Golang中国([www.golang.tc](http://www.golang.tc))源代码.

##Requirements

- Go1.0+
- MongoDB
- github.com/gorilla/mux
- github.com/gorilla/sessions
- labix.org/v2/mgo
- code.google.com/p/go-uuid/uuid
- github.com/jimmykuu/webhelpers
- github.com/jimmykuu/wtforms

##Install

    go get -u github.com/gorilla/mux
    go get -u github.com/gorilla/sessions
    go get -u labix.org/v2/mgo
	go get -u code.google.com/p/go-uuid/uuid
	go get -u github.com/jimmykuu/webhelpers
	go get -u github.com/jimmykuu/wtforms
    git clone git://github.com/jimmykuu/gopher.git

修改文件 *etc/config.json.default* 为 *etc/config.json* 作为配置文件

- superusers: 内容为用户名,如果没有管理员,内容为"",如果有多个,用英文逗号隔开
- analytics_file: 内容为统计分析代码的文件名

内容如下:

    {
        "host": "http://localhost:8888",
        "port": "8888",
        "db": "localhost:27017",
        "cookie_secret": "05e0ba2eca9411e18155109add4b8aac",
        "smtp_username": "username@example.com",
        "smtp_password": "password",
        "smtp_host": "smtp.example.com",
        "smtp_addr": "smtp.example.com:25",
        "from_email": "who@example.com",
        "superusers": "jimmykuu,another",
        "analytics_file": ""
    }

先启动MongoDB

Linux/Unix/Mac OS X:
    
    $ cd gopher
    $ ./build.sh
    $ ./bin/server

Windows:

    > cd gopher
    > build.bat
    > bin\server.exe

##Contributors

- [Contributors](https://github.com/jimmykuu/gopher/graphs/contributors)


##License

Copyright (c) 2012-2013

Released under the MIT license:

- [www.opensource.org/licenses/MIT](http://www.opensource.org/licenses/MIT)

