package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/fsnotify/fsnotify"
)

var excludes []string

func main() {
	excludeStr := flag.String("exclude", ".git,vendor,build", "comma-separated list of directories to exclude")
	excludes = strings.Split(*excludeStr, ",")
	flag.Parse()
	cmdargs := flag.Args()
	if len(cmdargs) < 1 {
		fmt.Fprintf(os.Stderr, "usage: progname [arg1 [arg2... ]]\n")
		os.Exit(1)
	}

	sigch := make(chan os.Signal)
	signal.Notify(sigch, syscall.SIGCHLD, syscall.SIGINT, syscall.SIGTERM)
	defer close(sigch)

	for {
		if err := execute(sigch, cmdargs[0], cmdargs[1:]...); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err.Error())
			os.Exit(0)
		}
	}
}

func execute(sigch chan os.Signal, progname string, args ...string) error {
	fmt.Fprintf(os.Stderr, "execute: %#v %#v\n", progname, args)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cmd := exec.CommandContext(ctx, progname, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true} // https://medium.com/@felixge/killing-a-child-process-and-all-of-its-children-in-go-54079af94773
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Start(); err != nil {
		return err
	}
	defer cmd.Wait()
	defer syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM) // negative pid = sending a signal to a Process Group

	filesch, err := watch(ctx, os.Getenv("PWD"))
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return context.Canceled
		case s := <-sigch:
			switch s {
			case syscall.SIGCHLD:
				return nil
			case syscall.SIGINT, syscall.SIGTERM:
				cancel()
			}
		case <-filesch:
			return nil
		}
	}
}

func excluded(path string) bool {
	if path == "/" {
		return false
	}
	basename := filepath.Base(path)
	for _, s := range excludes {
		if s == basename {
			return true
		}
	}
	return excluded(filepath.Dir(path))
}

func watch(ctx context.Context, root string) (chan string, error) {
	filesch := make(chan string)
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return filesch, err
	}
	err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() && !excluded(path) {
			watcher.Add(path)
		}
		return nil
	})
	go func() {
		defer watcher.Close()
		defer close(filesch)
		for {
			select {
			case <-ctx.Done():
				return
			case evt := <-watcher.Events:
				path := evt.Name
				switch evt.Op {
				case fsnotify.Create:
					if stat, _err := os.Stat(path); _err == nil && stat.IsDir() && !excluded(path) {
						watcher.Add(path)
					}
				case fsnotify.Remove:
					watcher.Remove(path)
				}
				filesch <- path
			}
		}
	}()
	return filesch, err
}
