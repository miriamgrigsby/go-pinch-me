package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	// "gobot.io/x/gobot"
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

type NewBot struct {
	NewBot string `json:"NewBot"`
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

// var robot1 *gobot.Robot

var runNum int = 0

func handleRequest() {
	myRouter := gin.Default()

	myRouter.Use(cors.Default())

	myRouter.Any("/bluetoothOn", handleBluetoothOn)
	myRouter.Any("/bluetoothOff", handleBluetoothOff)
	myRouter.Any("/slider", handleSlider)
	myRouter.Any("/dragNdrop", handleDragNDrop)
	myRouter.Any("/duck", handleDuckDuck)
	myRouter.Any("/newBot", handleNewBot)

	myRouter.Run(":3030")
}

func main() {
	handleRequest()
}

func handleBluetoothOn(c *gin.Context) {
	var g Grip
	c.Bind(&g)
	firmataAdaptor.Connect()
	firmataAdaptor.DigitalWrite("30", 1)
	handleRun(c, g)
}

func handleBluetoothOff(c *gin.Context) {
	firmataAdaptor.DigitalWrite("30", 0)
}

func handleSlider(c *gin.Context) {
	var g Grip
	c.Bind(&g)
	fmt.Println(g)
	firmataAdaptor.ServoWrite("10", g.Grip)
	firmataAdaptor.ServoWrite("9", g.WristPitch)
	// firmataAdaptor.ServoWrite("8", g.WristRoll)
	firmataAdaptor.ServoWrite("7", g.Elbow)
	firmataAdaptor.ServoWrite("6", g.Shoulder)
	firmataAdaptor.ServoWrite("5", g.Waist)
}

func handleRun(c *gin.Context, g Grip) {
	firmataAdaptor.Connect()
	firmataAdaptor.ServoWrite("10", g.Grip)
	firmataAdaptor.ServoWrite("9", g.WristPitch)
	firmataAdaptor.ServoWrite("8", g.WristRoll)
	firmataAdaptor.ServoWrite("7", g.Elbow)
	firmataAdaptor.ServoWrite("6", g.Shoulder)
	firmataAdaptor.ServoWrite("5", g.Waist)
}

func handleNewBot(c *gin.Context) {
	var n NewBot
	c.Bind(&n)
	fmt.Println(n)
}

func handleDragNDrop(c *gin.Context) {
	var g Grip
	c.Bind(&g)
	g.WristRoll = 5
	handleRun(c, g)
	time.Sleep(1 * time.Second)

	// Movement down, claw open
	var wg1 sync.WaitGroup
	wg1.Add(4)
	go func(i uint8, wg1 *sync.WaitGroup) {
		defer wg1.Done()
		for a := g.Grip; a < i; a++ {
			firmataAdaptor.ServoWrite("10", a)
			time.Sleep(5 * time.Millisecond)
		}
	}(uint8(150), &wg1)
	time.Sleep(1 * time.Second)
	go func(i uint8, wg1 *sync.WaitGroup) {
		defer wg1.Done()
		for a := g.WristPitch; a < i; a++ {
			firmataAdaptor.ServoWrite("9", a)
			time.Sleep(7 * time.Millisecond)
		}
	}(uint8(50), &wg1)
	go func(i uint8, wg1 *sync.WaitGroup) {
		defer wg1.Done()
		for a := g.Elbow; a < i; a++ {
			firmataAdaptor.ServoWrite("7", a)
			time.Sleep(9 * time.Millisecond)
		}
	}(uint8(165), &wg1)
	go func(i uint8, wg1 *sync.WaitGroup) {
		defer wg1.Done()
		for a := g.Shoulder; a > i; a-- {
			firmataAdaptor.ServoWrite("6", a)
			time.Sleep(15 * time.Millisecond)
		}
	}(uint8(25), &wg1)
	wg1.Wait()

	// Close claw, movement up
	var wg2 sync.WaitGroup
	wg2.Add(3)
	go func(j uint8, wg2 *sync.WaitGroup) {
		defer wg2.Done()
		for b := uint8(150); b > j; b-- {
			firmataAdaptor.ServoWrite("10", b)
			time.Sleep(8 * time.Millisecond)
		}
	}(uint8(55), &wg2)
	time.Sleep(1 * time.Second)
	go func(j uint8, wg2 *sync.WaitGroup) {
		defer wg2.Done()
		for b := uint8(15); b < j; b++ {
			firmataAdaptor.ServoWrite("6", b)
			time.Sleep(15 * time.Millisecond)
		}
	}(uint8(60), &wg2)
	go func(j uint8, wg2 *sync.WaitGroup) {
		defer wg2.Done()
		for b := uint8(20); b < j; b++ {
			firmataAdaptor.ServoWrite("9", b)
			servo2.Move(b)
			time.Sleep(7 * time.Millisecond)
		}
	}(uint8(110), &wg2)
	wg2.Wait()

	// Rotate to the side, release claw
	var wg3 sync.WaitGroup
	wg3.Add(2)
	go func(k uint8, wg3 *sync.WaitGroup) {
		defer wg3.Done()
		for d := uint8(160); d > k; d-- {
			firmataAdaptor.ServoWrite("5", d)
			time.Sleep(20 * time.Millisecond)
		}
	}(uint8(60), &wg3)
	time.Sleep(2500 * time.Millisecond)
	go func(k uint8, wg3 *sync.WaitGroup) {
		defer wg3.Done()
		for d := uint8(55); d < k; d++ {
			firmataAdaptor.ServoWrite("10", d)
			time.Sleep(5 * time.Millisecond)
		}
	}(uint8(150), &wg3)
	wg3.Wait()
	time.Sleep(750 * time.Millisecond)
	// Move back to origin
	var wg4 sync.WaitGroup
	wg4.Add(3)
	go func(l uint8, wg4 *sync.WaitGroup) {
		defer wg4.Done()
		for e := uint8(150); e > l; e-- {
			firmataAdaptor.ServoWrite("10", e)
			time.Sleep(3 * time.Millisecond)
		}
	}(uint8(70), &wg4)
	go func(l uint8, wg4 *sync.WaitGroup) {
		defer wg4.Done()
		for e := uint8(100); e < l; e++ {
			firmataAdaptor.ServoWrite("9", e)
			time.Sleep(3 * time.Millisecond)
		}
	}(uint8(120), &wg4)
	go func(l uint8, wg4 *sync.WaitGroup) {
		defer wg4.Done()
		for e := uint8(60); e < l; e++ {
			firmataAdaptor.ServoWrite("5", e)
			time.Sleep(20 * time.Millisecond)
		}
	}(uint8(160), &wg4)
	wg4.Wait()

}

func handleDuckDuck(c *gin.Context) {
	var g Grip
	g.Grip = 65
	g.WristPitch = 20
	g.WristRoll = 5
	g.Elbow = 120
	g.Shoulder = 60
	g.Waist = 160
	handleRun(c, g)
	time.Sleep(1 * time.Second)

	// From default to first duck boop
	var wg5 sync.WaitGroup
	wg5.Add(5)

	go func(m uint8, wg5 *sync.WaitGroup) {
		defer wg5.Done()
		for f := g.Waist; f > m; f-- {
			firmataAdaptor.ServoWrite("5", f)
			time.Sleep(20 * time.Millisecond)
		}
	}(uint8(60), &wg5)
	time.Sleep(1500 * time.Millisecond)

	go func(m uint8, wg5 *sync.WaitGroup) {
		defer wg5.Done()
		for f := g.WristPitch; f < m; f++ {
			firmataAdaptor.ServoWrite("9", f)
			time.Sleep(5 * time.Millisecond)
		}
	}(uint8(130), &wg5)

	go func(m uint8, wg5 *sync.WaitGroup) {
		defer wg5.Done()
		for f := g.Shoulder; f > m; f-- {
			firmataAdaptor.ServoWrite("6", f)
			time.Sleep(15 * time.Millisecond)
		}
	}(uint8(0), &wg5)
	time.Sleep(2500 * time.Millisecond)

	go func(m uint8, wg5 *sync.WaitGroup) {
		defer wg5.Done()
		for f := uint8(130); f > m; f-- {
			firmataAdaptor.ServoWrite("9", f)
			time.Sleep(5 * time.Millisecond)
		}
	}(uint8(95), &wg5)
	time.Sleep(500 * time.Millisecond)

	go func(m uint8, wg5 *sync.WaitGroup) {
		defer wg5.Done()
		for f := uint8(95); f < m; f++ {
			firmataAdaptor.ServoWrite("9", f)
			time.Sleep(5 * time.Millisecond)
		}
	}(uint8(130), &wg5)
	wg5.Wait()

	// From first duck to second duck boop
	var wg6 sync.WaitGroup
	wg6.Add(5)

	go func(m uint8, wg6 *sync.WaitGroup) {
		defer wg6.Done()
		for f := uint8(0); f < m; f++ {
			firmataAdaptor.ServoWrite("6", f)
			time.Sleep(15 * time.Millisecond)
		}
	}(uint8(60), &wg6)
	time.Sleep(500 * time.Millisecond)

	go func(m uint8, wg6 *sync.WaitGroup) {
		defer wg6.Done()
		for f := uint8(60); f < m; f++ {
			firmataAdaptor.ServoWrite("5", f)
			time.Sleep(20 * time.Millisecond)
		}
	}(uint8(110), &wg6)
	time.Sleep(1250 * time.Millisecond)

	go func(m uint8, wg6 *sync.WaitGroup) {
		defer wg6.Done()
		for f := uint8(60); f > m; f-- {
			firmataAdaptor.ServoWrite("6", f)
			time.Sleep(15 * time.Millisecond)
		}
	}(uint8(0), &wg6)
	time.Sleep(2500 * time.Millisecond)

	go func(m uint8, wg6 *sync.WaitGroup) {
		defer wg6.Done()
		for f := uint8(130); f > m; f-- {
			firmataAdaptor.ServoWrite("9", f)
			time.Sleep(5 * time.Millisecond)
		}
	}(uint8(95), &wg6)
	time.Sleep(500 * time.Millisecond)

	go func(m uint8, wg6 *sync.WaitGroup) {
		defer wg6.Done()
		for f := uint8(95); f < m; f++ {
			firmataAdaptor.ServoWrite("9", f)
			time.Sleep(5 * time.Millisecond)
		}
	}(uint8(130), &wg6)

	wg6.Wait()

	// From second duck boop to pick up "goose"
	var wg7 sync.WaitGroup
	wg7.Add(9)

	go func(m uint8, wg7 *sync.WaitGroup) {
		defer wg7.Done()
		for f := uint8(0); f < m; f++ {
			firmataAdaptor.ServoWrite("6", f)
			time.Sleep(15 * time.Millisecond)
		}
	}(uint8(60), &wg7)
	time.Sleep(500 * time.Millisecond)

	go func(m uint8, wg7 *sync.WaitGroup) {
		defer wg7.Done()
		for f := uint8(100); f < m; f++ {
			firmataAdaptor.ServoWrite("5", f)
			time.Sleep(20 * time.Millisecond)
		}
	}(uint8(160), &wg7)
	time.Sleep(1250 * time.Millisecond)

	go func(m uint8, wg7 *sync.WaitGroup) {
		defer wg7.Done()
		for f := uint8(60); f > m; f-- {
			firmataAdaptor.ServoWrite("6", f)
			time.Sleep(15 * time.Millisecond)
		}
	}(uint8(0), &wg7)
	time.Sleep(2500 * time.Millisecond)

	go func(m uint8, wg7 *sync.WaitGroup) {
		defer wg7.Done()
		for f := g.Grip; f < m; f++ {
			firmataAdaptor.ServoWrite("10", f)
			time.Sleep(5 * time.Millisecond)
		}
	}(uint8(150), &wg7)
	time.Sleep(750 * time.Millisecond)

	go func(m uint8, wg7 *sync.WaitGroup) {
		defer wg7.Done()
		for f := uint8(130); f > m; f-- {
			firmataAdaptor.ServoWrite("9", f)
			time.Sleep(5 * time.Millisecond)
		}
	}(uint8(77), &wg7)
	time.Sleep(500 * time.Millisecond)

	go func(m uint8, wg7 *sync.WaitGroup) {
		defer wg7.Done()
		for f := uint8(150); f > m; f-- {
			firmataAdaptor.ServoWrite("10", f)
			time.Sleep(5 * time.Millisecond)
		}
	}(uint8(93), &wg7)
	time.Sleep(2000 * time.Millisecond)

	go func(m uint8, wg7 *sync.WaitGroup) {
		defer wg7.Done()
		for f := uint8(70); f < m; f++ {
			firmataAdaptor.ServoWrite("9", f)
			time.Sleep(5 * time.Millisecond)
		}
	}(uint8(130), &wg7)

	go func(m uint8, wg7 *sync.WaitGroup) {
		defer wg7.Done()
		for f := uint8(0); f < m; f++ {
			firmataAdaptor.ServoWrite("6", f)
			time.Sleep(15 * time.Millisecond)
		}
	}(uint8(40), &wg7)

	go func(m uint8, wg7 *sync.WaitGroup) {
		defer wg7.Done()
		for f := uint8(120); f < m; f++ {
			firmataAdaptor.ServoWrite("7", f)
			time.Sleep(15 * time.Millisecond)
		}
	}(uint8(140), &wg7)

	wg7.Wait()

	// "Goose" down and release
	var wg8 sync.WaitGroup
	wg8.Add(8)

	time.Sleep(2 * time.Second)
	go func(m uint8, wg8 *sync.WaitGroup) {
		defer wg8.Done()
		for f := uint8(130); f > m; f-- {
			firmataAdaptor.ServoWrite("9", f)
			time.Sleep(5 * time.Millisecond)
		}
	}(uint8(80), &wg8)

	go func(m uint8, wg8 *sync.WaitGroup) {
		defer wg8.Done()
		for f := uint8(40); f > m; f-- {
			firmataAdaptor.ServoWrite("6", f)
			time.Sleep(15 * time.Millisecond)
		}
	}(uint8(0), &wg8)

	go func(m uint8, wg8 *sync.WaitGroup) {
		defer wg8.Done()
		for f := uint8(140); f > m; f-- {
			firmataAdaptor.ServoWrite("7", f)
			time.Sleep(5 * time.Millisecond)
		}
	}(uint8(120), &wg8)
	time.Sleep(1500 * time.Millisecond)

	go func(m uint8, wg8 *sync.WaitGroup) {
		defer wg8.Done()
		for f := uint8(90); f < m; f++ {
			firmataAdaptor.ServoWrite("10", f)
			time.Sleep(5 * time.Millisecond)
		}
	}(uint8(150), &wg8)
	time.Sleep(2 * time.Second)

	go func(m uint8, wg8 *sync.WaitGroup) {
		defer wg8.Done()
		for f := uint8(0); f < m; f++ {
			firmataAdaptor.ServoWrite("6", f)
			time.Sleep(15 * time.Millisecond)
		}
	}(uint8(40), &wg8)

	go func(m uint8, wg8 *sync.WaitGroup) {
		defer wg8.Done()
		for f := uint8(120); f < m; f++ {
			firmataAdaptor.ServoWrite("7", f)
			time.Sleep(15 * time.Millisecond)
		}
	}(uint8(140), &wg8)
	time.Sleep(1750*time.Millisecond)

	go func(m uint8, wg8 *sync.WaitGroup) {
		defer wg8.Done()
		for f := uint8(80); f < m; f++ {
			firmataAdaptor.ServoWrite("9", f)
			time.Sleep(5 * time.Millisecond)
		}
	}(uint8(130), &wg8)

	go func(m uint8, wg8 *sync.WaitGroup) {
		defer wg8.Done()
		for f := uint8(150); f > m; f-- {
			firmataAdaptor.ServoWrite("10", f)
			time.Sleep(5 * time.Millisecond)
		}
	}(uint8(70), &wg8)

	wg8.Wait()
	// work := func() {

	// 	gobot.After(1*time.Second, func() {
	// 		servo1.Move(55)
	// 		servo2.Move(85)
	// 		servo4.Move(120)
	// 		time.Sleep(500 * time.Millisecond)
	// 		servo5.Move(0)
	// 		servo6.Move(105)
	// 	})

	// 	gobot.After(5*time.Second, func() {
	// 		servo1.Move(55)
	// 		servo2.Move(110)
	// 		servo4.Move(120)
	// 		servo5.Move(60)
	// 		servo6.Move(125)
	// 	})

	// 	gobot.After(8*time.Second, func() {
	// 		servo1.Move(55)
	// 		servo2.Move(85)
	// 		servo4.Move(120)
	// 		servo5.Move(0)
	// 		servo6.Move(125)
	// 	})

	// 	gobot.After(11*time.Second, func() {
	// 		servo1.Move(55)
	// 		servo2.Move(110)
	// 		servo4.Move(120)
	// 		servo5.Move(60)
	// 		servo6.Move(140)
	// 	})

	// 	gobot.After(14*time.Second, func() {
	// 		servo1.Move(55)
	// 		servo2.Move(85)
	// 		servo4.Move(120)
	// 		servo5.Move(0)
	// 		servo6.Move(140)
	// 		time.Sleep(500 * time.Millisecond)
	// 		servo1.Move(150)
	// 		time.Sleep(1500 * time.Millisecond)
	// 		servo1.Move(55)
	// 	})

	// 	gobot.After(17*time.Second, func() {
	// 		servo1.Move(55)
	// 		servo2.Move(20)
	// 		servo4.Move(120)
	// 		servo5.Move(50)
	// 		servo6.Move(150)
	// 		time.Sleep(1500 * time.Millisecond)
	// 		servo1.Move(150)
	// 	})

	// }
	// robot := gobot.NewRobot("servoBot",
	// 	[]gobot.Connection{firmataAdaptor},
	// 	[]gobot.Device{servo1, servo2, servo4, servo5, servo6},
	// 	work,
	// )
	// robot.Start(false)
}
