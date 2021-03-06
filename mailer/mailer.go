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
	d       *gomail.Dialer
	s       gomail.SendCloser
	m       *gomail.Message
	enc     gomail.Encoding
	charset string
	opt     MailOption
}

// Attachment struct
type Attachment struct {
	Name string `json:"name,omitempty" bson:"name,omitempty"`
	Type string `json:"type,omitempty" bson:"type,omitempty"`
	Data string `json:"data,omitempty" bson:"data,omitempty"`
	Url  string `json:"url,omitempty" bson:"url,omitempty"`
	Path string `json:"path,omitempty" bson:"path,omitempty"`
}

// Embed struct
type Embed struct {
	Name string `json:"name,omitempty" bson:"name,omitempty"`
	Type string `json:"type,omitempty" bson:"type,omitempty"`
	Data string `json:"data,omitempty" bson:"data,omitempty"`
	Url  string `json:"url,omitempty" bson:"url,omitempty"`
	Path string `json:"path,omitempty" bson:"path,omitempty"`
}

const (
	// QuotedPrintable represents the quoted-printable encoding as defined in
	// RFC 2045.
	QuotedPrintable gomail.Encoding = "quoted-printable"
	// Base64 represents the base64 encoding as defined in RFC 2045.
	Base64 gomail.Encoding = "base64"
	// Unencoded can be used to avoid encoding the body of an email. The headers
	// will still be encoded using quoted-printable encoding.
	Unencoded gomail.Encoding = "8bit"

	// DefaultCharset email
	DefaultCharset string = "iso-8859-1"
)

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
		d:   d,
		s:   s,
		m:   nil,
		opt: mailOpt,
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
		d:   d,
		s:   s,
		m:   nil,
		opt: mailOpt,
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

func (mailer *Mailer) SetEncoding(enc gomail.Encoding) *Mailer {
	mailer.enc = enc
	return mailer
}

func (mailer *Mailer) SetCharset(charset string) *Mailer {
	mailer.charset = charset
	return mailer
}

func (mailer *Mailer) getMessageSettings() []gomail.MessageSetting {
	var setting []gomail.MessageSetting
	if mailer.enc != "" {
		setting = append(setting, gomail.SetEncoding(mailer.enc))
	} else {
		setting = append(setting, gomail.SetEncoding(Base64))
	}
	if mailer.charset != "" {
		setting = append(setting, gomail.SetCharset(mailer.charset))
	} else {
		setting = append(setting, gomail.SetCharset(DefaultCharset))
	}
	return setting
}

// Send method
func (mailer *Mailer) Send(to string, cc string, subject string, contentType string, content string, attachment ...string) error {
	m := gomail.NewMessage(mailer.getMessageSettings()...)
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
	m := gomail.NewMessage(mailer.getMessageSettings()...)
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

	mailer.m = m

	// set attachments
	if err := mailer.setAttachments(attachments...); err != nil {
		log.Println(err)
	}

	return gomail.Send(mailer.s, mailer.m)
}

// SendWithAttachments method
func (mailer *Mailer) SendWithAttachmentsEmbeds(to string, cc string, subject string, contentType string, content string, attachments []Attachment, embeds []Embed) error {
	m := gomail.NewMessage(mailer.getMessageSettings()...)
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

	mailer.m = m

	// set attachments
	if err := mailer.setAttachments(attachments...); err != nil {
		log.Println(err)
	}
	// set embeds
	if err := mailer.setEmbeds(embeds...); err != nil {
		log.Println(err)
	}

	return gomail.Send(mailer.s, mailer.m)
}

// setAttachments method
func (mailer *Mailer) setAttachments(attachments ...Attachment) error {
	for _, attachment := range attachments {
		if attachment.Data != "" {
			var f = func(a Attachment) func(w io.Writer) error {
				return func(w io.Writer) error {
					var rawImage string
					coI := strings.Index(a.Data, ",")
					if coI < 0 {
						rawImage = a.Data
					} else {
						rawImage = a.Data[coI+1:]
					}
					unBased, err := base64.StdEncoding.DecodeString(rawImage)
					if err != nil {
						return err
					}
					res := bytes.NewReader(unBased)
					if _, err := io.Copy(w, res); err != nil {
						return err
					}
					return nil
				}
			}
			originAttachmentName := attachment.Name
			if attachment.Name == "" {
				attachment.Name = "attachment"
				originAttachmentName = "attachment"
			}
			attachment.Name = strings.TrimSuffix(attachment.Name, filepath.Ext(attachment.Name))
			var contentType string
			coI := strings.Index(attachment.Data, ",")
			if coI < 0 {
				if attachment.Type != "" {
					contentType = attachment.Type
				} else if t := mime.TypeByExtension(filepath.Ext(originAttachmentName)); t != "" {
					contentType = t
				} else {
					log.Printf("unknown Content-Type of attachment %s\n", originAttachmentName)
					mailer.m.Attach(originAttachmentName, gomail.SetCopyFunc(f(attachment)))
					continue
				}
			} else {
				contentType = strings.TrimSuffix(attachment.Data[5:coI], ";base64")
			}

			ext, err := mime.ExtensionsByType(contentType)
			if err != nil {
				log.Println(err)
				continue
			}
			if len(ext) > 0 {
				attachment.Name = attachment.Name + ext[len(ext)-1]
			}

			mailer.m.Attach(attachment.Name, gomail.SetCopyFunc(f(attachment)))
		} else if attachment.Url != "" {
			var f = func(r io.ReadCloser) func(w io.Writer) error {
				return func(w io.Writer) error {
					defer r.Close()
					if _, err := io.Copy(w, r); err != nil {
						return err
					}
					return nil
				}
			}
			cl := client.LoadHttp(attachment.Url)
			cl.SetRequest("GET", "", nil)
			resp, err := cl.Start()
			if err != nil {
				log.Println(err)
				continue
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
				continue
			}
			if len(ext) > 0 {
				cExt := strings.ToLower(ext[len(ext)-1])
				if extName := strings.ToLower(filepath.Ext(name)); extName != "" && extName == cExt {
					name = strings.TrimSuffix(name, filepath.Ext(name)) + cExt
				} else {
					name = name + cExt
				}
			}

			mailer.m.Attach(name, gomail.SetCopyFunc(f(resp.Body)))
		} else if attachment.Path != "" {
			mailer.m.Attach(attachment.Path)
		}
	}

	return nil
}

// setEmbeds method
func (mailer *Mailer) setEmbeds(embeds ...Embed) error {
	for _, embed := range embeds {
		if embed.Data != "" {
			var f = func(a Embed) func(w io.Writer) error {
				return func(w io.Writer) error {
					var rawImage string
					coI := strings.Index(a.Data, ",")
					if coI < 0 {
						rawImage = a.Data
					} else {
						rawImage = a.Data[coI+1:]
					}
					unBased, err := base64.StdEncoding.DecodeString(rawImage)
					if err != nil {
						return err
					}
					res := bytes.NewReader(unBased)
					if _, err := io.Copy(w, res); err != nil {
						return err
					}
					return nil
				}
			}
			if embed.Name == "" {
				log.Println("error embed, required name for:", embed.Data)
				continue
			}
			/*
				embed.Name = strings.TrimSuffix(embed.Name, filepath.Ext(embed.Name))
				coI := strings.Index(embed.Data, ",")
				contentType := strings.TrimSuffix(embed.Data[5:coI], ";base64")
				ext, err := mime.ExtensionsByType(contentType)
				if err != nil {
					log.Println(err)
					continue
				}
				if len(ext) > 0 {
					embed.Name = embed.Name + ext[len(ext)-1]
				}
				embeds[i].Name = embed.Name
			*/
			mailer.m.Embed(embed.Name, gomail.SetCopyFunc(f(embed)))
		} else if embed.Url != "" {
			var f = func(r io.ReadCloser) func(w io.Writer) error {
				return func(w io.Writer) error {
					defer r.Close()
					if _, err := io.Copy(w, r); err != nil {
						return err
					}
					return nil
				}
			}
			if embed.Name == "" {
				log.Println("error embed, required name for:", embed.Url)
				continue
			}

			cl := client.LoadHttp(embed.Url)
			cl.SetRequest("GET", "", nil)
			resp, err := cl.Start()
			if err != nil {
				log.Println(err)
				continue
			}
			/*
				ext, err := mime.ExtensionsByType(resp.Header.Get("Content-Type"))
				if err != nil {
					log.Println(err)
					continue
				}
				if len(ext) > 0 {
					cExt := strings.ToLower(ext[len(ext)-1])
					if extName := strings.ToLower(filepath.Ext(embed.Name)); extName != "" && extName == cExt {
						embed.Name = strings.TrimSuffix(embed.Name, filepath.Ext(embed.Name)) + cExt
					} else {
						embed.Name = embed.Name + cExt
					}
				}
				embeds[i].Name = embed.Name
			*/
			mailer.m.Embed(embed.Name, gomail.SetCopyFunc(f(resp.Body)))
		} else if embed.Path != "" {
			mailer.m.Embed(embed.Path)
		}
	}

	return nil
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
