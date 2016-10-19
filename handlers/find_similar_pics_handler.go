package handlers

import (
	"context"
	"net/http"
	"time"

	"pixur.org/pixur/schema"
	"pixur.org/pixur/schema/db"
	"pixur.org/pixur/status"
	"pixur.org/pixur/tasks"
)

type FindSimilarPicsHandler struct {
	// embeds
	http.Handler

	// deps
	DB     db.DB
	Runner *tasks.TaskRunner
	Now    func() time.Time
}

func (h *FindSimilarPicsHandler) FindSimilarPics(
	ctx context.Context, req *FindSimilarPicsRequest) (*FindSimilarPicsResponse, status.S) {

	ctx, sts := fillUserIDFromCtx(ctx)
	if sts != nil {
		return nil, sts
	}

	var requestedPicID schema.Varint
	if req.PicId != "" {
		if err := requestedPicID.DecodeAll(req.PicId); err != nil {
			return nil, status.InvalidArgument(err, "bad pic id")
		}
	}

	var task = &tasks.FindSimilarPicsTask{
		DB:    h.DB,
		PicID: int64(requestedPicID),
		Ctx:   ctx,
	}
	if sts := h.Runner.Run(task); sts != nil {
		return nil, sts
	}

	resp := FindSimilarPicsResponse{}
	for _, id := range task.SimilarPicIDs {
		resp.PicId = append(resp.PicId, schema.Varint(id).Encode())
	}

	return &resp, nil
}

// TODO: test this
func (h *FindSimilarPicsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rc := &requestChecker{r: r, now: h.Now}
	rc.checkPost()
	rc.checkXsrf()
	if rc.code != 0 {
		http.Error(w, rc.message, rc.code)
		return
	}

	ctx := r.Context()
	if token, present := authTokenFromReq(r); present {
		ctx = tasks.CtxFromAuthToken(ctx, token)
	}

	resp, sts := h.FindSimilarPics(ctx, &FindSimilarPicsRequest{
		PicId: r.FormValue("pic_id"),
	})
	if sts != nil {
		http.Error(w, sts.Message(), sts.Code().HttpStatus())
		return
	}

	returnProtoJSON(w, r, resp)
}

func init() {
	register(func(mux *http.ServeMux, c *ServerConfig) {
		mux.Handle("/api/findSimilarPics", &FindSimilarPicsHandler{
			DB:  c.DB,
			Now: time.Now,
		})
	})
}
