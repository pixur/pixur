package pixur

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"pixur.org/pixur/schema"
)

var _ Task = &DeletePicTask{}

type DeletePicTask struct {
	// Deps
	pixPath string
	db      *sql.DB

	// input
	PicId int64
}

func (task *DeletePicTask) Run() error {
	tx, err := task.db.Begin()
	if err != nil {
		return InternalError("Unable to Begin TX", err)
	}
	defer tx.Rollback()

	p, err := lookupPicToDelete(task.PicId, tx)
	if err != nil {
		return err
	}

	pts, err := findPicTagsToDelete(task.PicId, tx)
	if err != nil {
		return err
	}

	ts, err := findTagsToDelete(pts, tx)
	if err != nil {
		return err
	}

	if err := deletePicTags(pts, tx); err != nil {
		return err
	}

	now := time.Now()
	if err := upleteTags(ts, now, tx); err != nil {
		return err
	}

	if err := p.Delete(tx); err != nil {
		return InternalError("Unable to Delete Pic", err)
	}

	if err := tx.Commit(); err != nil {
		return InternalError("Unable to Commit", err)
	}

	if err := os.Remove(p.Path(task.pixPath)); err != nil {
		log.Println("Warning, unable to delete pic data", p, err)
	}

	if err := os.Remove(p.ThumbnailPath(task.pixPath)); err != nil {
		log.Println("Warning, unable to delete pic data", p, err)
	}

	return nil
}

func findPicTagsToDelete(picId int64, tx *sql.Tx) ([]*schema.PicTag, Status) {
	stmt, err := schema.PicTagPrepare("SELECT * FROM_ WHERE %s = ? FOR UPDATE;", tx, schema.PicTagColPicId)
	if err != nil {
		return nil, InternalError("Unable to Prepare Lookup", err)
	}
	defer stmt.Close()
	pts, err := schema.FindPicTags(stmt, picId)
	if err != nil {
		return nil, InternalError("Error Looking up Pic Tags", err)
	}
	return pts, nil
}

func findTagsToDelete(pts []*schema.PicTag, tx *sql.Tx) ([]*schema.Tag, Status) {
	stmt, err := schema.TagPrepare("SELECT * FROM_ WHERE %s = ? FOR UPDATE;", tx, schema.TagColId)
	if err != nil {
		return nil, InternalError("Unable to Prepare Lookup", err)
	}
	defer stmt.Close()

	ts := make([]*schema.Tag, 0, len(pts))
	for _, pt := range pts {
		t, err := schema.LookupTag(stmt, pt.TagId)
		if err != nil {
			return nil, InternalError(fmt.Sprintf("Error Looking up Tag: %d", pt.TagId), err)
		}
		ts = append(ts, t)
	}
	return ts, nil
}

func deletePicTags(pts []*schema.PicTag, tx *sql.Tx) Status {
	for _, pt := range pts {
		if err := pt.Delete(tx); err != nil {
			return InternalError("Unable to Delete PicTag", err)
		}
	}
	return nil
}

func lookupPicToDelete(picId int64, tx *sql.Tx) (*schema.Pic, Status) {
	stmt, err := schema.PicPrepare("SELECT * FROM_ WHERE %s = ? FOR UPDATE;", tx, schema.PicColId)
	if err != nil {
		return nil, InternalError("Unable to Prepare Lookup", err)
	}
	defer stmt.Close()

	p, err := schema.LookupPic(stmt, picId)
	if err == sql.ErrNoRows {
		return nil, NotFound(fmt.Sprintf("Could not find pic %d", picId), nil)
	} else if err != nil {
		return nil, InternalError("Error Looking up Pic", err)
	}
	return p, nil
}

// if Update|Insert = Upsert, then Update|Delete = uplete?
func upleteTags(ts []*schema.Tag, now time.Time, tx *sql.Tx) Status {
	for _, t := range ts {
		if t.UsageCount > 1 {
			t.UsageCount--
			t.SetModifiedTime(now)
			if err := t.Update(tx); err != nil {
				return InternalError("Unable to Update Tag", err)
			}
		} else {
			if err := t.Delete(tx); err != nil {
				return InternalError("Unable to Delete Tag", err)
			}
		}
	}
	return nil
}
