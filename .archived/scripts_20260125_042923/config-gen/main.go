package main

import (
	"flag"
	"os"

	"github.com/xpzouying/xiaohongshu-mcp/internal/configgen"
)

func parseFlags(args []string) configgen.Options {
	fs := flag.NewFlagSet("config-gen", flag.ContinueOnError)
	opt := configgen.DefaultOptions()
	fs.StringVar(&opt.OutputPath, "output", opt.OutputPath, "输出配置文件路径")
	fs.BoolVar(&opt.Backup, "backup", opt.Backup, "是否备份旧配置")
	fs.BoolVar(&opt.Headless, "headless", opt.Headless, "是否无头模式")
	fs.BoolVar(&opt.DryRun, "dry-run", opt.DryRun, "仅生成报告不写入")
	fs.BoolVar(&opt.VerifyOnly, "verify-only", opt.VerifyOnly, "仅校验不写入")
	_ = fs.Parse(args)
	return opt
}

func main() {
	_ = parseFlags(os.Args[1:])
}
