package ui

import (
	"github.com/gdamore/tcell"
	"github.com/kanopeld/go-socket"
	"github.com/rivo/tview"
	"github.com/shilangyu/typer-go/game"
	"github.com/shilangyu/typer-go/utils"

	// "log"
	"fmt"
	"strconv"
	"time"
)

type setup struct {
	RoomIP, Nickname, Port string
	IsServer               bool
	// Server will be nil if IsServer is false
	Server *socket.Server
	Client socket.Client
}

// CreateMultiplayerSetup creates multiplayer room
func CreateMultiplayerSetup(app *tview.Application) error {
	IP, _ := utils.IPv4()
	// mimi local
	IP = "127.0.0.1"
	setup := setup{IP, "", "9001", true, nil, nil}

	formWi := tview.NewForm().
		AddInputField("Room IP", setup.RoomIP, 20, nil, func(text string) { setup.RoomIP = text }).
		AddInputField("Port", setup.Port, 20, nil, func(text string) { setup.Port = text }).
		AddCheckbox("Server", setup.IsServer, func(checked bool) { setup.IsServer = checked }).
		AddButton("CONNECT", func() {
			if setup.IsServer {
				var err error
				setup.Server, err = game.NewServer(setup.Port)
				utils.Check(err)
			}

			c, err := socket.NewDial(setup.RoomIP + ":" + setup.Port)
			utils.Check(err)
			setup.Client = c

			utils.Check(CreateMultiplayerRoom(app, setup))
		}).
		AddButton("CANCEL", func() {
			utils.Check(CreateWelcome(app))
		})

	app.SetRoot(Center(28, 11, formWi), true)
	keybindings(app, CreateWelcome)
	return nil
}

// CreateMultiplayerRoom creates multiplayer room
func CreateMultiplayerRoom(app *tview.Application, setup setup) error {
	const maxNicknameLength int = 10

	players := make(game.Players)

	roomWi := tview.NewTextView()
	roomWi.SetBorder(true).SetTitle("ROOM")
	renderRoom := func() {
		ps := ""
		for _, p := range players {
			ps += p.Nickname + "\n"
		}
		app.QueueUpdateDraw(func() {
			roomWi.SetText(ps)
		})
	}
	setup.Client.On(socket.CONNECTION_NAME, func(ccc socket.Client) {
		setup.Client.On(game.ChangeName, func(payload string) {
			ID, nickname := game.ExtractChangeName(payload)
			players.Add(ID, nickname)
			renderRoom()
		})
		setup.Client.On(game.ExitPlayer, func(payload string) {
			ID := game.ExtractExitPlayer(payload)
			delete(players, ID)
			renderRoom()
		})
		setup.Client.Emit(game.ChangeName, setup.Client.ID()+":"+setup.Nickname)
		// [mimi]: local players : add yourself
		players[setup.Client.ID()] = &game.Player{Nickname: setup.Nickname}
		renderRoom()
	})

	formWi := tview.NewForm().
		AddInputField("Nickname", setup.Nickname, 20, func(textToCheck string, lastChar rune) bool {
			return len(textToCheck) <= maxNicknameLength
		}, func(text string) {
			setup.Nickname = text
			players[setup.Client.ID()].Nickname = setup.Nickname
			setup.Client.Emit(game.ChangeName, setup.Client.ID()+":"+setup.Nickname)
		}).
		AddButton("BACK", func() {
			utils.Check(CreateMultiplayerSetup(app))
		}).
		AddButton("Enter", func() {
			utils.Check(CreateMultiplayer(app, setup))
		})

	layout := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(tview.NewBox(), 0, 1, false).
		AddItem(tview.NewFlex().
			AddItem(Center(28, 11, formWi), 0, 1, true).
			AddItem(Center(28, 11, roomWi), 0, 1, false),
			0, 1, true).
		AddItem(tview.NewBox(), 0, 1, false)

	app.SetRoot(layout, true)
	keybindings(app, func(app *tview.Application) error {
		setup.Client.Emit(game.ExitPlayer, setup.Client.ID())
		return CreateMultiplayerSetup(app)
	})

	return nil
}

func CreateMultiplayer(app *tview.Application, setup setup) error {

	// [TODO] : socket
	text, err := game.ChooseText()
	if err != nil {
		return err
	}

	state := game.NewState(text)

	players := make(game.Players)

	statsWis := [...]*tview.TextView{
		tview.NewTextView().SetText("wpm: 0"),
		tview.NewTextView().SetText("time: 0s"),
		tview.NewTextView().SetText("Player numeber: 0"),
		tview.NewTextView().SetText("CntDown: 10s"), // count down clock
		tview.NewTextView().SetText(""),             // players list
	}

	renderPlayers := func() {
		// log.Println(len(players))
		ps := ""
		// TODO: sort players by progress
		for _, p := range players {
			ps += p.Nickname + ": " + strconv.Itoa(p.Progress) + "%\n"
			// ps += fmt.Sprintf("%s: 0\n", p.Nickname)
		}
		app.QueueUpdateDraw(func() {
			statsWis[2].SetText(fmt.Sprintf("Num: %d", len(players)))
			statsWis[4].SetText(ps)
		})
	}

	startGame := func() {
		// start game
		if state.StartTime.IsZero() {
			state.Start()
			setup.Client.On(game.Progress, func(payload string) {
				ID, progress := game.ExtractProgress(payload)
				players[ID].Progress = progress
				renderPlayers()
			})
			go func() {
				ticker := time.NewTicker(100 * time.Millisecond)
				for range ticker.C {
					if state.CurrWord == len(state.Words) {
						ticker.Stop()
						return
					}
					app.QueueUpdateDraw(func() {
						statsWis[0].SetText(fmt.Sprintf("wpm: %.0f", state.Wpm()))
						statsWis[1].SetText(fmt.Sprintf("time: %.02fs", time.Since(state.StartTime).Seconds()))
					})

					// broadcast progress
					setup.Client.Emit(game.Progress, setup.Client.ID()+":"+strconv.Itoa(int(state.Progress())))
					players[setup.Client.ID()].Progress = int(state.Progress())
					renderPlayers()
				}
			}()
		}
	}

	setup.Client.On(game.EnterGame, func(payload string) {
		ID, nickname := game.ExtractChangeName(payload)
		players.Add(ID, nickname)
		renderPlayers()
		// check player number
		if len(players) >= 2 {
			if state.StartCountDownTime.IsZero() {
				state.StartCountDownTime = time.Now()
				go func() {
					ticker_cnt := time.NewTicker(1000 * time.Millisecond)
					for range ticker_cnt.C {
						if int(time.Since(state.StartCountDownTime).Seconds()) > 10 { // start game
							ticker_cnt.Stop()
							startGame()
							return
						}
						app.QueueUpdateDraw(func() {
							statsWis[3].SetText(fmt.Sprintf("CntDown: %ds", 10-int(time.Since(state.StartCountDownTime).Seconds())))
						})
					}
				}()
			}
		}
	})

	setup.Client.Emit(game.EnterGame, setup.Client.ID()+":"+setup.Nickname)
	players[setup.Client.ID()] = &game.Player{Nickname: setup.Nickname}
	renderPlayers()

	pages := tview.NewPages().
		AddPage("modal", tview.NewModal().
			SetText("Play again?").
			SetBackgroundColor(tcell.ColorDefault).
			AddButtons([]string{"yes", "exit"}).
			SetDoneFunc(func(index int, label string) {
				switch index {
				case 0:
					utils.Check(CreateSingleplayer(app))
				case 1:
					utils.Check(CreateWelcome(app))
				}
			}), false, false)

	var textWis []*tview.TextView
	for _, word := range state.Words {
		textWis = append(textWis, tview.NewTextView().SetText(word).SetDynamicColors(true))
	}

	currInput := ""
	inputWi := tview.NewInputField().
		SetFieldBackgroundColor(tcell.ColorDefault)
	inputWi.
		SetChangedFunc(func(text string) {
			if !state.StartTime.IsZero() {
				if len(currInput) < len(text) {
					if len(text) > len(state.Words[state.CurrWord]) || state.Words[state.CurrWord][len(text)-1] != text[len(text)-1] {
						state.IncError()
					}
				}

				app.QueueUpdateDraw(func(i int) func() {
					return func() {
						textWis[i].SetText(paintDiff(state.Words[i], text))
					}
				}(state.CurrWord))

				if text == state.Words[state.CurrWord] {
					state.NextWord()
					if state.CurrWord == len(state.Words) {
						state.End()

						pages.ShowPage("modal")
					} else {
						inputWi.SetText("")
					}
				}

				currInput = text
			}
		})

	// mimi layout design
	layout := tview.NewFlex()
	statsFrame := tview.NewFlex().SetDirection(tview.FlexRow)
	statsFrame.SetBorder(true).SetBorderPadding(1, 1, 1, 1).SetTitle("STATS")
	for _, statsWi := range statsWis {
		// statsFrame.AddItem(statsWi, 1, 1, false)
		// mimi flexible
		statsFrame.AddItem(statsWi, 0, 1, false)
	}
	layout.AddItem(statsFrame, 0, 1, false)

	secondColumn := tview.NewFlex().SetDirection(tview.FlexRow)
	textsLayout := tview.NewFlex()
	for _, textWi := range textWis {
		textsLayout.AddItem(textWi, len(textWi.GetText(true)), 1, false)
	}
	textsLayout.SetBorder(true)
	secondColumn.AddItem(textsLayout, 0, 3, false)
	inputWi.SetBorder(true)
	secondColumn.AddItem(inputWi, 0, 1, true)
	layout.AddItem(secondColumn, 0, 3, true)

	pages.AddPage("game", layout, true, true).SendToBack("game")
	app.SetRoot(pages, true)

	keybindings(app, CreateWelcome)
	return nil

}
