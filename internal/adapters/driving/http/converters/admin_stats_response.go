package converters

import admindomain "github.com/anuarkuanysh/dental_project/internal/domain/admin"

type statisticsResponse struct {
	TotalUsers               int64 `json:"total_users"`
	TotalPatients            int64 `json:"total_patients"`
	TotalDoctors             int64 `json:"total_doctors"`
	TotalAdmins              int64 `json:"total_admins"`
	TotalPhotoSubmissions    int64 `json:"total_photo_submissions"`
	PendingPhotoSubmissions  int64 `json:"pending_photo_submissions"`
	AnsweredPhotoSubmissions int64 `json:"answered_photo_submissions"`
	TotalAppointments        int64 `json:"total_appointments"`
	PendingAppointments      int64 `json:"pending_appointments"`
	ConfirmedAppointments    int64 `json:"confirmed_appointments"`
	CancelledAppointments    int64 `json:"cancelled_appointments"`
}

func ToStatisticsResponse(stats admindomain.Statistics) statisticsResponse {
	return statisticsResponse{
		TotalUsers:               stats.TotalUsers,
		TotalPatients:            stats.TotalPatients,
		TotalDoctors:             stats.TotalDoctors,
		TotalAdmins:              stats.TotalAdmins,
		TotalPhotoSubmissions:    stats.TotalPhotoSubmissions,
		PendingPhotoSubmissions:  stats.PendingPhotoSubmissions,
		AnsweredPhotoSubmissions: stats.AnsweredPhotoSubmissions,
		TotalAppointments:        stats.TotalAppointments,
		PendingAppointments:      stats.PendingAppointments,
		ConfirmedAppointments:    stats.ConfirmedAppointments,
		CancelledAppointments:    stats.CancelledAppointments,
	}
}
