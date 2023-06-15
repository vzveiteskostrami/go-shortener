package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

var cfg Cfg

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

func getAddrAndPort(s string) (string, int, error) {
	args := strings.Split(s, ":")
	if len(args) != 2 {
		return "", -1, errors.New("неверный формат строки, требуется host:port")
	}

	p, err := strconv.Atoi(args[1])
	if err != nil {
		return "", -1, errors.New("неверный номер порта, " + err.Error())
	}
	return args[0], p, nil
}

type Cfg struct {
	InAddr  *NetAddress
	OutAddr *NetAddress
}

func configStart() {
	cfg.InAddr = new(NetAddress)
	cfg.OutAddr = new(NetAddress)
	cfg.InAddr.Host = ""
	cfg.InAddr.Port = 8080
	cfg.OutAddr.Host = "http://127.0.0.1"
	cfg.OutAddr.Port = 8080
	_ = flag.Value(cfg.InAddr)
	flag.Var(cfg.InAddr, "a", "In net address host:port")
	_ = flag.Value(cfg.OutAddr)
	flag.Var(cfg.OutAddr, "b", "Out net address host:port")
	var err error
	s := os.Getenv("SERVER_ADDRESS")
	if s != "" {
		cfg.InAddr.Host, cfg.InAddr.Port, err = getAddrAndPort(s)
		if err != nil {
			fmt.Println("Неудачный парсинг переменной окружения SERVER_ADDRESS")
		}
	}
	s = os.Getenv("BASE_URL")
	if s != "" {
		cfg.OutAddr.Host, cfg.InAddr.Port, err = getAddrAndPort(s)
		if err != nil {
			fmt.Println("Неудачный парсинг переменной окружения BASE_URL")
		}
	}
}
