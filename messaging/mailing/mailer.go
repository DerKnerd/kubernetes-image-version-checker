package mailing

import (
	"bytes"
	"github.com/scorredoira/email"
	"html/template"
	"io/ioutil"
	"kubernetes-pod-version-checker/messaging"
	"net/smtp"
)

type Mailer struct {
	To       []string `yaml:"to"`
	From     string   `yaml:"from"`
	Username string   `yaml:"username"`
	Password string   `yaml:"password"`
	Host     string   `yaml:"host"`
	Port     string   `yaml:"port"`
}

func New(to []string, from string, username string, password string, host string, port string) *Mailer {
	return &Mailer{
		To:       to,
		From:     from,
		Username: username,
		Password: password,
		Host:     host,
		Port:     port,
	}
}

func (mailer Mailer) SendMail(message messaging.Message) error {
	tmpl, err := template.New("email").ParseFiles("mailing/mail-body.gohtml")
	if err != nil {
		return err
	}

	buffer := bytes.NewBuffer([]byte{})
	err = tmpl.Execute(buffer, message)

	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(buffer)
	if err != nil {
		return err
	}

	htmlMessage := email.NewHTMLMessage("New version for image "+message.Image, string(body))
	htmlMessage.To = mailer.To
	htmlMessage.From.Address = mailer.From
	htmlMessage.BodyContentType = "text/html"

	var auth smtp.Auth
	if mailer.Username != "" && mailer.Password != "" {
		auth = smtp.PlainAuth("", mailer.Username, mailer.Password, mailer.Host)
	}

	return email.Send(mailer.Host+":"+mailer.Port, auth, htmlMessage)
}
