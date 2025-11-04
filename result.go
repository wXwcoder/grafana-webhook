package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func Success(c *gin.Context, msg string, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"code":    1,
		"message": msg,
		"data":    data,
	})
}

func SuccessNoData(c *gin.Context, msg string) {
	c.JSON(http.StatusOK, gin.H{
		"code":    1,
		"message": msg,
		"data":    nil,
	})
}

func Fail(c *gin.Context, msg string) {
	c.JSON(http.StatusBadRequest, gin.H{
		"code":    -1,
		"message": msg,
		"data":    nil,
	})
}

func FailWithSys(c *gin.Context, msg string) {
	c.JSON(http.StatusBadRequest, gin.H{
		"code":    -1,
		"message": "服务器错误",
		"data":    nil,
	})
}

func FailAndCode(c *gin.Context, code int8, msg string) {
	c.JSON(http.StatusBadRequest, gin.H{
		"code":    code,
		"message": msg,
		"data":    nil,
	})
}
