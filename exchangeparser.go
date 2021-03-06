package xchango

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"github.com/sgoertzen/html2text"
	"log"
	"time"
)

type Organizer struct {
	Mailbox Mailbox
}

type Mailbox struct {
	Name string
}

type ItemId struct {
	Id        string `xml:"Id,attr"`
	ChangeKey string `xml:"ChangeKey,attr"`
}

type CalendarItem struct {
	ItemId         ItemId
	Subject        string
	DisplayCc      string
	DisplayTo      string
	Start          string
	End            string
	IsAllDayEvent  bool
	Location       string
	MyResponseType string
	Organizer      Organizer
	Body           Body
}

type Body struct {
	BodyType string `xml:"BodyType,attr"`
	Body     string `xml:",chardata"`
}

type Appointment struct {
	ItemId         string
	ChangeKey      string
	Subject        string
	Cc             string
	To             string
	Start          time.Time
	End            time.Time
	IsAllDayEvent  bool
	Location       string
	MyResponseType string
	Organizer      string
	Body           string
	BodyType       string
}

func parseCalendarFolder(soap string) ItemId {
	decoder := xml.NewDecoder(bytes.NewBufferString(soap))
	var itemId ItemId

	for {
		// Read tokens from the XML document in a stream.
		t, _ := decoder.Token()
		if t == nil {
			break
		}
		switch se := t.(type) {
		case xml.StartElement:
			if se.Name.Local == "FolderId" {
				decoder.DecodeElement(&itemId, &se)
				break
			}
		}
	}
	return itemId
}

func parseAppointments(soap string) []Appointment {
	decoder := xml.NewDecoder(bytes.NewBufferString(soap))

	appointments := make([]Appointment, 0)

	for {
		// Read tokens from the XML document in a stream.
		t, _ := decoder.Token()
		if t == nil {
			break
		}
		switch se := t.(type) {
		case xml.StartElement:
			if se.Name.Local == "CalendarItem" {
				var item CalendarItem
				decoder.DecodeElement(&item, &se)
				appointments = append(appointments, item.ToAppointment())
			}
		}
	}
	return appointments
}

func (c CalendarItem) ToAppointment() Appointment {
	app := Appointment{
		ItemId:         c.ItemId.Id,
		ChangeKey:      c.ItemId.ChangeKey,
		Subject:        c.Subject,
		Cc:             c.DisplayCc,
		To:             c.DisplayTo,
		IsAllDayEvent:  c.IsAllDayEvent,
		Location:       c.Location,
		MyResponseType: c.MyResponseType,
		Organizer:      c.Organizer.Mailbox.Name,
		Body:           c.Body.Body,
		BodyType:       c.Body.BodyType,
	}
	if len(c.Start) > 0 {
		t1, err := time.Parse(time.RFC3339, c.Start)
		if err != nil {
			log.Printf("Error while parsing time.  Start time string wass: %v", c.Start)
			log.Println(err)
		}
		app.Start = t1
	}

	if len(c.End) > 0 {
		t1, err := time.Parse(time.RFC3339, c.End)
		if err != nil {
			log.Printf("Error while parsing time.  End time string wass: %v", c.End)
			log.Println(err)
		}
		app.End = t1
	}
	return app
}

func (a Appointment) String() string {
	return fmt.Sprintf("%s starting %d", a.Subject, a.Start)
}

func (a *Appointment) BuildDesc() string {
	desc := ""

	addField := func(field string, label string) {
		if len(field) > 0 {
			desc += label + " " + field + "\n"
		}
	}
	addField(a.Organizer, "Organizer:")
	addField(a.To, "To:")
	addField(a.Cc, "Cc:")
	addField(a.MyResponseType, "Response:")
	body, err := html2text.Textify(a.Body)
	if err != nil {
		log.Println("Error while building text description.")
		log.Println(a.Body)
		log.Println(err)
		body = "Unable to create text from appointment description."
	}
	desc += body
	return desc
}
