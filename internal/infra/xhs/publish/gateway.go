package publish

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/xpzouying/xiaohongshu-mcp/internal/domain/publish"
	"github.com/xpzouying/xiaohongshu-mcp/internal/infra/browser"
)

var ErrNotReady = errors.New("publish not implemented")

type Config struct {
	PublishImageURL string
	PublishVideoURL string
	Selectors       map[string]string
}

type Gateway struct {
	cfg    Config
	engine browser.Engine
}

func NewGateway(cfg Config, engine browser.Engine) (*Gateway, error) {
	if cfg.PublishImageURL == "" || cfg.PublishVideoURL == "" {
		return nil, errors.New("publish url missing")
	}
	if engine == nil {
		return nil, errors.New("engine missing")
	}
	return &Gateway{cfg: cfg, engine: engine}, nil
}

func (g *Gateway) PublishImage(ctx context.Context, content publish.ImageContent) error {
	if err := g.engine.Start(); err != nil {
		return err
	}
	defer g.engine.Close()

	page, err := g.engine.NewPage()
	if err != nil {
		return err
	}
	defer page.Close()

	if err := page.Goto(g.cfg.PublishImageURL); err != nil {
		return fmt.Errorf("publish image goto url: %w", err)
	}
	// 等待上传输入框可见
	if err := page.WaitVisible(g.cfg.Selectors["upload_input"]); err != nil {
		return fmt.Errorf("publish image wait upload_input(%s): %w", g.cfg.Selectors["upload_input"], err)
	}
	if err := page.SetFiles(g.cfg.Selectors["upload_input"], content.ImagePaths); err != nil {
		return fmt.Errorf("publish image upload_input(%s): %w", g.cfg.Selectors["upload_input"], err)
	}
	// 等待标题输入框可见（图片上传后才出现）
	if err := page.WaitVisible(g.cfg.Selectors["title_input"]); err != nil {
		return fmt.Errorf("publish image wait title_input(%s): %w", g.cfg.Selectors["title_input"], err)
	}
	if err := page.Fill(g.cfg.Selectors["title_input"], content.Title); err != nil {
		return fmt.Errorf("publish image title_input(%s): %w", g.cfg.Selectors["title_input"], err)
	}
	// 等待内容编辑器可见
	if err := page.WaitVisible(g.cfg.Selectors["content"]); err != nil {
		return fmt.Errorf("publish image wait content(%s): %w", g.cfg.Selectors["content"], err)
	}
	if err := page.Fill(g.cfg.Selectors["content"], content.Content); err != nil {
		return fmt.Errorf("publish image content(%s): %w", g.cfg.Selectors["content"], err)
	}

	// 提交前短暂等待，确保内容已输入完成
	logrus.Info("内容填写完成，等待2秒让页面渲染完成...")
	time.Sleep(2 * time.Second)

	// 点击发布按钮
	submitSelector := g.cfg.Selectors["submit"]
	logrus.Infof("=== 准备点击发布按钮 ===")
	logrus.Infof("选择器: %s", submitSelector)

	// 先检查按钮是否可见
	isVisible, err := page.IsVisible(submitSelector)
	if err != nil {
		logrus.Errorf("检查发布按钮可见性失败: %v", err)
	} else {
		logrus.Infof("发布按钮可见性: %v", isVisible)
	}

	// 滚动到按钮位置（对无头模式很重要）
	logrus.Info("滚动到发布按钮位置...")
	if err := page.ScrollIntoView(submitSelector); err != nil {
		logrus.Warnf("滚动到发布按钮失败: %v (继续尝试)", err)
	}

	// 等待按钮稳定
	time.Sleep(1 * time.Second)

	logrus.Info(">>> 使用强制点击发布按钮（JavaScript）...")
	if err := page.ClickForce(submitSelector); err != nil {
		return fmt.Errorf("publish image submit(%s): %w", submitSelector, err)
	}
	logrus.Info(">>> 发布按钮已点击！")

	// 等待足够长的时间让小红书处理提交
	logrus.Info("等待10秒，让小红书处理提交...")
	time.Sleep(10 * time.Second)

	logrus.Info("提交完成，关闭浏览器")

	return nil
}

func (g *Gateway) PublishVideo(ctx context.Context, content publish.VideoContent) error {
	if err := g.engine.Start(); err != nil {
		return err
	}
	defer g.engine.Close()

	page, err := g.engine.NewPage()
	if err != nil {
		return err
	}
	defer page.Close()

	if err := page.Goto(g.cfg.PublishVideoURL); err != nil {
		return fmt.Errorf("publish video goto url: %w", err)
	}
	// 等待上传输入框可见
	if err := page.WaitVisible(g.cfg.Selectors["upload_input"]); err != nil {
		return fmt.Errorf("publish video wait upload_input(%s): %w", g.cfg.Selectors["upload_input"], err)
	}
	if err := page.SetFiles(g.cfg.Selectors["upload_input"], []string{content.VideoPath}); err != nil {
		return fmt.Errorf("publish video upload_input(%s): %w", g.cfg.Selectors["upload_input"], err)
	}
	// 等待标题输入框可见（视频上传后才出现）
	if err := page.WaitVisible(g.cfg.Selectors["title_input"]); err != nil {
		return fmt.Errorf("publish video wait title_input(%s): %w", g.cfg.Selectors["title_input"], err)
	}
	if err := page.Fill(g.cfg.Selectors["title_input"], content.Title); err != nil {
		return fmt.Errorf("publish video title_input(%s): %w", g.cfg.Selectors["title_input"], err)
	}
	// 等待内容编辑器可见
	if err := page.WaitVisible(g.cfg.Selectors["content"]); err != nil {
		return fmt.Errorf("publish video wait content(%s): %w", g.cfg.Selectors["content"], err)
	}
	if err := page.Fill(g.cfg.Selectors["content"], content.Content); err != nil {
		return fmt.Errorf("publish video content(%s): %w", g.cfg.Selectors["content"], err)
	}

	// 提交前短暂等待，确保内容已输入完成
	logrus.Info("内容填写完成，等待2秒让页面渲染完成...")
	time.Sleep(2 * time.Second)

	// 点击发布按钮
	submitSelector := g.cfg.Selectors["submit"]
	logrus.Infof("=== 准备点击发布按钮 ===")
	logrus.Infof("选择器: %s", submitSelector)

	// 检查按钮可见性
	isVisible, err := page.IsVisible(submitSelector)
	if err != nil {
		logrus.Errorf("检查发布按钮可见性失败: %v", err)
	} else {
		logrus.Infof("发布按钮可见性: %v", isVisible)
	}

	// 滚动到按钮位置
	logrus.Info("滚动到发布按钮位置...")
	if err := page.ScrollIntoView(submitSelector); err != nil {
		logrus.Warnf("滚动到发布按钮失败: %v (继续尝试)", err)
	}

	// 等待按钮稳定
	time.Sleep(1 * time.Second)

	logrus.Info(">>> 使用强制点击发布按钮...")
	if err := page.ClickForce(submitSelector); err != nil {
		return fmt.Errorf("publish video submit(%s): %w", submitSelector, err)
	}
	logrus.Info(">>> 发布按钮已点击！")

	// 等待提交完成
	logrus.Info("等待10秒，让小红书处理提交...")
	time.Sleep(10 * time.Second)

	logrus.Info("提交完成，关闭浏览器")

	return nil
}

func (g *Gateway) SaveImageDraft(ctx context.Context, content publish.ImageContent) error {
	if err := g.engine.Start(); err != nil {
		return err
	}
	defer g.engine.Close()

	page, err := g.engine.NewPage()
	if err != nil {
		return err
	}
	defer page.Close()

	if err := page.Goto(g.cfg.PublishImageURL); err != nil {
		return fmt.Errorf("save image draft goto url: %w", err)
	}
	// 等待上传输入框可见
	if err := page.WaitVisible(g.cfg.Selectors["upload_input"]); err != nil {
		return fmt.Errorf("save image draft wait upload_input(%s): %w", g.cfg.Selectors["upload_input"], err)
	}
	if err := page.SetFiles(g.cfg.Selectors["upload_input"], content.ImagePaths); err != nil {
		return fmt.Errorf("save image draft upload_input(%s): %w", g.cfg.Selectors["upload_input"], err)
	}
	// 等待标题输入框可见（图片上传后才出现）
	if err := page.WaitVisible(g.cfg.Selectors["title_input"]); err != nil {
		return fmt.Errorf("save image draft wait title_input(%s): %w", g.cfg.Selectors["title_input"], err)
	}
	if err := page.Fill(g.cfg.Selectors["title_input"], content.Title); err != nil {
		return fmt.Errorf("save image draft title_input(%s): %w", g.cfg.Selectors["title_input"], err)
	}
	// 等待内容编辑器可见
	if err := page.WaitVisible(g.cfg.Selectors["content"]); err != nil {
		return fmt.Errorf("save image draft wait content(%s): %w", g.cfg.Selectors["content"], err)
	}
	if err := page.Fill(g.cfg.Selectors["content"], content.Content); err != nil {
		return fmt.Errorf("save image draft content(%s): %w", g.cfg.Selectors["content"], err)
	}

	// 等待页面渲染完成
	logrus.Info("内容填写完成，等待2秒...")
	time.Sleep(2 * time.Second)

	// 点击暂存按钮
	saveDraftSelector := g.cfg.Selectors["save_draft"]
	logrus.Infof("准备点击暂存按钮: %s", saveDraftSelector)

	// 滚动并强制点击
	if err := page.ScrollIntoView(saveDraftSelector); err != nil {
		logrus.Warnf("滚动到暂存按钮失败: %v", err)
	}
	time.Sleep(500 * time.Millisecond)

	if err := page.ClickForce(saveDraftSelector); err != nil {
		return fmt.Errorf("save image draft save_draft(%s): %w", saveDraftSelector, err)
	}
	logrus.Info("已点击暂存按钮")

	// 等待草稿保存完成
	time.Sleep(3 * time.Second)

	return nil
}

func (g *Gateway) SaveVideoDraft(ctx context.Context, content publish.VideoContent) error {
	if err := g.engine.Start(); err != nil {
		return err
	}
	defer g.engine.Close()

	page, err := g.engine.NewPage()
	if err != nil {
		return err
	}
	defer page.Close()

	if err := page.Goto(g.cfg.PublishVideoURL); err != nil {
		return fmt.Errorf("save video draft goto url: %w", err)
	}
	// 等待上传输入框可见
	if err := page.WaitVisible(g.cfg.Selectors["upload_input"]); err != nil {
		return fmt.Errorf("save video draft wait upload_input(%s): %w", g.cfg.Selectors["upload_input"], err)
	}
	if err := page.SetFiles(g.cfg.Selectors["upload_input"], []string{content.VideoPath}); err != nil {
		return fmt.Errorf("save video draft upload_input(%s): %w", g.cfg.Selectors["upload_input"], err)
	}
	// 等待标题输入框可见（视频上传后才出现）
	if err := page.WaitVisible(g.cfg.Selectors["title_input"]); err != nil {
		return fmt.Errorf("save video draft wait title_input(%s): %w", g.cfg.Selectors["title_input"], err)
	}
	if err := page.Fill(g.cfg.Selectors["title_input"], content.Title); err != nil {
		return fmt.Errorf("save video draft title_input(%s): %w", g.cfg.Selectors["title_input"], err)
	}
	// 等待内容编辑器可见
	if err := page.WaitVisible(g.cfg.Selectors["content"]); err != nil {
		return fmt.Errorf("save video draft wait content(%s): %w", g.cfg.Selectors["content"], err)
	}
	if err := page.Fill(g.cfg.Selectors["content"], content.Content); err != nil {
		return fmt.Errorf("save video draft content(%s): %w", g.cfg.Selectors["content"], err)
	}

	// 等待页面渲染完成
	logrus.Info("内容填写完成，等待2秒...")
	time.Sleep(2 * time.Second)

	// 点击暂存按钮
	saveDraftSelector := g.cfg.Selectors["save_draft"]
	logrus.Infof("准备点击暂存按钮: %s", saveDraftSelector)

	// 滚动并强制点击
	if err := page.ScrollIntoView(saveDraftSelector); err != nil {
		logrus.Warnf("滚动到暂存按钮失败: %v", err)
	}
	time.Sleep(500 * time.Millisecond)

	if err := page.ClickForce(saveDraftSelector); err != nil {
		return fmt.Errorf("save video draft save_draft(%s): %w", saveDraftSelector, err)
	}
	logrus.Info("已点击暂存按钮")

	// 等待草稿保存完成
	time.Sleep(3 * time.Second)

	return nil
}
