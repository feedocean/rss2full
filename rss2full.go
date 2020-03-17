package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"syscall"

	"github.com/judwhite/go-svc/svc"
	"github.com/julienschmidt/httprouter"
	"github.com/sirupsen/logrus"
)

var (
	Version = "0.2.1"

	aAddr              = flag.String("a", "", "Bind address")
	aPort              = flag.Int("p", 8088, "Port to listen")
	aVers              = flag.Bool("v", false, "Show version")
	aVersl             = flag.Bool("version", false, "Show version")
	aHelp              = flag.Bool("h", false, "Show help")
	aHelpl             = flag.Bool("help", false, "Show help")
	aItemCount         = flag.Int("item-count", 10, "Define number of items in feed")
	aConnectionPerFeed = flag.Int("connection-per-feed", 2, "Define number of parallel connections per feed")
)

const usage = `rss2full %s

Usage:
  rss2full -p 80
  rss2full -h | -help
  rss2full -v | -version

Options:
  -a <addr>                   Bind address [default: *]
  -p <port>                   Bind port [default: 8088]
  -h, -help                   Show help
  -v, -version                Show version
  -item-count <num>           Define number of items in feed
  -connection-per-feed <num>  Define number of parallel connections(workers) per feed
`

type program struct {
	quit chan struct{}
}

func (p *program) Init(env svc.Environment) error {
	return nil
}

func (p *program) Start() error {
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, fmt.Sprintf(usage, Version))
	}
	flag.Parse()

	if *aHelp || *aHelpl {
		showUsage()
	}
	if *aVers || *aVersl {
		showVersion()
	}

	port := getPort(*aPort)
	addr := *aAddr + ":" + strconv.Itoa(port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	go func() {
		wwwroot := http.Dir("./wwwroot")
		fs := http.FileServer(wwwroot)

		router := httprouter.New()
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
	logrus.Infof("version: %s", Version)
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

func getPort(port int) int {
	if portEnv := os.Getenv("PORT"); portEnv != "" {
		newPort, _ := strconv.Atoi(portEnv)
		if newPort > 0 {
			port = newPort
		}
	}
	return port
}

func showVersion() {
	fmt.Println(Version)
	os.Exit(1)
}

func showUsage() {
	flag.Usage()
	os.Exit(1)
}
