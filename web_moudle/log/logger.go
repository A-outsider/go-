package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"new/web_moudle/settings"

	"os"
)

var Logger *zap.Logger

//var SugarLogger *zap.SugaredLogger

func InitLogger(conf *settings.LogConfig) (err error) { //核心需要的三部分
	encoder := getEncoder()

	writeSyncer := getLogWriter(
		conf.Filename,
		conf.MaxSize,
		conf.MaxBackups,
		conf.MaxAge)

	var l = new(zapcore.Level)
	err = l.UnmarshalText([]byte(conf.Level))
	if err != nil {
		return
	}

	core1 := zapcore.NewCore(encoder, writeSyncer, l) // 设置写入等级

	//这边默认配置了error等级
	core2 := zapcore.NewCore(encoder, getLogErrWriter(conf.Filename), zapcore.ErrorLevel)

	core := zapcore.NewTee(core1, core2)
	Lg := zap.New(core, zap.AddCaller()) //addCall添加调用方的信息

	zap.ReplaceGlobals(Lg) // 全局的logger太长 ,替换掉

	//	SugarLogger = Logger.Sugar()
	return
}

func getEncoder() zapcore.Encoder { //自定义写入日志的格式(编译器)
	encoderConfig := zap.NewProductionEncoderConfig() //自定义时间的返回类型
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	return zapcore.NewConsoleEncoder(encoderConfig) //使用普通的log的信息返回
}

func getLogWriter(filename string, maxSize, maxBackup, maxAge int) zapcore.WriteSyncer { //写入的配置文件所在地
	//workdir, _ := os.Getwd()
	//file, _ := os.Create(workdir + filename)
	lumberJackLogger := &lumberjack.Logger{ //实现自动分割
		Filename:   filename,
		MaxAge:     maxAge,
		MaxBackups: maxBackup,
		MaxSize:    maxSize,
	}
	return zapcore.AddSync(lumberJackLogger)
}

func getLogErrWriter(filename string) zapcore.WriteSyncer {
	workdir, _ := os.Getwd()
	file, _ := os.Create(workdir + filename)
	return zapcore.AddSync(file)
}

//// GinLogger 接收gin框架默认的日志
//func GinLogger() gin.HandlerFunc {
//	return func(c *gin.Context) {
//		start := time.Now()
//		path := c.Request.URL.Path
//		query := c.Request.URL.RawQuery
//		c.Next()
//
//		cost := time.Since(start)
//		zap.L().Info(path,
//			zap.Int("status", c.Writer.Status()),
//			zap.String("method", c.Request.Method),
//			zap.String("path", path),
//			zap.String("query", query),
//			zap.String("ip", c.ClientIP()),
//			zap.String("user-agent", c.Request.UserAgent()),
//			zap.String("errors", c.Errors.ByType(gin.ErrorTypePrivate).String()),
//			zap.Duration("cost", cost),
//		)
//	}
//}

// GinRecovery recover掉项目可能出现的panic，并使用zap记录相关日志
//func GinRecovery(stack bool) gin.HandlerFunc {
//	return func(c *gin.Context) {
//		defer func() {
//			if err := recover(); err != nil {
//				// Check for a broken connection, as it is not really a
//				// condition that warrants a panic stack trace.
//				var brokenPipe bool
//				var ne *net.OpError
//				if errors.As(err.(error), &ne) {
//					var se *os.SyscallError
//					if errors.As(ne.Err, &se) {
//						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") || strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
//							brokenPipe = true
//						}
//					}
//				}
//
//				httpRequest, _ := httputil.DumpRequest(c.Request, false)
//				if brokenPipe {
//					zap.L().Error(c.Request.URL.Path,
//						zap.Any("error", err),
//						zap.String("request", string(httpRequest)),
//					)
//					// If the connection is dead, we can't write a status to it.
//					c.Error(err.(error)) // nolint: errcheck
//					c.Abort()
//					return
//				}

//				if stack {
//					lg.Error("[Recovery from panic]",
//						zap.Any("error", err),
//						zap.String("request", string(httpRequest)),
//						zap.String("stack", string(debug.Stack())),
//					)
//				} else {
//					lg.Error("[Recovery from panic]",
//						zap.Any("error", err),
//						zap.String("request", string(httpRequest)),
//					)
//				}
//				c.AbortWithStatus(http.StatusInternalServerError)
//			}
//		}()
//		c.Next()
//	}
//}
