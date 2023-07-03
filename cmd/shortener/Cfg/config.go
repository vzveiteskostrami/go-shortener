package config

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

var (
	Addresses   InOutAddresses
	FileStorage FileStorageAttr
)

type NetAddress struct {
	Host string
	Port int
}

func (na *NetAddress) String() string {
	return na.Host + ":" + strconv.Itoa(na.Port)
}

func (na *NetAddress) Set(flagValue string) error {
	var err error
	na.Host, na.Port, err = getAddrAndPort(flagValue)
	return err
}

type FileStorageAttr struct {
	FileName string
}

func (fs *FileStorageAttr) String() string {
	return fs.FileName
}

func (fs *FileStorageAttr) Set(fn string) error {
	fs.FileName = fn
	return nil
}

func getAddrAndPort(s string) (string, int, error) {
	var err error
	h := ""
	p := int(0)
	args := strings.Split(s, ":")
	if len(args) == 2 || len(args) == 3 {
		if len(args) == 3 {
			h = args[0] + ":" + args[1]
			args[1] = args[2]
		} else {
			h = args[0]
		}

		if args[1] == "" {
			return "", -1, errors.New("неверный формат строки, требуется host:port")
		}
		p, err = strconv.Atoi(args[1])
		if err != nil {
			return "", -1, errors.New("неверный номер порта, " + err.Error())
		}
	} else {
		return "", -1, errors.New("неверный формат строки, требуется host:port")
	}
	return h, p, nil
}

type InOutAddresses struct {
	In  *NetAddress
	Out *NetAddress
}

func ReadData() {
	Addresses.In = new(NetAddress)
	Addresses.Out = new(NetAddress)
	Addresses.In.Host = "localhost"
	Addresses.In.Port = 8080
	Addresses.Out.Host = "http://127.0.0.1"
	Addresses.Out.Port = 8080

	FileStorage.FileName = `/tmp/short-url-db.json`

	_ = flag.Value(Addresses.In)
	flag.Var(Addresses.In, "a", "In net address host:port")
	_ = flag.Value(Addresses.Out)
	flag.Var(Addresses.Out, "b", "Out net address host:port")
	_ = flag.Value(&FileStorage)
	flag.Var(&FileStorage, "f", "Storage file name")

	flag.Parse()

	var err error
	s := os.Getenv("SERVER_ADDRESS")
	if s != "" {
		Addresses.In.Host, Addresses.In.Port, err = getAddrAndPort(s)
		if err != nil {
			fmt.Println("Неудачный парсинг переменной окружения SERVER_ADDRESS")
		}
	}
	s = os.Getenv("BASE_URL")
	if s != "" {
		Addresses.Out.Host, Addresses.In.Port, err = getAddrAndPort(s)
		if err != nil {
			fmt.Println("Неудачный парсинг переменной окружения BASE_URL")
		}
	}
	s = os.Getenv("FILE_STORAGE_PATH")
	if s != "" {
		FileStorage.FileName = s
		if err != nil {
			fmt.Println("Неудачный парсинг переменной окружения FILE_STORAGE_PATH")
		}
	}
}
