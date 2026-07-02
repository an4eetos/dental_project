package converters

import (
	"time"

	"github.com/anuarkuanysh/dental_project/internal/domain/booking"
	bookingvalidate "github.com/anuarkuanysh/dental_project/internal/service/booking"
)

type PatientBriefResponse struct {
	ID         int64  `json:"id"`
	TelegramID int64  `json:"telegram_id"`
	Username   string `json:"username,omitempty"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name,omitempty"`
}

type DoctorAppointmentResponse struct {
	ID                 int64                `json:"id"`
	PreferredDate      string               `json:"preferred_date"`
	PreferredTime      string               `json:"preferred_time"`
	PreferredVisitType string               `json:"preferred_visit_type,omitempty"`
	Status             string               `json:"status"`
	VisitType          string               `json:"visit_type,omitempty"`
	ZoomLink           string               `json:"zoom_link,omitempty"`
	DoctorNotes        string               `json:"doctor_notes,omitempty"`
	NeedsZoomLink      bool                 `json:"needs_zoom_link,omitempty"`
	CreatedAt          time.Time            `json:"created_at"`
	Patient            PatientBriefResponse `json:"patient"`
}

func ToDoctorAppointmentResponse(item booking.AppointmentWithPatient) DoctorAppointmentResponse {
	resp := DoctorAppointmentResponse{
		ID:            item.Appointment.ID,
		PreferredDate: item.Appointment.PreferredDate.UTC().Format(bookingvalidate.DateLayout),
		PreferredTime: item.Appointment.PreferredTime.Format(bookingvalidate.TimeLayout),
		Status:        item.Appointment.Status.String(),
		CreatedAt:     item.Appointment.CreatedAt.UTC(),
		Patient: PatientBriefResponse{
			ID:         item.Patient.ID,
			TelegramID: item.Patient.TelegramID,
			Username:   item.Patient.Username,
			FirstName:  item.Patient.FirstName,
			LastName:   item.Patient.LastName,
		},
	}
	if item.Appointment.PreferredVisitType.Valid() {
		resp.PreferredVisitType = item.Appointment.PreferredVisitType.String()
	}
	if item.Appointment.VisitType.Valid() {
		resp.VisitType = item.Appointment.VisitType.String()
	}
	if item.Appointment.ZoomLink != "" {
		resp.ZoomLink = item.Appointment.ZoomLink
	}
	if item.Appointment.DoctorNotes != "" {
		resp.DoctorNotes = item.Appointment.DoctorNotes
	}
	if item.Appointment.Status == booking.StatusConfirmed &&
		item.Appointment.VisitType == booking.VisitTypeVideo &&
		item.Appointment.ZoomLink == "" {
		resp.NeedsZoomLink = true
	}
	return resp
}

func ToDoctorAppointmentListResponse(items []booking.AppointmentWithPatient) []DoctorAppointmentResponse {
	out := make([]DoctorAppointmentResponse, 0, len(items))
	for _, item := range items {
		out = append(out, ToDoctorAppointmentResponse(item))
	}
	return out
}
