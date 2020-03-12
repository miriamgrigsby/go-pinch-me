package main

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
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
	NewBot     string `json:"NewBot"`
	Grip       uint8  `json:"Grip"`
	WristPitch uint8  `json:"WristPitch"`
	WristRoll  uint8  `json:"WristRoll"`
	Elbow      uint8  `json:"Elbow"`
	Shoulder   uint8  `json:"Shoulder"`
	Waist      uint8  `json:"Waist"`
}

type savingActions struct {
	ID    int
	Names []string `json:"Names"`
	Bots  []NewBot `json:"Bots"`
}

type Name struct {
	Name string `json:"Name"`
}

type savingEachAction struct {
	ID    int
	Attrs NewBot
}

var dbBotActions []NewBot
var NewBotActions []NewBot

var db *sql.DB

var firmataAdaptor = firmata.NewAdaptor("/dev/cu.usbmodem14201")

func (a NewBot) Value() (driver.Value, error) {
	return json.Marshal(a)
}

func (a *NewBot) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &a)
}

func handleRequest() {
	myRouter := gin.Default()

	myRouter.Use(cors.Default())

	myRouter.Any("/bluetoothOn", handleBluetoothOn)
	myRouter.Any("/bluetoothOff", handleBluetoothOff)
	myRouter.Any("/slider", handleSlider)
	myRouter.Any("/dragNdrop", handleDragNDrop)
	myRouter.Any("/duck", handleDuckDuck)
	myRouter.Any("/newBot", handleNewBot)
	myRouter.Any("/onSave", handleSave)
	myRouter.Any("/findName", handleFindName)
	myRouter.Any("/showNewRobot", handleShowNewRobot)
	myRouter.Any("/deleteRobot", handleDeleteRobot)
	myRouter.Any("/onReset", handleReset)

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
	firmataAdaptor.ServoWrite("10", g.Grip)
	firmataAdaptor.ServoWrite("9", g.WristPitch)
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

func handleReset(c *gin.Context) {
	NewBotActions = nil
}

func handleNewBot(c *gin.Context) {
	var n NewBot
	c.Bind(&n)

	NewBotActions = append(NewBotActions, n)

}

func handleSave(c *gin.Context) {
	db, err := sql.Open("postgres", "user=miriamgrigsby dbname=robot sslmode=disable")
	if err != nil {
		panic(err.Error())
	}
	for i := 0; i < len(NewBotActions); i++ {
		_, err = db.Exec("INSERT INTO robotactions (actions) VALUES($1)", NewBotActions[i])
	}

	if err != nil {
		panic(err.Error())
	}

}

func unique(intSlice []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range intSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func handleFindName(c *gin.Context) {
	db, err := sql.Open("postgres", "user=miriamgrigsby dbname=robot sslmode=disable")
	if err != nil {
		panic(err.Error())
	}
	var savingActions savingActions
	rows, err := db.Query("SELECT actions FROM robotactions")
	defer rows.Close()
	for rows.Next() {
		var savedBotActions NewBot
		err = rows.Scan(&savedBotActions)
		savingActions.Names = append(savingActions.Names, savedBotActions.NewBot)
		savingActions.Bots = append(savingActions.Bots, savedBotActions)
	}
	savingActions.Names = unique(savingActions.Names)
	c.JSON(200, savingActions.Names)
}

func handleDeleteRobot(c *gin.Context) {
	var n Name
	c.Bind(&n)
	db, err := sql.Open("postgres", "user=miriamgrigsby dbname=robot sslmode=disable")
	res, err := db.Exec("DELETE FROM robotactions WHERE actions -> 'NewBot' ? $1", n.Name)
	if err != nil {
		panic(err.Error())
	}
	count, _ := res.RowsAffected()
	fmt.Println(count)
}

func handleShowNewRobot(c *gin.Context) {
	var moveAction savingActions
	var n Name
	c.Bind(&n)
	db, err := sql.Open("postgres", "user=miriamgrigsby dbname=robot sslmode=disable")
	rows, err := db.Query("SELECT * FROM robotactions WHERE actions -> 'NewBot' ? $1", n.Name)
	if err != nil {
		panic(err.Error())
	}
	defer rows.Close()
	for rows.Next() {
		var savingActions savingEachAction
		err = rows.Scan(&savingActions.ID, &savingActions.Attrs)
		if err != nil {
			panic(err.Error())
		}
		moveAction.Bots = append(moveAction.Bots, savingActions.Attrs)
	}

	firmataAdaptor.Connect()
	firmataAdaptor.ServoWrite("10", 55)
	firmataAdaptor.ServoWrite("9", 20)
	firmataAdaptor.ServoWrite("8", 5)
	firmataAdaptor.ServoWrite("7", 120)
	firmataAdaptor.ServoWrite("6", 60)
	firmataAdaptor.ServoWrite("5", 160)
	time.Sleep(1*time.Second)

	var newBot NewBot
	newBot = moveAction.Bots[0]
	
	var servo1, servo2, servo3, servo4, servo5, servo6 uint8 = 0, 0, 0, 0, 0, 0
	for i := 1; i < len(moveAction.Bots); i++ {
		fmt.Println(newBot)
		if newBot.Grip > moveAction.Bots[i].Grip {
			servo1 = 1
		} else {
			servo1 = 0
		}
		if newBot.WristPitch > moveAction.Bots[i].WristPitch {
			servo2 = 1
		} else {
			servo2 = 0
		}
		if newBot.WristRoll > moveAction.Bots[i].WristRoll {
			servo3 = 1
		} else {
			servo3 = 0
		}
		if newBot.Elbow > moveAction.Bots[i].Elbow {
			servo4 = 1
		} else {
			servo4 = 0
		}
		if newBot.Shoulder > moveAction.Bots[i].Shoulder {
			servo5 = 1
		} else {
			servo5 = 0
		}
		if newBot.Waist > moveAction.Bots[i].Waist {
			servo6 = 1
		} else {
			servo6 = 0
		}

		var wg9 sync.WaitGroup
		wg9.Add(6)
		go func(a uint8, wg9 *sync.WaitGroup) {
			defer wg9.Done()
			for j := newBot.Grip; j != a; {
				firmataAdaptor.ServoWrite("10", j)
				time.Sleep(5 * time.Millisecond)
				if servo1 == 1 {
					j--
				} else {
					j++
				}
			}
		}(uint8(moveAction.Bots[i].Grip), &wg9)

		go func(a uint8, wg9 *sync.WaitGroup) {
			defer wg9.Done()
			for j := newBot.WristPitch; j != a; {
				firmataAdaptor.ServoWrite("9", j)
				time.Sleep(5 * time.Millisecond)
				if servo2 == 1 {
					j--
				} else {
					j++
				}
			}
		}(uint8(moveAction.Bots[i].WristPitch), &wg9)

		go func(a uint8, wg9 *sync.WaitGroup) {
			defer wg9.Done()
			for j := newBot.WristRoll; j != a; {
				firmataAdaptor.ServoWrite("8", j)
				time.Sleep(5 * time.Millisecond)
				if servo3 == 1 {
					j--
				} else {
					j++
				}
			}
		}(uint8(moveAction.Bots[i].WristRoll), &wg9)

		go func(a uint8, wg9 *sync.WaitGroup) {
			defer wg9.Done()
			for j := newBot.Elbow; j != a; {
				firmataAdaptor.ServoWrite("7", j)
				time.Sleep(5 * time.Millisecond)
				if servo4 == 1 {
					j--
				} else {
					j++
				}
			}
		}(uint8(moveAction.Bots[i].Elbow), &wg9)

		go func(a uint8, wg9 *sync.WaitGroup) {
			defer wg9.Done()
			for j := newBot.Shoulder; j != a; {
				firmataAdaptor.ServoWrite("6", j)
				time.Sleep(15 * time.Millisecond)
				if servo5 == 1 {
					j--
				} else {
					j++
				}
			}
		}(uint8(moveAction.Bots[i].Shoulder), &wg9)

		go func(a uint8, wg9 *sync.WaitGroup) {
			defer wg9.Done()
			for j := newBot.Waist; j != a; {
				firmataAdaptor.ServoWrite("5", j)
				time.Sleep(20 * time.Millisecond)
				if servo6 == 1 {
					j--
				} else {
					j++
				}
			}
		}(uint8(moveAction.Bots[i].Waist), &wg9)
		wg9.Wait()
		if n.Name != "Lickity Splickty" {
			time.Sleep(1*time.Second)
		}
		newBot = moveAction.Bots[i]
	}
}

func handleDragNDrop(c *gin.Context) {
	var g Grip
	c.Bind(&g)
	g.WristRoll = 5
	g.Waist = 160
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
	}(uint8(168), &wg1)
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
		for b := uint8(50); b < j; b++ {
			firmataAdaptor.ServoWrite("9", b)
			time.Sleep(7 * time.Millisecond)
		}
	}(uint8(110), &wg2)
	wg2.Wait()

	// Rotate to the side, release claw
	var wg3 sync.WaitGroup
	wg3.Add(2)
	go func(k uint8, wg3 *sync.WaitGroup) {
		defer wg3.Done()
		for d := uint8(163); d > k; d-- {
			firmataAdaptor.ServoWrite("5", d)
			time.Sleep(20 * time.Millisecond)
		}
	}(uint8(55), &wg3)
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
	}(uint8(157), &wg4)
	wg4.Wait()

}

func handleDuckDuck(c *gin.Context) {
	var g Grip
	g.Grip = 65
	g.WristPitch = 20
	g.WristRoll = 5
	g.Elbow = 120
	g.Shoulder = 60
	g.Waist = 157
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
	}(uint8(55), &wg5)
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
		for f := uint8(55); f < m; f++ {
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
	}(uint8(64), &wg7)
	time.Sleep(500 * time.Millisecond)

	go func(m uint8, wg7 *sync.WaitGroup) {
		defer wg7.Done()
		for f := uint8(110); f < m; f++ {
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
	}(uint8(75), &wg7)
	time.Sleep(500 * time.Millisecond)

	go func(m uint8, wg7 *sync.WaitGroup) {
		defer wg7.Done()
		for f := uint8(150); f > m; f-- {
			firmataAdaptor.ServoWrite("10", f)
			time.Sleep(5 * time.Millisecond)
		}
	}(uint8(90), &wg7)
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
	}(uint8(143), &wg7)

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
		for f := uint8(143); f > m; f-- {
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
	time.Sleep(1750 * time.Millisecond)

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
}
