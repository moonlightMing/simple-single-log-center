package conf

import (
    "encoding/json"
    "fmt"
    "github.com/sirupsen/logrus"
    "io/ioutil"
    "strings"
    "sync"
    "errors"
)

var (
    once   sync.Once
    config Config
)

type Config struct {
    Server      `json:"server"`
    SSH         `json:"ssh"`
    Allow       `json:"allow"`
    Ansible     `json:"ansible"`
    WebTerminal `json:"web_terminal"`
}

//"server": {
//"addr": "0.0.0.0",
//"port": 9090
//}
type Server struct {
    Addr string `json:"addr"`
    Port int    `json:"port"`
}

//"ssh": {
//"user": "root",
//"port": 22,
//"private_key": "id_rsa_2048",
//"private_key_password": ""
//}
type SSH struct {
    User               string `json:"user"`
    Port               int    `json:"port"`
    PrivateKey         string `json:"private_key"`
    PrivateKeyPassword string `json:"private_key_password"`
}

type WebTerminal struct {
    Open bool `json:"open"`
}

type cArray []string

func (c cArray) isFormatOK() bool {
    for _, e := range c {
        if !strings.HasPrefix(e, "/") {
            return false
        }
    }
    return true
}

func (c cArray) Has(s string) bool {
    for _, e := range c {
        if s == e {
            return true
        }
    }
    return false
}

func (c cArray) HasPre(s string) bool {
    for _, e := range c {
        arr := strings.Split(s, "/")
        if len(arr) != 0 && "/"+arr[1]+"/" == e+"/" {
            return true
        }
    }
    return false
}

func (c cArray) HasSuffix(s string) bool {
    for _, e := range c {
        if strings.HasSuffix(s, "."+e) {
            return true
        }
    }
    return false
}

type Allow struct {
    Dir      cArray `json:"dir"`
    FileType cArray `json:"file_type"`
}

//"ansible": {
//"hosts_file": "hosts"
//}
type Ansible struct {
    HostsFile string `json:"hosts_file"`
}

func NewConfig() Config {
    once.Do(func() {
        fd, err := ioutil.ReadFile("config.json")
        if err != nil {
            logrus.Fatal(err.Error())
        }
        json.Unmarshal(fd, &config)

        // 允许列表格式校验
        if !config.Allow.Dir.isFormatOK() {
            logrus.Fatal(errors.New("allow dir syntax error").Error())
        }

        // 不允许将根目录写入允许列表中
        if config.Allow.Dir.Has("/") {
            logrus.Fatal(errors.New("not allow '/' dir").Error())
        }
    })
    return config
}

func (c *Config) GetListenHost() string {
    return fmt.Sprintf("%s:%d", c.Server.Addr, c.Server.Port)
}
