package binding

import (
	"mime/multipart"
	"net/http"
	"reflect"
)

type setter interface {
	TrySet(v reflect.Value, f reflect.StructField, key string) (bool, error)
}

type formSetter map[string][]string

var _ setter = (formSetter)(nil)

func (s formSetter) TrySet(v reflect.Value, f reflect.StructField, key string) (bool, error) {
	vs, ok := s[key]
	if !ok || len(vs) == 0 {
		return false, nil
	}

	if v.Kind() == reflect.Slice {
		vSlide := reflect.MakeSlice(v.Type(), 0, len(vs))
		elemKind := v.Type().Elem().Kind()

		for _, value := range vs {
			if convertValue, err := convert(elemKind, value); err != nil {
				return false, err
			} else {
				vSlide = reflect.Append(vSlide, convertValue)
			}
		}

		v.Set(vSlide)

		return true, nil
	}

	if convertValue, err := convert(v.Kind(), vs[0]); err != nil {
		return false, err
	} else {
		v.Set(convertValue)
		return true, nil
	}
}

type multipartSetter http.Request

var _ setter = (*multipartSetter)(nil)

func (s *multipartSetter) TrySet(v reflect.Value, f reflect.StructField, key string) (bool, error) {
	if files := s.MultipartForm.File[key]; len(files) != 0 {
		if _, ok := v.Interface().(*multipart.FileHeader); ok {
			v.Set(reflect.ValueOf(files[0]))
			return true, nil
		}

		if _, ok := v.Interface().([]*multipart.FileHeader); ok {
			vSlice := reflect.MakeSlice(v.Type(), 0, len(files))
			for _, fh := range files {
				vSlice = reflect.Append(vSlice, reflect.ValueOf(fh))
			}
			v.Set(vSlice)
			return true, nil
		}
	}

	return formSetter(s.MultipartForm.Value).TrySet(v, f, key)
}
