package main

import (
	"flag"
	"fmt"
	"github.com/goburrow/modbus"
)

const (
	Success = "Successful communication"
)

var (
	Address  uint16 = 1
	Quantity uint16 = 1
	// SerialNamePtr 串口号 BaudPtr 波特率 ReadTimeoutPtr 读取时间
	ModPtr         = flag.String("m", "RTU", "MOD")
	SerialNamePtr  = flag.String("sn", "COM1", "SerialName")
	BaudPtr        = flag.Int("b", 2400, "Baud")
	ReadTimeoutPtr = flag.Duration("rt", 3000000000, "ReadTimeout")
	SlaveIdPtrA    = flag.Uint("ida", 1, "SlaveIdA")
	SlaveIdPtrB    = flag.Uint("idb", 10, "SlaveIdB")
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
}

func main() {
	// Modbus RTU/ASCII
	if *ModPtr == "RTU" {
		handler := modbus.NewRTUClientHandler(*SerialNamePtr)
		handler.BaudRate = *BaudPtr
		if *DataBitsPtr < 5 || *DataBitsPtr > 8 {
			fmt.Println("DataBitsPtr ERR :", *DataBitsPtr)
			return
		}
		handler.DataBits = *DataBitsPtr
		if *ParityPtr != "N" && *ParityPtr != "E" && *ParityPtr != "O" {
			fmt.Println("ParityPtr ERR :", *ParityPtr)
			return
		}
		handler.Parity = *ParityPtr
		if *StopBitsPtr < 1 || *StopBitsPtr > 2 {
			fmt.Println("StopBitsPtr ERR :", *StopBitsPtr)
			return
		}
		handler.StopBits = *StopBitsPtr
		handler.Timeout = *ReadTimeoutPtr
		if *SlaveIdPtrA > 255 || *SlaveIdPtrB > 255 {
			fmt.Println("SlaveIdPtr ERR :", byte(*SlaveIdPtrA), byte(*SlaveIdPtrB))
			return
		}
		for i := *SlaveIdPtrA; i <= *SlaveIdPtrB; i++ {
			handler.SlaveId = byte(i)
			err := handler.Connect()
			if err != nil {
				fmt.Println("Connect ERR :", err)
			}
			defer handler.Close()
			client := modbus.NewClient(handler)
			_, err = client.ReadCoils(Address, Quantity)
			if err != nil {
				fmt.Println("SlaveId : ", i, "ReadCoils ERR :", err)
			} else {
				fmt.Println("SlaveId : ", i, Success)
			}
			_, err = client.ReadDiscreteInputs(Address, Quantity)
			if err != nil {
				fmt.Println("SlaveId : ", i, "ReadDiscreteInputs ERR :", err)
			} else {
				fmt.Println("SlaveId : ", i, Success)
			}
			_, err = client.ReadInputRegisters(Address, Quantity)
			if err != nil {
				fmt.Println("SlaveId : ", i, "ReadInputRegisters ERR :", err)
			} else {
				fmt.Println("SlaveId : ", i, Success)
			}
			_, err = client.ReadHoldingRegisters(Address, Quantity)
			if err != nil {
				fmt.Println("SlaveId : ", i, "ReadHoldingRegisters ERR :", err)
			} else {
				fmt.Println("SlaveId : ", i, Success)
			}
		}
	} else if *ModPtr == "TCP" {
		handler := modbus.NewTCPClientHandler(*SerialNamePtr)
		handler.Timeout = *ReadTimeoutPtr
		if *SlaveIdPtrA > 255 || *SlaveIdPtrB > 255 {
			fmt.Println("SlaveIdPtr ERR :", byte(*SlaveIdPtrA), byte(*SlaveIdPtrB))
			return
		}
		for i := *SlaveIdPtrA; i <= *SlaveIdPtrB; i++ {
			handler.SlaveId = byte(i)
			err := handler.Connect()
			if err != nil {
				fmt.Println("Connect ERR :", err)
			}
			defer handler.Close()
			client := modbus.NewClient(handler)
			_, err = client.ReadCoils(Address, Quantity)
			if err != nil {
				fmt.Println("SlaveId : ", i, "ReadCoils ERR :", err)
			} else {
				fmt.Println("SlaveId : ", i, Success)
			}
			_, err = client.ReadDiscreteInputs(Address, Quantity)
			if err != nil {
				fmt.Println("SlaveId : ", i, "ReadDiscreteInputs ERR :", err)
			} else {
				fmt.Println("SlaveId : ", i, Success)
			}
			_, err = client.ReadInputRegisters(Address, Quantity)
			if err != nil {
				fmt.Println("SlaveId : ", i, "ReadInputRegisters ERR :", err)
			} else {
				fmt.Println("SlaveId : ", i, Success)
			}
			_, err = client.ReadHoldingRegisters(Address, Quantity)
			if err != nil {
				fmt.Println("SlaveId : ", i, "ReadHoldingRegisters ERR :", err)
			} else {
				fmt.Println("SlaveId : ", i, Success)
			}
		}
	} else {
		fmt.Println("ModPtr ERR :", *ModPtr)
		return
	}
}
