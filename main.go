package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"net"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"

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
		conn, err := net.Dial("tcp", server)
		if err != nil {
			// 显示连接失败的弹框
			dialog.ShowError(
				errors.New("the connection failed"), // 修复后的错误信息
				parent,                              // 父窗口
			)
			return
		}
		form.Hide()

		reconnectButton := widget.NewButtonWithIcon("Reconnect", theme.MediaPauseIcon(), func() {
			parentContainer.RemoveIndex(0)
			parentContainer.Append(container.NewTabItem("Client", form))
		})

		parentContainer.RemoveIndex(0)

		sendForm, respose := clientSend(conn)
		parentContainer.Append(container.NewTabItem("Client", container.NewVBox(reconnectButton, sendForm, respose)))
	}
	return form
}

func clientSend(conn net.Conn) (*widget.Form, *widget.Entry) {
	headerVar := widget.NewEntry()
	tailerVar := widget.NewEntry()
	lengthVar := widget.NewEntry()
	endianVar := widget.NewSelect([]string{"Big", "Small"}, func(value string) {})
	endianVar.Selected = "Big"
	// 创建一个大的文本框
	textArea := widget.NewMultiLineEntry()
	responseArea := widget.NewMultiLineEntry()

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Header", Widget: headerVar},
			{Text: "Tailer", Widget: tailerVar},
			{Text: "Length", Widget: lengthVar},
			{Text: "Endian", Widget: endianVar},
			{Text: "Body", Widget: textArea},
		},
		SubmitText: "Send",
	}

	form.OnSubmit = func() {
		// 将字符串转换为字节序列
		header := []byte(headerVar.Text)
		trailer := []byte(tailerVar.Text)
		body := []byte(textArea.Text) // 假设 utf8.encodeRuneInBytes 是一个函数，将字符串转换为字节序列
		// Length 是 Body 的长度
		length := uint32(len(body))

		var endian binary.ByteOrder

		// Endian 决定了字节序，这里假设是 big endian
		var buf bytes.Buffer

		// 写入 Length，转换为 big endian
		if endianVar.Selected == "Big" {
			endian = binary.BigEndian
		} else {
			endian = binary.LittleEndian
		}

		// 写入 Header
		binary.Write(&buf, endian, header)
		// 写入 Length
		binary.Write(&buf, endian, length)
		// 写入 Body
		binary.Write(&buf, endian, body)
		// 写入 Tailer
		binary.Write(&buf, endian, trailer)
		// 发送构造好的消息
		_, err := conn.Write(buf.Bytes())
		if err != nil {
			fmt.Println("发送消息失败:", err)
			return
		}

		fmt.Println("消息发送成功")
	}

	return form, responseArea
}

func setHeader() *fyne.Container {
	radio := widget.NewRadioGroup([]string{"Tcp", "Udp"}, func(s string) {})
	radio.Required = true
	radio.Selected = "Tcp"
	radio.Horizontal = true

	header := container.New(layout.NewCenterLayout(), radio)
	return header
}
