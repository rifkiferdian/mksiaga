package blmkc

import "testing"

func TestEvidenceFilenameUsesCategoryAndKeepsUniqueSuffix(t *testing.T) {
	tests := map[string]string{
		"incident_photo":    "foto-kejadian-12345678.jpg",
		"subject_photo":     "foto-pelaku-12345678.jpg",
		"stolen_item_photo": "foto-barang-dicuri-12345678.jpg",
		"vehicle_photo":     "foto-kendaraan-12345678.jpg",
		"cctv_video":        "video-cctv-12345678.mp4",
	}
	for category, want := range tests {
		extension := ".JPG"
		if category == "cctv_video" {
			extension = ".MP4"
		}
		if got := evidenceFilename(category, "1234567890abcdef", extension); got != want {
			t.Errorf("evidenceFilename(%q) = %q, want %q", category, got, want)
		}
	}
}
