package config

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/vzveiteskostrami/go-shortener/internal/logging"
)

type PConfig struct {
	ServerAddress   *string `json:"server_address,omitempty"`
	BaseUrl         *string `json:"base_url,omitempty"`
	FileStoragePath *string `json:"file_storage_path,omitempty"`
	EnableHttps     *bool   `json:"enable_https,omitempty"`
	DatabaseDSN     *string `json:"database_dsn,omitempty"`
	TrustedSubnet   *string `json:"trusted_subnet,omitempty"`
	EnablegRPC      *bool   `json:"enable_grpc,omitempty"`
}

var (
	Addresses InOutAddresses
	Storage   StorageAttr
	UseHTTPS  bool
	UsegRPC   bool
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
	FileName      string
	DBConnect     string
	TrustedSubnet string
	TrustedIPNet  *net.IPNet
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
	Addresses.In.Host = ""
	Addresses.In.Port = -1
	Addresses.Out.Host = ""
	Addresses.Out.Port = -1

	_ = flag.Value(Addresses.In)
	flag.Var(Addresses.In, "a", "In net address host:port")
	_ = flag.Value(Addresses.Out)
	flag.Var(Addresses.Out, "b", "Out net address host:port")
	fn := flag.String("f", "", "Storage text file name")
	dbc := flag.String("d", "", "Database connect string")
	uh := flag.Bool("s", false, "HTTPS connect enabled")
	trustedSubnet := flag.String("t", "", "Trusted subnet")
	cfgFileName := flag.String("c", "", "Config file name")
	cfgFileName1 := flag.String("config", "", "Config file name")
	gRPC := flag.Bool("gRPC", false, "gRPC connect enabled (-s doesn't work in this case)")

	flag.Parse()

	if *cfgFileName == "" && *cfgFileName1 != "" {
		*cfgFileName = *cfgFileName1
	}

	var (
		ok  bool
		err error
		cfg PConfig
	)

	*cfgFileName1, ok = getConfigFileName()
	if ok && *cfgFileName1 != "" {
		*cfgFileName = *cfgFileName1
	}

	// Получение конфига из JSON в структуру
	if *cfgFileName != "" {
		cfg = cfgReadFromJSON(*cfgFileName)
	}

	// Разбираемся с "In net address host:port"
	if Addresses.In.Host == "" || Addresses.In.Port == -1 {
		if cfg.ServerAddress != nil {
			tmpNA := NetAddress{}
			tmpNA.Set(*cfg.ServerAddress)
			if Addresses.In.Host == "" {
				Addresses.In.Host = tmpNA.Host
			}
			if Addresses.In.Port == -1 {
				Addresses.In.Port = tmpNA.Port
			}
		}
	}
	err = setServerAddress()
	if err != nil {
		fmt.Println(err)
	}
	if Addresses.In.Host == "" {
		Addresses.In.Host = "localhost"
	}
	if Addresses.In.Port == -1 {
		Addresses.In.Port = 8080
	}

	// Разбираемся с "Out net address host:port"
	if Addresses.Out.Host == "" || Addresses.Out.Port == -1 {
		if cfg.BaseUrl != nil {
			tmpNA := NetAddress{}
			tmpNA.Set(*cfg.BaseUrl)
			if Addresses.Out.Host == "" {
				Addresses.Out.Host = tmpNA.Host
			}
			if Addresses.Out.Port == -1 {
				Addresses.Out.Port = tmpNA.Port
			}
		}
	}
	err = setBaseURL()
	if err != nil {
		fmt.Println(err)
	}
	if Addresses.Out.Host == "" {
		Addresses.Out.Host = "http://127.0.0.1"
	}
	if Addresses.Out.Port == -1 {
		Addresses.Out.Port = 8080
	}

	// Разбираемся с "Storage text file name"
	if *fn != "" {
		Storage.FileName = *fn
	} else if cfg.FileStoragePath != nil && *cfg.FileStoragePath != "" {
		Storage.FileName = *cfg.FileStoragePath
	} else {
		Storage.FileName = "/tmp/short-url-db.json"
	}
	err = setFileStoragePath()
	if err != nil {
		fmt.Println(err)
	}

	// Разбираемся с "Database connect string"
	if *dbc != "" {
		Storage.DBConnect = *dbc
	} else if cfg.DatabaseDSN != nil && *cfg.DatabaseDSN != "" {
		Storage.DBConnect = *cfg.DatabaseDSN
	}
	err = setDatabaseDSN()
	if err != nil {
		fmt.Println(err)
	}

	// Разбираемся с "Trusted subnet"
	if *trustedSubnet != "" {
		Storage.TrustedSubnet = *trustedSubnet
	} else if cfg.TrustedSubnet != nil && *cfg.TrustedSubnet != "" {
		Storage.TrustedSubnet = *cfg.TrustedSubnet
	}
	err = setTrustedSubnet()
	if err != nil {
		fmt.Println(err)
	}

	if Storage.TrustedSubnet != "" {
		_, Storage.TrustedIPNet, err = net.ParseCIDR(Storage.TrustedSubnet)
		if err != nil {
			panic("Это что такое? " + Storage.TrustedSubnet)
		}
	}

	// Узнаём в принципе были ли флаги -s и -gRPC или нет. Необходимо для приоритетности
	// установки значений. Если их не было, значит нужно читать значение из
	// конфига в json. Если был, значение в json конфиге не имеет значения.
	hasUseHTTPSArg := false
	hasUsegRPCArg := false
	for _, s := range os.Args {
		if s == "-s" || strings.HasPrefix(s, "-s=") {
			hasUseHTTPSArg = true
		} else if s == "-gRPC" || strings.HasPrefix(s, "-gRPC=") {
			hasUsegRPCArg = true
		}
	}

	// Разбираемся с "Enable https"
	if hasUseHTTPSArg {
		UseHTTPS = *uh
	} else {
		if cfg.EnableHttps != nil {
			UseHTTPS = *cfg.EnableHttps
		}
	}
	err = setEnableHttps()
	if err != nil {
		fmt.Println(err)
	}

	// Разбираемся с "Enable gRPC"
	if hasUsegRPCArg {
		UsegRPC = *gRPC
	} else {
		if cfg.EnablegRPC != nil {
			UsegRPC = *cfg.EnablegRPC
		}
	}
	err = setEnablegRPC()
	if err != nil {
		fmt.Println(err)
	}
	UsegRPC = true

	// сохранена/закомментирована эмуляция указания БД в параметрах вызова.
	// Необходимо для быстрого перехода тестирования работы приложения с
	// Postgres.
	//Storage.DBConnect = "host=127.0.0.1 port=5432 user=videos password=masterkey dbname=videos sslmode=disable"
	Storage.DBConnect = "host=127.0.0.1 port=5432 user=executor password=executor dbname=gophermart sslmode=disable"
	//Storage.FileName = ""
}

func setServerAddress() (err error) {
	if s, ok := os.LookupEnv("SERVER_ADDRESS"); ok && s != "" {
		Addresses.In.Host, Addresses.In.Port, err = getAddrAndPort(s)
		if err != nil {
			fmt.Println("Неудачный парсинг переменной окружения SERVER_ADDRESS")
		}
	}
	return
}

func setBaseURL() (err error) {
	if s, ok := os.LookupEnv("BASE_URL"); ok && s != "" {
		Addresses.Out.Host, Addresses.In.Port, err = getAddrAndPort(s)
		if err != nil {
			fmt.Println("Неудачный парсинг переменной окружения BASE_URL")
		}
	}
	return
}

func setFileStoragePath() (err error) {
	if s, ok := os.LookupEnv("FILE_STORAGE_PATH"); ok && s != "" {
		Storage.FileName = s
	}
	return
}

func setDatabaseDSN() (err error) {
	if s, ok := os.LookupEnv("DATABASE_DSN"); ok && s != "" {
		Storage.DBConnect = s
	}
	return
}

func setTrustedSubnet() (err error) {
	if s, ok := os.LookupEnv("TRUSTED_SUBNET"); ok && s != "" {
		Storage.TrustedSubnet = s
	}
	return
}

func setEnableHttps() (err error) {
	if _, ok := os.LookupEnv("ENABLE_HTTPS"); ok {
		UseHTTPS = ok
	}
	return
}

func setEnablegRPC() (err error) {
	if _, ok := os.LookupEnv("ENABLE_GRPC"); ok {
		UsegRPC = ok
	}
	return
}

func getConfigFileName() (s string, ok bool) {
	s, ok = os.LookupEnv("CONFIG")
	return
}

func cfgReadFromJSON(fname string) PConfig {
	var r PConfig
	content, err := os.ReadFile(fname)
	if err != nil {
		logging.S().Errorln("Ошибка чтения config файла", fname, err)
		return r
	}
	err = json.Unmarshal(content, &r)
	if err != nil {
		logging.S().Errorln("Ошибка парсинга config файла", fname, err)
	}
	return r
}
