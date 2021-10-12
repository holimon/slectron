package slectron

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type cmdstdCall struct {
	// cmdBuffer *bytes.Buffer
	stdCaller func(b []byte) error
}

func newcmdstdCall(call func(p []byte) error) *cmdstdCall {
	return &cmdstdCall{stdCaller: call}
}

func (cs *cmdstdCall) Write(p []byte) (n int, err error) {
	defer func() {
		n = len(p)
	}()
	err = cs.stdCaller(p)
	return
}

type cmdWatch struct {
	cmd    *exec.Cmd
	ctx    context.Context
	cancel context.CancelFunc
	//wg block electron subprocess; ch block until start is ready
	wg, ch     *sync.WaitGroup
	or, ow, os sync.Once
}

func newcwdWatch() (cw *cmdWatch) {
	cw = &cmdWatch{wg: &sync.WaitGroup{}, ch: &sync.WaitGroup{}}
	cw.ch.Add(1)
	cw.ctx, cw.cancel = context.WithCancel(context.Background())
	return cw
}

func (cw *cmdWatch) cmdSigHandler() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGABRT, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
	go func() {
		defer cw.cmdStop()
		for {
			select {
			case <-ch:
				return
			case <-cw.ctx.Done():
				return
			}
		}
	}()
}

func (cw *cmdWatch) cmdRun(process string, args ...string) {
	cw.or.Do(func() {
		cw.wg.Add(1)
		cw.ch.Done()
		defer func() {
			cw.wg.Done()
			cw.cmd = nil
		}()
		cw.cmd = exec.CommandContext(cw.ctx, process, args...)
		cw.cmd.Stdout = newcmdstdCall(func(p []byte) error {
			fmt.Println("Stdout", string(p))
			return nil
		})
		cw.cmd.Stderr = newcmdstdCall(func(p []byte) error {
			fmt.Println("Stderr", string(p))
			return nil
		})
		fmt.Println("Starting ", cw.cmd.Args)
		cw.cmd.Start()
		cw.cmdSigHandler()
		fmt.Println("Waitting ", cw.cmd.Args)
		cw.cmd.Wait()
		fmt.Println(cw.cmd.Path, " exited with code: ", cw.cmd.ProcessState.ExitCode())
	})
}

func (cw *cmdWatch) cmdWait() {
	cw.ow.Do(func() {
		done := make(chan struct{})
		go func() {
			cw.ch.Wait()
			done <- struct{}{}
		}()
		select {
		case <-done:
			//nothing
		case <-time.After(time.Second * 5):
			fmt.Println("The Wait method may have been used before the Start method.")
			return
		}
		cw.wg.Wait()
	})
}

func (cw *cmdWatch) cmdStop() {
	cw.os.Do(func() {
		fmt.Println("Stopping...")
		cw.cancel()
	})
}

func (cw *cmdWatch) isRun() bool {
	return cw.cmd != nil
}
