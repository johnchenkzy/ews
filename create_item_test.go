package ews

import (
	"encoding/xml"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_marshal_Message_Attachments(t *testing.T) {
	mails := make([]Mailbox, 0)
	mails = append(mails,
		Mailbox{EmailAddress: "sadie@contoso.com"},
		Mailbox{EmailAddress: "ronnie@contoso.com"},
	)

	attachments := CreateAttachmentsByPaths("C:\\Users\\JohnChen\\Desktop\\需求分析说明书.md", "./FileAttachment2.txt", "FileAttachment3.txt")

	msg := &Message{
		Subject: "Meeting Cancellation",
		Body: Body{
			BodyType: "Text",
			Body:     []byte("The meeting scheduled for tomorrow has been canceled."),
		},
		Attachments:  attachments,
		ToRecipients: &XMailbox{
			mails,
		},
	};
	xmlbytes, err := xml.MarshalIndent(msg, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(xmlbytes))
}

func Test_marshal_CalendarItem(t *testing.T) {

	attendee := make([]Attendee, 0)
	attendee = append(attendee,
		Attendee{Mailbox: Mailbox{EmailAddress: "User1@example.com"}},
		Attendee{Mailbox: Mailbox{EmailAddress: "User2@example.com"}},
	)
	attendees := make([]Attendees, 0)
	attendees = append(attendees, Attendees{Attendee: attendee})

	start, _ := time.Parse(time.RFC3339, "2006-11-02T14:00:00Z")
	end, _ := time.Parse(time.RFC3339, "2006-11-02T15:00:00Z")

	citem := &CalendarItem{
		Subject: "Planning Meeting",
		Body: Body{
			BodyType: "Text",
			Body:     []byte("Plan the agenda for next week's meeting."),
		},
		ReminderIsSet:              true,
		ReminderMinutesBeforeStart: 60,
		Start:                      start,
		End:                        end,
		IsAllDayEvent:              false,
		LegacyFreeBusyStatus:       "Busy",
		Location:                   "Conference Room 721",
		RequiredAttendees:          attendees,
	}

	xmlBytes, err := xml.MarshalIndent(citem, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	assert.Equal(t, `<CalendarItem>
  <t:Subject>Planning Meeting</t:Subject>
  <t:Body BodyType="Text">Plan the agenda for next week&#39;s meeting.</t:Body>
  <t:ReminderIsSet>true</t:ReminderIsSet>
  <t:ReminderMinutesBeforeStart>60</t:ReminderMinutesBeforeStart>
  <t:Start>2006-11-02T14:00:00Z</t:Start>
  <t:End>2006-11-02T15:00:00Z</t:End>
  <t:IsAllDayEvent>false</t:IsAllDayEvent>
  <t:LegacyFreeBusyStatus>Busy</t:LegacyFreeBusyStatus>
  <t:Location>Conference Room 721</t:Location>
  <t:RequiredAttendees>
    <t:Attendee>
      <t:Mailbox>
        <t:EmailAddress>User1@example.com</t:EmailAddress>
      </t:Mailbox>
    </t:Attendee>
    <t:Attendee>
      <t:Mailbox>
        <t:EmailAddress>User2@example.com</t:EmailAddress>
      </t:Mailbox>
    </t:Attendee>
  </t:RequiredAttendees>
</CalendarItem>`, string(xmlBytes))
}
