#Gopher

Golang中国([www.golang.tc](http://www.golang.tc))源代码.

##Requirements

- Go1.0+
- MongoDB
- code.google.com/p/gorilla/mux
- code.google.com/p/gorilla/sessions
- labix.org/v2/mgo

##Install

    go get code.google.com/p/gorilla/mux
    go get code.google.com/p/gorilla/sessions
    go get labix.org/v2/mgo
    git clone git://github.com/jimmykuu/gopher.git
	
创建配置文件config.json文件,内容如下:

    {
        "host": "http://localhost:8888",
        "port": "8888",
        "cookie_secret": "05e0ba2eca9411e18155109add4b8aac",
        "smtp_username": "username@example.com",
        "smtp_password": "password",
        "smtp_host": "smtp.example.com",
        "smtp_addr": "smtp.example.com:25",
        "from_email": "who@example.com"
    }

先启动MongoDB

然后运行命令

	go run *.go

或者

    go build -o gopher *.go
    ./gopher

##Contributors

- [Contributors](https://github.com/jimmykuu/gopher/graphs/contributors)


##License

Copyright (c) 2012

Released under the MIT license:

- [www.opensource.org/licenses/MIT](http://www.opensource.org/licenses/MIT)

