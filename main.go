package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"github.com/goburrow/modbus"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/olekukonko/tablewriter"
	log "github.com/sirupsen/logrus"
	"io"
	"math"
	"os"
	"strconv"
	"time"
)

const (
	Success = "Successful communication"
)

var (
	//Address 开始地址位 Quantity 读取寄存器数量
	Address  = flag.Int("a", 0, "Address")
	Quantity = flag.Int("q", 8, "Quantity")
	// SerialNamePtr 串口号 BaudPtr 波特率 ReadTimeoutPtr 读取时间
	ModPtr         = flag.String("m", "RTU", "MOD (TCP/RTU)")
	DataTypePtr    = flag.String("dt", "uint16", "uint16, int16, uint32, int32, float32, float64...")
	SerialNamePtr  = flag.String("sn", "COM1", "SerialName (COM1.../localhost:502)")
	BaudPtr        = flag.Int("b", 9600, "Baud (1200/2400/4800/9600...)")
	ReadTimeoutPtr = flag.Duration("rt", 2000000000, "ReadTimeout (1s/2s...)")
	SlaveIdPtrA    = flag.Uint("ida", 1, "SlaveIdA")
	SlaveIdPtrB    = flag.Uint("idb", 3, "SlaveIdB")
	DataBitsPtr    = flag.Int("d", 8, "Data bits: 5, 6, 7 or 8")
	ParityPtr      = flag.String("p", "N", "Parity: N - None, E - Even, O - Odd")
	StopBitsPtr    = flag.Int("sb", 1, "Stop bits: 1 or 2")
	SuccessDevice  string
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
	start := time.Now() // 获取当前时间
	//Table Init
	Table := tablewriter.NewWriter(os.Stdout)
	header := []string{"ADDRESS", "ReadInputRegisters 3x", "ReadHoldingRegisters 4x"}
	Table.SetHeader(header)
	Table.SetAutoMergeCellsByColumnIndex([]int{0})
	Table.SetRowLine(true)

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
			Scanner(modbus.RTUClient(*SerialNamePtr), SlaveId, Table)
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
			Scanner(modbus.TCPClient(*SerialNamePtr), SlaveId, Table)
		}
	} else {
		log.Errorln("ModPtr ERR :", *ModPtr)
		return
	}

	Table.SetFooter([]string{"", "Time-consuming", time.Since(start).String()})
	log.Infoln("SuccessDevice:", SuccessDevice)
	Table.Render()
}

func Scanner(client modbus.Client, SlaveId uint, Table *tablewriter.Table) {
	//创建客户端
	var tableflag []string
	var tabletwins []string
	tableflag = append(tableflag, strconv.Itoa(int(SlaveId)))
	tabletwins = append(tabletwins, strconv.Itoa(int(SlaveId)))
	result, err := client.ReadInputRegisters(uint16(*Address), uint16(*Quantity))
	tableflag, tabletwins = SlaveError(result, "ReadInputRegisters", SlaveId, err, tableflag, tabletwins)
	result, err = client.ReadHoldingRegisters(uint16(*Address), uint16(*Quantity))
	tableflag, tabletwins = SlaveError(result, "ReadHoldingRegisters", SlaveId, err, tableflag, tabletwins)
	Table.Append(tableflag)
	Table.Append(tabletwins)
}

func SlaveError(result []byte, modname string, SlaveId uint, err error, Table []string, Twins []string) ([]string, []string) {
	if err != nil {
		log.Errorln("SlaveId : ", SlaveId, modname, "ERR :", err)
		Table = append(Table, "Fail")
		Twins = append(Twins, err.Error())
	} else {
		SuccessDevice += strconv.FormatUint(uint64(SlaveId), 10) + ","
		log.Infoln("SlaveId :", SlaveId, Success, "Result:", TransferData(result, *DataTypePtr))
		Table = append(Table, "Success")
		//TODO: result data to string.
		Twins = append(Twins, TransferData(result, *DataTypePtr))
	}
	return Table, Twins
}

func TransferData(value []byte, dataType string) string {
	switch dataType {
	case "uint16":
		data := binary.BigEndian.Uint16(value)
		return strconv.FormatUint(uint64(data), 10)
	case "int16":
		data := int16(binary.BigEndian.Uint16(value))
		return strconv.FormatInt(int64(data), 10)
	case "uint32":
		data := binary.BigEndian.Uint32(value)
		return strconv.FormatUint(uint64(data), 10)
	case "int32":
		data := int32(binary.BigEndian.Uint32(value))
		return strconv.FormatInt(int64(data), 10)
	case "uint64":
		data := binary.BigEndian.Uint64(value)
		return strconv.FormatUint(data, 10)
	case "int64":
		data := int32(binary.BigEndian.Uint64(value))
		return strconv.FormatInt(int64(data), 10)
	case "float32":
		bits := binary.BigEndian.Uint32(value)
		f := math.Float32frombits(bits)
		return strconv.FormatFloat(float64(f), 'f', -1, 32)
	case "float64":
		bits := binary.BigEndian.Uint64(value)
		f := math.Float64frombits(bits)
		return strconv.FormatFloat(f, 'f', -1, 64)
	default:
		return ""
	}
}
