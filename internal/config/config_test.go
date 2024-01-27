package config

import (
	"testing"
)

func Test_getAddrAndPort(t *testing.T) {
	tests := []struct {
		name    string
		arg     string
		want    string
		want1   int
		wantErr bool
	}{
		{name: "Empty string", arg: "", want: "", want1: -1, wantErr: true},
		{name: "Without :", arg: "jhgjhd", want: "", want1: -1, wantErr: true},
		{name: "OKey", arg: "siteaddress:1122", want: "siteaddress", want1: 1122, wantErr: false},
		{name: "Wrong port", arg: "1122:eeeee", want: "", want1: -1, wantErr: true},
		{name: "OKey without address", arg: ":1122", want: "", want1: 1122, wantErr: false},
		{name: "Wrong address and port", arg: ":", want: "", want1: -1, wantErr: true},
		{name: "Wrong port 2", arg: "1122:", want: "", want1: -1, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := getAddrAndPort(tt.arg)
			if (err != nil) != tt.wantErr {
				t.Errorf("getAddrAndPort() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getAddrAndPort() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("getAddrAndPort() got = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestNetAddress_String(t *testing.T) {
	tests := []struct {
		name string
		na   *NetAddress
		want string
	}{
		{name: "OKey address 1", want: "address:1122"},
		{name: "OKey address 2", want: ":1122"},
	}

	tests[0].na = new(NetAddress)
	tests[0].na.Host = "address"
	tests[0].na.Port = 1122

	tests[1].na = new(NetAddress)
	tests[1].na.Host = ""
	tests[1].na.Port = 1122

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.na.String(); got != tt.want {
				t.Errorf("NetAddress.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNetAddress_Set(t *testing.T) {
	tests := []struct {
		name      string
		na        *NetAddress
		flagValue string
		wantErr   bool
	}{
		{name: "a", na: nil, flagValue: "dfdfdff:sdfdfdf", wantErr: true},
		{name: "b", na: nil, flagValue: ":", wantErr: true},
		{name: "c", na: nil, flagValue: "aaa:111", wantErr: false},
	}
	tests[0].na = new(NetAddress)
	tests[1].na = new(NetAddress)
	tests[2].na = new(NetAddress)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.na.Set(tt.flagValue); (err != nil) != tt.wantErr {
				t.Errorf("NetAddress.Set() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_setSERVER_ADDRESS(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{"a", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := setSERVER_ADDRESS(); (err != nil) != tt.wantErr {
				t.Errorf("setSERVER_ADDRESS() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_setBASE_URL(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{"a", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := setBASE_URL(); (err != nil) != tt.wantErr {
				t.Errorf("setBASE_URL() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_setFILE_STORAGE_PATH(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{"a", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := setFILE_STORAGE_PATH(); (err != nil) != tt.wantErr {
				t.Errorf("setFILE_STORAGE_PATH() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_setDATABASE_DSN(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{"a", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := setDATABASE_DSN(); (err != nil) != tt.wantErr {
				t.Errorf("setDATABASE_DSN() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
