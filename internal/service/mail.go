package service

import (
	"net/smtp"
	"strconv"
)

type Mail struct {
	Host     string
	Port     string
	FromAddr string
	auth     smtp.Auth
}

func NewMailService(username, password, host, fromAddr string, port int) *Mail {
	return &Mail{
		Host:     host,
		Port:     strconv.Itoa(port),
		FromAddr: fromAddr,
		auth:     smtp.PlainAuth("", username, password, host),
	}
}

func (s *Mail) SendMail(toAddrs []string, msg string) error {
	return smtp.SendMail(s.Host+":"+s.Port, s.auth, s.FromAddr, toAddrs, []byte(msg))
}
