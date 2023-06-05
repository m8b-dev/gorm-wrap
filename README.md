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

type PreloadedSingleArticle struct {
	gorm.Model
	Title string
	Content string
	CreatedByUserId uint `gorm:""`
	CreatedByUser   User `gorm:"foreignKey:CreatedByUserId;references:ID"`
}

type PreloadedMultipleArticle struct {
	gorm.Model
	Title string
	Content string
	CreatedByUserId uint `gorm:""`
	CreatedByUser   User `gorm:"foreignKey:CreatedByUserId;references:ID"`
	WebsiteId uint `gorm:""`
	Website Website `gorm:"foreignKey:WebsiteId;references:ID"`
}

type User struct {
	gorm.Model
	Username string
}

type Website struct {
	gorm.Model
	Address string
}

// Declare single preload
func (c *PreloadedSingleArticle) RequiresPreload() (string, func(*gorm.DB) *gorm.DB) {
	return "CreatedByUser", nil
}

// Declare multiple preload
func (c *PreloadedMultipleArticle) RequiresPreload() ([]string, []func(orm *gorm.DB) *gorm.DB) {
	return []string{"CreatedByUser", "Website"}, nil
}

```
