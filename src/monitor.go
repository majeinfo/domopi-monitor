package main

import (
        "fmt"
	"flag"
	"log"
	"time"
	"os"
        "gopkg.in/mgo.v2"
        "gopkg.in/mgo.v2/bson"
	"gopkg.in/gomail.v2"
)

/*
type Person struct {
        Name string
        Phone string
}
*/

type User struct {
	Login string
	Password string
	Email string
	Address string
	Lat float64
	Lng float64
	Phonenu string
	Timezone string
}

type Controller struct {
	Description string
	DoVersion string
	Key string
	Zid string
	Login string
	LocalIP string
}

type Metric struct {
	ProbeTitle string
	ScaleTitle string
	IsLevelNumber bool
	Level int
	OnOff bool
	Change bool
}

type SensorData struct {
	Key string
	Zid string
	Devid string
	Instid string
	Sid string
	LastUpdate time.Time
	Description string
	DevType string
	Metrics Metric
}

type HistoryEvent struct {
	Key string
	Zid string
	EvtType string
	Updated time.Time
	Data SensorData
}

// Globals
var (
	smtpServer string
	smtpPort int
	mailFrom string
	mailSubject string
	maxHours int
	flag_no_alert bool = false
	flag_debug = false
)

func main() {
	_parse_cmdline()

	//fmt.Println("hello")
	session, err := mgo.Dial("localhost")
        if err != nil {
                panic(err)
        }
        defer session.Close()

        // Optional. Switch the session to a monotonic behavior.
        session.SetMode(mgo.Monotonic, true)

	/*
        c := session.DB("test").C("people")
        err = c.Insert(&Person{"Ale", "+55 53 8116 9639"},
	               &Person{"Cla", "+55 53 8402 8510"})
        if err != nil {
                log.Fatal(err)
        }

        result := Person{}
        err = c.Find(bson.M{"name": "Ale"}).One(&result)
        if err != nil {
                log.Fatal(err)
        }

        fmt.Println("Phone:", result.Phone)
	*/

	/*
	var sales_his []Sale
	err = c.Find(
    		bson.M{
        	"sale_date": bson.M{
            		"$gt": fromDate,
            		"$lt": toDate,
        	},
    	}).All(&sales_his)
	*/

	curtime := time.Now()
	fmt.Println(curtime)

	// Get access to Collections
	coll_users := session.DB("domopi").C("users")
	coll_controllers := session.DB("domopi").C("controllers")
	coll_histories := session.DB("domopi").C("histories")

	// For each User, we search for dead Controller
	var users []User
	err = coll_users.Find(bson.M{}).All(&users)
	if err != nil {
		log.Fatal(err)
	}
	//fmt.Println(users)
	for _, user := range users {

		fmt.Println("USER:", user)

		// For each Controller, find the last sensorevent et check its updated datetime
		// [db.histories.find({key:'a_key', evttype:'sensorevt'}).sort({updated:-1})]
		// Note: key missing in mongodb

		/* Loop on the Controllers: */
		var controllers []Controller
		err = coll_controllers.Find(bson.M{"login": user.Login}).All(&controllers)
		if err != nil {
			//log.Fatal(err)
			fmt.Println(err)
			continue
		}
		//fmt.Println(controllers)
	
		for _, cont := range controllers {
			fmt.Println("CONTROLLER:", "Key=", cont.Key, "Zid=", cont.Zid)
			var hist_evt HistoryEvent
			err = coll_histories.Find(
				bson.M{"key": cont.Key, "evttype": "sensorevt"}).Sort("-updated").One(&hist_evt)
			if err != nil {
				//log.Fatal("NO EVENT IN HISTORY", err)
				fmt.Println("NO EVENT IN HISTORY", err)
				continue
			}
			fmt.Println(hist_evt)

			updated := hist_evt.Updated
			//fmt.Println(updated)

			difftime := curtime.Sub(updated)
			//var delta int64
			//delta = int64(difftime)
			//fmt.Println(difftime, delta, maxHours*60*60*1000*1000*1000)

			if difftime > time.Duration(maxHours)*60*60*1000*1000*1000 {
				fmt.Println("TOO OLD VALUE")

				// Send mail to alert Customer
				if flag_no_alert == false {
					_send_alert(user, cont)
				}
			}
		}
	}
}

func _send_alert(user User, cont Controller) {
	m := gomail.NewMessage()
	m.SetHeader("From", mailFrom)
	m.SetHeader("To", user.Email)
	m.SetHeader("Subject", mailSubject)
	m.SetBody("text/plain", "Your Domopi Box " + cont.Key + fmt.Sprintf(" has not sent data since %dh", maxHours))

	d := gomail.Dialer{Host: smtpServer, Port: smtpPort}
	if err := d.DialAndSend(m); err != nil {
    		panic(err)
	}
}

func _parse_cmdline() {
	flag.IntVar(&maxHours, "hours", 24, "Maximum number of hours before alerting")
	flag.StringVar(&smtpServer, "smtp", "localhost", "Name or IP Address of SMTP Server (default: localhost)")
	flag.IntVar(&smtpPort, "port", 25, "SMTP Server Port (default: 25)")
	flag.StringVar(&mailFrom, "from", "domopi@delamarche.com", "Mail From value")
	flag.StringVar(&mailSubject, "subject", "No Data sent by your Domopi Box", "Mail Subject")
	flag.BoolVar(&flag_no_alert, "no-alert", false, "Do not send Alerts")
	flag.BoolVar(&flag_debug, "debug", true, "Set Debug Mode")
	help := flag.Bool( "help", false, "Display help")
	flag.Parse()
	if *help {
		flag.PrintDefaults()
		os.Exit(0)
	}
}

// EOF
