package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/mattn/go-isatty"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var rootCtx = context.Background()

func listenAndServe(ctx context.Context, addr string) error {
	l, err := (&net.ListenConfig{}).Listen(ctx, "tcp", addr)
	if err != nil {
		return err
	}

	server := &http.Server{
		Addr:    addr,
		Handler: http.HandlerFunc(iamService.Handle),
	}
	return server.Serve(l)
}

func start(addr string) error {
	ctx, cancel := context.WithCancel(rootCtx)
	defer cancel()
	return listenAndServe(ctx, "127.0.0.1:9000")
}

var progname = filepath.Base(os.Args[0])

func cmdlineErr(msg string) {
	fmt.Fprintf(os.Stderr, "%s: %s\n", progname, msg)
}

func initializerLogger() {
	if isatty.IsTerminal(os.Stdout.Fd()) {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339Nano})
	}
}

func main() {
	initializerLogger()
	var addr string
	flag.StringVar(&addr, "bind", "127.0.0.1:9000", "bind to `ADDRESS`")
	flag.Parse()
	if len(flag.Args()) < 1 {
		flag.PrintDefaults()
		cmdlineErr("specify a path to the YAML file")
		os.Exit(255)
	}
	b, err := ioutil.ReadFile(flag.Args()[0])
	if err != nil {
		cmdlineErr(err.Error())
		os.Exit(1)
	}
	reg, err := buildRegistryFromYAML(b)
	if err != nil {
		cmdlineErr(err.Error())
		os.Exit(1)
	}
	registerAPISet(reg)
	err = start(addr)
	if err != nil {
		cmdlineErr(err.Error())
		os.Exit(1)
	}
}
