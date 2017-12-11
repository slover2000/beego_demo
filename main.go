package main

import (
	"log"
	"os"
	"fmt"
	"time"
	"net/http"
	"syscall"
	"os/signal"
	"golang.org/x/net/context"
	
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	
    "github.com/sirupsen/logrus"	
	"github.com/rifflock/lfshook"
	"github.com/lestrrat/go-file-rotatelogs"
		
	"github.com/slover2000/prisma"
    "github.com/slover2000/prisma/logging"
    "github.com/slover2000/prisma/discovery"
    "github.com/slover2000/prisma/trace"
	"github.com/slover2000/prisma/trace/zipkin"
	
	_ "github.com/slover2000/beego_demo/routers"
	"github.com/slover2000/beego_demo/services"
)

func initInterceptor() (*prisma.InterceptorClient, error) {
	initLog("access.log")
	// initialize logger
	
	sampleRate := beego.AppConfig.DefaultFloat("tracesamplerate", 0.01)
	sampleQPS := beego.AppConfig.DefaultFloat("traceqps", 100)
	policy, _ := trace.NewLimitedSampler(sampleRate, sampleQPS)
	zipkinHost := beego.AppConfig.String("zipkinhost")
	collector := trace.NewMultiCollector(
					trace.NewConsoleCollector(), 
					zipkin.NewHTTPCollector(zipkinHost, zipkin.HTTPBatchSize(10), zipkin.HTTPMaxBacklog(3), zipkin.HTTPBatchInterval(3 * time.Second)))

	interceptorClient, err := prisma.NewInterceptorClient(
		context.Background(),
		prisma.EnableTracing(beego.BConfig.AppName, policy, collector),
		prisma.EnableLogging(logging.InfoLevel),
		prisma.EnableAllMetrics(),
		prisma.EnableMetricsExportHTTPServer(9100)) 					
	if err != nil {
		logs.Error("create http interceptor failed:%s", err.Error())
		return nil, err
	}
	
	return interceptorClient, nil
}

func initLog(filename string) {
	writer, err := rotatelogs.New(
		filename + ".%Y%m%d",
		rotatelogs.WithLinkName(filename),
		rotatelogs.WithMaxAge(time.Duration(7 * 24) * time.Hour),
		rotatelogs.WithRotationTime(time.Duration(24) * time.Hour),
	)

	if err != nil {
		panic("can't create rotatelogs writer")
	}

	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetLevel(logrus.InfoLevel)
	logrus.AddHook(lfshook.NewHook(lfshook.WriterMap{
		logrus.InfoLevel: writer,
		logrus.WarnLevel: writer,
		logrus.ErrorLevel: writer,
		logrus.FatalLevel: writer,
	}, &logrus.JSONFormatter{}))
}

func main() {
	if beego.BConfig.RunMode == "dev" {
		beego.BConfig.WebConfig.DirectoryIndex = true
		beego.BConfig.WebConfig.StaticDir["/swagger"] = "swagger"
	}	
	beego.BConfig.RunMode = "prod"
	beego.BConfig.Log.AccessLogs = false
	beego.BConfig.WebConfig.AutoRender = false

	// init log
	//logs.SetLogger(logs.AdapterFile,`{"filename":"access.log","level":6,"maxlines":0,"maxsize":0,"daily":true,"maxdays":7}`)
	//logs.SetLogFuncCall(false)
	
	interceptorClient, err := initInterceptor()
	if err != nil {
		log.Fatal("create interceptor with fail:%s", err.Error())
		return
	}

	err = services.InitHelloServiceClient("http://10.98.16.215:2379", interceptorClient)
	if err != nil {
		log.Fatal("init hello service failed:%s", err.Error())
		return
	}

	// intercept http handler
	httpHandler := interceptorClient.HTTPHandler(beego.BeeApp.Handlers) // 使用beego handlers 的结构处理 HTTP 请求
	server := &http.Server{
        Handler: httpHandler,
        Addr:    fmt.Sprintf(":%d", beego.BConfig.Listen.HTTPPort),
	}

	ch := make(chan os.Signal, 1)
    signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL, syscall.SIGHUP, syscall.SIGQUIT)
    go func(c chan os.Signal) {
		// register server into etcd
		endpoint := discovery.Endpoint{Host: "127.0.0.1", Port: beego.BConfig.Listen.HTTPPort, EnvType: discovery.Product}
		register, err := discovery.NewEtcdRegister(
			beego.AppConfig.String("etcdhost"),
			discovery.WithRegisterSystem(discovery.HTTPSystem),
			discovery.WithRegisterService(beego.BConfig.AppName),
			discovery.WithRegisterEndpoint(endpoint),
			discovery.WithRegisterDialTimeout(5 * time.Second),
			discovery.WithRegisterInterval(time.Duration(beego.AppConfig.DefaultInt("freshinterval", 10)) * time.Second),
			discovery.WithRegisterTTL(time.Duration(beego.AppConfig.DefaultInt("servicettl", 15)) * time.Second))
		if err != nil {
			logs.Error("init etcd register failed with:%s", err.Error())
			return
		}
		register.Register()

        s := <-c
		logs.Info("receive signal '%s', stop http server", s.String())		
        register.Unregister()
		interceptorClient.Close()
		services.CloseHelloServiceClient()
		ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
		server.Shutdown(ctx)
		cancel()
        os.Exit(1)
    }(ch)

	if err := server.ListenAndServe(); err != nil {
        fmt.Println("http server exit:", err)
    }
}