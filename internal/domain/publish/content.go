package publish

import (
	"fmt"
	"time"
)

type ImageContent struct {
	Title        string
	Content      string
	Tags         []string
	ImagePaths   []string
	ScheduleTime *time.Time
}

type VideoContent struct {
	Title        string
	Content      string
	Tags         []string
	VideoPath    string
	ScheduleTime *time.Time
}

type Limits struct {
	MaxTags   int
	MinImages int
	MaxImages int
}

func ValidateImageContent(c ImageContent, limits Limits) error {
	if len(c.ImagePaths) < limits.MinImages {
		return fmt.Errorf("图片数量不足: %d", len(c.ImagePaths))
	}
	if len(c.ImagePaths) > limits.MaxImages {
		return fmt.Errorf("图片数量过多: %d", len(c.ImagePaths))
	}
	if len(c.Tags) > limits.MaxTags {
		return fmt.Errorf("标签数量过多: %d", len(c.Tags))
	}
	return nil
}
