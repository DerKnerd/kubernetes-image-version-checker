package mailing

import (
	"bytes"
	"github.com/hashicorp/go-version"
	"github.com/scorredoira/email"
	"html/template"
	"io/ioutil"
	appsv1 "k8s.io/api/apps/v1"
	"net/smtp"
	"os"
)

func SendMail(usedVersion version.Version, latestVersion version.Version, image string, deployment appsv1.Deployment) error {
	tmpl, err := template.New("email").ParseFiles("mailing/mail-body.gohtml")
	if err != nil {
		return err
	}

	type tmplData struct {
		UsedVersion   string
		LatestVersion string
		Image         string
		Deployment    string
	}

	buffer := bytes.NewBuffer([]byte{})
	err = tmpl.Execute(buffer, tmplData{
		UsedVersion:   usedVersion.String(),
		LatestVersion: latestVersion.String(),
		Image:         image,
		Deployment:    deployment.GetName(),
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
