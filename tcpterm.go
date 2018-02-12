package main

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/gdamore/tcell"
	"github.com/google/gopacket"
	"github.com/sachaos/tview"
)

type Tcpterm struct {
	src        *gopacket.PacketSource
	view       *tview.Application
	primitives []tview.Primitive
	table      *tview.Table
	detail     *tview.TextView
	dump       *tview.TextView
	packets    []gopacket.Packet
}

const (
	timestampFormt = "2006-01-02 15:04:03.000000"
)

func NewTcpterm(src *gopacket.PacketSource) *Tcpterm {
	view := tview.NewApplication()

	packetList := preparePacketList()
	packetDetail := preparePacketDetail()
	packetDump := preparePacketDump()

	layout := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(packetList, 0, 1, true).
		AddItem(packetDetail, 0, 1, false).
		AddItem(packetDump, 0, 1, false)

	view.SetRoot(layout, true).SetFocus(packetList)

	app := &Tcpterm{
		src:        src,
		view:       view,
		primitives: []tview.Primitive{packetList, packetDetail, packetDump},
		table:      packetList,
		detail:     packetDetail,
		dump:       packetDump,
	}

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
		if key == tcell.KeyEnter {
			packetList.SetSelectable(true, false)
			row, column := packetList.GetOffset()
			packetList.Select(row+1, column)
			app.displayDetailOf(row + 1)
		}

		if key == tcell.KeyEsc {
			packetList.SetSelectable(false, false)
		}
	})

	packetList.SetSelectionChangedFunc(func(row int, column int) {
		app.displayDetailOf(row)
	})

	return app
}

func (app *Tcpterm) Run() {
	go func() {
		cnt := 0
		for packet := range app.src.Packets() {
			cnt++
			rowCount := app.table.GetRowCount()

			flow := packet.NetworkLayer().NetworkFlow()
			app.table.SetCell(rowCount, 0, tview.NewTableCell(strconv.Itoa(cnt)))
			app.table.SetCell(rowCount, 1, tview.NewTableCell(packet.Metadata().Timestamp.Format(timestampFormt)))
			app.table.SetCell(rowCount, 2, tview.NewTableCell(flow.String()))
			app.table.SetCell(rowCount, 3, tview.NewTableCell(strconv.Itoa(packet.Metadata().Length)))
			app.table.SetCell(rowCount, 4, tview.NewTableCell(packet.LinkLayer().LayerType().String()))
			app.table.SetCell(rowCount, 5, tview.NewTableCell(packet.NetworkLayer().LayerType().String()))
			app.table.SetCell(rowCount, 6, tview.NewTableCell(packet.TransportLayer().LayerType().String()))

			app.packets = append(app.packets, packet)
			app.view.Draw()
		}
	}()

	if app.view.Run(); err != nil {
		panic(err)
	}
}

func (app *Tcpterm) Stop() {
	app.view.Stop()
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
