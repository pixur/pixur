package handlers

import (
	"context"

	"pixur.org/pixur/api"
	"pixur.org/pixur/be/schema"
	"pixur.org/pixur/be/status"
	"pixur.org/pixur/be/tasks"
)

// TODO: test
func (s *serv) handleFindUserEvents(ctx context.Context, req *api.FindUserEventsRequest) (
	*api.FindUserEventsResponse, status.S) {
	var userId schema.Varint
	if req.UserId != "" {
		if err := userId.DecodeAll(req.UserId); err != nil {
			return nil, status.InvalidArgument(err, "bad user id")
		}
	}
	var keyUserId, keyCreatedTs, keyIndex schema.Varint
	if req.StartUserEventId != "" {
		var i int
		if n, err := keyUserId.Decode(req.StartUserEventId[i:]); err != nil {
			return nil, status.InvalidArgument(err, "bad user event id")
		} else {
			i += int(n)
		}
		if n, err := keyCreatedTs.Decode(req.StartUserEventId[i:]); err != nil {
			return nil, status.InvalidArgument(err, "bad user event id")
		} else {
			i += int(n)
		}
		if req.StartUserEventId[i:] != "" {
			if n, err := keyIndex.Decode(req.StartUserEventId[i:]); err != nil {
				return nil, status.InvalidArgument(err, "bad user event id")
			} else {
				i += int(n)
			}
		}
		if req.StartUserEventId[i:] != "" {
			// too much input
			return nil, status.InvalidArgument(nil, "bad user event id")
		}
	}

	var task = &tasks.FindUserEventsTask{
		Beg:            s.db,
		Now:            s.now,
		Ascending:      req.Ascending,
		ObjectUserId:   int64(userId),
		StartUserId:    int64(keyUserId),
		StartCreatedTs: int64(keyCreatedTs),
		StartIndex:     int64(keyIndex),
	}

	if sts := s.runner.Run(ctx, task); sts != nil {
		return nil, sts
	}

	resp := &api.FindUserEventsResponse{
		UserEvent: apiUserEvents(nil, task.UserEvents, nil),
	}
	if task.NextUserId != 0 {
		resp.NextUserEventId = apiUserEventId(task.NextUserId, task.NextCreatedTs, task.NextIndex)
	}
	if task.PrevUserId != 0 {
		resp.PrevUserEventId = apiUserEventId(task.PrevUserId, task.PrevCreatedTs, task.PrevIndex)
	}
	return resp, nil
}
