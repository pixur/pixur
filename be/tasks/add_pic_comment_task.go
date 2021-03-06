package tasks

import (
	"context"
	"math"
	"time"

	any "github.com/golang/protobuf/ptypes/any"

	"pixur.org/pixur/be/schema"
	"pixur.org/pixur/be/schema/db"
	tab "pixur.org/pixur/be/schema/tables"
	"pixur.org/pixur/be/status"
	"pixur.org/pixur/be/text"
)

type AddPicCommentTask struct {
	// Deps
	Beg tab.JobBeginner
	Now func() time.Time

	// Inputs
	PicId           int64
	CommentParentId int64
	Text            string

	// Ext is additional extra data associated with this comment.
	Ext map[string]*any.Any

	// Outs
	UnfilteredPicComment *schema.PicComment
	PicComment           *schema.PicComment
}

func (t *AddPicCommentTask) Run(ctx context.Context) (stscap status.S) {
	now := t.Now()
	j, u, sts := authedJob(ctx, t.Beg, now)
	if sts != nil {
		return sts
	}
	defer revert(j, &stscap)

	conf, sts := GetConfiguration(ctx)
	if sts != nil {
		return sts
	}
	var minCommentLen, maxCommentLen int64
	if conf.MinCommentLength != nil {
		minCommentLen = conf.MinCommentLength.Value
	} else {
		minCommentLen = math.MinInt64
	}
	if conf.MaxCommentLength != nil {
		maxCommentLen = conf.MaxCommentLength.Value
	} else {
		maxCommentLen = math.MaxInt64
	}
	txt, err := text.DefaultValidateAndNormalize(t.Text, "comment", minCommentLen, maxCommentLen)
	if err != nil {
		return status.From(err)
	}

	if sts := validateCapability(u, conf, schema.User_PIC_COMMENT_CREATE); sts != nil {
		return sts
	}
	userId := schema.AnonymousUserId
	if u != nil {
		userId = u.UserId
	}

	if len(t.Ext) != 0 {
		if sts := validateCapability(u, conf, schema.User_PIC_COMMENT_EXTENSION_CREATE); sts != nil {
			return sts
		}
	}

	pics, err := j.FindPics(db.Opts{
		Prefix: tab.PicsPrimary{&t.PicId},
	})
	if err != nil {
		return status.Internal(err, "can't lookup pic")
	}
	if len(pics) != 1 {
		return status.NotFound(nil, "can't find pic")
	}
	p := pics[0]

	if p.HardDeleted() {
		return status.InvalidArgument(nil, "can't comment on deleted pic")
	}

	var commentParent *schema.PicComment
	if t.CommentParentId != 0 {
		comments, err := j.FindPicComments(db.Opts{
			Prefix: tab.PicCommentsPrimary{PicId: &t.PicId, CommentId: &t.CommentParentId},
		})
		if err != nil {
			return status.Internal(err, "can't lookup comment")
		}
		if len(comments) != 1 {
			return status.NotFound(nil, "can't find comment")
		}
		commentParent = comments[0]

		if conf.EnablePicCommentSelfReply != nil && !conf.EnablePicCommentSelfReply.Value {
			if userId == commentParent.UserId && userId != schema.AnonymousUserId {
				return status.InvalidArgument(nil, "can't self reply")
			}
		}
	}
	if conf.EnablePicCommentSiblingReply != nil && !conf.EnablePicCommentSiblingReply.Value {
		if userId != schema.AnonymousUserId {
			comments, err := j.FindPicComments(db.Opts{
				Prefix: tab.PicCommentsPrimary{PicId: &t.PicId},
			})
			if err != nil {
				return status.Internal(err, "can't lookup comments")
			}
			for _, c := range comments {
				if c.CommentParentId == t.CommentParentId && c.UserId == userId {
					return status.InvalidArgument(nil, "can't double reply")
				}
			}
		}
	}

	commentId, err := j.AllocId()
	if err != nil {
		return status.Internal(err, "can't allocate id")
	}

	pc := &schema.PicComment{
		PicId:           p.PicId,
		CommentId:       commentId,
		CommentParentId: t.CommentParentId,
		Text:            txt,
		UserId:          userId,
		Ext:             t.Ext,
	}

	pc.SetCreatedTime(now)
	pc.SetModifiedTime(now)

	if err := j.InsertPicComment(pc); err != nil {
		return status.Internal(err, "can't insert comment")
	}

	createdTs := schema.UserEventCreatedTsCol(pc.CreatedTs)
	next := func(uid int64) (int64, status.S) {
		return nextUserEventIndex(j, uid, createdTs)
	}

	var iues []*schema.UserEvent
	var oue *schema.UserEvent
	notifications := make(map[int64]bool)
	if userId != schema.AnonymousUserId {
		idx, sts := next(userId)
		if sts != nil {
			return sts
		}
		notifications[userId] = true
		oue = &schema.UserEvent{
			UserId:     userId,
			Index:      idx,
			CreatedTs:  pc.CreatedTs,
			ModifiedTs: pc.ModifiedTs,
			Evt: &schema.UserEvent_OutgoingPicComment_{
				OutgoingPicComment: &schema.UserEvent_OutgoingPicComment{
					PicId:     p.PicId,
					CommentId: commentId,
				},
			},
		}
	}
	if commentParent != nil && commentParent.UserId != schema.AnonymousUserId &&
		!notifications[commentParent.UserId] {
		idx, sts := next(commentParent.UserId)
		if sts != nil {
			return sts
		}
		notifications[commentParent.UserId] = true
		iues = append(iues, &schema.UserEvent{
			UserId:     commentParent.UserId,
			Index:      idx,
			CreatedTs:  pc.CreatedTs,
			ModifiedTs: pc.ModifiedTs,
			Evt: &schema.UserEvent_IncomingPicComment_{
				IncomingPicComment: &schema.UserEvent_IncomingPicComment{
					PicId:     p.PicId,
					CommentId: commentId,
				},
			},
		})
	}
	// If we aren't notifying the parent comment because it doesn't exist, then create a notification
	// for each of the "uploaders" of the pic.
	if commentParent == nil {
		for _, fs := range p.Source {
			if fs.UserId != schema.AnonymousUserId && !notifications[fs.UserId] {
				idx, sts := next(fs.UserId)
				if sts != nil {
					return sts
				}
				notifications[fs.UserId] = true
				iues = append(iues, &schema.UserEvent{
					UserId:     fs.UserId,
					Index:      idx,
					CreatedTs:  pc.CreatedTs,
					ModifiedTs: pc.ModifiedTs,
					Evt: &schema.UserEvent_IncomingPicComment_{
						IncomingPicComment: &schema.UserEvent_IncomingPicComment{
							PicId:     p.PicId,
							CommentId: commentId,
						},
					},
				})
			}
		}
	}
	// In the future, these notifications could be done outside of the transaction.
	if oue != nil {
		if err := j.InsertUserEvent(oue); err != nil {
			return status.Internal(err, "can't create outgoing user event")
		}
	}
	for _, iue := range iues {
		if err := j.InsertUserEvent(iue); err != nil {
			return status.Internal(err, "can't create incoming user event")
		}
	}

	if err := j.Commit(); err != nil {
		return status.Internal(err, "can't commit job")
	}
	t.UnfilteredPicComment = pc
	t.PicComment = filterPicComment(t.UnfilteredPicComment, u, conf)

	// TODO: ratelimit
	return nil
}

func nextUserEventIndex(j *tab.Job, userId, createdTs int64) (int64, status.S) {
	ues, err := j.FindUserEvents(db.Opts{
		// We don't actually intend to write, but this prevents other transactions
		// from trying to use the same index.
		Lock: db.LockWrite,
		Prefix: tab.UserEventsPrimary{
			UserId:    &userId,
			CreatedTs: &createdTs,
		},
	})
	if err != nil {
		return 0, status.Internal(err, "can't lookup user events")
	}
	biggest := int64(-1)
	for _, ue := range ues {
		if ue.Index > biggest {
			biggest = ue.Index
		}
	}
	if biggest == math.MaxInt64 {
		return 0, status.Internal(nil, "overflow of user event index")
	}
	return biggest + 1, nil
}

func filterPicComment(
	pc *schema.PicComment, su *schema.User, conf *schema.Configuration) *schema.PicComment {
	uc := userCredOf(su, conf)
	return filterPicCommentInternal(pc, uc)
}

type userCred struct {
	subjectUserId int64
	cs            *schema.CapSet
}

func userCredOf(su *schema.User, conf *schema.Configuration) *userCred {
	if su != nil {
		return &userCred{
			subjectUserId: su.UserId,
			cs:            schema.CapSetOf(su.Capability...),
		}
	} else {
		return &userCred{
			subjectUserId: schema.AnonymousUserId,
			cs:            schema.CapSetOf(conf.AnonymousCapability.Capability...),
		}
	}
}

func filterPicComments(
	pcs []*schema.PicComment, su *schema.User, conf *schema.Configuration) []*schema.PicComment {
	uc := userCredOf(su, conf)
	dst := make([]*schema.PicComment, 0, len(pcs))
	for _, pc := range pcs {
		dst = append(dst, filterPicCommentInternal(pc, uc))
	}
	return dst
}

func filterPicCommentInternal(pc *schema.PicComment, uc *userCred) *schema.PicComment {
	dpc := *pc
	if !uc.cs.Has(schema.User_PIC_COMMENT_EXTENSION_READ) {
		dpc.Ext = nil
	}
	switch {
	case uc.cs.Has(schema.User_USER_READ_ALL):
	case uc.cs.Has(schema.User_USER_READ_PUBLIC) && uc.cs.Has(schema.User_USER_READ_PIC_COMMENT):
	case uc.subjectUserId == dpc.UserId && uc.cs.Has(schema.User_USER_READ_SELF):
	default:
		dpc.UserId = schema.AnonymousUserId
	}

	return &dpc
}
