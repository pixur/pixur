package tasks

import (
	"context"
	"time"

	"pixur.org/pixur/be/schema"
	"pixur.org/pixur/be/schema/db"
	tab "pixur.org/pixur/be/schema/tables"
	"pixur.org/pixur/be/status"
)

var _ Task = &PurgePicTask{}

type PurgePicTask struct {
	// deps
	PixPath string
	Beg     tab.JobBeginner
	Remove  func(name string) error
	Now     func() time.Time

	// input
	PicId int64
}

func (t *PurgePicTask) Run(ctx context.Context) (stscap status.S) {
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
	if sts := validateCapability(u, conf, schema.User_PIC_PURGE); sts != nil {
		return sts
	}

	pics, err := j.FindPics(db.Opts{
		Prefix: tab.PicsPrimary{&t.PicId},
		Limit:  1,
		Lock:   db.LockWrite,
	})
	if err != nil {
		return status.Internal(err, "can't find pics")
	}
	if len(pics) != 1 {
		return status.NotFound(nil, "can't lookup pic")
	}
	p := pics[0]

	if err := j.DeletePic(tab.KeyForPic(p)); err != nil {
		return status.Internal(err, "can't delete pic")
	}

	pis, err := j.FindPicIdents(db.Opts{
		Prefix: tab.PicIdentsPrimary{PicId: &t.PicId},
		Lock:   db.LockWrite,
	})
	if err != nil {
		return status.Internal(err, "can't find pic idents")
	}

	for _, pi := range pis {
		if err := j.DeletePicIdent(tab.KeyForPicIdent(pi)); err != nil {
			return status.Internal(err, "can't delete pic ident")
		}
	}

	pts, err := j.FindPicTags(db.Opts{
		Prefix: tab.PicTagsPrimary{PicId: &t.PicId},
		Lock:   db.LockWrite,
	})
	if err != nil {
		return status.Internal(err, "can't find pic tags")
	}

	for _, pt := range pts {
		if err := j.DeletePicTag(tab.KeyForPicTag(pt)); err != nil {
			return status.Internal(err, "can't delete pic tag")
		}
	}

	var ts []*schema.Tag
	for _, pt := range pts {
		tags, err := j.FindTags(db.Opts{
			Prefix: tab.TagsPrimary{&pt.TagId},
			Lock:   db.LockWrite,
			Limit:  1,
		})
		if err != nil {
			return status.Internal(err, "can't find tag")
		}
		if len(tags) != 1 {
			return status.Internal(nil, "can't lookup tag")
		}
		ts = append(ts, tags[0])
	}

	for _, t := range ts {
		if t.UsageCount > 1 {
			t.UsageCount--
			t.SetModifiedTime(now)
			if err := j.UpdateTag(t); err != nil {
				return status.Internal(err, "can't update tag")
			}
		} else {
			if err := j.DeleteTag(tab.KeyForTag(t)); err != nil {
				return status.Internal(err, "can't delete tag")
			}
		}
	}

	pcs, err := j.FindPicComments(db.Opts{
		Prefix: tab.PicCommentsPrimary{PicId: &t.PicId},
		Lock:   db.LockWrite,
	})
	if err != nil {
		return status.Internal(err, "can't find pic comments")
	}

	for _, pc := range pcs {
		if err := j.DeletePicComment(tab.KeyForPicComment(pc)); err != nil {
			return status.Internal(err, "can't delete pic comment")
		}
	}

	pvs, err := j.FindPicVotes(db.Opts{
		Prefix: tab.PicVotesPrimary{PicId: &t.PicId},
		Lock:   db.LockWrite,
	})
	if err != nil {
		return status.Internal(err, "can't find pic votes")
	}

	for _, pv := range pvs {
		if err := j.DeletePicVote(tab.KeyForPicVote(pv)); err != nil {
			return status.Internal(err, "can't delete pic vote")
		}
	}

	if err := j.Commit(); err != nil {
		return status.Internal(err, "Unable to Commit")
	}

	oldpath, sts := schema.PicFilePath(t.PixPath, p.PicId, p.File.Mime)
	if sts != nil {
		defer status.ReplaceOrSuppress(&stscap, sts)
	} else if err := t.Remove(oldpath); err != nil {
		defer status.ReplaceOrSuppress(&stscap, status.DataLoss(err, "unable to delete pic data", oldpath))
	}

	for _, th := range p.Thumbnail {
		oldthumbpath, sts := schema.PicFileDerivedPath(t.PixPath, p.PicId, th.Index, th.Mime)
		if sts != nil {
			defer status.ReplaceOrSuppress(&stscap, sts)
		} else if err := t.Remove(oldthumbpath); err != nil {
			defer status.ReplaceOrSuppress(&stscap, status.DataLoss(err, "unable to delete pic data", oldthumbpath))
		}
	}

	// TODO: purge any user events that contain the PicId

	return nil
}
