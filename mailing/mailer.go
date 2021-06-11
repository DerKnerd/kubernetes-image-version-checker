package mailing

import (
	"bytes"
	"github.com/scorredoira/email"
	"html/template"
	"io/ioutil"
	"net/smtp"
	"os"
)

func SendMail(usedVersion string, latestVersion string, image string, parentName string, entityType string) error {
	tmpl, err := template.New("email").ParseFiles("mailing/mail-body.gohtml")
	if err != nil {
		return err
	}

	type tmplData struct {
		UsedVersion   string
		LatestVersion string
		Image         string
		ParentName    string
		EntityType    string
	}

	buffer := bytes.NewBuffer([]byte{})
	err = tmpl.Execute(buffer, tmplData{
		UsedVersion:   usedVersion,
		LatestVersion: latestVersion,
		Image:         image,
		ParentName:    parentName,
		EntityType:    entityType,
	})

	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(buffer)
	if err != nil {
		return err
	}

	to := os.Getenv("MAILING_TO")
	from := os.Getenv("MAILING_FROM")
	username := os.Getenv("MAILING_USERNAME")
	password := os.Getenv("MAILING_PASSWORD")
	host := os.Getenv("MAILING_HOST")
	port := os.Getenv("MAILING_PORT")

	message := email.NewHTMLMessage("New version for image "+image, string(body))
	message.To = []string{to}
	message.From.Address = from
	message.BodyContentType = "text/html"

	var auth smtp.Auth
	if username != "" && password != "" {
		auth = smtp.PlainAuth("", username, password, host)
	}

	return email.Send(host+":"+port, auth, message)
}
