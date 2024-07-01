package main

import (
	"errors"
	"fmt"
	"net"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"

	"fyne.io/fyne/v2/widget"
)

func main() {
	myApp := app.New()

	mainWin := myApp.NewWindow("socket debugger")
	mainWin.CenterOnScreen()
	mainWin.Resize(fyne.NewSize(1000, 600))

	header := setHeader()

	tabs := container.NewAppTabs()
	
	form := client(mainWin, tabs)
	tabs.Append(container.NewTabItem("Client", form))
	// tabs.Append(container.NewTabItem("Server", ""))

	tabs.SelectIndex(0)

	tabs.SetTabLocation(container.TabLocationLeading)

	mainWin.SetContent(container.New(layout.NewVBoxLayout(), header, tabs))
	mainWin.ShowAndRun()
}

func client(parent fyne.Window, parentContainer *container.AppTabs) *widget.Form {
	hostVar := widget.NewEntry()
	portVar := widget.NewEntry()

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Host", Widget: hostVar},
			{Text: "Port", Widget: portVar},
		},
		SubmitText: "Connect",
	}
	form.OnSubmit = func() {
		// 定义服务器地址和端口
		server := fmt.Sprintf("%s:%s", hostVar.Text, portVar.Text)

		// 连接服务器
		_, err := net.Dial("tcp", server)
		if err == nil {
			// 显示连接失败的弹框
			dialog.ShowError(
				errors.New("the connection failed"), // 修复后的错误信息
				parent,                              // 父窗口
			)
			return
		}
		dialog.ShowInformation("Connection Success", "The connection is successful", parent)
		form.Hide()
	}
	return form
}

func setHeader() *fyne.Container {
	radio := widget.NewRadioGroup([]string{"Tcp", "Udp"}, func(s string) {})
	radio.Required = true
	radio.Selected = "Tcp"
	radio.Horizontal = true

	header := container.New(layout.NewCenterLayout(), radio)
	return header
}
