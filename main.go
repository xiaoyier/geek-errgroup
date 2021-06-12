package main

import (
	"context"
	"errors"
	"fmt"
	"golang.org/x/sync/errgroup"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {

	addrs := [...]string{":8000", ":8001"}
	group, ctx := errgroup.WithContext(context.Background())
	for i, addr := range addrs {
		handleFunc := handleHello1
		if i == 1 {
			handleFunc = hanldeHello2
		}
		server := newHttpServer(addr, handleFunc)
		group.Go(func() error {
			<-ctx.Done()
			timeOutCtx, cancel := context.WithTimeout(context.Background(), time.Second*20)
			defer cancel()
			err := server.Shutdown(timeOutCtx)
			fmt.Println("http server stopped ", server.Addr)
			return err
		})
		group.Go(func() error {
			// warning: 此处可能存在问题，若当前server能正常启动，另一server启动失败(比如Listen的时候返回了错误)，
			// 则当前server可能先shutdown,后serve, 在main goroutine退出前，可能会放进来几个request
			fmt.Println("http server listening on ", server.Addr)
			return server.ListenAndServe()
		})
	}

	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)
	group.Go(func() error {
		select {
		case <-ctx.Done():
			return nil
		case s := <-sig:
			return errors.New(fmt.Sprintf("terminate by signal: %s", s.String()))
		}
	})

	if err := group.Wait(); err != nil {
		fmt.Println(err)
	}
}

func newHttpServer(addr string, handleFunc func(w http.ResponseWriter, r *http.Request)) *http.Server {

	mux := http.NewServeMux()
	mux.HandleFunc("/hello", handleFunc)
	return &http.Server{
		Addr: addr,
		Handler: mux,
	}
}

func handleHello1(o http.ResponseWriter, r *http.Request) {
	fmt.Fprint(o, "hello world 1")
}

func hanldeHello2(o http.ResponseWriter, r *http.Request) {
	fmt.Fprint(o, "hello world 2")
}



