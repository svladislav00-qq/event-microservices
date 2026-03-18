package attendee

import "github.com/svladislav00-qq/event-microservices/attendee/pb"

func pbAttendeeToModel(e *pb.Attendee) *Attendee {
	if e == nil {
		return nil
	}

	return &Attendee{
		ID:           e.Id,
		EventID:      e.EventId,
		UserID:       e.UserId,
		Status:       Status(e.Status.String()),
		RegisteredAt: e.RegisteredAt.AsTime(),
	}
}

func mapStatusFromPB(s pb.Status) Status {
	switch s {
	case pb.Status_STATUS_REGISTERED:
		return StatusRegistered
	case pb.Status_STATUS_CANCELED:
		return StatusCanceled
	default:
		return Status("")
	}
}

func mapStatusToPB(status Status) pb.Status {
	switch status {
	case StatusRegistered:
		return pb.Status_STATUS_REGISTERED
	case StatusCanceled:
		return pb.Status_STATUS_CANCELED
	default:
		return pb.Status_STATUS_UNKNOWN
	}
}
