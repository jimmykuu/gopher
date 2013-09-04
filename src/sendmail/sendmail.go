package main

import (
	"fmt"
	"github.com/jimmykuu/webhelpers"
	"gopher"
)

func main() {
	c := gopher.DB.C("users")
	var users []gopher.User
	c.Find(nil).All(&users)

	smtpConfig := webhelpers.SmtpConfig{
		Username: gopher.Config.SmtpUsername,
		Password: gopher.Config.SmtpPassword,
		Host:     gopher.Config.SmtpHost,
		Addr:     gopher.Config.SmtpAddr,
	}

	for _, user := range users {
		email := "jimmy.kuu@gmail.com"
		subject := "Golang中国域名改为golangtc.com"
		message := user.Username + "，您好！\n\n由于golang.tc域名被Godaddy没收，现已不可继续使用，现在开始使用golangtc.com域名。希望继续参与到社区建设中来。\n\n Golang中国"
		webhelpers.SendMail(subject, message, gopher.Config.FromEmail, []string{email}, smtpConfig, false)

		fmt.Println("send to:", user.Username, user.Email)
	}
}
