package main

import (
	"math/rand"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func main() {
	var name string
	var user Client

	a := app.New()
	w := a.NewWindow("VChat")
	ipw := widget.NewEntry()
	ipw.SetPlaceHolder("IP:PORT")
	namew := widget.NewEntry()
	rand.Seed(time.Now().UnixNano())
	namew.SetText("Anon" + strconv.Itoa(rand.Intn(10000)))
	channelw := widget.NewEntry()
	channelw.SetText("chat")

	data := []string{"--- VCHAT ---"}

	list := widget.NewList(
		func() int {
			return len(data)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(data[i])
		})

	entry := widget.NewEntry()
	addtolist := func(text string) {
		data = append(data, text)
		entry.SetText("")
		list.ScrollToBottom()
		list.Refresh()
	}
	send := func(text string) {
		err := user.Send(text, user.GetChannel())
		if err != nil {
			panic(err)
		}
		addtolist(name + ": " + text)
	}

	button := widget.NewButtonWithIcon("", theme.MailSendIcon(), func() {
		send(entry.Text)
	})

	entry.OnSubmitted = func(text string) {
		send(text)
	}
	entry.SetPlaceHolder("Message to #" + user.GetChannel())

	split := container.NewVSplit(
		list,
		container.NewBorder(nil, nil, nil, button, entry),
	)
	split.SetOffset(1)

	form := widget.NewForm(
		&widget.FormItem{Text: "IP", Widget: ipw},
		&widget.FormItem{Text: "Name", Widget: namew},
		&widget.FormItem{Text: "Channel", Widget: channelw},
		&widget.FormItem{Text: "", Widget: widget.NewButton("Connect", func() {
			name = namew.Text
			user = Client{Name: namew.Text, IP: "ws://" + ipw.Text}
			err := user.Connect()
			if err != nil {
				panic(err)
			}
			w.SetContent(split)
			go user.Listen(func(msg Msg, err error) {
				if err != nil {
					panic(err)
				}
				addtolist(msg.Name + ": " + msg.Message)
			})
		})},
	)

	w.SetContent(form)

	w.ShowAndRun()

}
