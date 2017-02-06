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
	"sara/benchmark"
	"sara/config"
	"sara/node"
	"sara/service"
	"sara/utils"
	"strconv"
	"syscall"
	"time"

	"github.com/alecthomas/log4go"
	"github.com/urfave/cli"
)

var (
	version     string         = "0.0.1"
	logLevel    []log4go.Level = []log4go.Level{log4go.ERROR, log4go.WARNING, log4go.INFO, log4go.DEBUG}
	app         *cli.App
	cpu_log     string = "/tmp/sara_cpu.out"
	mem_log     string = "/tmp/sara_mem.out"
	blk_log     string = "/tmp/sara_blk.out"
	pidf        string = "/tmp/sara.pid"
	currentnode *node.Node
)

func init() {
	cpu_core := runtime.NumCPU()
	runtime.GOMAXPROCS(cpu_core)
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
			Name:   "version",
			Action: vsn,
		},
		{
			Name:   "stop",
			Usage:  "停止服务，尽量避免直接 kill 服务",
			Action: stop,
		},
		{
			Name:   "setup",
			Usage:  "生成默认配置文件",
			Action: setupConf,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "out,o",
					Usage: "配置文件全路径",
					Value: "/etc/sara/conf.json",
				},
			},
		},
		{
			Name:     "makeconn",
			Usage:    "创建指定个数的连接，测试最大连接数",
			Category: "benchmark",
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
		// init pprof
		if ctx.GlobalBool("debug") {
			log4go.Warn("start collection cpu and mem profile ... ")
			startCpuProfiles()
			startMemProfiles()
			startBlockProfile()
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
	// init config
	config.Load(ctx)
	// init log4go
	filepath := config.GetString("logfile", "")
	idx := config.GetInt("loglevel", 3)
	level := logLevel[idx]
	if filepath != "" {
		fmt.Println("logfile =", filepath, "level =", level)
		log4go.AddFilter("file", log4go.Level(level), log4go.NewFileLogWriter(filepath, false))
	}
	log4go.AddFilter("stdout", log4go.Level(level), log4go.NewConsoleLogWriter())
	//log4go.Debug(">> listener on port = %d", config.GetInt("port", 4222))
	signalHandler(ctx)
	currentnode = node.New(ctx)
	service.StartRPC(currentnode)
	if config.GetBool("enable_tcp", true) {
		currentnode.StartTCP()
	}
	if config.GetBool("enable_ws", true) {
		currentnode.StartWS()
	}
	if config.GetBool("enable_tls", true) {
		currentnode.StartTLS()
	}
	if config.GetBool("enable_wss", true) {
		currentnode.StartWSS()
	}
	if ctx.GlobalBool("debug") {
		go func() {
			for {
				log4go.Info("[debug] num_goroutine : %d", runtime.NumGoroutine())
				time.Sleep(time.Duration(time.Second * 5))
			}
		}()
	}
	savePid()
	currentnode.Wait()
	log4go.Info("👋  server shutdown success.")
	return nil
}

func makeconnForTest(ctx *cli.Context) error {
	l := ctx.String("laddr")
	a := ctx.String("raddr")
	t := ctx.Int("total")
	h := ctx.Int("heartbeat")
	mg := ctx.Int("messagegap")
	ms := ctx.Int("messagesize")
	benchmark.MakeConn(l, a, t, h, mg, ms)
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

func stop(ctx *cli.Context) error {
	b, err := ioutil.ReadFile(pidf)
	if err == nil {
		pid, _ := strconv.Atoi(string(b))
		log4go.Debug("sara_pid==>%d", pid)
		p, e := os.FindProcess(pid)
		if e != nil {
			return e
		}
		if e := p.Signal(syscall.SIGINT); e != nil {
			return e
		}
		log4go.Info("server shutdown success.")
	}
	return nil
}

func startCpuProfiles() {
	f, _ := os.Create(cpu_log)
	if err := pprof.StartCPUProfile(f); err != nil {
		f.Close()
		return
	}
}

func startBlockProfile() {
	runtime.SetBlockProfileRate(1)
}

func startMemProfiles() {
	runtime.MemProfileRate = 1 * 1024
}

func stopCpuProfiles() {
	pprof.StopCPUProfile()
}

func stopBlockProfile() {
	f, _ := os.Create(blk_log)
	if err := pprof.Lookup("block").WriteTo(f, 0); err != nil {
		fmt.Fprintf(os.Stderr, "Can not write %s: %s", *f, err)
	}
	f.Close()
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

//处理信号量
func signalHandler(ctx *cli.Context) {
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c)
		//signal.Notify(c, syscall.SIGUSR1, syscall.SIGKILL)
		for sig := range c {
			//log4go.Warn("signal: %v", sig)
			switch sig {
			case syscall.SIGUSR1:
				if ctx.GlobalBool("debug") {
					stopCpuProfiles()
					stopMemProfiles()
					stopBlockProfile()
					log4go.Warn("stop collection cpu and mem profile ... ")
				}
			case syscall.SIGTSTP:
			case syscall.SIGINT:
				log4go.Warn("stop server.")
				currentnode.Stop()
			}
		}
	}()
}

func setupConf(ctx *cli.Context) error {
	outpath := ctx.String("out")
	b := []byte(config.Template)
	if err := ioutil.WriteFile(outpath, b, 0644); err != nil {
		fmt.Printf("💔  [fail] setup config (%s) error : %v", outpath, err)
		return err
	}
	fmt.Printf("😄  [success] setup config to : %s\n", outpath)
	return nil
}

func vsn(ctx *cli.Context) error {
	fmt.Println("version:", version)
	fmt.Println("source: https://github.com/emsg/sara")
	return nil
}
