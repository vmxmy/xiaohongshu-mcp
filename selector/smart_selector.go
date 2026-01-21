package selector

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/go-rod/rod"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// SelectorConfig 选择器配置
type SelectorConfig struct {
	Version     string                       `yaml:"version"`
	LastUpdated string                       `yaml:"last_updated"`
	Description string                       `yaml:"description"`
	Elements    map[string]*ElementSelectors `yaml:",inline"`
}

// ElementSelectors 元素选择器集合
type ElementSelectors struct {
	Name       string              `yaml:"name"`
	Primary    []SelectorItem      `yaml:"primary"`
	Fallback   []SelectorItem      `yaml:"fallback"`
	TextMatch  []string            `yaml:"text_match"`
	Validation *ValidationRules    `yaml:"validation"`
}

// SelectorItem 单个选择器项
type SelectorItem struct {
	Selector    string  `yaml:"selector"`
	Version     string  `yaml:"version"`
	Confidence  float64 `yaml:"confidence"`
	Description string  `yaml:"description"`
}

// ValidationRules 验证规则
type ValidationRules struct {
	MustBeVisible   bool              `yaml:"must_be_visible"`
	MustBeClickable bool              `yaml:"must_be_clickable"`
	MustBeEditable  bool              `yaml:"must_be_editable"`
	ParentContains  []string          `yaml:"parent_contains"`
	TextContains    []string          `yaml:"text_contains"`
	Attributes      map[string]string `yaml:"attributes"`
}

// SmartSelector 智能选择器引擎
type SmartSelector struct {
	config     *SelectorConfig
	configPath string
	configMu   sync.RWMutex
	stats      *SelectorStats
	page       *rod.Page
}

// SelectorStats 选择器统计信息
type SelectorStats struct {
	mu      sync.RWMutex
	records map[string]*ElementStats
}

// ElementStats 元素统计
type ElementStats struct {
	ElementName    string                  `json:"element_name"`
	TotalAttempts  int                     `json:"total_attempts"`
	SuccessCount   int                     `json:"success_count"`
	SelectorStats  map[string]*UsageStats  `json:"selector_stats"`
	LastUsed       time.Time               `json:"last_used"`
	LastSuccessful string                  `json:"last_successful"`
}

// UsageStats 使用统计
type UsageStats struct {
	Selector     string    `json:"selector"`
	SuccessCount int       `json:"success_count"`
	FailCount    int       `json:"fail_count"`
	AvgTime      float64   `json:"avg_time_ms"`
	LastUsed     time.Time `json:"last_used"`
}

// NewSmartSelector 创建智能选择器
func NewSmartSelector(configPath string, page *rod.Page) (*SmartSelector, error) {
	s := &SmartSelector{
		configPath: configPath,
		page:       page,
		stats:      NewSelectorStats(),
	}

	if err := s.LoadConfig(); err != nil {
		return nil, fmt.Errorf("加载配置失败: %w", err)
	}

	// 启动配置热更新监控
	go s.watchConfigChanges()

	return s, nil
}

// LoadConfig 加载配置
func (s *SmartSelector) LoadConfig() error {
	s.configMu.Lock()
	defer s.configMu.Unlock()

	data, err := os.ReadFile(s.configPath)
	if err != nil {
		return fmt.Errorf("读取配置文件失败: %w", err)
	}

	var config SelectorConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("解析配置文件失败: %w", err)
	}

	s.config = &config
	logrus.Infof("✓ 加载选择器配置成功，版本: %s", config.Version)
	return nil
}

// watchConfigChanges 监控配置文件变化
func (s *SmartSelector) watchConfigChanges() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	var lastModTime time.Time

	for range ticker.C {
		info, err := os.Stat(s.configPath)
		if err != nil {
			continue
		}

		if info.ModTime().After(lastModTime) {
			lastModTime = info.ModTime()
			if err := s.LoadConfig(); err != nil {
				logrus.Errorf("热更新配置失败: %v", err)
			} else {
				logrus.Info("✓ 配置文件已热更新")
			}
		}
	}
}

// FindElement 智能查找元素
func (s *SmartSelector) FindElement(elementName string) (*rod.Element, error) {
	s.configMu.RLock()
	elementConfig, exists := s.config.Elements[elementName]
	s.configMu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("未找到元素配置: %s", elementName)
	}

	logrus.Infof("开始智能查找元素: %s", elementConfig.Name)

	// 记录尝试
	s.stats.RecordAttempt(elementName)

	// 1. 尝试主选择器
	for i, item := range elementConfig.Primary {
		logrus.Debugf("尝试主选择器 [%d/%d]: %s (置信度: %.2f)",
			i+1, len(elementConfig.Primary), item.Selector, item.Confidence)

		elem, duration, err := s.trySelector(item.Selector)
		if err != nil {
			s.stats.RecordFailure(elementName, item.Selector, duration)
			logrus.Debugf("主选择器失败: %v", err)
			continue
		}

		// 验证元素
		if elementConfig.Validation != nil {
			if !s.validateElement(elem, elementConfig.Validation) {
				s.stats.RecordFailure(elementName, item.Selector, duration)
				logrus.Debugf("元素验证失败")
				continue
			}
		}

		// 成功
		s.stats.RecordSuccess(elementName, item.Selector, duration)
		logrus.Infof("✓ 成功找到元素，使用主选择器: %s (耗时: %.2fms)", item.Selector, duration)
		return elem, nil
	}

	// 2. 尝试备用选择器
	logrus.Info("主选择器全部失败，尝试备用选择器...")
	for i, item := range elementConfig.Fallback {
		logrus.Debugf("尝试备用选择器 [%d/%d]: %s (置信度: %.2f)",
			i+1, len(elementConfig.Fallback), item.Selector, item.Confidence)

		elem, duration, err := s.trySelector(item.Selector)
		if err != nil {
			s.stats.RecordFailure(elementName, item.Selector, duration)
			continue
		}

		if elementConfig.Validation != nil {
			if !s.validateElement(elem, elementConfig.Validation) {
				s.stats.RecordFailure(elementName, item.Selector, duration)
				continue
			}
		}

		s.stats.RecordSuccess(elementName, item.Selector, duration)
		logrus.Infof("✓ 成功找到元素，使用备用选择器: %s (耗时: %.2fms)", item.Selector, duration)
		return elem, nil
	}

	// 3. 尝试文本匹配
	if len(elementConfig.TextMatch) > 0 {
		logrus.Info("尝试通过文本匹配查找元素...")
		elem, err := s.findByText(elementConfig.TextMatch)
		if err == nil && elem != nil {
			logrus.Infof("✓ 通过文本匹配找到元素")
			return elem, nil
		}
	}

	return nil, fmt.Errorf("所有选择器都失败了: %s", elementName)
}

// trySelector 尝试单个选择器
func (s *SmartSelector) trySelector(selector string) (*rod.Element, float64, error) {
	start := time.Now()
	elem, err := s.page.Timeout(3 * time.Second).Element(selector)
	duration := float64(time.Since(start).Milliseconds())

	return elem, duration, err
}

// validateElement 验证元素
func (s *SmartSelector) validateElement(elem *rod.Element, rules *ValidationRules) bool {
	// 1. 检查可见性
	if rules.MustBeVisible {
		visible, err := elem.Visible()
		if err != nil || !visible {
			logrus.Debug("验证失败: 元素不可见")
			return false
		}
	}

	// 2. 检查可点击性
	if rules.MustBeClickable {
		// 检查元素是否可交互
		_, err := elem.Interactable()
		if err != nil {
			logrus.Debug("验证失败: 元素不可点击")
			return false
		}
	}

	// 3. 检查可编辑性
	if rules.MustBeEditable {
		// 检查 contenteditable 属性
		attr, err := elem.Attribute("contenteditable")
		if err != nil || (attr == nil || *attr != "true") {
			logrus.Debug("验证失败: 元素不可编辑")
			return false
		}
	}

	// 4. 检查父元素
	if len(rules.ParentContains) > 0 {
		parent, err := elem.Parent()
		if err != nil {
			return false
		}

		html, err := parent.HTML()
		if err != nil {
			return false
		}

		for _, keyword := range rules.ParentContains {
			if !strings.Contains(html, keyword) {
				logrus.Debugf("验证失败: 父元素不包含 '%s'", keyword)
				return false
			}
		}
	}

	// 5. 检查文本内容
	if len(rules.TextContains) > 0 {
		text, err := elem.Text()
		if err != nil {
			return false
		}

		found := false
		for _, keyword := range rules.TextContains {
			if strings.Contains(text, keyword) {
				found = true
				break
			}
		}

		if !found {
			logrus.Debug("验证失败: 文本不匹配")
			return false
		}
	}

	return true
}

// findByText 通过文本查找元素
func (s *SmartSelector) findByText(texts []string) (*rod.Element, error) {
	for _, text := range texts {
		// 尝试查找包含指定文本的按钮
		buttons, err := s.page.Elements("button")
		if err != nil {
			continue
		}

		for _, btn := range buttons {
			btnText, _ := btn.Text()
			if strings.Contains(btnText, text) {
				return btn, nil
			}
		}
	}

	return nil, fmt.Errorf("未找到匹配文本的元素")
}

// GetStats 获取统计信息
func (s *SmartSelector) GetStats() map[string]*ElementStats {
	return s.stats.GetAll()
}

// SaveStats 保存统计信息
func (s *SmartSelector) SaveStats(path string) error {
	data, err := json.MarshalIndent(s.stats.GetAll(), "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// NewSelectorStats 创建统计对象
func NewSelectorStats() *SelectorStats {
	return &SelectorStats{
		records: make(map[string]*ElementStats),
	}
}

// RecordAttempt 记录尝试
func (s *SelectorStats) RecordAttempt(elementName string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.records[elementName]; !exists {
		s.records[elementName] = &ElementStats{
			ElementName:   elementName,
			SelectorStats: make(map[string]*UsageStats),
		}
	}

	s.records[elementName].TotalAttempts++
	s.records[elementName].LastUsed = time.Now()
}

// RecordSuccess 记录成功
func (s *SelectorStats) RecordSuccess(elementName, selector string, duration float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	elem := s.records[elementName]
	elem.SuccessCount++
	elem.LastSuccessful = selector

	if _, exists := elem.SelectorStats[selector]; !exists {
		elem.SelectorStats[selector] = &UsageStats{
			Selector: selector,
		}
	}

	stat := elem.SelectorStats[selector]
	stat.SuccessCount++
	stat.LastUsed = time.Now()

	// 更新平均时间
	if stat.AvgTime == 0 {
		stat.AvgTime = duration
	} else {
		stat.AvgTime = (stat.AvgTime + duration) / 2
	}
}

// RecordFailure 记录失败
func (s *SelectorStats) RecordFailure(elementName, selector string, duration float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	elem := s.records[elementName]

	if _, exists := elem.SelectorStats[selector]; !exists {
		elem.SelectorStats[selector] = &UsageStats{
			Selector: selector,
		}
	}

	elem.SelectorStats[selector].FailCount++
	elem.SelectorStats[selector].LastUsed = time.Now()
}

// GetAll 获取所有统计
func (s *SelectorStats) GetAll() map[string]*ElementStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string]*ElementStats)
	for k, v := range s.records {
		result[k] = v
	}

	return result
}
