package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

//日志信息种类
type MsgType uint8
const (
	Error = iota
	Info
	Warning
)

//日志记录
func Logs(level MsgType, msg string, pos string) {
	var msgType string;var text string
	switch level {
	case Error:	msgType = "Error"
	case Info:	msgType = "Info"
	case Warning:	msgType = "Warning"
	default: msgType = "Others"
	}

	curdir,_ := os.Getwd()
	currentTime := time.Now()
	file, _ := os.OpenFile(filepath.Join(curdir,"/logs/" + currentTime.Format("2006-01-02") + ".log"), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	defer file.Close()

	if pos == ""{
		text = fmt.Sprintf("%s \t %s：%s \r\n", currentTime.Format("2006-01-02 15:04:05"), msgType,msg)
	}else{
		text = fmt.Sprintf("%s \t %s：%s \t Pos：%s \r\n", currentTime.Format("2006-01-02 15:04:05"), msgType, msg,pos)
	}
	file.Write([]byte(text))
}

