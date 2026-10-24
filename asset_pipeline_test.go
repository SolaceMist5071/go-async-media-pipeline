package main

import "testing"

func TestAssetKeyAndState(t *testing.T) {
	cases := []struct{ name, id, file, want string }{{"nested", "a1", "photo.jpg", "a1/photo.jpg"}, {"slash", "/a2", "clip.png", "a2/clip.png"}}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Asset{ID: tc.id, ObjectKey: assetKey(tc.id, tc.file), State: "upload-url-issued"}
			if got.ObjectKey != tc.want {
				t.Fatalf("key=%q", got.ObjectKey)
			}
			if got.State != "upload-url-issued" {
				t.Fatal("state transition missing")
			}
		})
	}
}
