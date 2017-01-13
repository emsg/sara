package main

/* TODO
https://github.com/tidwall/gjson 处理json
https://github.com/alecthomas/log4go 日志
*/
import (
	"fmt"
	"os"
	"runtime"
	"sara/node"
	"sara/utils"

	"github.com/alecthomas/log4go"
	"github.com/urfave/cli"
)

var (
	logLevel []log4go.Level = []log4go.Level{log4go.ERROR, log4go.WARNING, log4go.INFO, log4go.DEBUG}
	app      *cli.App
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
	}
	app.Before = func(ctx *cli.Context) error {
		filepath := ctx.GlobalString("logfile")
		idx := ctx.GlobalInt("loglevel")
		level := logLevel[idx]
		if filepath != "" {
			fmt.Println("logfile =", filepath, "level =", level)
			log4go.AddFilter("file", log4go.Level(level), log4go.NewFileLogWriter(filepath, false))
		}
		log4go.AddFilter("stdout", log4go.Level(level), log4go.NewConsoleLogWriter())
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
	log4go.Debug(">> listener on port = %d", ctx.GlobalInt("port"))
	//service.StartRPC(ctx)
	n := node.New(ctx)
	n.StartTCP()
	n.StartWS()

	n.Wait()
	return nil
}

func makeconnForTest(ctx *cli.Context) error {
	a := ctx.String("addr")
	t := ctx.Int("total")
	utils.MakeConn(a, t)
	return nil
}
