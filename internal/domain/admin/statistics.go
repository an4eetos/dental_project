package admin

// Statistics aggregates clinic usage metrics for the admin panel.
type Statistics struct {
	TotalUsers               int64
	TotalPatients            int64
	TotalDoctors             int64
	TotalAdmins              int64
	TotalPhotoSubmissions    int64
	PendingPhotoSubmissions  int64
	AnsweredPhotoSubmissions int64
	TotalAppointments        int64
	PendingAppointments      int64
	ConfirmedAppointments    int64
	CancelledAppointments    int64
}
