package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
	"github.com/gorilla/mux"
	"gobot.io/x/gobot"
	"gobot.io/x/gobot/drivers/gpio"
	"gobot.io/x/gobot/platforms/firmata"
)

type Body struct {
	Body uint8 `json:"Body"`
}

type Grip struct {
	Grip       uint8 `json:"Grip"`
	WristPitch uint8 `json:"WristPitch"`
	WristRoll  uint8 `json:"WristRoll"`
	Elbow      uint8 `json:"Elbow"`
	Shoulder   uint8 `json:"Shoulder"`
	Waist      uint8 `json:"Waist"`
}

var firmataAdaptor = firmata.NewAdaptor("/dev/cu.usbmodem14201")
var servo1 = gpio.NewServoDriver(firmataAdaptor, "10")
var servo2 = gpio.NewServoDriver(firmataAdaptor, "9")
var servo3 = gpio.NewServoDriver(firmataAdaptor, "8")
var servo4 = gpio.NewServoDriver(firmataAdaptor, "7")
var servo5 = gpio.NewServoDriver(firmataAdaptor, "6")
var servo6 = gpio.NewServoDriver(firmataAdaptor, "5")

var runNum int = 0

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}

func handleRequest() {
	myRouter := mux.NewRouter().StrictSlash(true)

	// myRouter.HandleFunc("/")
	myRouter.HandleFunc("/bluetooth", handleBluetooth)
	myRouter.HandleFunc("/grip", handleGrip)
	myRouter.HandleFunc("/wristPitch", handleWristPitch)
	myRouter.HandleFunc("/wristRoll", handleWristRoll)
	myRouter.HandleFunc("/elbow", handleElbow)
	myRouter.HandleFunc("/shoulder", handleShoulder)
	myRouter.HandleFunc("/waist", handleWaist)
	// myRouter.HandleFunc("/speed", handleSpeed)
	// myRouter.HandleFunc("/save", handleSave)
	// myRouter.HandleFunc("/run", handleRun)
	// myRouter.HandleFunc("/reset", handleReset)

	log.Fatal(http.ListenAndServe(":3030", myRouter))
}

func main() {
	handleRequest()
}

func handleBluetooth(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	var g Grip
	err := json.NewDecoder(r.Body).Decode(&g)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if g.Grip != 69 {

		fmt.Println(g.Grip, g.WristPitch, g.WristRoll, g.Elbow, g.Shoulder, g.Waist)

		led := gpio.NewLedDriver(firmataAdaptor, "30")

		work := func() {

			gobot.Every(1*time.Second, func() {
				led.Toggle()
				servo1.Move(g.Grip)
				servo2.Move(g.WristPitch)
				servo3.Move(g.WristRoll)
				servo4.Move(g.Elbow)
				servo5.Move(g.Shoulder)
				servo6.Move(g.Waist)
			})

		}

		robot := gobot.NewRobot("servoBot",
			[]gobot.Connection{firmataAdaptor},
			[]gobot.Device{servo1, servo2, servo3, servo4, servo5, servo6, led},
			
			work,
		)
	
		robot.Start(false)

	}
}


func handleGrip(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	var b Body
	
	err := json.NewDecoder(r.Body).Decode(&b)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Println(b.Body)
	
	work := func() {
		gobot.Every(1*time.Second, func() {
				servo1.Move(b.Body)
		})
	}

		robot := gobot.NewRobot("servoBot", []gobot.Connection{firmataAdaptor}, []gobot.Device{servo1}, work)	
		robot.Start(false)
		 

}

func handleWristPitch(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	var b Body
	err := json.NewDecoder(r.Body).Decode(&b)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Println(b.Body)

	work := func() {
		gobot.Every(1*time.Second, func() {
			servo2.Move(b.Body)
		})
	}

	robot := gobot.NewRobot("servoBot",
		[]gobot.Connection{firmataAdaptor},
		[]gobot.Device{servo2},
		work,
	)

	robot.Start(false)

}

func handleWristRoll(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	var b Body
	err := json.NewDecoder(r.Body).Decode(&b)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Println(b.Body)

	work := func() {
		gobot.Every(1*time.Second, func() {
			servo3.Move(b.Body)
		})
	}

	robot := gobot.NewRobot("servoBot",
		[]gobot.Connection{firmataAdaptor},
		[]gobot.Device{servo3},
		work,
	)

	robot.Start(false)

}

func handleElbow(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	var b Body
	err := json.NewDecoder(r.Body).Decode(&b)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Println(b.Body)

	work := func() {
		gobot.Every(1*time.Second, func() {
			servo4.Move(b.Body)
		})
	}

	robot := gobot.NewRobot("servoBot",
		[]gobot.Connection{firmataAdaptor},
		[]gobot.Device{servo4},
		work,
	)

	robot.Start(false)

}

func handleShoulder(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	var b Body
	err := json.NewDecoder(r.Body).Decode(&b)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Println(b.Body)

	work := func() {
		gobot.Every(1*time.Second, func() {
			servo5.Move(b.Body)
		})
	}

	robot := gobot.NewRobot("servoBot",
		[]gobot.Connection{firmataAdaptor},
		[]gobot.Device{servo5},
		work,
	)

	robot.Start(false)

}

func handleWaist(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	var b Body
	err := json.NewDecoder(r.Body).Decode(&b)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Println(b.Body)

	work := func() {
		gobot.Every(1*time.Second, func() {
			servo6.Move(b.Body)
		})
	}

	robot := gobot.NewRobot("servoBot",
		[]gobot.Connection{firmataAdaptor},
		[]gobot.Device{servo6},
		work,
	)

	robot.Start(false)

}
