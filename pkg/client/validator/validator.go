package validator

import (
	"github.com/gin-gonic/gin/binding"
	"reflect"
	"sync"
	"time"

	"github.com/go-playground/validator/v10"
)

func Setup() {
	binding.Validator = new(owlValidator)
}

// 入参验证
type owlValidator struct {
	once     sync.Once
	validate *validator.Validate
}

func (c *owlValidator) ValidateStruct(obj interface{}) error {

	if kindOfData(obj) == reflect.Struct {

		c.lazyInit()

		return c.validate.Struct(obj)
	}

	return nil
}

func (c *owlValidator) Engine() interface{} {
	c.lazyInit()
	return c.validate
}
func (c *owlValidator) lazyInit() {
	c.once.Do(func() {
		c.validate = validator.New()
		c.validate.SetTagName("binding")

		c.validate.RegisterValidation("date", func(fl validator.FieldLevel) bool {
			_, err := time.Parse("2006-01-02", fl.Field().String())
			if err != nil {
				return false
			}
			return true
		})

		c.validate.RegisterValidation("time", func(fl validator.FieldLevel) bool {
			_, err := time.Parse("2006-01-02 15:04:05", fl.Field().String())
			if err != nil {
				return false
			}
			return true
		})

		c.validate.RegisterValidation("page", func(fl validator.FieldLevel) bool {
			return fl.Field().Int() > 0
		})

		c.validate.RegisterValidation("id", func(fl validator.FieldLevel) bool {
			return fl.Field().Int() > 0
		})
	})
}

func kindOfData(data interface{}) reflect.Kind {
	value := reflect.ValueOf(data)
	valueType := value.Kind()

	if valueType == reflect.Ptr {
		valueType = value.Elem().Kind()
	}
	return valueType
}
