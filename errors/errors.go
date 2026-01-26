package errors

import "errors"

var ErrNoFeeds = errors.New("没有捕获到 feeds 数据")
var ErrNoFeedDetail = errors.New("没有捕获到 feed 详情数据")

// ErrFeedNotAccessible 笔记不可访问错误
type ErrFeedNotAccessible struct {
	Reason string
}

func (e *ErrFeedNotAccessible) Error() string {
	return "笔记不可访问: " + e.Reason
}

// NewErrFeedNotAccessible 创建笔记不可访问错误
func NewErrFeedNotAccessible(reason string) *ErrFeedNotAccessible {
	return &ErrFeedNotAccessible{Reason: reason}
}
