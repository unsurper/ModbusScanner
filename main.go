package main

import (
	"flag"
	"fmt"
	"github.com/fatih/color"
	"github.com/goburrow/modbus"
	"github.com/gosuri/uitable"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"time"
)

const (
	Success = "Successful communication"
)

var (
	Address  uint16 = 1
	Quantity uint16 = 1
	// SerialNamePtr 串口号 BaudPtr 波特率 ReadTimeoutPtr 读取时间
	ModPtr         = flag.String("m", "RTU", "MOD (TCP/RTU)")
	SerialNamePtr  = flag.String("sn", "COM1", "SerialName (COM1.../localhost:502)")
	BaudPtr        = flag.Int("b", 9600, "Baud (1200/2400/4800/9600...)")
	ReadTimeoutPtr = flag.Duration("rt", 2000000000, "ReadTimeout (1s/2s...)")
	SlaveIdPtrA    = flag.Uint("ida", 1, "SlaveIdA")
	SlaveIdPtrB    = flag.Uint("idb", 3, "SlaveIdB")
	DataBitsPtr    = flag.Int("d", 8, "Data bits: 5, 6, 7 or 8")
	ParityPtr      = flag.String("p", "N", "Parity: N - None, E - Even, O - Odd")
	StopBitsPtr    = flag.Int("sb", 1, "Stop bits: 1 or 2")
)

func init() {
	flag.Usage = func() {
		fmt.Println("Usage:ModbusScannerDemo")
		flag.PrintDefaults()
	}
	flag.Parse()

	log.SetFormatter(&log.TextFormatter{ForceColors: true, FullTimestamp: true})
	path := "log/GBlog"
	// 下面配置日志每隔 10 分钟轮转一个新文件，保留最近 3 分钟的日志文件，多余的自动清理掉。
	writer, err := rotatelogs.New(
		path+".%Y%m%d%H%M"+".log",
		rotatelogs.WithLinkName(path),
		//rotatelogs.WithMaxAge(time.Duration(180)*time.Second),
		rotatelogs.WithRotationTime(time.Duration(600)*time.Second),
	)
	writers := []io.Writer{writer, os.Stdout}
	fileAndStdoutWriter := io.MultiWriter(writers...)
	if err == nil {
		log.SetOutput(fileAndStdoutWriter)
	} else {
		log.Info("failed to log to file.")
	}

}

func main() {
	//Table Init
	Table := uitable.New()
	Table.MaxColWidth = 50
	Table.AddRow("ADDRESS", "ReadCoils", "ReadDiscreteInputs", "ReadInputRegisters", "ReadHoldingRegisters")
	// Modbus RTU/ASCII
	if *ModPtr == "RTU" {
		handler := modbus.NewRTUClientHandler(*SerialNamePtr)
		handler.BaudRate = *BaudPtr
		if *DataBitsPtr < 5 || *DataBitsPtr > 8 {
			log.Errorln("DataBitsPtr ERR :", *DataBitsPtr)
			return
		}
		handler.DataBits = *DataBitsPtr
		if *ParityPtr != "N" && *ParityPtr != "E" && *ParityPtr != "O" {
			log.Errorln("ParityPtr ERR :", *ParityPtr)
			return
		}
		handler.Parity = *ParityPtr
		if *StopBitsPtr < 1 || *StopBitsPtr > 2 {
			log.Errorln("StopBitsPtr ERR :", *StopBitsPtr)
			return
		}
		handler.StopBits = *StopBitsPtr
		handler.Timeout = *ReadTimeoutPtr
		if *SlaveIdPtrA > 255 || *SlaveIdPtrB > 255 {
			log.Errorln("SlaveIdPtr ERR :", byte(*SlaveIdPtrA), byte(*SlaveIdPtrB))
			return
		}
		for SlaveId := *SlaveIdPtrA; SlaveId <= *SlaveIdPtrB; SlaveId++ {
			handler.SlaveId = byte(SlaveId)
			err := handler.Connect()
			if err != nil {
				log.Errorln("Connect ERR :", err)
			}
			defer handler.Close()
			Scanner(handler, SlaveId, Table)
		}
	} else if *ModPtr == "TCP" {
		handler := modbus.NewTCPClientHandler(*SerialNamePtr)
		handler.Timeout = *ReadTimeoutPtr
		if *SlaveIdPtrA > 255 || *SlaveIdPtrB > 255 {
			log.Errorln("SlaveIdPtr ERR :", byte(*SlaveIdPtrA), byte(*SlaveIdPtrB))
			return
		}
		for SlaveId := *SlaveIdPtrA; SlaveId <= *SlaveIdPtrB; SlaveId++ {
			handler.SlaveId = byte(SlaveId)
			err := handler.Connect()
			if err != nil {
				log.Errorln("Connect ERR :", err)
			}
			defer handler.Close()
			Scanner(handler, SlaveId, Table)
		}
	} else {
		log.Errorln("ModPtr ERR :", *ModPtr)
		return
	}
	fmt.Println(Table)
}

func Scanner(handler modbus.ClientHandler, SlaveId uint, Table *uitable.Table) {
	//创建客户端
	client := modbus.NewClient(handler)
	var tableflag []interface{}
	tableflag = append(tableflag, SlaveId)
	result, err := client.ReadCoils(Address, Quantity)
	tableflag = SlaveError(result, "ReadCoils", SlaveId, err, tableflag)
	result, err = client.ReadDiscreteInputs(Address, Quantity)
	tableflag = SlaveError(result, "ReadDiscreteInputs", SlaveId, err, tableflag)
	result, err = client.ReadInputRegisters(Address, Quantity)
	tableflag = SlaveError(result, "ReadInputRegisters", SlaveId, err, tableflag)
	result, err = client.ReadHoldingRegisters(Address, Quantity)
	tableflag = SlaveError(result, "ReadHoldingRegisters", SlaveId, err, tableflag)
	Table.AddRow(tableflag[0], tableflag[1], tableflag[2], tableflag[3], tableflag[4])
}

func SlaveError(result []byte, modname string, SlaveId uint, err error, Table []interface{}) []interface{} {
	if err != nil {
		log.Errorln("SlaveId : ", SlaveId, modname, "ERR :", err)
		Table = append(Table, color.RedString("Fail"))
	} else {
		log.Infoln("SlaveId :", SlaveId, Success, "Result:", result)
		Table = append(Table, color.GreenString("Success"))
	}
	return Table
}
