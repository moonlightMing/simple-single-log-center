package controller

import (
    "github.com/gin-gonic/gin"
    "net/http"
)

func AutoTerminalAllow() gin.HandlerFunc {
    return func(c *gin.Context) {
        if Config.WebTerminal.Open {
            c.Next()
        } else {
            c.Abort()
            c.JSON(
                http.StatusForbidden,
                gin.H{"data": "WebTerminal功能未开启"},
            )
        }
    }
}
