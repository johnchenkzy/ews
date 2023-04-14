package ewsutil

import (
	"fmt"

	"github.com/johnchenkzy/ews"
)

type Email struct {
	To      []string
	Subject string
	Body    string
	BodyType string
	Attachments []string
	InlineAttachments InlineAttachment
	Cc []string
	Bcc []string
}

/// key is the cid, value is the path of the attachment
type InlineAttachment map[string]string

const (
	BodyTypeText = "Text"
	BodyTypeHTML = "HTML"
)

// SendEmail helper method to send Message
func SendEmail(c ews.Client, to []string, subject, body string) error {
	fmt.Println("send email")
	m := ews.Message{
		ItemClass: "IPM.Note",
		Subject:   subject,
		Body: ews.Body{
			BodyType: "Text",
			Body:     []byte(body),
		},
		Sender: &ews.OneMailbox{
			Mailbox: ews.Mailbox{
				EmailAddress: c.GetUsername(),
			},
		},
	}
	mb := make([]ews.Mailbox, len(to))
	for i, addr := range to {
		mb[i].EmailAddress = addr
	}
	m.ToRecipients.Mailbox = append(m.ToRecipients.Mailbox, mb...)

	return ews.CreateMessageItem(c, m)
}

func SendEmails(c ews.Client, email Email) error {
	m := ews.Message{
		ItemClass: "IPM.Note",
		Subject: email.Subject,
		Body: ews.Body{
			BodyType: email.BodyType,
			Body:     []byte(email.Body),
		},
		Sender: &ews.OneMailbox{
			Mailbox: ews.Mailbox{
				EmailAddress: c.GetUsername(),
			},
		},
	}
	// deal ToRecipients
	mb := make([]ews.Mailbox, len(email.To))
	for i, addr := range email.To {
		mb[i].EmailAddress = addr
	}
	if len(mb) > 0 {
		m.ToRecipients = &ews.XMailbox{};
		m.ToRecipients.Mailbox = append(m.ToRecipients.Mailbox, mb...)
	}

	// deal CcRecipients
	mb = make([]ews.Mailbox, len(email.Cc))
	for i, addr := range email.Cc {
		mb[i].EmailAddress = addr
	}
	if len(mb)>0 {
		m.CcRecipients = &ews.XMailbox{};
		m.CcRecipients.Mailbox = append(m.CcRecipients.Mailbox, mb...)
	}

	// deal BccRecipients
	mb = make([]ews.Mailbox, len(email.Bcc))
	for i, addr := range email.Bcc {
		mb[i].EmailAddress = addr
	}

	
	if len(mb)>0 {
		m.BccRecipients = &ews.XMailbox{};
		m.BccRecipients.Mailbox = append(m.BccRecipients.Mailbox, mb...)
	}

	// deal file Attachments
	if len(email.Attachments) > 0 {
		m.Attachments = ews.CreateAttachmentsByPaths(email.Attachments...)
	}
	// deal inline attachments
	if len(email.InlineAttachments) > 0 {
		if m.Attachments != nil {
			newAttachments := ews.CreateInlineAttachments(email.InlineAttachments)
			m.Attachments.FileAttachment = append(m.Attachments.FileAttachment, newAttachments.FileAttachment...)
		} else {
			m.Attachments = ews.CreateInlineAttachments(email.InlineAttachments)
		}
	}

	return ews.CreateMessageItem(c, m)
}

// send email with attachment
func SendEmailWithFileAttachment(c ews.Client, to []string, subject, body string, attachmentPaths []string) error {
	m := ews.Message{
		ItemClass: "IPM.Note",
		Subject:   subject,
		Body: ews.Body{
			BodyType: "Text",
			Body:     []byte(body),
		},
		Sender: &ews.OneMailbox{
			Mailbox: ews.Mailbox{
				EmailAddress: c.GetUsername(),
			},
		},
	}
	mb := make([]ews.Mailbox, len(to))
	for i, addr := range to {
		mb[i].EmailAddress = addr
	}
	m.ToRecipients.Mailbox = append(m.ToRecipients.Mailbox, mb...)
	if len(attachmentPaths) > 0 {
		m.Attachments = ews.CreateAttachmentsByPaths(attachmentPaths...)
	}

	return ews.CreateMessageItem(c, m)
}