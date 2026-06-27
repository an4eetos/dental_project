package converters

import (
	"time"

	"github.com/anuarkuanysh/dental_project/internal/domain/booking"
	bookingvalidate "github.com/anuarkuanysh/dental_project/internal/service/booking"
)

type AppointmentResponse struct {
	ID            int64     `json:"id"`
	PreferredDate string    `json:"preferred_date"`
	PreferredTime string    `json:"preferred_time"`
	Status        string    `json:"status"`
	VisitType     string    `json:"visit_type,omitempty"`
	ZoomLink      string    `json:"zoom_link,omitempty"`
	DoctorNotes   string    `json:"doctor_notes,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

func ToAppointmentResponse(a booking.Appointment) AppointmentResponse {
	resp := AppointmentResponse{
		ID:            a.ID,
		PreferredDate: a.PreferredDate.UTC().Format(bookingvalidate.DateLayout),
		PreferredTime: a.PreferredTime.Format(bookingvalidate.TimeLayout),
		Status:        a.Status.String(),
		CreatedAt:     a.CreatedAt.UTC(),
	}
	if a.VisitType.Valid() {
		resp.VisitType = a.VisitType.String()
	}
	if a.Status == booking.StatusConfirmed && a.VisitType == booking.VisitTypeVideo && a.ZoomLink != "" {
		resp.ZoomLink = a.ZoomLink
	}
	if a.Status == booking.StatusRejected && a.DoctorNotes != "" {
		resp.DoctorNotes = a.DoctorNotes
	}
	return resp
}

func ToAppointmentListResponse(items []booking.Appointment) []AppointmentResponse {
	out := make([]AppointmentResponse, 0, len(items))
	for _, a := range items {
		out = append(out, ToAppointmentResponse(a))
	}
	return out
}
