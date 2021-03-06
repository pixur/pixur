package handlers

import (
	"context"
	"testing"

	oldctx "golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	gstatus "google.golang.org/grpc/status"

	"pixur.org/pixur/api"
	"pixur.org/pixur/be/status"
	"pixur.org/pixur/be/tasks"
)

func TestServerInterceptorSucceedsOnNoAuth(t *testing.T) {
	si := &serverInterceptor{}
	req := 1
	var ctxcap context.Context
	handler := grpc.UnaryHandler(func(ctx oldctx.Context, req interface{}) (interface{}, error) {
		ctxcap = ctx
		return nil, nil
	})
	_, err := si.intercept(context.Background(), req, nil, handler)

	if err != nil {
		t.Fatal(err)
	}
	if ctxcap == nil {
		t.Fatal("ctx cap is nil")
	}
}

func TestServerInterceptorFailsOnBadAuth(t *testing.T) {
	si := &serverInterceptor{}
	req := 1
	handler := grpc.UnaryHandler(func(ctx oldctx.Context, req interface{}) (interface{}, error) {
		return nil, nil
	})

	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(authPwtHeaderKey, "bogus"))

	_, err := si.intercept(ctx, req, nil, handler)

	if err == nil {
		t.Fatal("expected err")
	}
	gsts, ok := gstatus.FromError(err)
	if !ok {
		t.Fatal("not a gstatus", gsts)
	}
	if have, want := gsts.Code(), codes.Unauthenticated; have != want {
		t.Error("have", have, "want", want)
	}
}

func TestServerInterceptorIgnoresOnAuthHandlers(t *testing.T) {
	si := &serverInterceptor{}
	var ctxcap context.Context
	handler := grpc.UnaryHandler(func(ctx oldctx.Context, req interface{}) (interface{}, error) {
		ctxcap = ctx
		return nil, nil
	})

	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(authPwtHeaderKey, "bogus"))

	for _, req := range []interface{}{&api.GetRefreshTokenRequest{}} {
		_, err := si.intercept(ctx, req, nil, handler)

		if err != nil {
			t.Fatal(err)
		}
		if ctxcap == nil {
			t.Fatal("ctx cap is nil")
		}
		token, present := tasks.AuthTokenFromCtx(ctxcap)
		if !present {
			t.Fatal("missing token")
		}
		if token != "bogus" {
			t.Error("have", token, "want", "bogus")
		}

		ctxcap = nil
	}
}

func TestServerInterceptor(t *testing.T) {
	si := &serverInterceptor{}
	req := 1
	var ctxcap context.Context
	handler := grpc.UnaryHandler(func(ctx oldctx.Context, req interface{}) (interface{}, error) {
		ctxcap = ctx
		return nil, status.Unimplemented(nil, "no go")
	})

	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(authPwtHeaderKey, testAuthToken))

	_, err := si.intercept(ctx, req, nil, handler)

	if err == nil {
		t.Fatal(err)
	}
	gsts, ok := gstatus.FromError(err)
	if !ok {
		t.Fatal("not a gstatus", gsts)
	}
	if have, want := gsts.Code(), codes.Unimplemented; have != want {
		t.Error("have", have, "want", want)
	}
	if ctxcap == nil {
		t.Fatal("nil ctx")
	}
	tok, present := tasks.UserTokenFromCtx(ctxcap)
	if !present {
		t.Fatal("missing user token")
	}
	if have, want := tok.UserId, testAuthSubject; have != want {
		t.Error("have", have, "want", want)
	}
	if have, want := tok.TokenId, testUserToken; have != want {
		t.Error("have", have, "want", want)
	}
}
