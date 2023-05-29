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

err := ghlpr.W(&MyModel{
	Foo: "hello",
	Bar: "world!",
}).Insert(orm)


// R

mod, err := ghlpr.W(&MyModel{
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

err = ghlpr.W(mod).Update(orm)

// D

err = ghlpr.W(mod).Delete(orm)


```
