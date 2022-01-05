package ui

import (
	"sort"
	"strings"

	"github.com/gdamore/tcell"
	"github.com/kanopeld/go-socket"
	"github.com/rivo/tview"
	"github.com/shilangyu/typer-go/game"
	"github.com/shilangyu/typer-go/utils"

	"encoding/json"
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
func CreateMultiplayerConnection(app *tview.Application, setup setup) error {
	// IP, _ := utils.IPv4()
	// mimi local
	// IP = "127.0.0.1"
	// setup := setup{IP, "", "9001", true, nil, nil}

	// [mimi] NO UI
	c, err := socket.NewDial(setup.RoomIP + ":" + setup.Port)
	if err != nil { // build server
		var serv_err error
		setup.IsServer = true
		setup.Server, serv_err = game.NewServer(setup.Port)
		utils.Check(serv_err)
		_c, _err := socket.NewDial(setup.RoomIP + ":" + setup.Port)
		utils.Check(_err)
		setup.Client = _c
	} else {
		setup.Client = c
	}
	// utils.Check(CreateMultiplayer(app, setup))

	return nil
}

// CreateMultiplayerRoom creates multiplayer room
func CreateMultiplayerSetup(app *tview.Application) error {
	const maxNicknameLength int = 6
	IP := "140.114.89.109" // "127.0.0.1"
	setup := setup{IP, "", "9001", false, nil, nil}

	// build connection
	c, err := socket.NewDial(setup.RoomIP + ":" + setup.Port)
	utils.Check(err)
	setup.Client = c

	formWi := tview.NewForm().
		AddInputField("Nickname", setup.Nickname, 20, func(textToCheck string, lastChar rune) bool {
			return len(textToCheck) <= maxNicknameLength
		}, func(text string) {
			setup.Nickname = text
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
			AddItem(Center(28, 11, formWi), 0, 1, true),
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

	// statsWis := [...]*tview.TextView{
	// 	tview.NewTextView().SetText("wpm: 0"),
	// 	tview.NewTextView().SetText("time: 0s"),
	// 	tview.NewTextView().SetText("Player numeber: 0"),
	// 	tview.NewTextView().SetText("CntDown: 10s"), // count down clock
	// 	tview.NewTextView().SetText(""),             // players list
	// }
	stateText := tview.NewTextView().SetText(ascii_waiting)

	// engine: car race animation
	const carSign = `
.-'---\._
'-O---O--'`

	carWi := tview.NewTextView().SetText(carSign)
	carWi.SetBorder(false).SetBorderPadding(1, 1, 1, 1)

	pages := tview.NewPages().
		AddPage("modal", tview.NewModal().
			SetText("Game End").
			SetBackgroundColor(tcell.ColorDefault).
			AddButtons([]string{"exit"}).
			SetDoneFunc(func(index int, label string) {
				switch index {
				case 0:
					utils.Check(CreateWelcome(app))
				}
			}), false, false)

	renderPlayers := func() {
		// _, _, w, _ := carWi.GetRect()
		// log.Println("renderPlayers", w)

		ps := ""
		// TODO: sort players by progress

		// engine: sort players by name
		keys := make([]string, 0, len(players))
		for k := range players {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		// engine TODO: [bug] cannot get correct width first
		_, _, trackWidth, _ := carWi.GetRect()
		if trackWidth == 15 {
			trackWidth = 40
		}
		drawtext := "Player:"
		for _, k := range keys {
			p := players[k]
			ps += p.Nickname + ": " + strconv.Itoa(p.Progress) + "%\n"

			step := strings.Repeat(" ", p.Progress*(trackWidth-24)/100)
			cs := strings.Replace(carSign, "\n", "\n        "+step, 1)
			nickname := fmt.Sprintf("%-8s", p.Nickname)
			cs = strings.Replace(cs, "_\n", "_\n"+nickname+step, 1)
			drawtext += cs
			padding := strings.Repeat(" ", trackWidth-18-len(step)-8)
			drawtext += padding
			drawtext += fmt.Sprintf("%2s", strconv.Itoa(p.WPM)) + " wpm\n"
			drawtext += strings.Repeat("▔", trackWidth-2)
		}
		app.QueueUpdateDraw(func() {
			// statsWis[2].SetText(fmt.Sprintf("Num: %d", len(players)))
			// statsWis[4].SetText(ps)
			carWi.SetText(drawtext)
		})
	}

	startGame := func() {
		// start game
		if state.StartTime.IsZero() {
			state.Start()
			go func() {
				ticker := time.NewTicker(100 * time.Millisecond)
				for range ticker.C {
					if state.CurrWord == len(state.Words) || state.EndGame {
						ticker.Stop()
						return
					}
					app.QueueUpdateDraw(func() {
						// statsWis[0].SetText(fmt.Sprintf("wpm: %.0f", state.Wpm()))
						// statsWis[1].SetText(fmt.Sprintf("time: %.02fs", time.Since(state.StartTime).Seconds()))
						s_time := fmt.Sprintf("%.0f", time.Since(state.StartTime).Seconds())
						stateText.SetText(string_to_ascii(s_time))
					})

					// broadcast progress
					setup.Client.Emit(game.Progress, setup.Client.ID()+
						":"+strconv.Itoa(int(state.Progress()))+
						":"+strconv.Itoa(int(state.Wpm())))
					players[setup.Client.ID()].Progress = int(state.Progress())
					players[setup.Client.ID()].WPM = int(state.Wpm())
					renderPlayers()
				}
			}()
		}
	}

	// emit your information
	setup.Client.Emit(game.EnterGame, setup.Client.ID()+":"+setup.Nickname)

	setup.Client.On(game.EnterGame, func(_players []byte) {
		json.Unmarshal(_players, &players)
		renderPlayers()

	})

	setup.Client.On(game.StartCountDown, func(payload string) {
		if state.StartCountDownTime.IsZero() {
			state.StartCountDownTime, _ = time.Parse(time.RFC3339, payload)
			go func() {
				ticker_cnt := time.NewTicker(10 * time.Millisecond)
				for range ticker_cnt.C {
					if int(time.Since(state.StartCountDownTime).Seconds()) > 10 { // start game
						ticker_cnt.Stop()
						startGame()
						return
					}
					app.QueueUpdateDraw(func() {
						cnt := fmt.Sprintf("%d", 10-int(time.Since(state.StartCountDownTime).Seconds()))
						stateText.SetText(concat_ascii(ascii_countdown, string_to_ascii(cnt)))
					})
				}
			}()
		}
	})

	setup.Client.On(game.Progress, func(_players []byte) {
		json.Unmarshal(_players, &players)
		renderPlayers()
	})

	setup.Client.On(game.EndGame, func() {
		state.End()
		pages.ShowPage("modal")
	})

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
						// first
						setup.Client.Emit(game.EndGame, nil)
						state.End()
						pages.ShowPage("modal")
					} else {
						inputWi.SetText("")
					}
				}

				currInput = text
			} else {
				// no input text
			}
		})

	// mimi layout design
	layout := tview.NewFlex()

	signW, _ := utils.StringDimensions(ascii_waiting)
	statsFrame := tview.NewFlex().
		AddItem(tview.NewBox(), 0, 1, false).
		AddItem(stateText, signW, 1, true).
		AddItem(tview.NewBox(), 0, 1, false)
	statsFrame.SetBorder(true).SetBorderPadding(1, 1, 1, 1) //.SetTitle("STATS")
	/*
		statsFrame := tview.NewFlex().SetDirection(tview.FlexRow)
		statsFrame.SetBorder(true).SetBorderPadding(1, 1, 1, 1).SetTitle("STATS")
		for _, statsWi := range statsWis {
			// statsFrame.AddItem(statsWi, 1, 1, false)
			// mimi flexible
			statsFrame.AddItem(statsWi, 0, 1, false)
		}
	*/

	secondColumn := tview.NewFlex().SetDirection(tview.FlexRow)

	textsLayout := tview.NewFlex().SetDirection(tview.FlexRow)
	textsLayout.SetBorder(true)

	// _, _, secondColumnWidth, _ := secondColumn.GetRect()
	var textLines []*tview.Flex
	textLines = append(textLines, tview.NewFlex())
	curLine := 0
	curTextWidth := 0
	for _, textWi := range textWis {
		textLen := len(textWi.GetText(true))
		textLines[curLine].AddItem(textWi, textLen, 1, false)
		curTextWidth += textLen
		if curTextWidth > 70 {
			textsLayout.AddItem(textLines[curLine], 1, 1, false)
			curTextWidth = 0
			curLine += 1
			textLines = append(textLines, tview.NewFlex())
		}
	}

	setup.Client.Emit(game.GetText, nil)
	setup.Client.On(game.GetText, func(text string) {
		words := strings.Split(text, " ")
		for i := range words[:len(words)-1] {
			words[i] += " "
		}
		state.Words = words

		textWis = nil
		for _, word := range state.Words {
			textWis = append(textWis, tview.NewTextView().SetText(word).SetDynamicColors(true))
		}

		_, _, secondColumnWidth, _ := secondColumn.GetRect()
		var textLines []*tview.Flex
		textLines = append(textLines, tview.NewFlex())
		curLine := 0
		curTextWidth := 0
		for _, textWi := range textWis {
			textLen := len(textWi.GetText(true))
			if curTextWidth+textLen > secondColumnWidth-2 {
				textsLayout.AddItem(textLines[curLine], 1, 1, false)
				curTextWidth = 0
				curLine += 1
				textLines = append(textLines, tview.NewFlex())
			}
			textLines[curLine].AddItem(textWi, textLen, 1, false)
			curTextWidth += textLen
		}
		textsLayout.AddItem(textLines[curLine], 1, 1, false)
		renderPlayers()
	})

	layout.AddItem(tview.NewBox().SetBorder(false), 0, 1, false) // left margin
	secondColumn.AddItem(statsFrame, 9, 1, false)
	secondColumn.AddItem(carWi, 0, 3, false)
	secondColumn.AddItem(textsLayout, 8, 1, false)
	inputWi.SetBorder(true)
	secondColumn.AddItem(inputWi, 3, 1, true)
	secondColumn.AddItem(tview.NewBox().SetBorder(false), 0, 1, false) // buttom margin

	layout.AddItem(secondColumn, 0, 3, true)
	layout.AddItem(tview.NewBox().SetBorder(false), 0, 1, false) // right margin

	pages.AddPage("game", layout, true, true).SendToBack("game")
	app.SetRoot(pages, true)

	keybindings(app, CreateWelcome)
	return nil

}
