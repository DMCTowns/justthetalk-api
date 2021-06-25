// This file is part of the JUSTtheTalkAPI distribution (https://github.com/jdudmesh/justthetalk-api).
// Copyright (c) 2021 John Dudmesh.

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, version 3.

// This program is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU
// General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package businesslogic

import (
	"bytes"
	"html/template"
	"justthetalk/model"
	"os"
	"strconv"
	"sync"

	gomail "gopkg.in/gomail.v2"
)

var onceTemplateMap sync.Once

const (
	NewSignupTemplate            = 1
	PasswordResetRequestTemplate = 2
	ReportSubmittedTemplate      = 3
	CharSet                      = "UTF-8"
)

func getTemplate(filename string) *template.Template {
	tpl, err := template.ParseFiles(filename)
	if err != nil {
		panic(err)
	}
	return tpl
}

type templateSpec struct {
	subject      string
	htmlTemplate *template.Template
	textTemplate *template.Template
}

var templateMap map[int]templateSpec

func getTemplateMap() map[int]templateSpec {
	onceTemplateMap.Do(func() {
		templateMap = map[int]templateSpec{
			NewSignupTemplate: {
				subject:      "Welcome to JUSTtheTalk",
				htmlTemplate: getTemplate("./email_templates/signup.html.tpl"),
				textTemplate: getTemplate("./email_templates/signup.text.tpl"),
			},
			PasswordResetRequestTemplate: {
				subject:      "JUSTtheTalk - Password Reset Request",
				htmlTemplate: getTemplate("./email_templates/password_reset.html.tpl"),
				textTemplate: getTemplate("./email_templates/password_reset.text.tpl"),
			},
			ReportSubmittedTemplate: {
				subject:      "JUSTtheTalk - Report Submitted",
				htmlTemplate: getTemplate("./email_templates/report_submitted.html.tpl"),
				textTemplate: getTemplate("./email_templates/report_submitted.text.tpl"),
			},
		}
	})
	return templateMap
}

func SendEmailToUser(user *model.User, params interface{}, templateType int) {
	SendEmail(user.Email, params, templateType)
}

func SendEmail(toAddress string, params interface{}, templateType int) {

	var buf bytes.Buffer
	config := getTemplateMap()[templateType]

	template := config.htmlTemplate
	if err := template.Execute(&buf, params); err != nil {
		panic(err)
	}

	htmlBody := string(buf.Bytes())

	buf.Reset()
	template = config.textTemplate
	if err := template.Execute(&buf, params); err != nil {
		panic(err)
	}
	textBody := string(buf.Bytes())

	mailFromAddress := os.Getenv("MAIL_FROM_ADDRESS")
	mailFromName := os.Getenv("MAIL_FROM_NAME")
	mailBccAddress := os.Getenv("MAIL_BCC_ADDRESS")
	mailBccName := os.Getenv("MAIL_BCC_NAME")
	mailHost := os.Getenv("MAIL_HOST")
	mailPort, _ := strconv.Atoi(os.Getenv("MAIL_PORT"))
	mailUser := os.Getenv("MAIL_USERNAME")
	mailPassword := os.Getenv("MAIL_PASSWORD")

	m := gomail.NewMessage()
	m.SetAddressHeader("From", mailFromAddress, mailFromName)
	m.SetHeader("To", toAddress)
	m.SetAddressHeader("Bcc", mailBccAddress, mailBccName)
	m.SetHeader("Subject", config.subject)
	m.SetBody("text/plain", textBody)
	m.AddAlternative("text/html", htmlBody)

	d := gomail.NewDialer(mailHost, mailPort, mailUser, mailPassword)

	// Send the email to Bob, Cora and Dan.
	if err := d.DialAndSend(m); err != nil {
		panic(err)
	}

}
