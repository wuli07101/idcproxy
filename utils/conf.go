package utils

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"os"
)

type ServerConf struct {
	Address  string `json:"address"`
	Port     int    `json:"port"`
	Timeout  int    `json:"timeout"`
	Password string `json:"password"`
	Method   string `json:"method"`
	Protocol string `json:"protocol"`
}

type LocalConf struct {
	Address  string `json:"address"`
	Port     int    `json:"port"`
	Protocol string `json:"protocol"`
}

type Conf struct {
	Server      []*ServerConf `json:"server"`
	Local       []*LocalConf  `json:"local"`
	Timeout     int64         `json:"timeout"`
	TcpFastOpen bool          `json:"tcp_fastopen"`
}

func newConf() *Conf {
	c := &Conf{}
	c.Server = make([]*ServerConf, 1)
	c.Server[0] = new(ServerConf)
	c.Local = make([]*LocalConf, 1)
	c.Local[0] = new(LocalConf)
	return c
}

func ParseSeverConf() *Conf {
	var confFile string
	var conf = newConf()
	var help bool

	flag.StringVar(&confFile, "c", "", "path to config file")
	flag.StringVar(&conf.Server[0].Address, "s", "", "server address")
	flag.IntVar(&conf.Server[0].Port, "p", 8388, "server port")
	flag.StringVar(&conf.Server[0].Password, "k", "password", "password")
	flag.StringVar(&conf.Server[0].Method, "m", "aes-256-cfb", "encryption method")
	conf.Server[0].Protocol = "tcp"
	conf.Local[0].Protocol = "http"
	flag.StringVar(&conf.Local[0].Address, "b", "127.0.0.1", "local binding address")
	flag.IntVar(&conf.Local[0].Port, "l", 1080, "local port")
	flag.Int64Var(&conf.Timeout, "t", 300, "timeout in seconds")
	flag.BoolVar(&help, "-help", false, "print usage")
	// flag.BoolVar(&conf.TcpFastOpen, "-fast-open", false, "use TCP_FASTOPEN, requires Linux 3.7+")
	flag.Parse()

	if help {
		flag.Usage()
		os.Exit(0)
	}

	c, err := parseConf(confFile)
	if err != nil {
		return conf
	}
	return c
}

func parseConf(confFile string) (*Conf, error) {
	rawConf, err := ioutil.ReadFile(confFile)
	if err != nil {
		return nil, err
	}
	v := &Conf{}

	err = json.Unmarshal(rawConf, &v)
	return v, nil
}
