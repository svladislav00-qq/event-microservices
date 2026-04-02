package graphql

import "net/http"

func DownloadAttendeesHandler(server *Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		eventID := r.URL.Query().Get("event_id")
		if eventID == "" {
			http.Error(w, "event_id required", http.StatusBadRequest)
			return
		}

		data, filename, err := server.attendeeClient.ExportAttendeeTable(r.Context(), eventID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Disposition", "attachment; filename="+filename)
		w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")

		w.Write(data)
	}
}
