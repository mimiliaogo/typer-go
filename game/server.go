package game

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/antchfx/htmlquery"
	"github.com/gocolly/colly"
	"github.com/kanopeld/go-socket"

	// "strconv"
	"log"
	"time"
)

// NewServer initializes a server that just broadcasts all events
func NewServer(port string) (*socket.Server, error) {
	s, err := socket.NewServer(":" + port)
	if err != nil {
		return nil, err
	}
	players := make(Players)
	game_state := GameState{} // startcntdown, start
	text := ""
	numEndPlayers := 0
	s.On(socket.CONNECTION_NAME, func(c socket.Client) {

		onExit := func() {
			delete(players, c.ID())
			c.Broadcast(ExitPlayer, []byte(c.ID()))
			if len(players) == 0 {
				game_state = GameState{}
			}
		}
		c.On(socket.DISCONNECTION_NAME, onExit)
		c.On(ExitPlayer, onExit)

		c.On(EnterGame, func(data []byte) {
			ID, nickname := ExtractChangeName(string(data))
			players.Add(ID, nickname)
			jsonString, _ := json.Marshal(players)
			log.Println(string(jsonString))
			c.Emit(EnterGame, jsonString)
			c.Broadcast(EnterGame, jsonString)

			if game_state.StartCountDownTime.IsZero() {
				if len(players) >= 2 {
					game_state.StartCountDownTime = time.Now()
					c.Emit(StartCountDown, []byte(game_state.StartCountDownTime.Format(time.RFC3339)))
					c.Broadcast(StartCountDown, []byte(game_state.StartCountDownTime.Format(time.RFC3339)))
				}
			} else {
				c.Emit(StartCountDown, []byte(game_state.StartCountDownTime.Format(time.RFC3339)))
				c.Broadcast(StartCountDown, []byte(game_state.StartCountDownTime.Format(time.RFC3339)))
			}

		})
		c.On(Progress, func(data []byte) {
			ID, progress, wpm := ExtractProgress(string(data))
			players[ID].Progress = progress
			players[ID].WPM = wpm
			jsonString, _ := json.Marshal(players)
			c.Emit(Progress, jsonString)
			c.Broadcast(Progress, jsonString)
		})
		c.On(EndGame, func() {
			numEndPlayers += 1
			if numEndPlayers == len(players) {
				game_state = GameState{}
				players = make(Players)
				text = ""
				numEndPlayers = 0
			} else if numEndPlayers == 1 {
				go func() {
					ticker := time.NewTicker(1000 * time.Millisecond)
					start_cnt_down_end := time.Now()
					for range ticker.C {
						if int(time.Since(start_cnt_down_end).Seconds()) > 10 {
							ticker.Stop()
							c.Emit(EndGame, nil)
							c.Broadcast(EndGame, nil)
							game_state = GameState{}
							players = make(Players)
							text = ""
							numEndPlayers = 0
							return
						}
					}
				}()
			}
		})
		c.On(GetText, func() {
			if len(players) == 1 {

				col := colly.NewCollector()

				col.OnRequest(func(r *colly.Request) {
					fmt.Println("Visiting", r.URL.String())
				})

				col.OnResponse(func(r *colly.Response) {
					doc, err := htmlquery.Parse(strings.NewReader(string(r.Body)))
					if err != nil {
						log.Fatal(err)
					}
					// nodes := htmlquery.FindOne(doc, `//*[@id="mw-content-text"]/div[1]/p[2]`)
					nodes := htmlquery.FindOne(doc, `//*[@id="mw-content-text"]/div[1]/p[1]`)
					s := htmlquery.InnerText(nodes)
					fmt.Println(s)

					// remove non ascii
					re := regexp.MustCompile("[[:^ascii:]]")
					t := re.ReplaceAllLiteralString(s, "")
					fmt.Println(t)

					// remove consecutive spaces
					re = regexp.MustCompile("[ ]{2,}")
					text = re.ReplaceAllLiteralString(t, " ")
					text = strings.Replace(text, "\n", " ", -1)
					text = strings.Replace(text, "[", " ", -1)
					text = strings.Replace(text, "]", " ", -1)
					if len(text) > 200 {
						text = text[:200]
					} else if len(text) < 10 {
						text = "A wiki is a hypertext publication collaboratively edited and managed by its own audience directly using a web browser."
					}
					text = strings.TrimSpace(text)

					c.Emit(GetText, text)
				})
				col.Visit("https://en.wikipedia.org/wiki/Special:Random")

			} else {
				c.Emit(GetText, text)
			}
		})
	})

	go s.Start()

	return s, nil
}
