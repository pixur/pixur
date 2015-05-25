package pixur

import (
	"encoding/json"
	"net/http"
	"strconv"
)

// TODO: add tests
// TODO: Add csrf protection

func (s *Server) deletePicHandler(w http.ResponseWriter, r *http.Request) error {
	requestedRawPicID := r.FormValue("pic_id")
	var requestedPicId int64
	if requestedRawPicID != "" {
		if picId, err := strconv.Atoi(requestedRawPicID); err != nil {
			return err
		} else {
			requestedPicId = int64(picId)
		}
	}

	var task = &DeletePicTask{
		db:      s.db,
		pixPath: s.pixPath,
		PicId:   requestedPicId,
	}
	runner := new(TaskRunner)
	if err := runner.Run(task); err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(true); err != nil {
		return err
	}
	return nil
}
