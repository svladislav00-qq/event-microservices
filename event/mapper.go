package event

import (
	"github.com/svladislav00-qq/event-microservices/event/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func eventToPB(e *Event) *pb.Event {
	fileUrls := make([]string, 0, len(e.Files))

	for _, f := range e.Files {
		fileUrls = append(fileUrls, f.FileKey)
	}

	return &pb.Event{
		Id:          e.ID,
		Name:        e.EventName,
		Description: e.Description,
		Department:  e.Department,
		CreatedBy:   e.CreatedBy,
		StartTime:   timestamppb.New(e.StartTime),
		EndTime:     timestamppb.New(e.EndTime),
		CreatedAt:   timestamppb.New(e.CreatedAt),
		FileUrls:    fileUrls,
		Capacity:    uint32(*e.Capacity),
	}
}

func eventsToPB(events []Event) []*pb.Event {
	result := make([]*pb.Event, 0, len(events))

	for i := range events {
		result = append(result, eventToPB(&events[i]))
	}

	return result
}

func pbEventToModel(e *pb.Event) *Event {
	if e == nil {
		return nil
	}

	return &Event{
		ID:          e.Id,
		EventName:   e.Name,
		Description: e.Description,
		Department:  e.Department,
		CreatedBy:   e.CreatedBy,
		StartTime:   e.StartTime.AsTime(),
		EndTime:     e.EndTime.AsTime(),
		CreatedAt:   e.CreatedAt.AsTime(),
		Capacity:    uint32PtrToIntPtr(&e.Capacity),
	}
}

func uint32PtrToIntPtr(v *uint32) *int {
	if v == nil {
		return nil
	}

	i := int(*v)
	return &i
}

func eventsFromPB(events []*pb.Event) []*Event {
	result := make([]*Event, 0, len(events))

	for _, e := range events {
		result = append(result, pbEventToModel(e))
	}

	return result
}
