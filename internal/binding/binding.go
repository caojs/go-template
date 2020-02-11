package binding

import (
	"github.com/pkg/errors"
	"mime/multipart"
	"net/http"
	"reflect"
)

func Bind(r *http.Request, o interface{}) (bool, error) {
	v := reflect.ValueOf(o)
	if v.Kind() != reflect.Ptr && v.Elem().Kind() != reflect.Struct {
		return false, errors.New("interface must be a pointer of struct")
	}

	return bind(v.Elem(), reflect.StructField{}, r)
}

func bind(v reflect.Value, f reflect.StructField, r *http.Request) (bool, error) {
	vKind := v.Kind()
	vType := v.Type()

	_, isFileHeader := v.Interface().(*multipart.FileHeader)
	_, isFileHeaderSlice := v.Interface().([]*multipart.FileHeader)
	if isFileHeader || isFileHeaderSlice {
		return setMultipartValue(v, f, r)
	}

	switch vKind {
	case reflect.Ptr:
		var isNew bool
		vPtr := v
		if v.IsNil() {
			isNew = true
			vPtr = reflect.New(vType.Elem())
		}

		setted, err := bind(vPtr.Elem(), f, r)
		if err != nil {
			return false, err
		}
		if isNew && setted {
			v.Set(vPtr)
		}
		return setted, nil

	case reflect.Struct:
		setted := false

		for i := 0; i < v.NumField(); i++ {
			sf := vType.Field(i)
			s, err := bind(v.Field(i), sf, r)
			if err != nil {
				return false, err
			}
			setted = setted || s
		}

		return setted, nil

	default:
		return setValue(v, f, r)
	}
}

func tagName(f reflect.StructField) string {
	tag := f.Tag.Get("json")
	if tag == "" {
		tag = f.Name
	}
	return tag
}

func setValue(v reflect.Value, f reflect.StructField, r *http.Request) (bool, error) {
	tag := tagName(f)
	if tag == "" {
		return false, nil
	}

	if v.Kind() == reflect.Slice {
		setted := false
		if vs, ok := r.PostForm[tag]; ok {
			vSlide := reflect.MakeSlice(v.Type(), 0, len(vs))
			elemKind := v.Type().Elem().Kind()

			for _, value := range vs {
				if convertValue, err := convert(elemKind, value); err != nil {
					return false, err
				} else {
					vSlide = reflect.Append(vSlide, convertValue)
					setted = true
				}
			}

			if setted {
				v.Set(vSlide)
			}
		}

		return setted, nil
	}

	if r.PostForm.Get(tag) == "" {
		return false, nil
	}

	if convertValue, err := convert(v.Kind(), r.PostForm.Get(tag)); err != nil {
		return false, err
	} else {
		v.Set(convertValue)
		return true, nil
	}
}

func convert(vKind reflect.Kind, value string) (reflect.Value, error) {
	if convert, ok := builtinConverters[vKind]; ok {
		return convert(value), nil
	}

	return reflect.Value{}, errors.New("unsupported converter")
}

func setMultipartValue(v reflect.Value, f reflect.StructField, r *http.Request) (bool, error) {
	tag := tagName(f)
	if tag == "" {
		return false, nil
	}

	fhs := r.MultipartForm.File[tag]

	if len(fhs) == 0 {
		return false, nil
	}

	if _, ok := v.Interface().(*multipart.FileHeader); ok {
		v.Set(reflect.ValueOf(fhs[0]))
		return true, nil
	}

	if _, ok := v.Interface().([]*multipart.FileHeader); ok {
		vSlice := reflect.MakeSlice(v.Type(), 0, len(fhs))
		for _, fh := range fhs {
			vSlice = reflect.Append(vSlice, reflect.ValueOf(fh))
		}
		v.Set(vSlice)
		return true, nil
	}

	return false, nil
}
