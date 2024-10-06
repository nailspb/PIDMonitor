package main

import (
	"bufio"
	"fmt"
	"github.com/TwiN/go-color"
	"github.com/eiannone/keyboard"
	"go.bug.st/serial"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"test/parser"
	"time"
)

func main() {
	fmt.Println(color.InWhiteOverBlack("---------------------------------------------------------"))
	fmt.Println(color.InYellow("Oven pid monitor v1.0"))
	fmt.Println(color.InWhiteOverBlack("---------------------------------------------------------\r\n"))
	port, err := serial.Open(selectPort(), &serial.Mode{
		BaudRate: 115200 * 8,
		DataBits: 8,
		StopBits: serial.OneStopBit,
	})
	if err != nil {
		log.Fatal("not open port: ", err)
	}

	dataChannel := make(chan parser.DataPackage)
	osSignal := make(chan os.Signal, 1)
	signal.Notify(osSignal, os.Interrupt)
	signal.Notify(osSignal, os.Kill)
	//запуск чтения порта в отдельном потоке
	go parser.NewParser(dataChannel).Parse(port)
	//чтение и парсинг данных с порта
	fileName := "./" + strconv.FormatInt(time.Now().Unix(), 10) + "_output.csv"
	for {
		select {
		case dp := <-dataChannel:
			fmt.Printf("%ds | Current=%.3fC | Target=%.0fC | pulse=%d | vp=%.3f | vi=%.3f | vd=%.3f\r\n", dp.Time, float32(dp.CurrentTemp)/1000, float32(dp.TargetTemp)/1000, dp.Pulse, dp.VP, dp.VI, dp.VD)
			f, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE, 666)
			if err != nil {
				PrintErrorf("не удалось открыть файл для записи %s\r\n", fileName)
			} else {
				res := fmt.Sprintf("%d;%d;%d;%d;%f;%f;%f\r\n", dp.Time, dp.CurrentTemp, dp.TargetTemp, dp.Pulse, dp.VP, dp.VI, dp.VD)
				res = strings.ReplaceAll(res, ".", ",") //for excel (rus)
				_, err = f.WriteString(res)
				if err != nil {
					PrintErrorf("не удалось записать данные в файл %s\r\n", fileName)
				}
				_ = f.Close()
			}
		case <-osSignal:
			fmt.Println("Stop signal received")
			err := port.Close()
			if err != nil {
				PrintError("Не удалось корректно закрыть COM порт\r\n")
			}
			fmt.Println("bye, bye...")
			os.Exit(0)
		}

	}

}

func PrintErrorf(format string, a ...any) {
	fmt.Printf(time.DateTime+color.InRedOverBlack(" Ошибка: ")+color.InWhiteOverBlack(format), a...)
}

func PrintError(format string) {
	fmt.Println(time.DateTime + color.InRedOverBlack(" Ошибка: ") + color.InWhiteOverBlack(format))
}

func selectPort() string {
	for {
		ports, err := serial.GetPortsList()
		if err != nil {
			log.Fatal(err)
		}
		if len(ports) == 0 {
			PrintError("В системе не найдены com порты!")
		} else {
			fmt.Println("Выберите номер порта для подключения:")
			for index, port := range ports {
				fmt.Printf("    %d - %v\n", index+1, port)
			}
			var pIndex int
			stdin := bufio.NewReader(os.Stdin)
			_, _ = stdin.Discard(stdin.Buffered())
			if _, err = fmt.Fscanln(stdin, &pIndex); err != nil {
				PrintError("укажите число - порядковый номер порта")
			} else if pIndex < 0 || pIndex > len(ports) {
				PrintError("указан не существующий номер порта")
			} else {
				return ports[pIndex-1]
			}
		}
		fmt.Println("\r\nНажмите любую клавишу для повторного ввода...")
		_, _, _ = keyboard.GetSingleKey()
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		_ = cmd.Run()
	}

}
