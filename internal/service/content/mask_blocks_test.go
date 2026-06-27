package content

import (
	"encoding/json"
	"testing"

	contentdomain "github.com/anuarkuanysh/dental_project/internal/domain/content"
)

func TestMaskBlocks_RedactsPremiumFields(t *testing.T) {
	t.Parallel()

	blocks := []contentdomain.Block{
		{Type: contentdomain.BlockTypeText, Data: json.RawMessage(`{"html":"<p>hi</p>"}`)},
		{Type: contentdomain.BlockTypeYouTube, Data: json.RawMessage(`{"youtube_id":"abc12345678"}`)},
		{Type: contentdomain.BlockTypeImage, Data: json.RawMessage(`{"media_id":5}`)},
	}

	masked := MaskBlocks(blocks)
	if len(masked) != 3 {
		t.Fatalf("expected 3 blocks, got %d", len(masked))
	}

	var yt contentdomain.YouTubeBlockData
	if err := json.Unmarshal(masked[1].Data, &yt); err != nil {
		t.Fatal(err)
	}
	if yt.YouTubeID != "" {
		t.Fatalf("expected empty youtube id, got %q", yt.YouTubeID)
	}

	var img contentdomain.ImageBlockData
	if err := json.Unmarshal(masked[2].Data, &img); err != nil {
		t.Fatal(err)
	}
	if img.MediaID != 0 {
		t.Fatalf("expected media id 0, got %d", img.MediaID)
	}
}

func TestNormalizeBlocks_YouTube(t *testing.T) {
	t.Parallel()

	blocks := []contentdomain.Block{
		{
			Type: contentdomain.BlockTypeYouTube,
			Data: json.RawMessage(`{"youtube_id":"https://youtu.be/zQZ3SGSwGBI"}`),
		},
	}

	out, err := NormalizeBlocks(blocks)
	if err != nil {
		t.Fatal(err)
	}

	var data contentdomain.YouTubeBlockData
	if err := json.Unmarshal(out[0].Data, &data); err != nil {
		t.Fatal(err)
	}
	if data.YouTubeID != "zQZ3SGSwGBI" {
		t.Fatalf("got %q", data.YouTubeID)
	}
}
