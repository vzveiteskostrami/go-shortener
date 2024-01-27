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
