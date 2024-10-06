package parser

import (
	"log"
	"unsafe"
)

type DataPackage struct {
	Time        uint32
	CurrentTemp int32
	TargetTemp  int32
	Pulse       int32
	VP          float32
	VI          float32
	VD          float32
	_           uint32
}

func mapData(data []byte) DataPackage {
	//b := []byte{1, 2, 3, 4, 0x42, 0xf6, 0, 0}
	//a := math.Float32frombits(binary.BigEndian.Uint32(b[4:]))
	//a := binary.BigEndian.Uint32(b[4:])
	//fmt.Printf("0x%.8x", a)
	//fmt.Printf("%v", a)
	return DataPackage{
		Time:        *((*uint32)(unsafe.Pointer(&data[0]))),
		CurrentTemp: *((*int32)(unsafe.Pointer(&data[4]))),
		TargetTemp:  *((*int32)(unsafe.Pointer(&data[8]))),
		Pulse:       *((*int32)(unsafe.Pointer(&data[12]))),
		VP:          *((*float32)(unsafe.Pointer(&data[16]))),
		VI:          *((*float32)(unsafe.Pointer(&data[20]))),
		VD:          *((*float32)(unsafe.Pointer(&data[24]))),
	}
}

type Data struct {
	length   int
	position int
	buffer   []byte
	channel  chan DataPackage
}

type PortReader interface {
	Read(buf []byte) (int, error)
	ResetInputBuffer() error
}

func NewParser(channel chan DataPackage) *Data {
	return &Data{
		position: 0,
		length:   int(unsafe.Sizeof(DataPackage{})),
		buffer:   make([]byte, int(unsafe.Sizeof(DataPackage{}))),
		channel:  channel,
	}
}

func (data *Data) Parse(p PortReader) {
	//очистить не прочитанные с порта данные
	err := p.ResetInputBuffer()
	if err != nil {
		log.Println(err)
	}
	//начать чтение данных
	for {
		buff := make([]byte, data.length)
		pos := 0
		count, err := p.Read(buff)
		if err != nil {
			log.Fatal(err)
		} else {
			if count > data.length {
				log.Fatal("Data length is greater than length")
			} else {
				for pos < count {
					data.buffer[data.position] = buff[pos]
					pos++
					data.position++
					if data.position == data.length {
						data.position = 0
						data.channel <- mapData(data.buffer)
					}
				}
			}
		}
	}
}
