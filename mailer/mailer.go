// Copyright (c) 2019 JasaCloud.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package mailer

import (
	"bytes"
	"crypto/sha1"
	"crypto/tls"
	"encoding/base64"
	"encoding/hex"
	"github.com/jasacloud/go-libraries/client"
	"github.com/jasacloud/go-libraries/config"
	"github.com/jasacloud/go-libraries/system"
	"gopkg.in/gomail.v2"
	"io"
	"log"
	"mime"
	"path/filepath"
	"strings"
)

// MailOption struct
type MailOption struct {
	Name       string `json:"name" bson:"name"`
	Protocol   string `json:"protocol" bson:"protocol"`
	Host       string `json:"host" bson:"host"`
	Port       int    `json:"port" bson:"port"`
	SecureType string `json:"secureType" bson:"secureType"`
	Sender     string `json:"sender" bson:"sender"`
	SenderName string `json:"senderName" bson:"senderName"`
	Auth       bool   `json:"auth" bson:"auth"`
	UserName   string `json:"userName" bson:"userName"`
	Password   string `json:"password" bson:"password"`
}

// MailConf struct
type MailConf struct {
	MailOption []MailOption `json:"mailResources" bson:"mailResources"`
}

// Mailer struct
type Mailer struct {
	d   *gomail.Dialer
	s   gomail.SendCloser
	m   *gomail.Message
	opt MailOption
}

// Attachment struct
type Attachment struct {
	Name string `json:"name,omitempty" bson:"name,omitempty"`
	Type string `json:"type,omitempty" bson:"type,omitempty"`
	Data string `json:"data,omitempty" bson:"data,omitempty"`
	Url  string `json:"url,omitempty" bson:"url,omitempty"`
	Path string `json:"path,omitempty" bson:"path,omitempty"`
}

var (
	// Mail variable
	Mail MailConf
)

// GetMailResource function
func GetMailResource(resourceName string) MailOption {
	c := config.GetConfig()
	config.GetConf(c.ByteConfig, &Mail)
	for _, v := range Mail.MailOption {
		if v.Name == resourceName {

			return v
		}
	}

	return MailOption{}
}

// EmailDial function
func EmailDial(resourceName string) *Mailer {
	mailOpt := GetMailResource(resourceName)

	d := gomail.NewDialer(mailOpt.Host, mailOpt.Port, mailOpt.UserName, mailOpt.Password)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	s, err := d.Dial()
	if err != nil {
		log.Println("Error dial email :", err)
	}
	return &Mailer{
		d,
		s,
		nil,
		mailOpt,
	}
}

// SetMailOption function
func SetMailOption(mailOpt MailOption) *Mailer {
	d := gomail.NewDialer(mailOpt.Host, mailOpt.Port, mailOpt.UserName, mailOpt.Password)
	if d == nil {
		log.Println("Error on create gomail.NewDialer")
		return nil
	}
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	s, err := d.Dial()
	if err != nil {
		log.Println("Error dial email :", err)
		return nil
	}
	return &Mailer{
		d,
		s,
		nil,
		mailOpt,
	}
}

// getHashString function
func getHashString(s string) string {
	h := sha1.New()
	h.Write([]byte(s))

	return hex.EncodeToString(h.Sum(nil))
}

// getMessageId function
func getMessageId(email string) string {
	if email == "" {
		email = "mail@unknown.com"
	}
	components := strings.Split(email, "@")
	_, domain := components[0], components[1]

	return "<" + getHashString(system.GetUID()) + ".jc@" + domain + ">"
}

// Send method
func (mailer *Mailer) Send(to string, cc string, subject string, contentType string, content string, attachment ...string) error {
	m := gomail.NewMessage(gomail.SetEncoding("base64"), gomail.SetCharset("iso-8859-1"))
	m.SetHeader("Message-ID", getMessageId(mailer.opt.Sender))
	m.SetHeader("To", to)
	if cc != "" {
		m.SetHeader("Cc", cc)
	}
	//m.SetHeader("From", "dwi@jasacloud.com")
	//m.SetHeader("From", mailer.opt.Sender)
	m.SetAddressHeader("From", mailer.opt.Sender, mailer.opt.SenderName)
	m.SetHeader("Subject", subject)
	m.SetHeader("X-Priority", "1")
	m.SetHeader("X-Mailer", "GSkyMailer 1.0")
	m.SetHeader("X-MSMail-Priority", "High")
	m.SetHeader("Importance", "High")
	m.SetHeader("Mime-Version", "1.0")
	m.SetBody(contentType, content)

	for _, attach := range attachment {
		if attach != "" {
			m.Attach(attach)
		}
	}
	mailer.m = m
	return gomail.Send(mailer.s, mailer.m)
}

// SendWithAttachments method
func (mailer *Mailer) SendWithAttachments(to string, cc string, subject string, contentType string, content string, attachments ...Attachment) error {
	m := gomail.NewMessage(gomail.SetEncoding("base64"), gomail.SetCharset("iso-8859-1"))
	m.SetHeader("Message-ID", getMessageId(mailer.opt.Sender))
	m.SetHeader("To", to)
	if cc != "" {
		m.SetHeader("Cc", cc)
	}
	//m.SetHeader("From", "dwi@jasacloud.com")
	//m.SetHeader("From", mailer.opt.Sender)
	m.SetAddressHeader("From", mailer.opt.Sender, mailer.opt.SenderName)
	m.SetHeader("Subject", subject)
	m.SetHeader("X-Priority", "1")
	m.SetHeader("X-Mailer", "GSkyMailer 1.0")
	m.SetHeader("X-MSMail-Priority", "High")
	m.SetHeader("Importance", "High")
	m.SetHeader("Mime-Version", "1.0")
	m.SetBody(contentType, content)

	for _, attachment := range attachments {
		if attachment.Data != "" {
			var f = func(a Attachment) func(w io.Writer) error {
				return func(w io.Writer) error {
					var dataUri = a.Data
					log.Println("dataUri:", dataUri)
					coI := strings.Index(dataUri, ",")
					rawImage := dataUri[coI+1:]
					unBased, _ := base64.StdEncoding.DecodeString(string(rawImage))
					res := bytes.NewReader(unBased)
					if _, err := io.Copy(w, res); err != nil {
						return err
					}
					return nil
				}
			}
			if attachment.Name == "" {
				attachment.Name = "attachment"
			}
			attachment.Name = strings.TrimSuffix(attachment.Name, filepath.Ext(attachment.Name))
			coI := strings.Index(attachment.Data, ",")
			contentType := strings.TrimSuffix(attachment.Data[5:coI], ";base64")
			ext, err := mime.ExtensionsByType(contentType)
			if err != nil {
				log.Println(err)
				continue
			}
			if len(ext) > 0 {
				attachment.Name = attachment.Name + ext[0]
			}

			m.Attach(attachment.Name, gomail.SetCopyFunc(f(attachment)))
		} else if attachment.Url != "" {
			var url = attachment.Url
			log.Println("URL:", url)
			cl := client.LoadHttp(url)
			cl.SetRequest("GET", "", nil)
			resp, err := cl.Start()
			if err != nil {
				log.Println(err)
				return nil
			}
			name := filepath.Base(attachment.Url)
			if name == "" {
				name = "attachment"
			}
			if attachment.Name != "" {
				name = attachment.Name

			}
			ext, err := mime.ExtensionsByType(resp.Header.Get("Content-Type"))
			if err != nil {
				log.Println(err)
				return nil
			}
			if len(ext) > 0 {
				cExt := strings.ToLower(ext[0])
				if extName := strings.ToLower(filepath.Ext(name)); extName != "" && extName == cExt {
					name = strings.TrimSuffix(name, filepath.Ext(name)) + cExt
				} else {
					name = name + cExt
				}
			}

			var f = func(r io.ReadCloser) func(w io.Writer) error {
				return func(w io.Writer) error {
					defer r.Close()
					if _, err := io.Copy(w, r); err != nil {
						return err
					}
					return nil
				}
			}
			m.Attach(name, gomail.SetCopyFunc(f(resp.Body)))
		} else if attachment.Path != "" {
			m.Attach(attachment.Path)
		}
	}
	mailer.m = m
	return gomail.Send(mailer.s, mailer.m)
}

// SendEmail function
func SendEmail(resourceName string, to string, cc string, subject string, contentType string, content string, attachment ...string) error {

	mailOpt := GetMailResource(resourceName)

	d := gomail.NewDialer(mailOpt.Host, mailOpt.Port, mailOpt.UserName, mailOpt.Password)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	m := gomail.NewMessage()
	//m.SetHeader("From", "dwi@jasacloud.com")
	m.SetAddressHeader("From", mailOpt.Sender, mailOpt.SenderName)
	m.SetHeader("To", to)
	if cc != "" {
		m.SetAddressHeader("Cc", cc, cc)
	}
	m.SetHeader("Subject", subject)
	m.SetBody(contentType, content)

	for _, attach := range attachment {
		if attach != "" {
			m.Attach(attach)
		}
	}

	// Send the email to Bob, Cora and Dan.
	return d.DialAndSend(m)
}

// GetMessageId method
func (mailer *Mailer) GetMessageId() string {
	msgIds := mailer.m.GetHeader("Message-ID")
	if len(msgIds) > 0 {
		return msgIds[0]
	}
	return ""
}
