package sshsupport

import (
    "bufio"
    "bytes"
    "errors"
    "fmt"
    "github.com/gorilla/websocket"
    "github.com/moonlightming/simple-single-log-center/wsserver"
    "github.com/sirupsen/logrus"
    "golang.org/x/crypto/ssh"
    "io"
    "strconv"
    "strings"
    "time"
)

// ParseUnixFile: 将字符串数组转化为文件列表格式
func ParseUnixFile(s []string) []*UnixFile {
    var unixFiles []*UnixFile
    for _, e := range s {
        unixFiles = append(unixFiles, &UnixFile{
            Name: e,
            Type: FOLDER_TYPE,
            Size: 0,
        })
    }
    return unixFiles
}

// isParamsValid: 不允许参数带有管道符 防注入
// &、&&、; 符号都会被识别切割成新参数
func isParamsValid(s string) bool {
    banList := "|"
    return !strings.ContainsAny(s, banList)
}

func ListDir(host, password, dir string) ([]*UnixFile, error) {

    if !config.Dir.HasPre(dir) {
        return nil, errors.New("permission denied")
    }

    if !isParamsValid(dir) {
        return nil, errors.New("contains invalid string")
    }

    var (
        client   *ssh.Client
        session  *ssh.Session
        buf      bytes.Buffer
        fileList []*UnixFile
        err      error
    )

    client, err = getOrCreateClient(host, password)
    if err != nil {
        logrus.Error(err.Error())
        return nil, err
    }

    session, err = client.NewSession()
    if err != nil {
        return nil, err
    }
    defer session.Close()

    session.Stdout = &buf
    if err := session.Run(fmt.Sprintf("ls -l %s", dir)); err != nil {
        return nil, err
    }

    fileList = []*UnixFile{}
    items := strings.Split(buf.String(), "\n")
    if len(items) == 2 {
        return fileList, nil
    }
    for _, item := range items[1 : len(items)-1] {
        i := strings.Split(item, " ")
        // 多个space会造成数组元素异常多 因此部分信息从后面往前取
        iLens := len(i)
        name := i[iLens-1]
        types := func(fd string) int {
            switch fd[0] {
            case 'l':
                return LINK_TYPE
            case 'd':
                return FOLDER_TYPE
            case '-':
                return FILE_TYPE
            default:
                return OTHER_TYPE
            }
        }(i[0])

        if types == LINK_TYPE {
            continue
        }

        if types == FILE_TYPE && !config.FileType.HasSuffix(name) {
            continue
        }

        file := &UnixFile{
            Name: name,
            Size: func(s string) int64 {
                size, _ := strconv.Atoi(s)
                return int64(size)
            }(i[iLens-5]),
            Type:  types,
            Group: i[2],
            Owner: i[3],
        }
        fileList = append(fileList, file)
    }
    return fileList, nil
}

func TailLog(host, password, logFileDir string) (*ssh.Session, <-chan []byte, error) {

    if !config.Dir.HasPre(logFileDir) {
        return nil, nil, errors.New("permission denied")
    }

    if !isParamsValid(logFileDir) {
        return nil, nil, errors.New("contains invalid string")
    }

    var (
        client     *ssh.Client
        session    *ssh.Session
        LogChan    = make(chan []byte, 300)
        logScanner *bufio.Scanner
        cmdReader  io.Reader
        err        error
    )

    client, err = getOrCreateClient(host, password)
    if err != nil {
        return nil, nil, err
    }

    session, err = client.NewSession()
    if err != nil {
        return nil, nil, err
    }

    cmdReader, err = session.StdoutPipe()
    if err != nil {
        return nil, nil, err
    }
    logScanner = bufio.NewScanner(cmdReader)

    go func(logScan *bufio.Scanner, logChan chan<- []byte) {
        for logScan.Scan() {
            // 按行发送，行尾附加回车换行符
            LogChan <- []byte(logScan.Text() + "\r\n")
        }
    }(logScanner, LogChan)

    if err = session.Start(fmt.Sprintf("tail --pid=$$ -n 200 -f %s", logFileDir)); err != nil {
        return nil, nil, err
    }
    return session, LogChan, nil
}

const (
    TerminalTimeOut = 5 * time.Minute
)

func ShellTerminal(host, password string, wsConn *wsserver.WsConnection, termHeight, termWidth int) error {
    var (
        client  *ssh.Client
        session *ssh.Session
        err     error
    )
    client, err = getOrCreateClient(host, password)
    if err != nil {
        return err
    }

    session, err = client.NewSession()
    if err != nil {
        return err
    }
    defer session.Close()

    // 读取ssh的输出，往WebSocket发
    go func() {
        cmdReader, err := session.StdoutPipe()
        if err != nil {
            return
        }
        for {
            if wsConn.IsClosed {
                break
            }
            buf := make([]byte, 1024)
            _, err = cmdReader.Read(buf)
            if err != nil && err.Error() != "EOF" {
                return
            }
            wsConn.WsWrite(websocket.TextMessage, buf)
        }
    }()

    // 读取WebSocket方发来的命令，往ssh发
    go func() {
        cmdWriter, err := session.StdinPipe()
        if err != nil {
            return
        }
        timeC := time.NewTimer(TerminalTimeOut)
        defer timeC.Stop()
        for {
            if wsConn.IsClosed {
                break
            }
            select {
            // 规定时间内无操作就发信号退出当前Terminal
            case <-timeC.C:
                cmdWriter.Write([]byte("logout\r\n"))
            case wsMsg := <-wsConn.InChan:
                data := wsMsg.Data
                cmdWriter.Write(data)
                timeC.Reset(TerminalTimeOut)
            }
        }
    }()
    modes := ssh.TerminalModes{
        ssh.ECHO:          1,
        ssh.TTY_OP_ISPEED: 14400,
        ssh.TTY_OP_OSPEED: 14400,
    }

    if err = session.RequestPty("xterm-256color", termHeight, termWidth, modes); err != nil {
        return err
    }
    if err = session.Shell(); err != nil {
        return err
    }

    if err = session.Wait(); err != nil {
        return err
    }
    return nil
}
