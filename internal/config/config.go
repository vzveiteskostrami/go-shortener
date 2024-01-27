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
	Addresses InOutAddresses
	Storage   StorageAttr
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

type StorageAttr struct {
	FileName  string
	DBConnect string
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

	_ = flag.Value(Addresses.In)
	flag.Var(Addresses.In, "a", "In net address host:port")
	_ = flag.Value(Addresses.Out)
	flag.Var(Addresses.Out, "b", "Out net address host:port")
	fn := flag.String("f", "/tmp/short-url-db.json", "Storage text file name")
	dbc := flag.String("d", "", "Database connect string")

	flag.Parse()

	Storage.FileName = *fn
	Storage.DBConnect = *dbc

	var err error
	err = setSERVER_ADDRESS()
	if err != nil {
		fmt.Println(err)
	}
	err = setBASE_URL()
	if err != nil {
		fmt.Println(err)
	}
	err = setFILE_STORAGE_PATH()
	if err != nil {
		fmt.Println(err)
	}
	err = setDATABASE_DSN()
	if err != nil {
		fmt.Println(err)
	}
	// сохранена/закомментирована эмуляция указания БД в параметрах вызова.
	// Необходимо для быстрого перехода тестирования работы приложения с
	// Postgres.
	//Storage.DBConnect = "host=127.0.0.1 port=5432 user=videos password=masterkey dbname=videos sslmode=disable"
	//Storage.DBConnect = "host=127.0.0.1 port=5432 user=executor password=executor dbname=gophermart sslmode=disable"
	//Storage.FileName = ""
}

func setSERVER_ADDRESS() (err error) {
	if s, ok := os.LookupEnv("SERVER_ADDRESS"); ok && s != "" {
		Addresses.In.Host, Addresses.In.Port, err = getAddrAndPort(s)
		if err != nil {
			fmt.Println("Неудачный парсинг переменной окружения SERVER_ADDRESS")
		}
	}
	return
}

func setBASE_URL() (err error) {
	if s, ok := os.LookupEnv("BASE_URL"); ok && s != "" {
		Addresses.Out.Host, Addresses.In.Port, err = getAddrAndPort(s)
		if err != nil {
			fmt.Println("Неудачный парсинг переменной окружения BASE_URL")
		}
	}
	return
}

func setFILE_STORAGE_PATH() (err error) {
	if s, ok := os.LookupEnv("FILE_STORAGE_PATH"); ok && s != "" {
		Storage.FileName = s
	}
	return
}

func setDATABASE_DSN() (err error) {
	if s, ok := os.LookupEnv("DATABASE_DSN"); ok && s != "" {
		Storage.DBConnect = s
	}
	return
}
