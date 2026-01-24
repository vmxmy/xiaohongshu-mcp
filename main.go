package main

import (
	"flag"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/xpzouying/xiaohongshu-mcp/config"
	"github.com/xpzouying/xiaohongshu-mcp/configs"
)

func main() {
	var (
		headless   bool
		binPath    string // 浏览器二进制文件路径
		port       string
		configPath string // 配置文件路径
	)
	flag.BoolVar(&headless, "headless", true, "是否无头模式")
	flag.StringVar(&binPath, "bin", "", "浏览器二进制文件路径")
	flag.StringVar(&port, "port", ":18060", "端口")
	flag.StringVar(&configPath, "config", "", "配置文件路径（默认自动查找 config.yaml）")
	flag.Parse()

	// 加载配置文件
	var err error
	if configPath != "" {
		_, err = config.Load(configPath)
	} else {
		_, err = config.LoadDefault()
	}
	if err != nil {
		logrus.Warnf("加载配置文件失败（将使用默认值）: %v", err)
		// 不退出，继续使用代码中的默认值
	} else {
		logrus.Infof("配置文件加载成功")
	}

	if len(binPath) == 0 {
		binPath = os.Getenv("ROD_BROWSER_BIN")
	}

	configs.InitHeadless(headless)
	configs.SetBinPath(binPath)

	// 初始化服务
	publishUsecase := initPublishUsecase(headless)
	xiaohongshuService := NewXiaohongshuServiceWithUsecase(publishUsecase)

	// 创建并启动应用服务器
	appServer := NewAppServerWithPublish(xiaohongshuService, publishUsecase)
	if err := appServer.Start(port); err != nil {
		logrus.Fatalf("failed to run server: %v", err)
	}
}
