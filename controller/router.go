package controller

import (
    "github.com/gin-gonic/gin"
    "github.com/gorilla/websocket"
    "github.com/moonlightming/simple-single-log-center/conf"
    "github.com/moonlightming/simple-single-log-center/sshsupport"
    "github.com/moonlightming/simple-single-log-center/wsserver"
    "github.com/sirupsen/logrus"
    "golang.org/x/crypto/ssh"
    "net/http"
    "strconv"
)

func InitRouter() {
    // 静态资源映射
    Router.LoadHTMLGlob("web/index.html")
    Router.NoRoute(index)
    Router.Static("/static", "web/static")
    Router.StaticFile("/favicon.ico", "web/favicon.ico")

    // 交互接口
    Router.GET("/api/listSection", listSection)
    Router.GET("/api/listHost", listHost)
    Router.GET("/api/listDir", listDir)
    Router.GET("/api/listAllHosts", listAllHosts)
    Router.GET("/api/isWebTerminalOpen", isWebTerminalOpen)
    Router.GET("/api/tailLog", tailLog)

    // WebTerminal接口
    Router.Use(AutoTerminalAllow())
    Router.GET("/api/terminalShell", terminalShell)
}

func index(c *gin.Context) {
    c.HTML(http.StatusOK, "index.html", gin.H{})
}

func listSection(c *gin.Context) {
    ansibleInventory, err := conf.NewAnsibleInventory()
    if err != nil {
        logrus.
            WithFields(logrus.Fields{
            "func": "listSection",
        }).
            Error(err.Error())
        responseError(c, err)
    }

    responseOK(c, ansibleInventory.GetSections())
}

func listHost(c *gin.Context) {
    ansibleInventory, err := conf.NewAnsibleInventory()
    section := c.Query("section")

    hosts, err := ansibleInventory.GetHosts(section)
    if err != nil {
        logrus.
            WithFields(logrus.Fields{
            "host": hosts,
            "func": "listHost",
        }).
            Error(err.Error())
        responseError(c, err)
    }
    responseOK(c, hosts)
}

func listAllHosts(c *gin.Context) {
    ansibleInventory, err := conf.NewAnsibleInventory()
    if err != nil {
        responseError(c, err)
        return
    }

    allHosts, err := ansibleInventory.GetHostsAll()
    if err != nil {
        logrus.
            WithFields(logrus.Fields{
            "host": "all",
            "func": "listAllHosts",
        }).
            Error(err.Error())
        responseError(c, err)
        return
    }

    responseOK(c, allHosts)
}

func listDir(c *gin.Context) {
    host := c.Query("host")
    password := c.DefaultQuery("password", "")
    path := c.DefaultQuery("path", "")

    var (
        files []*sshsupport.UnixFile
        err   error
    )
    if path != "" {
        files, err = sshsupport.ListDir(host, password, path)
    } else {
        files = sshsupport.ParseUnixFile(Config.Allow.Dir)
    }

    if err != nil {
        logrus.
            WithFields(logrus.Fields{
            "host": host,
            "func": "listDir",
        }).
            Error(err.Error())
        responseError(c, err)
        return
    }
    responseOK(c, files)
}

func tailLog(c *gin.Context) {
    var (
        wsConn  *wsserver.WsConnection
        session *ssh.Session
        logChan <-chan []byte
        err     error
    )

    host := c.Query("host")
    password := c.DefaultQuery("password", "")
    path := c.Query("path")

    wsConn, err = wsserver.NewWsConnection(c.Writer, c.Request)
    if err != nil {
        logrus.
            WithFields(logrus.Fields{
            "host": host,
            "func": "tailLog",
        }).
            Error(err.Error())
        return
    }
    defer wsConn.Close()

    // 开启读写协程
    wsConn.WsHandler()

    session, logChan, err = sshsupport.TailLog(host, password, path)
    if err != nil {
        logrus.
            WithFields(logrus.Fields{
            "host": host,
            "func": "tailLog",
        }).
            Error(err.Error())
        responseError(c, err)
    }

    for {
        select {
        case tLog := <-logChan:
            wsConn.WsWrite(websocket.TextMessage, tLog)
        case <-wsConn.CloseChan:
            goto CLOSE
        }
    }

CLOSE:
    defer session.Close()
}

func isWebTerminalOpen(c *gin.Context) {
    if Config.WebTerminal.Open {
        responseOK(c, true)
    } else {
        responseOK(c, false)
    }
}

func terminalShell(c *gin.Context) {
    var (
        wsConn *wsserver.WsConnection
        err    error
    )

    host := c.Query("host")
    password := c.DefaultQuery("password", "")
    termHeight, err := strconv.Atoi(c.DefaultQuery("termHeight", "30"))
    termWidth, err := strconv.Atoi(c.DefaultQuery("termWidth", "10"))

    wsConn, err = wsserver.NewWsConnection(c.Writer, c.Request)
    if err != nil {
        logrus.Error(err.Error())
        return
    }
    defer wsConn.Close()

    // 开启读写协程
    wsConn.WsHandler()

    if err = sshsupport.ShellTerminal(host, password, wsConn, termHeight, termWidth); err != nil && err.Error() != "EOF" {
        logrus.
            WithFields(logrus.Fields{
            "host": host,
            "func": "terminalShell",
        }).
            Error(err.Error())
        responseError(c, err)
    }
}
