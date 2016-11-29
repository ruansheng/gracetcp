package gracetcp

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	GRACEFUL_ENVIRON_KEY    = "IS_GRACEFUL"
	GRACEFUL_ENVIRON_STRING = GRACEFUL_ENVIRON_KEY + "=1"

	DEFAULT_READ_TIMEOUT  = 60 * time.Second
	DEFAULT_WRITE_TIMEOUT = DEFAULT_READ_TIMEOUT
)

func ListenTCP(ip string, port int) (*Listener, error) {
	// 获取环境变量
	isGraceful := false
	if os.Getenv(GRACEFUL_ENVIRON_KEY) != "" {
		isGraceful = true
	}

	var ln net.Listener
	var err error

	// 如果是重启的，就新写一个文件描述符
	if isGraceful {
		file := os.NewFile(3, "")
		ln, err = net.FileListener(file)
	} else {
		ln, err = net.ListenTCP("tcp", &net.TCPAddr{net.ParseIP(ip), port, ""})
	}

	if err != nil {
		return nil, err
	}
	listener := ln.(*net.TCPListener)
	li := NewListener(listener)

	go listenSignals(li)
	return li, nil
}

func listenSignals(li *Listener) {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGUSR2)
	for {
		sig := <-c
		switch sig {
		case syscall.SIGUSR2:
			//启动新进程
			err := startNewProcess(li)
			if err != nil {
				fmt.Println(err)
			} else {
				// 关闭老进程
				stopOldProcess(li)
			}
		case syscall.SIGTERM:
			// 关闭老进程
			stopOldProcess(li)
		}
	}
}

func stopOldProcess(li *Listener) {
	fmt.Println("stopOldProcess")
	li.Close()
}

func startNewProcess(li *Listener) error {
	listenerFd, err := (*li).GetFd()
	fmt.Println("fd:", listenerFd)
	if err != nil {
		return fmt.Errorf("failed to get socket file descriptor: %v", err)
	}
	path := os.Args[0]

	// 设置标识优雅重启的环境变量
	environList := []string{}
	for _, value := range os.Environ() {
		if value != GRACEFUL_ENVIRON_STRING {
			environList = append(environList, value)
		}
	}
	environList = append(environList, GRACEFUL_ENVIRON_STRING)

	execSpec := &syscall.ProcAttr{
		Env:   environList,
		Files: []uintptr{os.Stdin.Fd(), os.Stdout.Fd(), os.Stderr.Fd(), listenerFd},
	}

	fork, err1 := syscall.ForkExec(path, os.Args, execSpec)
	if err1 != nil {
		return fmt.Errorf("failed to forkexec: %v", err1)
	}
	fmt.Println("start new process success, pid %d.", fork)
	return nil
}
