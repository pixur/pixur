package handlers

import (
	"compress/gzip"
	"crypto/rsa"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	oldctx "golang.org/x/net/context"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"

	"google.golang.org/grpc"
	gcodes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	gstatus "google.golang.org/grpc/status"

	"pixur.org/pixur/api"
	"pixur.org/pixur/schema/db"
	"pixur.org/pixur/status"
	"pixur.org/pixur/tasks"
)

func (s *serv) intercept(ctx oldctx.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	if md, present := metadata.FromIncomingContext(ctx); present {
		if token, present := authTokenFromMD(md); present {
			ctx = tasks.CtxFromAuthToken(ctx, token)
		}
	}

	resp, err := handler(ctx, req)
	if err != nil {
		sts := err.(status.S)
		err = gstatus.Error(gcodes.Code(sts.Code()), sts.Message())
	}
	return resp, err
}

var _ api.PixurServiceServer = &serv{}

type serv struct {
	db          db.DB
	pixpath     string
	tokenSecret []byte
	privkey     *rsa.PrivateKey
	pubkey      *rsa.PublicKey
	secure      bool
	runner      *tasks.TaskRunner
	now         func() time.Time
	rand        io.Reader
}

func (s *serv) AddPicComment(ctx oldctx.Context, req *api.AddPicCommentRequest) (*api.AddPicCommentResponse, error) {
	return s.handleAddPicComment(ctx, req)
}

func (s *serv) AddPicTags(ctx oldctx.Context, req *api.AddPicTagsRequest) (*api.AddPicTagsResponse, error) {
	return s.handleAddPicTags(ctx, req)
}

func (s *serv) CreatePic(ctx oldctx.Context, req *api.CreatePicRequest) (*api.CreatePicResponse, error) {
	return s.handleCreatePic(ctx, req)
}

func (s *serv) CreateUser(ctx oldctx.Context, req *api.CreateUserRequest) (*api.CreateUserResponse, error) {
	return s.handleCreateUser(ctx, req)
}

func (s *serv) DeleteToken(ctx oldctx.Context, req *api.DeleteTokenRequest) (*api.DeleteTokenResponse, error) {
	return s.handleDeleteToken(ctx, req)
}

func (s *serv) FindIndexPics(ctx oldctx.Context, req *api.FindIndexPicsRequest) (*api.FindIndexPicsResponse, error) {
	return s.handleFindIndexPics(ctx, req)
}

func (s *serv) FindSimilarPics(ctx oldctx.Context, req *api.FindSimilarPicsRequest) (*api.FindSimilarPicsResponse, error) {
	return s.handleFindSimilarPics(ctx, req)
}

func (s *serv) GetRefreshToken(ctx oldctx.Context, req *api.GetRefreshTokenRequest) (*api.GetRefreshTokenResponse, error) {
	return s.handleGetRefreshToken(ctx, req)
}

func (s *serv) GetXsrfToken(ctx oldctx.Context, req *api.GetXsrfTokenRequest) (*api.GetXsrfTokenResponse, error) {
	return s.handleGetXsrfToken(ctx, req)
}

func (s *serv) IncrementViewCount(ctx oldctx.Context, req *api.IncrementViewCountRequest) (*api.IncrementViewCountResponse, error) {
	return s.handleIncrementViewCount(ctx, req)
}

func (s *serv) LookupPicDetails(ctx oldctx.Context, req *api.LookupPicDetailsRequest) (*api.LookupPicDetailsResponse, error) {
	return s.handleLookupPicDetails(ctx, req)
}

func (s *serv) LookupUser(ctx oldctx.Context, req *api.LookupUserRequest) (*api.LookupUserResponse, error) {
	return nil, status.Unimplemented(nil, "Not implemented")
}

func (s *serv) PurgePic(ctx oldctx.Context, req *api.PurgePicRequest) (*api.PurgePicResponse, error) {
	return s.handlePurgePic(ctx, req)
}

func (s *serv) SoftDeletePic(ctx oldctx.Context, req *api.SoftDeletePicRequest) (*api.SoftDeletePicResponse, error) {
	return s.handleSoftDeletePic(ctx, req)
}

func (s *serv) UpdateUser(ctx oldctx.Context, req *api.UpdateUserRequest) (*api.UpdateUserResponse, error) {
	return s.handleUpdateUser(ctx, req)
}

func (s *serv) UpsertPic(ctx oldctx.Context, req *api.UpsertPicRequest) (*api.UpsertPicResponse, error) {
	return s.handleUpsertPic(ctx, req)
}

func (s *serv) UpsertPicVote(ctx oldctx.Context, req *api.UpsertPicVoteRequest) (*api.UpsertPicVoteResponse, error) {
	return s.handleUpsertPicVote(ctx, req)
}

type registerFunc func(mux *http.ServeMux, c *ServerConfig)

var (
	handlerFuncs []registerFunc
)

type ServerConfig struct {
	DB          db.DB
	PixPath     string
	TokenSecret []byte
	PrivateKey  *rsa.PrivateKey
	PublicKey   *rsa.PublicKey
	Secure      bool
}

func register(rf registerFunc) {
	handlerFuncs = append(handlerFuncs, rf)
}

func AddAllHandlers(mux *http.ServeMux, c *ServerConfig) {
	for _, rf := range handlerFuncs {
		rf(mux, c)
	}
}

var errorLog = log.New(os.Stderr, "", log.LstdFlags)

func httpError(w http.ResponseWriter, sts status.S) {
	w.Header().Set("Pixur-Status", strconv.Itoa(int(sts.Code())))
	w.Header().Set("Pixur-Message", sts.Message())

	code := sts.Code()
	http.Error(w, code.String()+": "+sts.Message(), code.HttpStatus())

	errorLog.Println(sts.String())
}

var protoJSONMarshaller = &jsonpb.Marshaler{}

func returnProtoJSON(w http.ResponseWriter, r *http.Request, pb proto.Message) {
	var writer io.Writer = w

	if encs := r.Header.Get("Accept-Encoding"); encs != "" {
		for _, enc := range strings.Split(encs, ",") {
			if strings.TrimSpace(enc) == "gzip" {
				if gw, err := gzip.NewWriterLevel(writer, gzip.BestSpeed); err != nil {
					panic(err)
				} else {
					defer gw.Close()
					// TODO: log this

					writer = gw
				}
				w.Header().Set("Content-Encoding", "gzip")
				break
			}
		}
	}
	if accept := r.Header.Get("Accept"); accept != "" {
		for _, acc := range strings.Split(accept, ",") {
			switch strings.TrimSpace(acc) {
			case "application/json":
				w.Header().Set("Content-Type", "application/json")
				if err := protoJSONMarshaller.Marshal(writer, pb); err != nil {
					httpError(w, status.InternalError(err, "error writing json"))
					return
				}
				return
			case "application/proto":
				w.Header().Set("Content-Type", "application/proto")
				raw, err := proto.Marshal(pb)
				if err != nil {
					httpError(w, status.InternalError(err, "error building proto"))
					return
				}
				if _, err := writer.Write(raw); err != nil {
					httpError(w, status.InternalError(err, "error writing proto"))
					return
				}
				return
			}
		}
	}
	// default
	w.Header().Set("Content-Type", "application/json")
	if err := protoJSONMarshaller.Marshal(writer, pb); err != nil {
		httpError(w, status.InternalError(err, "error writing json"))
	}
}
