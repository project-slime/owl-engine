package common

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// 入参校验
func CheckParam(c *gin.Context, obj interface{}) error {
	err := c.ShouldBindJSON(obj)
	if err != nil {
		return err
	}
	return binding.Validator.ValidateStruct(obj)
}

// 对于整型的参数校验
var PageAndSizeValid validator.Func = func(field validator.FieldLevel) bool {
	if num, ok := field.Field().Interface().(int); ok {
		if num <= 0 {
			return false
		}
	}
	return true
}
