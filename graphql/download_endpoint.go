package graphql

import (
	"net/http"

	attendeepb "github.com/svladislav00-qq/event-microservices/attendee/pb"
)

func DownloadAttendeesHandler(server *Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		eventID := r.URL.Query().Get("event_id")
		if eventID == "" {
			http.Error(w, "event_id required", http.StatusBadRequest)
			return
		}

		resp, err := server.attendeeClient.ExportAttendeesTable(r.Context(), &attendeepb.ExportAttendeesTableRequest{
			EventId: eventID,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Disposition", "attachment; filename="+resp.Filename)
		w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")

		w.Write(resp.File)
	}
}
