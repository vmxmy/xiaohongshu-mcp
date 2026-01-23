package testkit

import (
	"testing"

	"github.com/xpzouying/xiaohongshu-mcp/internal/app/ports"
)

func TestFakesImplementPorts(t *testing.T) {
	var _ ports.PublishGateway = (*FakePublishGateway)(nil)
	var _ ports.SelectorStore = (*FakeSelectorStore)(nil)
}
