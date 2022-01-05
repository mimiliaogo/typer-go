package game

import (
	"github.com/kanopeld/go-socket"
	"encoding/json"
	// "strconv"
	"time"
	"log"
)

// NewServer initializes a server that just broadcasts all events
func NewServer(port string) (*socket.Server, error) {
	s, err := socket.NewServer(":" + port)
	if err != nil {
		return nil, err
	}
	players := make(Players)
	game_state := GameState{} // startcntdown, start

	s.On(socket.CONNECTION_NAME, func(c socket.Client) {

		onExit := func() {
			delete(players, c.ID())
			c.Broadcast(ExitPlayer, []byte(c.ID()))
			if (len(players) == 0) {
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
				if (len(players) >= 2) {
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
			ID, progress := ExtractProgress(string(data))
			players[ID].Progress = progress
			jsonString, _ := json.Marshal(players)
			c.Emit(Progress, jsonString)
			c.Broadcast(Progress, jsonString)
		})
		c.On(EndGame, func() {
			go func() {
				ticker := time.NewTicker(1000 * time.Millisecond)
				start_cnt_down_end := time.Now()
				for range ticker.C {
					if int(time.Since(start_cnt_down_end).Seconds()) > 10 {
						ticker.Stop()
						c.Emit(EndGame, nil)
						c.Broadcast(EndGame, nil)
						return
					}
				}
			}()
		})
	})

	go s.Start()

	return s, nil
}
