package sshsupport

import (
    "fmt"
    "io/ioutil"
    "net"
    "time"

    "github.com/sirupsen/logrus"
    "golang.org/x/crypto/ssh"
)

type SSHTools struct {
    Host string
    Port int
    *ssh.ClientConfig
}

func NewSSHClientWithOptions(ops ...func(c *SSHTools) error) (*SSHTools, error) {

    sshTools := &SSHTools{
        Host: "127.0.0.1",
        Port: 22,
        ClientConfig: &ssh.ClientConfig{
            User:    "root",
            Auth:    []ssh.AuthMethod{},
            Timeout: 10 * time.Second,
            HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
                return nil
            },
        },
    }

    for _, op := range ops {
        if err := op(sshTools); err != nil {
            return sshTools, err
        }
    }

    return sshTools, nil
}

func WithHost(host string) func(c *SSHTools) error {
    return func(c *SSHTools) error {
        c.Host = host
        return nil
    }
}

func WithPort(port int) func(c *SSHTools) error {
    return func(c *SSHTools) error {
        c.Port = port
        return nil
    }
}

func WithUser(user string) func(c *SSHTools) error {
    return func(c *SSHTools) error {
        c.User = user
        return nil
    }
}

func WithPassword(password string) func(c *SSHTools) error {
    return func(c *SSHTools) error {
        c.Auth = append(c.Auth, ssh.Password(password))
        return nil
    }
}

func WithAuthKey(authKey, password string) func(c *SSHTools) error {
    return func(c *SSHTools) error {
        var (
            signer ssh.Signer
            err    error
        )
        pemBytes, err := ioutil.ReadFile(authKey)
        if err != nil {
            logrus.Fatalf("unable to read private key: %v", err)
            return err
        }

        if password == "" {
            if signer, err = ssh.ParsePrivateKey(pemBytes); err != nil {
                logrus.Fatalf("unable to parse private key: %v", err)
                return err
            }
        } else {
            if signer, err = ssh.ParsePrivateKeyWithPassphrase(pemBytes, []byte(password)); err != nil {
                logrus.Fatalf("unable to parse private key: %v", err)
                return err
            }
        }

        c.Auth = append(c.Auth, ssh.PublicKeys(signer))

        return nil
    }
}

func WithTimeout(i int64) func(c *SSHTools) error {
    return func(c *SSHTools) error {
        // 在Go里面int和int64是不同的，int是平台相关的，在32位平台和64位平台不同，a := 1 Go会自动推导成int型，a * time.Second肯定不行
        // 由于i是int64类型，虽然Duration是定义的int64，但是还是算两种类型，因此b * time.Second也不行
        // 正确的做法是转换，time.Duration(b)*time.Second
        c.Timeout = time.Duration(i) * time.Second
        return nil
    }
}

func (s *SSHTools) NewClient() (*ssh.Client, error) {
    var (
        err error
    )
    client, err := ssh.Dial(
        "tcp",
        fmt.Sprintf("%s:%d", s.Host, s.Port),
        s.ClientConfig,
    )
    if err != nil {
        return nil, err
    }
    return client, nil
}
