package ezg

import (
	"errors"
	"fmt"
	"gorm.io/gorm"
	"reflect"
)

// Database helper methods use wrapper function to apply general purpose functions. Most of the functions can be
// overridden if passed model implements called function (with variant functions instead of defining 2 methods, it uses second
// parameter for variant, ie. instead of ShallowFindOne(gorm) and FindOne(gorm), the override is FindOne(gorm, shallow bool))
// If no records are found, instead of returning error record not found, a nil object will be returned. For slices,
// instead of nil slice, empty slice is returned. Helper is bootstrapped with &model{} where it's fields will be passed
// to gorm as WHERE clause, allowing querying like this:
// result, err := W(&Model{UUID: "abc"}).FindOne(orm)
// Relations should be marked with function on model struct:
// RequiresPreload() ([]string, []func(orm *gorm.DB) *gorm.DB)
// or, for only single relation
// RequiresPreload() (string, func(orm *gorm.DB) *gorm.DB)
// where string is name of other model, and function is for things like order by. Function can be nil, and if it's not
// used for anything useful, should be nil. Not including RequiresPreload on model that does require preloading will
// break that relation.

// Q represents a generalized struct wrapper that is used for CRUD operations on any gorm.Model.
type Q[t any] struct{ obj *t }

// M is a short form for Model. It returns the underlying model.
func (q Q[t]) M() *t {
	return q.obj
}

// W is a short form for Wrap. It wraps the model with general purpose functions.
func W[t any](obj *t) Q[t] {
	return Q[t]{obj: obj}
}

// Insert inserts the underlying model object into the database using GORM.
// If the model implements a custom Insert method, it will be used instead.
func (q Q[t]) Insert(db *gorm.DB) error {
	if o, ok := interface{}(q.obj).(interface{ Insert(db *gorm.DB) error }); ok {
		return o.Insert(db)
	}

	return db.Create(q.obj).Error
}

// Update updates the underlying model object in the database using GORM.
// If the model implements a custom Update method, it will be used instead.
func (q Q[t]) Update(db *gorm.DB) error {
	if o, ok := interface{}(q.obj).(interface{ Update(db *gorm.DB) error }); ok {
		return o.Update(db)
	}

	return db.Save(q.obj).Error
}

// Delete deletes the underlying model object from the database using GORM.
// If the model implements a custom Delete method, it will be used instead.
// If the model does not use gorm.Model while not implementing custom model method, it will return an error.
func (q Q[t]) Delete(db *gorm.DB) error {
	if o, ok := interface{}(q.obj).(interface{ Delete(db *gorm.DB) error }); ok {
		return o.Delete(db)
	}

	metaValue := reflect.ValueOf(q.obj).Elem()
	field := metaValue.FieldByName("Model")
	if field == (reflect.Value{}) {
		return errors.New("LOGIC ERROR: model is not gorm.Model. Implement override function")
	}
	field = field.FieldByName("ID")
	if field == (reflect.Value{}) || !field.CanUint() {
		return errors.New("LOGIC ERROR: model is not gorm.Model. Implement override function")
	}
	value := field.Uint()
	return db.Model(q.obj).Delete("id", value).Error
}

// FindOne retrieves a single instance of the underlying model from the database using GORM.
// If the model implements a custom FindOne method, it will be used instead.
// Instead of using gorm.ErrRecordNotFound it will return nil model and nil error.
func (q Q[t]) FindOne(db *gorm.DB) (*t, error) {
	return q.findOne(db, false)
}

// ShallowFindOne retrieves a single instance of the underlying model from the database using GORM,
// without preloading any associations.
// If the model implements a custom FindOne method, it will be used instead.
// Instead of using gorm.ErrRecordNotFound it will return nil model and nil error.
func (q Q[t]) ShallowFindOne(db *gorm.DB) (*t, error) {
	return q.findOne(db, true)
}

func (q Q[t]) findOne(db *gorm.DB, shallow bool) (*t, error) {
	if o, ok := interface{}(q.obj).(interface {
		FindOne(db *gorm.DB, shallow bool) (*t, error)
	}); ok {
		return o.FindOne(db, shallow)
	}

	err := q.preload(db.Where(q.obj), shallow).First(q.obj).Error

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		err = fmt.Errorf("failed to read database: %w", err)
	}
	return q.obj, err
}

// FindOneSql retrieves a single instance of the underlying model from the database using GORM,
// using a custom SQL query.
// If the model implements a custom FindOneSql method, it will be used instead.
// Instead of using gorm.ErrRecordNotFound it will return nil model and nil error.
func (q Q[t]) FindOneSql(db *gorm.DB, sql string, sqlArgs ...interface{}) (*t, error) {
	return q.findOneSql(db, false, sql, sqlArgs...)
}

// ShallowFindOneSql retrieves a single instance of the underlying model from the database using GORM,
// using a custom SQL query and without preloading any associations.
// If the model implements a custom FindOneSql method, it will be used instead.
// Instead of using gorm.ErrRecordNotFound it will return nil model and nil error.
func (q Q[t]) ShallowFindOneSql(db *gorm.DB, sql string, sqlArgs ...interface{}) (*t, error) {
	return q.findOneSql(db, true, sql, sqlArgs...)
}

func (q Q[t]) findOneSql(db *gorm.DB, shallow bool, sql string, sqlArgs ...interface{}) (*t, error) {
	if o, ok := interface{}(q.obj).(interface {
		FindOneSql(db *gorm.DB, sql string, sqlArgs ...interface{}) (*t, error)
	}); ok {
		return o.FindOneSql(db, sql, sqlArgs...)
	}

	err := q.preload(db.Model(q.obj).Where(sql, sqlArgs...), shallow).First(q.obj).Error

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		err = fmt.Errorf("failed to read database: %w", err)
	}
	return q.obj, err
}

// Find retrieves all instances of the underlying model from the database using GORM.
// If the model implements a custom Find method, it will be used instead.
// Instead of using gorm.ErrRecordNotFound it will return empty slice and nil error.
func (q Q[t]) Find(db *gorm.DB) ([]t, error) {
	return q.find(db, false, false)
}

// FindReverse retrieves all instances of the underlying model from the database using GORM in reverse order.
func (q Q[t]) FindReverse(db *gorm.DB) ([]t, error) {
	return q.find(db, false, true)
}

// ShallowFind retrieves all instances of the underlying model from the database using GORM,
// without preloading any associations.
// If the model implements a custom Find method, it will be used instead.
// Instead of using gorm.ErrRecordNotFound it will return empty slice and nil error.
func (q Q[t]) ShallowFind(db *gorm.DB) ([]t, error) {
	return q.find(db, true, false)
}

func (q Q[t]) find(db *gorm.DB, shallow bool, reverseOrder bool) ([]t, error) {
	if o, ok := interface{}(q.obj).(interface {
		Find(db *gorm.DB) ([]t, error)
	}); ok {
		return o.Find(db)
	}

	orderStr := "id ASC"
	if reverseOrder {
		orderStr = "id DESC"
	}

	out := make([]t, 0)
	err := q.preload(db.Where(q.obj), shallow).Order(orderStr).Find(&out).Error

	if err == gorm.ErrRecordNotFound {
		return make([]t, 0), nil
	}
	return out, err
}

// FindSql retrieves all instances of the underlying model from the database using GORM,
// using a custom SQL query.
// If the model implements a custom FindSql method, it will be used instead.
// Instead of using gorm.ErrRecordNotFound it will return empty slice and nil error.
func (q Q[t]) FindSql(db *gorm.DB, sql string, sqlArgs ...interface{}) ([]t, error) {
	return q.findSql(db, false, sql, sqlArgs...)
}

// Join retrieves a single instance of the underlying model from the database using GORM,
// with a join on another table using a custom condition.
// Instead of using gorm.ErrRecordNotFound it will return nil model and nil error.
func (q Q[t]) Join(db *gorm.DB, table, condition string) (*t, error) {
	return q.join(db, table, condition)
}

func (q Q[t]) join(db *gorm.DB, table, condition string) (*t, error) {
	err := q.preload(
		db.Model(q.obj).Joins(fmt.Sprintf("INNER JOIN %s ON %s", table, condition)),
		false,
	).First(q.obj).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	if err != nil {
		err = fmt.Errorf("failed to read database: %w", err)
	}

	return q.obj, err
}

// ShallowFindSql retrieves all instances of the underlying model from the database using GORM,
// using a custom SQL query and without preloading any associations.
// If the model implements a custom FindSql method, it will be used instead.
// Instead of using gorm.ErrRecordNotFound it will return empty slice and nil error.
func (q Q[t]) ShallowFindSql(db *gorm.DB, sql string, sqlArgs ...interface{}) ([]t, error) {
	return q.findSql(db, true, sql, sqlArgs...)
}

func (q Q[t]) findSql(db *gorm.DB, shallow bool, sql string, sqlArgs ...interface{}) ([]t, error) {
	if o, ok := interface{}(q.obj).(interface {
		FindSql(db *gorm.DB, shallow bool, sql string, sqlArgs ...interface{}) ([]t, error)
	}); ok {
		return o.FindSql(db, shallow, sql, sqlArgs...)
	}

	out := make([]t, 0)
	err := q.preload(db.Model(q.obj).Where(sql, sqlArgs...), shallow).Order("id ASC").Find(&out).Error

	if err == gorm.ErrRecordNotFound {
		return make([]t, 0), nil
	}
	return out, err
}

// FindPaginated retrieves a slice of models from the database with pagination parameters (limit and offset).
// If the model implements a custom FindPaginated method, it will be used instead.
// Instead of using gorm.ErrRecordNotFound it will return empty slice and nil error.
func (q Q[t]) FindPaginated(db *gorm.DB, offset *uint64, limit *uint64, reverseOrder bool) ([]t, error) {
	return q.findPaginated(db, offset, limit, reverseOrder, false)
}

// ShallowFindPaginated retrieves a slice of models from the database with pagination parameters (limit and offset), without preloading.
// If the model implements a custom FindPaginated method, it will be used instead.
// Instead of using gorm.ErrRecordNotFound it will return empty slice and nil error.
func (q Q[t]) ShallowFindPaginated(db *gorm.DB, offset *uint64, limit *uint64, reverseOrder bool) ([]t, error) {
	return q.findPaginated(db, offset, limit, reverseOrder, true)
}
func (q Q[t]) findPaginated(db *gorm.DB, offset *uint64, limit *uint64, reverseOrder, shallow bool) ([]t, error) {
	if o, ok := interface{}(q.obj).(interface {
		FindPaginated(db *gorm.DB, offset *uint64, limit *uint64, reverseOrder bool, shallow bool) ([]t, error)
	}); ok {
		return o.FindPaginated(db, offset, limit, reverseOrder, shallow)
	}

	out := make([]t, 0)
	qry := q.preload(db.Where(q.obj), shallow)
	if offset != nil {
		qry = qry.Offset(int(*offset))
	}
	if limit != nil {
		qry = qry.Limit(int(*limit))
	}
	orderStr := "id ASC"
	if reverseOrder {
		orderStr = "id DESC"
	}
	err := qry.Order(orderStr).Find(&out).Error
	if err == gorm.ErrRecordNotFound {
		return make([]t, 0), nil
	}
	if err != nil {
		err = fmt.Errorf("failed to read database: %w", err)
	}
	return out, err
}

// FindPaginatedSql retrieves a slice of models from the database with pagination, optional reverse ordering, and with custom WHERE SQL.
// If the model implements a custom FindPaginatedSql method, it will be used instead.
// Instead of using gorm.ErrRecordNotFound it will return empty slice and nil error.
func (q Q[t]) FindPaginatedSql(db *gorm.DB, offset *uint64, limit *uint64, reverseOrder bool, sql string, sqlArgs ...interface{}) ([]t, error) {
	return q.findPaginatedSql(db, offset, limit, reverseOrder, false, sql, sqlArgs...)
}

// ShallowFindPaginatedSql retrieves a slice of models from the database with pagination, optional reverse ordering, without preloading and with custom WHERE SQL.
// If the model implements a custom FindPaginatedSql method, it will be used instead.
// Instead of using gorm.ErrRecordNotFound it will return empty slice and nil error.
func (q Q[t]) ShallowFindPaginatedSql(db *gorm.DB, offset *uint64, limit *uint64, reverseOrder bool, sql string, sqlArgs ...interface{}) ([]t, error) {
	return q.findPaginatedSql(db, offset, limit, reverseOrder, true, sql, sqlArgs...)
}
func (q Q[t]) findPaginatedSql(db *gorm.DB, offset *uint64, limit *uint64, reverseOrder, shallow bool, sql string, sqlArgs ...interface{}) ([]t, error) {
	if o, ok := interface{}(q.obj).(interface {
		FindPaginatedSql(db *gorm.DB, offset *uint64, limit *uint64, reverseOrder bool, sql string, sqlArgs ...interface{}) ([]t, error)
	}); ok {
		return o.FindPaginatedSql(db, offset, limit, reverseOrder, sql, sqlArgs...)
	}

	out := make([]t, 0)
	qry := q.preload(db.Model(q.obj).Where(sql, sqlArgs...), shallow)
	if offset != nil {
		qry = qry.Offset(int(*offset))
	}
	if limit != nil {
		qry = qry.Limit(int(*limit))
	}
	orderStr := "id ASC"
	if reverseOrder {
		orderStr = "id DESC"
	}
	err := qry.Order(orderStr).Find(&out).Error
	if err == gorm.ErrRecordNotFound {
		return make([]t, 0), nil
	}
	if err != nil {
		err = fmt.Errorf("failed to read database: %w", err)
	}

	return out, err
}

// CountSql counts the number of rows in the database that match the custom SQL query and arguments.
// If the model implements a custom CountSql method, it will be used instead.
func (q Q[t]) CountSql(db *gorm.DB, sql string, sqlArgs ...interface{}) (uint64, error) {
	if o, ok := interface{}(q.obj).(interface {
		CountSql(db *gorm.DB, sql string, sqlArgs ...interface{}) (uint64, error)
	}); ok {
		return o.CountSql(db, sql, sqlArgs...)
	}
	count := int64(0)
	err := db.Model(&q.obj).Where(sql, sqlArgs...).Count(&count).Error
	if err == gorm.ErrRecordNotFound {
		return 0, nil
	}

	return uint64(count), err
}

// Count counts the number of rows in the database that match the model.
// If the model implements a custom Count method, it will be used instead.
func (q Q[t]) Count(db *gorm.DB) (uint64, error) {
	if o, ok := interface{}(q.obj).(interface {
		Count(db *gorm.DB) (uint64, error)
	}); ok {
		return o.Count(db)
	}

	count := int64(0)
	err := db.Model(q.obj).Where(q.obj).Count(&count).Error
	if err == gorm.ErrRecordNotFound {
		return 0, nil
	}
	if err != nil {
		err = fmt.Errorf("failed to read database: %w", err)
	}
	return uint64(count), err
}

func (q Q[t]) preload(qry *gorm.DB, shallow bool) *gorm.DB {
	if shallow {
		return qry
	}
	if o, ok := interface{}(q.obj).(interface {
		RequiresPreload() (string, func(orm *gorm.DB) *gorm.DB)
	}); ok {
		a, b := o.RequiresPreload()
		if b == nil {
			return qry.Preload(a)
		}
		return qry.Preload(a, b)
	} else if o, ok := interface{}(q.obj).(interface { // support for multi table preload
		RequiresPreload() ([]string, []func(orm *gorm.DB) *gorm.DB)
	}); !ok {
		a := autoPreloads(q.obj)
		for i := range a {
			qry = qry.Preload(a[i])
		}
	} else {
		a, b := o.RequiresPreload()
		if b == nil || len(b) == 0 {
			for i := range a {
				qry = qry.Preload(a[i])
			}
			return qry
		}
		if len(a) != len(b) {
			name := reflect.ValueOf(q.obj).Type().Name()
			//           This is a logic error which will be obvious if it occurs - every read
			//           of the affected table will fail. As the RequiresPreload interface should be constant for a model,
			//           this error points to an inconsistency in the model's definition. Hence, it is justified to halt
			//           execution (panic) so this logic error can be detected and fixed during the development phase.
			//           Specifically, when a model declares multi-preload, the length of gorm functions (preload
			//           conditions or modifications) should be either 0 (meaning no specific conditions for all
			//           preloaded tables) or equal to the length of preloaded tables (meaning each preload has a
			//           corresponding condition or modification, even if it's nil - which is valid usage).
			panic(fmt.Sprintf("LOGIC ERROR: model %s declares multi-preload but does not define consistent "+
				"preload definition, length of gorm functions must be either 0 or equal to length of preloaded tables, "+
				"instead got len(tables) = %d, len(gormFuncs) = %d. Gorm function is allways allowable to be nil for"+
				" selective usage", name, len(a), len(b)))
		}
		for i := range a {
			if b[i] == nil {
				qry = qry.Preload(a[i])
			} else {
				qry = qry.Preload(a[i], b[i])
			}
		}
	}

	return qry
}
