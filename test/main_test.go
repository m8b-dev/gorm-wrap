package main

import (
	"github.com/m8b-dev/gorm-wrap/ezg"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"strings"
	"testing"
)

func Test_Run(t *testing.T) {
	orm, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}

	mkFakeDB(t, orm)
	createRetrieve(t, orm)
	update(t, orm)
	del(t, orm)
}

func createRetrieve(t *testing.T, orm *gorm.DB) {
	err := ezg.W(&Author{
		Username: "SomeUser",
	}).Insert(orm)
	if err != nil {
		t.Fatal(err)
	}

	usr, err := ezg.W(&Author{Username: "SomeUser"}).FindOne(orm)
	if err != nil {
		t.Fatal(err)
	}
	if usr == nil {
		t.Fatal("user not found")
	}

	usr.Posts = append(usr.Posts, Post{
		Title:   "Hello world post",
		Content: "Lorem ipsum",
		Images: []Img{
			{Title: "Lorem ipsum image"},
			{Title: "Lorem ipsum second image"},
			{Title: "Lorem ipsum third image"},
		},
		Videos: []Vid{},
	})
	// insert new post
	err = ezg.W(usr).Update(orm)
	if err != nil {
		t.Fatal(err)
	}

	// test retrieves
	cnt, err := ezg.W(&Img{}).Count(orm)
	if err != nil {
		t.Fatal(err)
	}
	if cnt != 3 {
		t.Fatalf("expected 3 images, got %d images", cnt)
	}

	cnt, err = ezg.W(&Img{}).CountSql(orm, "title LIKE 'Lorem ipsum % image'")
	if err != nil {
		t.Fatal(err)
	}
	if cnt != 2 {
		t.Fatalf("expected 2 images, got %d images", cnt)
	}

	cnt, err = ezg.W(&Img{}).CountSql(orm, "title = ?", "Lorem ipsum image")
	if err != nil {
		t.Fatal(err)
	}
	if cnt != 1 {
		t.Fatalf("expected 1 image, got %d images", cnt)
	}

	img, err := ezg.W(&Img{}).FindOne(orm)
	if err != nil {
		t.Fatal(err)
	}
	if usr == nil {
		t.Fatal("image not found")
	}
	if !strings.HasPrefix(img.Title, "Lorem ipsum") {
		// does not guarantee order
		t.Fatal("invalid data from db")
	}

	usrs, err := ezg.W(&Author{}).FindSql(orm, "username = ?", "SomeUser")
	if err != nil {
		t.Fatal(err)
	}
	if len(usrs) == 0 {
		t.Fatal("user not found")
	}

	usrs, err = ezg.W(&Author{Username: "SomeUser"}).Find(orm)
	if err != nil {
		t.Fatal(err)
	}
	if len(usrs) == 0 {
		t.Fatal("user not found")
	}
	usr, err = ezg.W(&Author{}).FindOneSql(orm, "username = ?", "SomeUser")
	if err != nil {
		t.Fatal(err)
	}
	if usr == nil {
		t.Fatal("user not found")
	}
	// shallow finds use same function, test if it's indeed shallow
	usr, err = ezg.W(&Author{Username: "SomeUser"}).ShallowFindOne(orm)
	if err != nil {
		t.Fatal(err)
	}
	if usr == nil {
		t.Fatal("user not found")
	}
	if usr.Posts != nil {
		t.Fatal("shallow find not shallow")
	}

	imgs, err := ezg.W(&Img{}).FindPaginated(orm, ptr(uint64(1)), ptr(uint64(1)), false)
	if err != nil {
		t.Fatal(err)
	}
	if len(imgs) != 1 {
		t.Fatal("post not found")
	}
	if imgs[0].ID != 2 {
		t.Fatal("pagination failed - expected second post for offset=1.")
	}
}
func update(t *testing.T, orm *gorm.DB) {
	usr, err := ezg.W(&Author{Username: "SomeUser"}).FindOne(orm)
	if err != nil {
		t.Fatal(err)
	}
	if usr == nil {
		t.Fatal("user not found")
	}
	usr.Posts[0].Title = "My first post!"
	err = ezg.W(&usr.Posts[0]).Update(orm)
	if err != nil {
		t.Fatal(err)
	}
	post, err := ezg.W(&Post{Model: gorm.Model{ID: usr.Posts[0].Model.ID}}).FindOne(orm)
	if err != nil {
		t.Fatal(err)
	}
	if post.Title != "My first post!" {
		t.Fatal("update failed")
	}
}
func del(t *testing.T, orm *gorm.DB) {
	usr, err := ezg.W(&Author{Username: "SomeUser"}).FindOne(orm)
	if err != nil {
		t.Fatal(err)
	}
	if usr == nil {
		t.Fatal("user not found")
	}
	err = ezg.W(&usr.Posts[0]).Delete(orm)
	if err != nil {
		t.Fatal(err)
	}

	usrRefreshed, err := ezg.W(&Author{Username: "SomeUser"}).FindOne(orm)
	if err != nil {
		t.Fatal(err)
	}
	if usrRefreshed == nil {
		t.Fatal("user not found")
	}

	if len(usr.Posts) == len(usrRefreshed.Posts) {
		t.Fatal("delete failed")
	}
}

func mkFakeDB(t *testing.T, orm *gorm.DB) {
	err := orm.AutoMigrate(&Vid{}, &Img{}, &Post{}, &Author{})
	if err != nil {
		t.Fatal(err)
	}
}

func ptr[T any](val T) *T {
	return &val
}
