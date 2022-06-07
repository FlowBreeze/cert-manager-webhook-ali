// Package exit
// 此包用于处理程序退出信号,当导入此包后所有功能即生效
// 该模块封装了一些退出方法,用于简化程序的退出逻辑
// 同时该模块会持续监听应用收到的信号(os.Signal)
// 当收到 quitSignals 中指定的信号后,或是调用该模块的退出方法后,程序不会立刻退出
// 而是会关闭 BackgroundCtx 等待程序内进行进一步处理
// 如果程序收到两次 quitSignals 的信号,该模块会调用 os.Exit 方法并返回退出码 130
package exit

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"runtime/debug"
	"sync"
	"syscall"
)

var (
	// 程序可通过 BackgroundCtx 的 Done 方法返回的 channel 判断该程序是否被要求退出
	background, cancel = context.WithCancel(context.Background())
	// StopWg 当程序准备退出时,会等待该集合中的协程执行完毕后退出
	StopWg sync.WaitGroup
	// 该模块所监听的退出信号
	quitSignals = []os.Signal{syscall.SIGINT, syscall.SIGTERM}
	// 调用退出方法时传递的参数,只可使用一次
	exitResultCh = make(chan interface{}, 1)
	// 退出处理方法,仅调用一次
	shut    sync.Once
	errExit = errors.New("the application is exit")
)

type errorStr string

// init 该包在导入后会开始追踪系统信号
// 在收到 quitSignals 中指定的信号时, 会通过 返回的 chain 通知程序处理退出请求
// 在第二次收到信号后会调用 os.Exit 方法强制退出
func init() {
	c := make(chan os.Signal, 2)
	signal.Notify(c, quitSignals...)
	go func() {
		s := <-c // 第一次收到退出信号
		exitResultCh <- s
		cancel()
		<-c // 第二次收到退出信号
		os.Exit(130)
	}()
}

func BackgroundCtx() context.Context {
	return background
}

// WaitForExit 表示该程序已经做好被强制中断的准备,可以通过 Exit, PanicIfNotNil 进行退出
// WaitForExit 调用后将阻塞,直到收到退出信号
// 然后该方法会通过 StopWg 等待程序中的退出后处理代码执行完毕,并返回相应的退出结果
// 	如果程序通过 Exit 中断,则返回对应的退出码
// 	如果程序通过 Panic 中断,则返回 2
// 	如果程序通过 quitSignals 中断,则返回 130
// 该方法需要在主协程调用,通常在 main 方法中以 defer 的形式调用
// 并在 main 方法执行的最后通过 Exit 方法显式退出
func WaitForExit() {
	// 如果在进入此方法时依旧没有调用过退出方法,则自动调用退出方法
	// 如果不调用会导致程序一直运行,出现死锁异常
	shut.Do(func() {
		cancel()
	})

	<-background.Done()
	StopWg.Wait()
	select {
	case result := <-exitResultCh:
		switch r := result.(type) {
		case int:
			os.Exit(r)
		case errorStr:
			print(r)
			os.Exit(2)
		case os.Signal:
			os.Exit(130)
		}
	default:
		// 进入此处可能是因为调用该方法的代码块已经执行完毕且后处理方法已经执行完毕
		// 这种情况下什么也不做,等待程序自行结束
		// 否则 panic 处理可能会被过早结束,从而不打印
	}
}

// Panic 退出程序,并展示堆栈信息,返回退出码 2
// 不会立即退出,而是会等到主协程执行 WaitForExit 时退出
func Panic(err error) error {
	stack := debug.Stack()
	shut.Do(func() {
		defer cancel()
		e := errorStr(fmt.Sprintf("%v\n%s", err, stack))
		exitResultCh <- e
	})
	return err
}

// Error 退出程序,并展示堆栈信息,返回退出码 2
// 不会立即退出,而是会等到主协程执行 WaitForExit 时退出
func Error(text string) error {
	return Panic(errors.New(text))
}

// Exit 退出程序,指定退出码
// 不会立即退出,而是会等到主协程执行 WaitForExit 时退出
func Exit(code int) error {
	shut.Do(func() {
		defer cancel()
		exitResultCh <- code
	})
	return errExit
}
