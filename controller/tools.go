package controller

import (
    "github.com/gin-gonic/gin"
    "net/http"
)

func responseOK(c *gin.Context, data interface{}) {
    c.JSON(
        http.StatusOK,
        gin.H{
            "result": data,
        },
    )
}

func responseError(c *gin.Context, err error) {
    c.JSON(
        http.StatusInternalServerError,
        gin.H{
            "err": err.Error(),
        },
    )
}
