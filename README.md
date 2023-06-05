# Gorm wrapper

This is convinience package for gorm, making the api slightly more friendly.

It allows simplified usage of most common usages of gorm, removes gorm's error not found and uses nils / empty slices instead.

It allows overriding convinience functions, the preloading requires minimal setup (implementing a function that defines what to use with .Preload function from gorm)

Usage example:

```go

type MyModel struct {
	gorm.Model // MUST be included or it will be broken.

	Foo string
	Bar string
}

// ...

var orm *gorm.DB

// C

err := ezg.W(&MyModel{
	Foo: "hello",
	Bar: "world!",
}).Insert(orm)


// R

mod, err := ezg.W(&MyModel{
	Foo: "hello",
}).FindOne(orm)

if err != nil {
	// actual error happened
}

if mod != nil {
	fmt.Printf("Found hello foo with bar = %s\n", mod.Bar)
} else {
	fmt.Println("Foo not found!")
}

// U

mod.Bar = "new bar"

err = ezg.W(mod).Update(orm)

// D

err = ezg.W(mod).Delete(orm)

// Preload

type Image struct {
    gorm.Model

    Slug      string
    StorePath string

    ArticleId uint
}

type Tag struct {
    gorm.Model

    Name string
    Slug string

    ArticleId uint
}

type Article struct {
    gorm.Model

    Title   string
    Content string
    Tags    []Tag
    Images  []Image

    UserId uint
}

type User struct {
    gorm.Model

    Username string
    Articles []Article
}

func (u *User) RequiresPreload() (string, func(*gorm.DB) *gorm.DB) {
    return "Article", nil
}

func (a *Article) RequiresPreload() ([]string, []func(orm *gorm.DB) *gorm.DB) {
    return []string{"Tag", "Image"}, []func(*gorm.DB) *gorm.DB{func(orm *gorm.DB) *gorm.DB {
        return orm.Order("tags.name ASC") // order tags alphabetically
    }, nil}
}

// example

func UserInfo(orm *gorm.DB) {
    user, err := W(&User{
        Username: "DubbaThony",
    }).FindOne(orm)
    if err != nil {
        // ..
        return
    }
    if user == nil {
        // ...
        return
    }
    fmt.Printf("User %s have written %d articles:\n", user.Username, len(user.Articles))
    for i := range user.Articles {
        fmt.Printf("  - %s (%d images, %d tags)\n", user.Articles[i].Title, len(user.Articles[i].Images), len(user.Articles[i].Tags))
    }
}

```
