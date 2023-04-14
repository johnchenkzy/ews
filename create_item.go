package ews

import (
	"encoding/base64"
	"encoding/xml"
	"errors"
	"io"
	"log"
	"mime"
	"os"
	"path/filepath"
	"time"
)

type InlineAttachment map[string]string

type CreateItem struct {
	XMLName                struct{}           `xml:"m:CreateItem"`
	MessageDisposition     string             `xml:"MessageDisposition,attr"`
	SendMeetingInvitations string             `xml:"SendMeetingInvitations,attr,omitempty"`
	SavedItemFolderId      *SavedItemFolderId `xml:"m:SavedItemFolderId,omitempty"`
	Items                  Items              `xml:"m:Items"`
}

type Items struct {
	Message      []Message      `xml:"t:Message"`
	CalendarItem []CalendarItem `xml:"t:CalendarItem"`
}

type SavedItemFolderId struct {
	DistinguishedFolderId DistinguishedFolderId `xml:"t:DistinguishedFolderId"`
}

type Message struct {
	ItemClass     string       `xml:"t:ItemClass,omitempty"`
	ItemId        *ItemId      `xml:"t:ItemId,omitempty"`
	Subject       string       `xml:"t:Subject"`
	Body          Body         `xml:"t:Body"`
	Attachments   *Attachments `xml:"t:Attachments,omitempty"`
	Sender        *OneMailbox  `xml:"t:Sender,omitempty"`
	ToRecipients  *XMailbox    `xml:"t:ToRecipients"`
	CcRecipients  *XMailbox    `xml:"t:CcRecipients,omitempty"`
	BccRecipients *XMailbox    `xml:"t:BccRecipients,omitempty"`
}

type Attachments struct {
	FileAttachment []FileAttachment `xml:"t:FileAttachment"`
}

type FileAttachment struct {
	Content        string `xml:"t:Content"`
	Name           string `xml:"t:Name"`
	ContentType    string `xml:"t:ContentType,omitempty"`
	IsInline       bool   `xml:"t:IsInline"`
	ContentId      string `xml:"t:ContentId,omitempty"`
	IsContactPhoto bool   `xml:"t:IsContactPhoto"`
}

type CalendarItem struct {
	Subject                    string      `xml:"t:Subject"`
	Body                       Body        `xml:"t:Body"`
	ReminderIsSet              bool        `xml:"t:ReminderIsSet"`
	ReminderMinutesBeforeStart int         `xml:"t:ReminderMinutesBeforeStart"`
	Start                      time.Time   `xml:"t:Start"`
	End                        time.Time   `xml:"t:End"`
	IsAllDayEvent              bool        `xml:"t:IsAllDayEvent"`
	LegacyFreeBusyStatus       string      `xml:"t:LegacyFreeBusyStatus"`
	Location                   string      `xml:"t:Location"`
	RequiredAttendees          []Attendees `xml:"t:RequiredAttendees"`
	OptionalAttendees          []Attendees `xml:"t:OptionalAttendees"`
	Resources                  []Attendees `xml:"t:Resources"`
}

type Body struct {
	BodyType string `xml:"BodyType,attr"`
	Body     []byte `xml:",chardata"`
}

type OneMailbox struct {
	Mailbox Mailbox `xml:"t:Mailbox"`
}

type XMailbox struct {
	Mailbox []Mailbox `xml:"t:Mailbox"`
}

type Mailbox struct {
	EmailAddress string `xml:"t:EmailAddress"`
}

type Attendee struct {
	Mailbox Mailbox `xml:"t:Mailbox"`
}

type Attendees struct {
	Attendee []Attendee `xml:"t:Attendee"`
}

type createItemResponseBodyEnvelop struct {
	XMLName struct{}               `xml:"Envelope"`
	Body    createItemResponseBody `xml:"Body"`
}
type createItemResponseBody struct {
	CreateItemResponse CreateItemResponse `xml:"CreateItemResponse"`
}

type CreateItemResponse struct {
	ResponseMessages ResponseMessages `xml:"ResponseMessages"`
}

type ResponseMessages struct {
	CreateItemResponseMessage Response `xml:"CreateItemResponseMessage"`
}

// Create Attachments By Paths
func CreateAttachmentsByPaths(paths ...string) *Attachments {
	var attachments Attachments
	for _, path := range paths {
		attachments.FileAttachment = append(attachments.FileAttachment, CreateFileAttachmentByPath(path))
	}
	return &attachments
}

func CreateInlineAttachments(attachmentInfos map[string]string) *Attachments {
	var attachments Attachments
	for name, path := range attachmentInfos {
		attachments.FileAttachment = append(attachments.FileAttachment, createFileAttachment(name, path, true))
	}
	return &attachments
}

// Create FileAttachment By Name and Path and Inline
func createFileAttachment(name string, path string, inline bool) FileAttachment {
	f, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	b, err2 := io.ReadAll(f)
	if err2 != nil {
		log.Fatal(err2)
	}
	if name == "" {
		_, name = filepath.Split(path)
	}
	ext := filepath.Ext(path)

	mimeType := mime.TypeByExtension(ext)
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	mimeType = mime.FormatMediaType(mimeType, map[string]string{"charset": "UTF-8"})
	fileAttachment := FileAttachment{
		Content:        base64.StdEncoding.EncodeToString(b),
		Name:           name,
		ContentType:    mimeType,
		IsInline:       inline,
		IsContactPhoto: false,
	}
	if inline {
		fileAttachment.ContentId = name
	}

	return fileAttachment
}

func CreateFileAttachmentByPath(path string) FileAttachment {
	return createFileAttachment("", path, false)
}

func createMessageItemWithAttachment(c Client, m ...Message) error {
	// 1 - Save the message without attachments
	item := &CreateItem{
		MessageDisposition: "SaveOnly",
	}
	item.Items.Message = append(item.Items.Message, m...)
	attachmentList := make([]FileAttachment, 0)
	for i := range item.Items.Message {
		attachmentList = append(attachmentList, item.Items.Message[i].Attachments.FileAttachment...)
		item.Items.Message[i].Attachments = nil
		item.Items.Message[i].ItemClass = ""
	}
	xmlBytes, err := xml.MarshalIndent(item, "", "  ")
	if err != nil {
		return err
	}
	bb, err := c.SendAndReceive(xmlBytes)
	if err != nil {
		return err
	}
	resp, err := checkCreateItemResponse(bb)
	if err != nil {
		return err
	}

	// 2 - Save the attachments
	if len(*resp.Items.Message) <= 0 {
		return errors.New("do not have ParentId")
	}
	attachments := &Attachments{
		FileAttachment: attachmentList,
	}
	messages := *resp.Items.Message
	if len(messages) > 0 {
		messages[0].ItemId.ChangeKey = ""
	}

	attachmentResp, err := SaveCreateAttachment(c, attachments, messages[0].ItemId)
	if err != nil {
		return err
	}

	// 3 - Send the mail
	if len(*attachmentResp) > 0 {
		messages[0].ItemId.ChangeKey = (*attachmentResp)[0].FileAttachmentId.RootItemChangeKey
	}
	err = SendSavedItem(c, &ItemIds{[]ItemId{*messages[0].ItemId}})
	if err != nil {
		return err
	}

	return nil
}

func createMessageItem(c Client, m ...Message) error {
	item := &CreateItem{
		MessageDisposition: "SendAndSaveCopy",
		SavedItemFolderId:  &SavedItemFolderId{DistinguishedFolderId{Id: "sentitems"}},
	}
	item.Items.Message = append(item.Items.Message, m...)

	xmlBytes, err := xml.MarshalIndent(item, "", "  ")
	if err != nil {
		return err
	}

	bb, err := c.SendAndReceive(xmlBytes)
	if err != nil {
		return err
	}

	if _, err := checkCreateItemResponse(bb); err != nil {
		return err
	}

	return nil
}

// CreateMessageItem
// https://docs.microsoft.com/en-us/exchange/client-developer/web-service-reference/createitem-operation-email-message
func CreateMessageItem(c Client, m ...Message) error {
	for i := range m {
		if m[i].Attachments != nil && len(m[i].Attachments.FileAttachment) > 0 {
			return createMessageItemWithAttachment(c, m...)
		}
	}

	return createMessageItem(c, m...)
}

// CreateCalendarItem
// https://docs.microsoft.com/en-us/exchange/client-developer/web-service-reference/createitem-operation-calendar-item
func CreateCalendarItem(c Client, ci ...CalendarItem) error {

	item := &CreateItem{
		SendMeetingInvitations: "SendToAllAndSaveCopy",
		SavedItemFolderId:      &SavedItemFolderId{DistinguishedFolderId{Id: "calendar"}},
	}
	item.Items.CalendarItem = append(item.Items.CalendarItem, ci...)

	xmlBytes, err := xml.MarshalIndent(item, "", "  ")
	if err != nil {
		return err
	}

	bb, err := c.SendAndReceive(xmlBytes)
	if err != nil {
		return err
	}

	if _, err := checkCreateItemResponse(bb); err != nil {
		return err
	}

	return nil
}

func checkCreateItemResponse(bb []byte) (*Response, error) {
	var soapResp createItemResponseBodyEnvelop
	if err := xml.Unmarshal(bb, &soapResp); err != nil {
		return nil, err
	}
	resp := soapResp.Body.CreateItemResponse.ResponseMessages.CreateItemResponseMessage
	if resp.ResponseClass == ResponseClassError {
		return nil, errors.New(resp.MessageText)
	}
	return &resp, nil
}
