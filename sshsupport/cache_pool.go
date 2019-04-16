package sshsupport

import (
    "fmt"
    "github.com/moonlightming/simple-single-log-center/conf"
    "github.com/patrickmn/go-cache"
    "golang.org/x/crypto/ssh"
    "time"
)

var (
    config        = conf.NewConfig()
    SSHClientPool = cache.New(20*time.Minute, 30*time.Minute)
)

func init() {
    SSHClientPool.OnEvicted(func(s string, i interface{}) {
        client := i.(*ssh.Client)
        client.Close()
    })
}

func getOrCreateClient(host, password string) (*ssh.Client, error) {

    var (
        sshTools *SSHTools
        client   *ssh.Client
        key      string
        err      error
    )

    key = fmt.Sprintf("%s:%s", host, password)

    if c, found := SSHClientPool.Get(key); found {
        return c.(*ssh.Client), nil
    }

    sshTools, err = NewSSHClientWithOptions(
        WithHost(host),
        WithPort(config.SSH.Port),
        func() (func(c *SSHTools) error) {
            if password == "" {
                return WithAuthKey(config.SSH.PrivateKey, config.SSH.PrivateKeyPassword)
            } else {
                return WithPassword(password)
            }
        }(),
    )

    if err != nil {
        return nil, err
    }

    client, err = sshTools.NewClient()
    if err != nil {
        return nil, err
    }

    SSHClientPool.Set(key, client, cache.DefaultExpiration)

    return client, nil
}
