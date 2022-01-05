package main
import (
	"github.com/kanopeld/go-socket"
	// "github.com/rivo/tview"
	"github.com/shilangyu/typer-go/game"
	"github.com/shilangyu/typer-go/utils"
	// "github.com/gdamore/tcell"
	// "log"
	// "strconv"
	// "fmt"
	// "time"
)

type setup struct {
	RoomIP, Nickname, Port string
	IsServer               bool
	// Server will be nil if IsServer is false
	Server *socket.Server
	Client socket.Client
}

func main() {
	
	IP := "127.0.0.1"
	setup := setup{IP, "", "9001", true, nil, nil}
	var err error
	setup.Server, err = game.NewServer(setup.Port)
	utils.Check(err)
	setup.Server.Stop()
	
}