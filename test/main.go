// Demo code for the Flex primitive.
package main

import (
	"strings"

	"github.com/rivo/tview"
	"github.com/shilangyu/typer-go/utils"
)

func Center(width, height int, p tview.Primitive) tview.Primitive {
	return tview.NewFlex().
		AddItem(tview.NewBox(), 0, 1, false).
		AddItem(tview.NewFlex().
			SetDirection(tview.FlexRow).
			AddItem(tview.NewBox(), 0, 1, false).
			AddItem(p, height, 1, true).
			AddItem(tview.NewBox(), 0, 1, false), width, 1, true).
		AddItem(tview.NewBox(), 0, 1, false)
}

func main() {
	carSign := `
.-'---\._
'-O---O--'
`
	step := strings.Repeat(" ", 3)
	carSign = strings.Replace(carSign, "\n", "\n"+step, 2)
	const trackSign = "▔"
	carSign += strings.Repeat(trackSign, 20)

	carSign = strings.Repeat(carSign, 3)

	app := tview.NewApplication()
	// signW, signH := utils.StringDimensions(carSign)
	// signWi := tview.NewTextView().SetText(carSign)
	// layout := tview.NewFlex().
	// 	SetDirection(tview.FlexRow).
	// 	AddItem(tview.NewBox().SetBorder(true), 0, 1, false).
	// 	AddItem(Center(signW, signH, signWi), 0, 1, false).
	// 	AddItem(tview.NewBox().SetBorder(true), 0, 1, false)

	// button := tview.NewButton("Hit Enter to close").SetSelectedFunc(func() {
	// 	app.Stop()
	// })
	// button.SetBorder(true).SetRect(0, 0, 22, 3)
	// if err := app.SetRoot(button, false).EnableMouse(true).Run(); err != nil {
	// 	panic(err)
	// }

	// layout := tview.NewBox()
	signW, signH := utils.StringDimensions(carSign)
	carWi := tview.NewTextView().SetText(carSign)
	carWi.SetBorder(false).SetRect(10, 10, signW, signH)

	button := tview.NewButton("Hit Enter to close").SetSelectedFunc(func() {
		app.Stop()
	})
	button.SetBorder(true).SetRect(0, 0, 22, 3)

	layout := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(button, 0, 1, false).
		AddItem(tview.NewBox().SetBorder(true), 0, 1, false).
		AddItem(carWi, 0, 1, false).
		AddItem(tview.NewBox().SetBorder(true), 0, 1, false)

	if err := app.SetRoot(layout, true).Run(); err != nil {
		panic(err)
	}
}
