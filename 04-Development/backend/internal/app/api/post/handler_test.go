package post

import "testing"

func TestSlugify(t *testing.T) {
	tests := map[string]string{
		"Cara Merawat Hoodie Heavyweight": "cara-merawat-hoodie-heavyweight",
		"  Drop 05: Void—Runner!! ":       "drop-05-void-runner",
		"Café Édition":                    "cafe-edition",
		"###":                             "",
	}
	for in, want := range tests {
		if got := Slugify(in); got != want {
			t.Errorf("Slugify(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestValidate(t *testing.T) {
	ok := Input{Title: "Panduan ukuran", Slug: "panduan-ukuran", Status: StatusDraft}
	if errs := validate(ok); len(errs) != 0 {
		t.Fatalf("valid draft rejected: %v", errs)
	}
	short := ok
	short.Status = StatusPublished
	short.Content = "pendek"
	if errs := validate(short); len(errs) != 1 || errs[0].Field != "content" {
		t.Errorf("publish with short content: %v", errs)
	}
	bad := ok
	bad.Slug = "Bad Slug"
	bad.ProductIDs = []string{"x"}
	if errs := validate(bad); len(errs) != 2 {
		t.Errorf("want slug & product_ids errors, got %v", errs)
	}
}
