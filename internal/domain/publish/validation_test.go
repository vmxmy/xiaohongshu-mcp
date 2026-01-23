package publish

import "testing"

func TestValidateImageContent_Limits(t *testing.T) {
	limits := Limits{MaxTags: 2, MinImages: 1, MaxImages: 3}
	err := ValidateImageContent(ImageContent{
		Title:      "t",
		Content:    "c",
		Tags:       []string{"a", "b", "c"},
		ImagePaths: []string{"1.jpg"},
	}, limits)
	if err == nil {
		t.Fatalf("expected tag limit error")
	}
}
