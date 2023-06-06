# Gorm wrapper

This is convenience package for gorm, making the api slightly more friendly.

It allows simplified usage of most common usages of gorm, removes gorm's error not found and uses nils / empty slices instead.

It allows overriding convenience functions, the preloading requires minimal setup (implementing a function that defines what to use with .Preload function from gorm)

Usage example:

```go

type MyModel struct {
	gorm.Model // MUST be included or it will be broken. It is OK to have custom struct instead of gorm.Model
	           // but field MUST be named "Model", and MUST contain `ID uint` field.

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


```
