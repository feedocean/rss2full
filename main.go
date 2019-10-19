package main

import (
	"flag"
	"net"
	"net/http"
	"os"
	"strconv"
	"syscall"

	"github.com/judwhite/go-svc/svc"
	"github.com/julienschmidt/httprouter"
	"github.com/sirupsen/logrus"
)

var appVersion = "0.1.2"

type program struct {
	quit chan struct{}
}

func (p *program) Init(env svc.Environment) error {
	return nil
}

func (p *program) Start() error {
	var (
		aAddr = flag.String("a", "", "Bind address")
		aPort = flag.Int("p", 8088, "Port to listen")
	)
	flag.Parse()
	addr := *aAddr + ":" + strconv.Itoa(*aPort)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	go func() {
		wwwroot := http.Dir("./wwwroot")
		fs := http.FileServer(wwwroot)

		router := httprouter.New()
		router.GET("/version", Version)
		router.GET("/feed/*feed", FullRss)
		router.Handler("GET", "/assets/*filepath", fs)
		router.Handler("GET", "/", fs)

		if err := http.Serve(listener, router); err != nil {
			logrus.Fatalf("http.Serve got error: %v", err)
		}
		<-p.quit
		listener.Close()
		logrus.Info("exit")
	}()
	logrus.Infof("version: %s", appVersion)
	logrus.Infof("listen on %s \n", addr)
	return nil
}

func (p *program) Stop() error {
	close(p.quit)
	return nil
}

func main() {
	logrus.SetOutput(os.Stdout)
	prg := &program{
		quit: make(chan struct{}),
	}
	if err := svc.Run(prg, syscall.SIGINT, syscall.SIGTERM); err != nil {
		logrus.Fatal(err)
	}
}
