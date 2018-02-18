package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"time"

	"io/ioutil"

	"github.com/gdamore/tcell"
	"github.com/google/gopacket"
	"github.com/sachaos/tview"
)

const (
	TailMode = iota
	SelectMode
)

type Tcpterm struct {
	src        *gopacket.PacketSource
	view       *tview.Application
	primitives []tview.Primitive
	table      *tview.Table
	detail     *tview.TextView
	dump       *tview.TextView
	frame      *tview.Frame
	packets    []gopacket.Packet
	mode       int

	logger *log.Logger
}

const (
	timestampFormt = "2006-01-02 15:04:05.000000"
)

func NewTcpterm(src *gopacket.PacketSource, debug bool) *Tcpterm {
	view := tview.NewApplication()

	packetList := preparePacketList()
	packetDetail := preparePacketDetail()
	packetDump := preparePacketDump()

	layout := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(packetList, 0, 1, true).
		AddItem(packetDetail, 0, 1, false).
		AddItem(packetDump, 0, 1, false)
	frame := prepareFrame(layout)

	view.SetRoot(frame, true).SetFocus(packetList)

	var w io.Writer
	if debug {
		w = os.Stderr
	} else {
		w = ioutil.Discard
	}

	app := &Tcpterm{
		src:        src,
		view:       view,
		primitives: []tview.Primitive{packetList, packetDetail, packetDump},
		table:      packetList,
		detail:     packetDetail,
		dump:       packetDump,
		frame:      frame,
		logger:     log.New(w, "[tcpterm]", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile),
	}
	app.SwitchToTailMode()

	view.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlC {
			app.Stop()
		}

		if event.Key() == tcell.KeyTAB {
			app.rotateView()
		}
		return event
	})

	packetList.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEsc {
			app.SwitchToTailMode()
		}

		if key == tcell.KeyEnter {
			app.SwitchToSelectMode()
		}
	})

	packetList.SetSelectionChangedFunc(func(row int, column int) {
		app.displayDetailOf(row)
	})

	return app
}

func (app *Tcpterm) PacketListGenerator(refreshTrigger chan bool) {
	cnt := 0
	for {
		packet, err := app.src.NextPacket()
		if err == io.EOF {
			return
		} else if err == nil {
			cnt++
			rowCount := app.table.GetRowCount()

			app.logger.Printf("count: %v start\n", cnt)

			app.table.SetCell(rowCount, 0, tview.NewTableCell(strconv.Itoa(cnt)))
			app.table.SetCell(rowCount, 1, tview.NewTableCell(packet.Metadata().Timestamp.Format(timestampFormt)))
			app.table.SetCell(rowCount, 2, tview.NewTableCell(flowOf(packet)))
			app.table.SetCell(rowCount, 3, tview.NewTableCell(strconv.Itoa(packet.Metadata().Length)))
			app.table.SetCell(rowCount, 4, tview.NewTableCell(packet.Layers()[1].LayerType().String()))
			if len(packet.Layers()) > 2 {
				app.table.SetCell(rowCount, 5, tview.NewTableCell(packet.Layers()[2].LayerType().String()))
			}

			app.packets = append(app.packets, packet)

			app.logger.Printf("count: %v end\n", cnt)
		}
	}
}

func (app *Tcpterm) Ticker(refreshTrigger chan bool) {
	for {
		time.Sleep(100 * time.Millisecond)
		refreshTrigger <- true
	}
}

func (app *Tcpterm) Refresh(refreshTrigger chan bool) {
	for {
		_, ok := <-refreshTrigger
		if ok {
			app.view.Draw()
		}
	}
}

func (app *Tcpterm) Run() {
	refreshTrigger := make(chan bool)

	go app.PacketListGenerator(refreshTrigger)
	go app.Ticker(refreshTrigger)
	go app.Refresh(refreshTrigger)

	if app.view.Run(); err != nil {
		panic(err)
	}
}

func (app *Tcpterm) Stop() {
	app.view.Stop()
}

func (app *Tcpterm) SwitchToTailMode() {
	app.mode = TailMode

	app.table.SetSelectable(false, false)
	app.table.ScrollToEnd()

	app.frame.Clear().AddText("**Tail**", true, tview.AlignLeft, tcell.ColorGreen)

	app.frame.AddText("g: page top, G: page end, TAB: rotate panel, Enter: Detail mode", true, tview.AlignRight, tcell.ColorDefault)
}

func (app *Tcpterm) SwitchToSelectMode() {
	app.mode = SelectMode

	app.table.SetSelectable(true, false)
	row, _ := app.table.GetOffset()
	app.table.Select(row+1, 0)
	app.displayDetailOf(row + 1)

	app.frame.Clear().AddText("*Detail*", true, tview.AlignLeft, tcell.ColorBlue)
	app.frame.AddText("g: page top, G: page end, TAB: rotate panel, ECS: Tail mode", true, tview.AlignRight, tcell.ColorDefault)
}

func (app *Tcpterm) displayDetailOf(row int) {
	if row < 1 || row > len(app.packets) {
		return
	}

	app.detail.Clear().ScrollToBeginning()
	app.dump.Clear().ScrollToBeginning()

	packet := app.packets[row-1]

	fmt.Fprint(app.detail, packet.String())
	fmt.Fprint(app.dump, packet.Dump())
}

func (app *Tcpterm) rotateView() {
	idx, err := app.findPrimitiveIdx(app.view.GetFocus())
	if err != nil {
		panic(err)
	}

	nextIdx := idx + 1
	if nextIdx >= len(app.primitives) {
		nextIdx = 0
	}
	app.view.SetFocus(app.primitives[nextIdx])
}

func (app *Tcpterm) findPrimitiveIdx(p tview.Primitive) (int, error) {
	for i, primitive := range app.primitives {
		if p == primitive {
			return i, nil
		}
	}
	return 0, errors.New("Primitive not found")
}

func flowOf(packet gopacket.Packet) string {
	if packet.NetworkLayer() == nil {
		return "-"
	} else {
		return packet.NetworkLayer().NetworkFlow().String()
	}
}
