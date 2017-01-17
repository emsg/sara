package main

/* TODO
https://github.com/tidwall/gjson 处理json
https://github.com/alecthomas/log4go 日志
*/
import (
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"sara/config"
	"sara/node"
	"sara/utils"
	"strconv"
	"syscall"

	"github.com/alecthomas/log4go"
	"github.com/urfave/cli"
)

var (
	logLevel []log4go.Level = []log4go.Level{log4go.ERROR, log4go.WARNING, log4go.INFO, log4go.DEBUG}
	app      *cli.App
	cpu_log  string = "/tmp/sara_cpu.out"
	mem_log  string = "/tmp/sara_mem.out"
	pidf     string = "/tmp/sara.pid"
)

func init() {
	cpu_core := runtime.NumCPU()
	runtime.GOMAXPROCS(cpu_core)
	fmt.Println("cpu_core_total", cpu_core)

	app = cli.NewApp()
	app.Name = os.Args[0]
	app.Usage = "SARA IM Server"
	app.Version = "0.0.1"
	app.Author = "liangc"
	app.Email = "cc14514@icloud.com"
	app.Flags = utils.InitFlags()
	app.Action = sara
	app.Commands = []cli.Command{
		{
			Name:     "makeconn",
			Usage:    "创建指定个数的连接，测试最大连接数",
			Category: "test",
			Flags:    utils.InitFlagsForTestOfMakeConn(),
			Action:   makeconnForTest,
		},
		{
			Name:     "pprof",
			Usage:    "将 cpu/mem 信息写入文件",
			Category: "debug",
			Action:   pprofForDebug,
		},
	}
	app.Before = func(ctx *cli.Context) error {
		// init log4go
		filepath := ctx.GlobalString("logfile")
		idx := ctx.GlobalInt("loglevel")
		level := logLevel[idx]
		if filepath != "" {
			fmt.Println("logfile =", filepath, "level =", level)
			log4go.AddFilter("file", log4go.Level(level), log4go.NewFileLogWriter(filepath, false))
		}
		log4go.AddFilter("stdout", log4go.Level(level), log4go.NewConsoleLogWriter())
		// init config
		if ctx.GlobalString("hostname") == "" {
			defer os.Exit(0)
			log4go.Error("❌❌❌  hostname or ip can not empty, use --hostname to set.")
		}
		config.LoadFromCtx(ctx)
		// init pprof
		if ctx.GlobalBool("debug") {
			log4go.Warn("start collection cpu and mem profile ... ")
			startCpuProfiles()
			startMemProfiles()
			go func() {
				c := make(chan os.Signal)
				signal.Notify(c, syscall.SIGUSR1)
				for sig := range c {
					log4go.Warn(" 📶  %v", sig)
					stopCpuProfiles()
					stopMemProfiles()
					log4go.Warn("stop collection cpu and mem profile ... ")
				}
			}()
		}
		return nil
	}
	app.After = func(ctx *cli.Context) error {
		log4go.Close()
		return nil
	}
}
func main() {
	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
	}
}

func sara(ctx *cli.Context) error {
	//log4go.Debug(">> listener on port = %d", config.GetInt("port", 4222))
	//service.StartRPC(ctx)
	n := node.New(ctx)
	n.StartTCP()
	n.StartWS()
	savePid()
	n.Wait()
	return nil
}

func makeconnForTest(ctx *cli.Context) error {
	a := ctx.String("addr")
	t := ctx.Int("total")
	h := ctx.Int("heartbeat")
	utils.MakeConn(a, t, h)
	return nil
}

//向 /tmp/sara.pid 进程发送 SIGNUSR1 信号，用来刷新 cpu / mem 日志文件
func pprofForDebug(ctx *cli.Context) error {
	b, err := ioutil.ReadFile(pidf)
	if err == nil {
		pid, _ := strconv.Atoi(string(b))
		log4go.Debug("sara_pid==>%d", pid)
		p, e := os.FindProcess(pid)
		if e != nil {
			log4go.Info("flush pprof of [%d] fail ; e=%v", pid, e)
			return e
		}
		if e := p.Signal(syscall.SIGUSR1); e != nil {
			log4go.Info("flush pprof of [%d] fail ; e=%v", pid, e)
			return e
		}
		log4go.Info("flush pprof of [%d] success", pid)
	}
	return err
}

func startCpuProfiles() {
	f, _ := os.Create(cpu_log)
	if err := pprof.StartCPUProfile(f); err != nil {
		f.Close()
		return
	}
}

func startMemProfiles() {
	runtime.MemProfileRate = 1 * 1024
}

func stopCpuProfiles() {
	pprof.StopCPUProfile()
}

func stopMemProfiles() {
	f, _ := os.Create(mem_log)
	defer f.Close()
	pprof.WriteHeapProfile(f)
}

func savePid() {
	pid := os.Getpid()
	f, err := os.Create(pidf)
	if err == nil {
		defer f.Close()
		f.WriteString(fmt.Sprintf("%d", pid))
	}
}
