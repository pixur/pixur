package tasks

import (
	"bytes"
	"crypto/sha256"
	"database/sql"
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"io"
	"io/ioutil"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"testing"

	"pixur.org/pixur/schema"
	ptest "pixur.org/pixur/testing"
)

type container struct {
	t  *testing.T
	db *sql.DB

	pixPath           string
	createdPicIds     []int64
	createdTagIds     []int64
	createdPicTagKeys []schema.PicTagKey
}

func (c *container) CreatePic() *schema.Pic {
	h := sha256.New()
	if err := binary.Write(h, binary.LittleEndian, rand.Int63()); err != nil {
		c.t.Fatal(err)
	}
	p := &schema.Pic{
		Sha256Hash: h.Sum(nil),
	}
	if err := p.Insert(c.db); err != nil {
		c.t.Fatal(err, p)
	}
	c.createdPicIds = append(c.createdPicIds, p.PicId)
	if err := c.writeImageData(p); err != nil {
		c.t.Fatal(err)
	}
	if err := c.writeThumbnailData(p); err != nil {
		c.t.Fatal(err)
	}

	return p
}

func (c *container) CreateTag() *schema.Tag {
	dictionary := "abcdefghijklmnopqrstuvwxyz"
	var name string
	for i := 0; i < 6; i++ {
		name += string(dictionary[rand.Intn(len(dictionary))])
	}
	t := &schema.Tag{Name: name}
	if err := t.Insert(c.db); err != nil {
		c.t.Fatal(err)
	}
	c.createdTagIds = append(c.createdTagIds, t.TagId)

	return t
}

func (c *container) CreatePicTag(p *schema.Pic, t *schema.Tag) *schema.PicTag {
	picTag := &schema.PicTag{
		PicId: p.PicId,
		TagId: t.TagId,
		Name:  t.Name,
	}
	if _, err := picTag.Insert(c.db); err != nil {
		c.t.Fatal(err)
	}
	t.UsageCount++
	if err := t.Update(c.db); err != nil {
		c.t.Fatal(err)
	}
	c.createdPicTagKeys = append(c.createdPicTagKeys, schema.PicTagKey{
		PicId: p.PicId,
		TagId: t.TagId,
	})
	return picTag
}

func (c *container) RefreshPic(p **schema.Pic) {
	stmt, err := schema.PicPrepare("SELECT * FROM_ WHERE %s = ?;", c.db, schema.PicColId)
	if err != nil {
		c.t.Fatal(err)
	}
	updated, err := schema.LookupPic(stmt, (*p).PicId)
	if err == sql.ErrNoRows {
		*p = nil
	} else if err != nil {
		c.t.Fatal(err)
	}
	*p = updated
}

func (c *container) RefreshTag(t **schema.Tag) {
	stmt, err := schema.TagPrepare("SELECT * FROM_ WHERE %s = ?;", c.db, schema.TagColId)
	if err != nil {
		c.t.Fatal(err)
	}
	updated, err := schema.LookupTag(stmt, (*t).TagId)
	if err == sql.ErrNoRows {
		*t = nil
	} else if err != nil {
		c.t.Fatal(err)
	}
	*t = updated
}

func (c *container) RefreshPicTag(pt **schema.PicTag) {
	stmt, err := schema.PicTagPrepare("SELECT * FROM_ WHERE %s = ? AND %s = ?;",
		c.db, schema.PicTagColPicId, schema.PicTagColTagId)
	if err != nil {
		c.t.Fatal(err)
	}
	updated, err := schema.LookupPicTag(stmt, (*pt).PicId, (*pt).TagId)
	if err == sql.ErrNoRows {
		*pt = nil
	} else if err != nil {
		c.t.Fatal(err)
	}
	*pt = updated
}

func (c *container) CleanUp() {
	for _, picTagKey := range c.createdPicTagKeys {
		if err := schema.DeletePicTag(picTagKey, c.db); err != nil {
			c.t.Error(err)
		}
	}
	c.createdPicTagKeys = nil

	for _, picId := range c.createdPicIds {
		if err := (&schema.Pic{PicId: picId}).Delete(c.db); err != nil {
			c.t.Error(err)
		}
	}
	c.createdPicIds = nil

	for _, tagId := range c.createdTagIds {
		if err := (&schema.Tag{TagId: tagId}).Delete(c.db); err != nil {
			c.t.Error(err)
		}
	}
	c.createdTagIds = nil

	if c.pixPath != "" {
		if err := os.RemoveAll(c.pixPath); err != nil {
			c.t.Error(err)
		}
	}
	c.pixPath = ""
}

func (c *container) mkPixPath() string {
	if c.pixPath != "" {
		return c.pixPath
	}
	if path, err := ioutil.TempDir("", "unitTestPixPath"); err != nil {
		c.t.Fatal(err)
	} else {
		c.pixPath = path
	}
	return c.pixPath
}

func (c *container) getRandomImageData() *bytes.Reader {
	bounds := image.Rect(0, 0, 5, 10)
	img := image.NewGray(bounds)
	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			img.SetGray(x, y, color.Gray{Y: uint8(rand.Int31n(math.MaxUint8))})
		}
	}
	f := bytes.NewBuffer(nil)

	if err := gif.Encode(f, img, &gif.Options{}); err != nil {
		c.t.Fatal(err)
	}
	return bytes.NewReader(f.Bytes())
}

func (c *container) writeImageData(p *schema.Pic) error {
	path := p.Path(c.mkPixPath())
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0770); err != nil {
		c.t.Fatal(err)
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := io.Copy(f, c.getRandomImageData()); err != nil {
		return err
	}
	return nil
}

func (c *container) writeThumbnailData(p *schema.Pic) error {
	path := p.ThumbnailPath(c.mkPixPath())
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0770); err != nil {
		c.t.Fatal(err)
	}
	f, err := os.Create(p.ThumbnailPath(c.mkPixPath()))
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := io.Copy(f, c.getRandomImageData()); err != nil {
		return err
	}
	return nil
}

var (
	testDB         *sql.DB
	_testSetups    []func() error
	_testTearDowns []func() error
)

func BeforeTestSuite(before func() error) {
	_testSetups = append(_testSetups, before)
}

func AfterTestSuite(after func() error) {
	_testTearDowns = append(_testTearDowns, after)
}

func init() {
	BeforeTestSuite(func() error {
		db, err := ptest.GetDB()
		if err != nil {
			return err
		}
		AfterTestSuite(func() error {
			ptest.CleanUp()
			return nil
		})
		testDB = db
		if err := schema.CreateTables(db); err != nil {
			return err
		}
		return nil
	})
}

func runTests(m *testing.M) int {
	defer func() {
		for _, after := range _testTearDowns {
			if err := after(); err != nil {
				fmt.Println("Error in teardown", err)
			}
		}
	}()

	for _, before := range _testSetups {
		if err := before(); err != nil {
			fmt.Println("Error in test setup", err)
			return 1
		}
	}

	return m.Run()
}

func TestMain(m *testing.M) {
	os.Exit(runTests(m))
}