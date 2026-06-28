package photoreview_test

import (
	"strings"
	"testing"

	photoreview "github.com/anuarkuanysh/dental_project/internal/domain/photo_review"
	photoreviewservice "github.com/anuarkuanysh/dental_project/internal/service/photo_review"
)

func TestFormatDoctorReply_MediaTypePrefix(t *testing.T) {
	t.Parallel()

	photo := photoreviewservice.FormatDoctorReply(photoreview.MediaTypePhoto, "Тест")
	if !strings.HasPrefix(photo, "Ответ врача клиники по вашему фото:") {
		t.Fatalf("unexpected photo prefix: %q", photo)
	}

	video := photoreviewservice.FormatDoctorReply(photoreview.MediaTypeVideo, "Тест")
	if !strings.HasPrefix(video, "Ответ врача клиники по вашему видео:") {
		t.Fatalf("unexpected video prefix: %q", video)
	}
}

func TestPatientAckMessage(t *testing.T) {
	t.Parallel()

	if got := photoreviewservice.PatientAckMessage(photoreview.MediaTypePhoto); !strings.HasPrefix(got, "Фото получено.") {
		t.Fatalf("unexpected photo ack: %q", got)
	}
	if got := photoreviewservice.PatientAckMessage(photoreview.MediaTypeVideo); !strings.HasPrefix(got, "Видео получено.") {
		t.Fatalf("unexpected video ack: %q", got)
	}
}
