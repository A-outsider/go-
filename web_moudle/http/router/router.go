package router

import (
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"io"
	"new/web_moudle/http/middleware"

	"os"
)

func Setup() *gin.Engine {
	workdir, _ := os.Getwd()
	logfile, err := os.Create(workdir + "/log/gin_http.log")
	if err != nil {
		zap.L().Error(" logfileCreate failed", zap.Error(err))
	}
	gin.SetMode(gin.DebugMode)
	gin.DefaultWriter = io.MultiWriter(logfile, os.Stdout)
	r := gin.Default()
	r.Use(middleware.Cors())
	pprof.Register(r)

	//...构建模块
	return r
}
