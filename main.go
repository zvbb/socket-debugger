package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"socketTest/play"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"

	"fyne.io/fyne/v2/widget"
)

const (
	HeaderLength = 2
	LengthSize   = 4
	TailerLength = 2
)

var conn net.Conn

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
	hostVar.PlaceHolder = "dev.caibeiyun.cn"
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
		var err error
		conn, err = net.Dial("tcp", server)
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

		sendForm, respose := clientSend(parent, conn)
		parentContainer.Append(container.NewTabItem("Client", container.NewVBox(reconnectButton, sendForm, respose)))
	}
	return form
}

func clientSend(parent fyne.Window, conn net.Conn) (*widget.Form, *widget.Entry) {
	// 头部
	headerVar := widget.NewEntry()
	headerVar.SetPlaceHolder("F1F2")

	// 尾部
	tailerVar := widget.NewEntry()
	tailerVar.SetPlaceHolder("F3F4")

	// 长度
	lengthVar := widget.NewEntry()
	lengthVar.SetPlaceHolder("4")

	// 大端模式（小端模式）
	endianVar := widget.NewSelect([]string{"Big", "Small"}, func(value string) {})
	endianVar.Selected = "Big"

	// body
	bodyArea := widget.NewMultiLineEntry()

	var fileBytes []byte

	fileSelector := dialog.NewFileOpen(func(uc fyne.URIReadCloser, err error) {
		if err != nil {
			dialog.ShowError(err, parent)
			return
		}
		if uc == nil {
			return
		}
		fileBytes, _ = io.ReadAll(uc)
		bodyArea.DisableableWidget.Disable()
		bodyArea.SetText(hex.EncodeToString(fileBytes))
	}, parent)

	fileBtn := widget.NewButtonWithIcon("", theme.FileIcon(), func() {
		fileSelector.Show()
	})

	// 相应
	responseArea := widget.NewMultiLineEntry()

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Header", Widget: headerVar},
			{Text: "Tailer", Widget: tailerVar},
			{Text: "Length(byte)", Widget: lengthVar},
			{Text: "Endian", Widget: endianVar},
			{Text: "Body", Widget: bodyArea},
			{Text: "select file", Widget: fileBtn},
		},
		SubmitText: "Send",
	}

	form.OnSubmit = func() {
		buf := socketPack(headerVar, tailerVar, bodyArea, fileBytes, endianVar)

		// 发送构造好的消息
		_, err := conn.Write(buf.Bytes())
		if err != nil {
			fmt.Println("发送消息失败:", err)
			return
		}

		fmt.Println("消息发送成功")

		// 创建bufio.Reader以方便读取
		reader := bufio.NewReader(conn)

		// // 同步响应
		// syncResponse(reader)

		// 流式相应
		streamResponse(reader)

		fmt.Println("文件接收完成")
	}

	return form, responseArea
}

func syncResponse(reader *bufio.Reader) {
	// 读取Header
	header := make([]byte, HeaderLength)
	_, err := io.ReadFull(reader, header)
	if err != nil {
		panic(err)
	}

	lengthBytes := make([]byte, LengthSize)
	_, err = io.ReadFull(reader, lengthBytes)
	if err != nil {
		panic(err)
	}

	// 转换为uint32
	fileSize := binary.BigEndian.Uint32(lengthBytes)

	// 读取文件数据
	fileData := make([]byte, fileSize)

	_, err = io.ReadFull(reader, fileData)
	if err != nil {
		panic(err)
	}

	play.Play(fileData)

	// 读取Tailer
	tailer := make([]byte, TailerLength)
	_, err = io.ReadFull(reader, tailer)
	if err != nil {
		panic(err)
	}

	// 确认Tailer是否正确，以确保数据包接收完成
	expectedTailer, _ := hex.DecodeString("F3F4") // 假设tailer是"F3F4"
	if !bytes.Equal(tailer, expectedTailer) {
		panic("无效的tailer，数据包可能已损坏")
	}
}

func streamResponse(reader *bufio.Reader) {
	// 读取Header
	header := make([]byte, HeaderLength)
	_, err := io.ReadFull(reader, header)
	if err != nil {
		panic(err)
	}

	// 确认Tailer是否正确，以确保数据包接收完成
	expectedTailer, _ := hex.DecodeString("F3F4")

	for {
		lengthBytes := make([]byte, LengthSize)
		_, err = io.ReadFull(reader, lengthBytes)
		if err != nil {
			panic(err)
		}

		// 转换为uint32
		fileSize := binary.BigEndian.Uint32(lengthBytes)
		// 读取文件数据
		fileData := make([]byte, fileSize)

		_, err = io.ReadFull(reader, fileData)
		if err != nil {
			panic(err)
		}

		if !bytes.Equal(fileData, expectedTailer) {
			play.Pcm(fileData)
			continue
		}

		break
	}
}

func socketPack(headerVar, tailerVar, bodyArea *widget.Entry,
	fileBytes []byte,
	endianVar *widget.Select) bytes.Buffer {
	headerStr := headerVar.Text
	if headerStr == "" {
		headerStr = headerVar.PlaceHolder
	}
	header, _ := hex.DecodeString(headerStr)

	var body []byte
	if fileBytes == nil {
		body = []byte(bodyArea.Text)
	} else {
		body = fileBytes
	}
	length := uint32(len(body))

	tailerStr := tailerVar.Text
	if tailerStr == "" {
		tailerStr = tailerVar.PlaceHolder
	}
	trailer, _ := hex.DecodeString(tailerStr)

	var endian binary.ByteOrder

	if endianVar.Selected == "Big" {
		endian = binary.BigEndian
	} else {
		endian = binary.LittleEndian
	}

	var buf bytes.Buffer

	// 写入 Header
	binary.Write(&buf, endian, header)
	// 写入 Length
	binary.Write(&buf, endian, length)
	// 写入 Body
	binary.Write(&buf, endian, body)
	// 写入 Tailer
	binary.Write(&buf, endian, trailer)

	return buf
}

func setHeader() *fyne.Container {
	radio := widget.NewRadioGroup([]string{"Tcp", "Udp"}, func(s string) {})
	radio.Required = true
	radio.Selected = "Tcp"
	radio.Horizontal = true

	header := container.New(layout.NewCenterLayout(), radio)
	return header
}
