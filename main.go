package main

import (
	"bytes"
	"fmt"
	"github.com/emersion/go-ical"
	"github.com/hduLib/hdu/skl"
	"github.com/teambition/rrule-go"
	"log"
	"os"
	"strconv"
	"time"
)

var hduTimeList = []time.Duration{
	time.Hour*8 + time.Minute*5,
	time.Hour*8 + time.Minute*55,
	time.Hour*10 + time.Minute*0,
	time.Hour*10 + time.Minute*50,
	time.Hour*11 + time.Minute*40,
	time.Hour*13 + time.Minute*30,
	time.Hour*14 + time.Minute*20,
	time.Hour*15 + time.Minute*15,
	time.Hour*16 + time.Minute*5,
	time.Hour*18 + time.Minute*30,
	time.Hour*19 + time.Minute*20,
	time.Hour*20 + time.Minute*10,
}

var weekday = []rrule.Weekday{rrule.SU, rrule.MO, rrule.TU, rrule.WE, rrule.TH, rrule.FR, rrule.SA, rrule.SU}

func main() {
	var account, passwd string
	fmt.Println("智慧杭电账号:")
	if _, err := fmt.Scan(&account); err != nil {
		log.Fatalln(err)
		return
	}
	fmt.Println("智慧杭电密码:")
	if _, err := fmt.Scan(&passwd); err != nil {
		log.Fatalln(err)
		return
	}

	user, err := skl.Login(account, passwd)
	if err != nil {
		log.Fatalln(err)
		return
	}
	startDay := time.Now()
	startDay = startDay.Add(-time.Duration(int(time.Hour)*startDay.Hour() + int(time.Minute)*startDay.Minute() + int(time.Second)*startDay.Second() + startDay.Nanosecond()))
	resp, err := user.Course(startDay)
	if err != nil {
		log.Fatalln(err)
		return
	}
	startDay = startDay.AddDate(0, 0, -((resp.Week-1)*7 + (int(startDay.Weekday())+6)%7))
	ics := ical.NewCalendar()
	ics.Props.SetText(ical.PropVersion, "2.0")
	ics.Props.SetText(ical.PropProductID, "-//student//hduLib//CN")
	ics.Props.SetText(ical.PropCalendarScale, "GREGORIAN")

	alarm := ical.NewComponent(ical.CompAlarm)
	alarm.Props.SetText(ical.PropAction, ical.ParamDisplay)
	alarm.Props.SetText(ical.PropDescription, "距离上课还有30min")
	alarm.Props.SetText(ical.PropTrigger, "-P0DT0H30M0S")

	for _, v := range resp.List {
		event := ical.NewEvent()
		event.Props.SetDateTime(ical.PropDateTimeStart, startDay.AddDate(0, 0, (v.StartWeek-1)*7+v.WeekDay-1).Add(hduTimeList[v.StartSection-1]))
		event.Props.SetDateTime(ical.PropDateTimeEnd, startDay.AddDate(0, 0, (v.StartWeek-1)*7+v.WeekDay-1).Add(hduTimeList[v.EndSection-1]).Add(time.Minute*45))
		event.Props.SetDateTime(ical.PropDateTimeStamp, startDay)
		event.Props.SetText(ical.PropSummary, fmt.Sprintf("%s %s %s %s", v.CourseName, v.ClassRoom, v.TeacherName, v.CourseSchema))
		event.Props.SetText(ical.PropUID, v.CourseCode+strconv.Itoa(v.WeekDay)+strconv.Itoa(v.StartWeek))
		count := v.EndWeek - v.StartWeek + 1
		interval := 1
		if v.Period != "" {
			count = (count + 1) / 2
			interval = 2
		}
		if err != nil {
			log.Fatalln(err)
			return
		}

		event.Props.SetRecurrenceRule(&rrule.ROption{
			Freq:      rrule.WEEKLY,
			Interval:  interval,
			Wkst:      rrule.SU,
			Count:     count,
			Byweekday: []rrule.Weekday{weekday[v.WeekDay]},
		})

		event.Component.Children = append(event.Component.Children, alarm)

		ics.Children = append(ics.Children, event.Component)
	}

	buf := new(bytes.Buffer)
	encoder := ical.NewEncoder(buf)
	if err := encoder.Encode(ics); err != nil {
		log.Fatalln(err)
		return
	}
	// yysy我不理解, 可能是库的问题
	out := bytes.ReplaceAll(buf.Bytes(), []byte(";TZID=Local"), []byte(""))
	if err := os.WriteFile("./course.ics", out, os.ModePerm); err != nil {
		log.Fatalln(err)
		return
	}

}
