package server

import (
	"go_collector_server/database"
	"go_collector_server/server/model"
	"go_collector_server/util"
	"net"
	"os"
	"syscall"

	"github.com/gin-gonic/gin"
)

// var appPath = "/home/dcim/agent"

var appPath = "/app"

func StartUnix() {
	sPath := appPath + "/run/collect-server.sock"
	if _, err := os.Stat(sPath); err == nil {
		os.Remove(sPath)
	}

	listener, err := net.Listen("unix", sPath)
	if err != nil {
		util.Log().Error(err.Error())
	}
	defer listener.Close()

	err = syscall.Chmod(sPath, 0777)
	if err != nil {
		util.Log().Error("无法设置文件权限: %v\n", err)
		return
	}

	router := gin.Default()

	router.POST("/report/sys-collect", func(c *gin.Context) {
		clientIP := c.Request.Header.Get("X-Real-IP")
		if clientIP == "" {
			clientIP = c.Request.Header.Get("X-Forwarded-For")
		}
		if clientIP == "" {
			clientIP = c.ClientIP() // Fallback to ClientIP method if headers are not set
		}

		util.Log().Debug("请求来自 IP: %s", clientIP)

		var collectData model.CollectDataStruct

		if err := c.ShouldBindJSON(&collectData); err != nil {
			// 如果解析出错，返回 400 状态码和错误信息
			c.String(401, "Invaild json data")
			return
		}
		if database.Cache[clientIP] {
			// 返回响应
			c.String(200, "OK")
			return
		}
		if database.WriterCache[clientIP] {
			// 返回响应
			c.String(200, "OK")
			return
		}
		database.Cache[clientIP] = true
		database.WriterChannel <- &database.WriterDataStruct{
			IP:   clientIP,
			Data: &collectData,
		}

		// 返回响应
		c.String(200, "OK")
	})

	util.Log().Info("Gin 正在通过 Unix Socket 监听: %s\n", sPath)

	// 使用 http.Serve 来通过 Unix Socket 启动 Gin 服务
	if err := router.RunListener(listener); err != nil {
		util.Log().Error("Gin 服务错误: %v", err)
	}
}
