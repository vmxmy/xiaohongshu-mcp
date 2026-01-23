package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Config 全局配置
type Config struct {
	URLs      URLsConfig      `yaml:"urls"`
	Selectors SelectorsConfig `yaml:"selectors"`
	Timeouts  TimeoutsConfig  `yaml:"timeouts"`
	Intervals IntervalsConfig `yaml:"intervals"`
	Limits    LimitsConfig    `yaml:"limits"`
	API       APIConfig       `yaml:"api"`
	Schedule  ScheduleConfig  `yaml:"schedule"`
	Search    SearchConfig    `yaml:"search"`
	Logging   LoggingConfig   `yaml:"logging"`
	Retry     RetryConfig     `yaml:"retry"`
}

// URLsConfig URL配置
type URLsConfig struct {
	Main    MainURLs    `yaml:"main"`
	Creator CreatorURLs `yaml:"creator"`
	// 模板URLs
	FeedDetail  string `yaml:"feed_detail"`
	UserProfile string `yaml:"user_profile"`
}

type MainURLs struct {
	Home    string `yaml:"home"`
	Explore string `yaml:"explore"`
	Search  string `yaml:"search"`
}

type CreatorURLs struct {
	Base             string `yaml:"base"`
	Home             string `yaml:"home"`
	PublishImage     string `yaml:"publish_image"`
	PublishVideo     string `yaml:"publish_video"`
	FanAnalytics     string `yaml:"fan_analytics"`
	ContentAnalytics string `yaml:"content_analytics"`
}

// SelectorsConfig 选择器配置
type SelectorsConfig struct {
	Publish PublishSelectors `yaml:"publish"`
}

type PublishSelectors struct {
	UploadContent        string   `yaml:"upload_content"`
	CreatorTab           string   `yaml:"creator_tab"`
	UploadInput          string   `yaml:"upload_input"`
	UploadedImages       string   `yaml:"uploaded_images"`
	TitleInput           string   `yaml:"title_input"`
	TitleMaxLength       string   `yaml:"title_max_length"`
	ContentEditorQL      string   `yaml:"content_editor_ql"`
	ContentEditorTextbox string   `yaml:"content_editor_textbox"`
	ContentPlaceholder   string   `yaml:"content_placeholder"`
	ContentMaxLength     string   `yaml:"content_max_length"`
	TopicContainer       string   `yaml:"topic_container"`
	TopicItem            string   `yaml:"topic_item"`
	SubmitButton         string   `yaml:"submit_button"`
	RadioLabel           string   `yaml:"radio_label"`
	DatetimePicker       string   `yaml:"datetime_picker"`
	DateInput            string   `yaml:"date_input"`
	TimeInput            string   `yaml:"time_input"`
	ConfirmButton        string   `yaml:"confirm_button"`
	Popover              string   `yaml:"popover"`
	ErrorMessage         []string `yaml:"error_message"`
	SuccessMessage       []string `yaml:"success_message"`
}

// TimeoutsConfig 超时配置 (秒)
type TimeoutsConfig struct {
	Navigate      int `yaml:"navigate"`
	PageLoad      int `yaml:"page_load"`
	ElementWait   int `yaml:"element_wait"`
	TabSearch     int `yaml:"tab_search"`
	ImageUpload   int `yaml:"image_upload"`
	PublishResult int `yaml:"publish_result"`
	APIResponse   int `yaml:"api_response"`
}

// IntervalsConfig 间隔配置 (毫秒)
type IntervalsConfig struct {
	Check         int `yaml:"check"`
	TabRetry      int `yaml:"tab_retry"`
	TagInputChar  int `yaml:"tag_input_char"`
	TagInputWait  int `yaml:"tag_input_wait"`
	TagSelectWait int `yaml:"tag_select_wait"`
	ArrowDown     int `yaml:"arrow_down"`
}

// LimitsConfig 限制配置
type LimitsConfig struct {
	MaxTags       int `yaml:"max_tags"`
	MaxTitleWidth int `yaml:"max_title_width"`
	MaxImages     int `yaml:"max_images"`
	MinImages     int `yaml:"min_images"`
	MaxImageSize  int `yaml:"max_image_size"`
	MaxVideoSize  int `yaml:"max_video_size"`
}

// APIConfig API配置
type APIConfig struct {
	InterceptPatterns []string           `yaml:"intercept_patterns"`
	SuccessIndicators []SuccessIndicator `yaml:"success_indicators"`
}

type SuccessIndicator struct {
	Field string      `yaml:"field"`
	Value interface{} `yaml:"value"`
}

// ScheduleConfig 定时发布配置
type ScheduleConfig struct {
	MinHours int `yaml:"min_hours"`
	MaxDays  int `yaml:"max_days"`
}

// SearchConfig 搜索配置
type SearchConfig struct {
	Filters SearchFilters `yaml:"filters"`
}

type SearchFilters struct {
	SortBy      FilterOptions `yaml:"sort_by"`
	NoteType    FilterOptions `yaml:"note_type"`
	PublishTime FilterOptions `yaml:"publish_time"`
	SearchScope FilterOptions `yaml:"search_scope"`
	Location    FilterOptions `yaml:"location"`
}

type FilterOptions struct {
	Default string   `yaml:"default"`
	Options []string `yaml:"options"`
}

// LoggingConfig 日志配置
type LoggingConfig struct {
	Debug               bool   `yaml:"debug"`
	Level               string `yaml:"level"`
	LogSelectorFailures bool   `yaml:"log_selector_failures"`
}

// RetryConfig 重试配置
type RetryConfig struct {
	MaxAttempts        int  `yaml:"max_attempts"`
	Interval           int  `yaml:"interval"`
	ExponentialBackoff bool `yaml:"exponential_backoff"`
}

var globalConfig *Config

// Load 加载配置文件
func Load(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	globalConfig = &cfg
	return &cfg, nil
}

// LoadDefault 加载默认配置文件
func LoadDefault() (*Config, error) {
	// 查找配置文件的可能位置
	possiblePaths := []string{
		"config.yaml",
		"config.yml",
		"configs/config.yaml",
		"configs/config.yml",
	}

	// 获取可执行文件所在目录
	if execPath, err := os.Executable(); err == nil {
		execDir := filepath.Dir(execPath)
		possiblePaths = append(possiblePaths,
			filepath.Join(execDir, "config.yaml"),
			filepath.Join(execDir, "config.yml"),
		)
	}

	// 获取当前工作目录
	if wd, err := os.Getwd(); err == nil {
		possiblePaths = append(possiblePaths,
			filepath.Join(wd, "config.yaml"),
			filepath.Join(wd, "config.yml"),
		)
	}

	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			return Load(path)
		}
	}

	return nil, fmt.Errorf("未找到配置文件，尝试的路径: %v", possiblePaths)
}

// Get 获取全局配置
func Get() *Config {
	if globalConfig == nil {
		panic("配置未加载，请先调用 Load() 或 LoadDefault()")
	}
	return globalConfig
}

// 辅助方法

// BuildFeedDetailURL 构建Feed详情URL
func (c *Config) BuildFeedDetailURL(feedID, xsecToken string) string {
	url := c.URLs.FeedDetail
	url = strings.ReplaceAll(url, "{feedID}", feedID)
	url = strings.ReplaceAll(url, "{xsecToken}", xsecToken)
	return url
}

// BuildUserProfileURL 构建用户主页URL
func (c *Config) BuildUserProfileURL(userID, xsecToken string) string {
	url := c.URLs.UserProfile
	url = strings.ReplaceAll(url, "{userID}", userID)
	url = strings.ReplaceAll(url, "{xsecToken}", xsecToken)
	return url
}

// GetTimeout 获取超时时间 (返回 time.Duration)
func (t *TimeoutsConfig) GetNavigate() time.Duration {
	return time.Duration(t.Navigate) * time.Second
}

func (t *TimeoutsConfig) GetPageLoad() time.Duration {
	return time.Duration(t.PageLoad) * time.Second
}

func (t *TimeoutsConfig) GetElementWait() time.Duration {
	return time.Duration(t.ElementWait) * time.Second
}

func (t *TimeoutsConfig) GetTabSearch() time.Duration {
	return time.Duration(t.TabSearch) * time.Second
}

func (t *TimeoutsConfig) GetImageUpload() time.Duration {
	return time.Duration(t.ImageUpload) * time.Second
}

func (t *TimeoutsConfig) GetPublishResult() time.Duration {
	return time.Duration(t.PublishResult) * time.Second
}

func (t *TimeoutsConfig) GetAPIResponse() time.Duration {
	return time.Duration(t.APIResponse) * time.Second
}

// GetInterval 获取间隔时间 (返回 time.Duration)
func (i *IntervalsConfig) GetCheck() time.Duration {
	return time.Duration(i.Check) * time.Millisecond
}

func (i *IntervalsConfig) GetTabRetry() time.Duration {
	return time.Duration(i.TabRetry) * time.Millisecond
}

func (i *IntervalsConfig) GetTagInputChar() time.Duration {
	return time.Duration(i.TagInputChar) * time.Millisecond
}

func (i *IntervalsConfig) GetTagInputWait() time.Duration {
	return time.Duration(i.TagInputWait) * time.Millisecond
}

func (i *IntervalsConfig) GetTagSelectWait() time.Duration {
	return time.Duration(i.TagSelectWait) * time.Millisecond
}

func (i *IntervalsConfig) GetArrowDown() time.Duration {
	return time.Duration(i.ArrowDown) * time.Millisecond
}
