package controller

import (
    "github.com/gin-gonic/gin"
    "github.com/moonlightming/simple-single-log-center/conf"
    "log"
)

var (
    Router *gin.Engine
    Config conf.Config
)

func InitWebServer() {
    Config = conf.NewConfig()
    gin.SetMode(gin.ReleaseMode)
    Router = gin.Default()
}

func StartWebServer() {
    log.Fatalln(Router.Run(Config.GetListenHost()))
}
