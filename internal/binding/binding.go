package binding

import (
	"github.com/pkg/errors"
	"mime/multipart"
	"net/http"
	"reflect"
)

func Bind(r *http.Request, o interface{}) error {
	v := reflect.ValueOf(o)
	if v.Kind() != reflect.Ptr && v.Elem().Kind() != reflect.Struct {
		return errors.New("interface must be a pointer of struct")
	}

	return bind(v.Elem(), reflect.StructField{}, r)
}

func bind(v reflect.Value, f reflect.StructField, r *http.Request) error {
	vKind := v.Kind()
	vType := v.Type()

	_, isFileHeader := v.Interface().(*multipart.FileHeader)
	_, isFileHeaderSlice := v.Interface().([]*multipart.FileHeader)
	if isFileHeader || isFileHeaderSlice {
		return setMultipartValue(v, f, r)
	}

	switch vKind {
	case reflect.Ptr:
		vPtr := v
		if v.IsNil() {
			vPtr = reflect.New(v.Elem().Type())
		}

		if err := bind(vPtr.Elem(), f, r); err != nil {
			return err
		}

		v.Set(vPtr)
		return nil

	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			sf := vType.Field(i)
			if err := bind(v.Field(i), sf, r); err != nil {
				return err
			}
		}
		return nil

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

func setValue(v reflect.Value, f reflect.StructField, r *http.Request) error {
	tag := tagName(f)
	if tag == "" {
		return nil
	}

	if v.Kind() == reflect.Slice {
		if vs, ok := r.PostForm[tag]; ok {
			vSlide := reflect.MakeSlice(v.Type(), 0, len(vs))
			elemKind := v.Type().Elem().Kind()

			for _, value := range vs {
				if convertValue, err := convert(elemKind, value); err != nil {
					return err
				} else {
					vSlide = reflect.Append(vSlide, convertValue)
				}
			}

			v.Set(vSlide)
		}
		return nil
	}

	if r.PostForm.Get(tag) == "" {
		return nil
	}

	if convertValue, err := convert(v.Kind(), r.PostForm.Get(tag)); err != nil {
		return err
	} else {
		v.Set(convertValue)
		return nil
	}
}

func convert(vKind reflect.Kind, value string) (reflect.Value, error) {
	if convert, ok := builtinConverters[vKind]; ok {
		return convert(value), nil
	}

	return reflect.Value{}, errors.New("unsupported converter")
}

func setMultipartValue(v reflect.Value, f reflect.StructField, r *http.Request) error {
	tag := tagName(f)
	if tag == "" {
		return nil
	}

	fhs := r.MultipartForm.File[tag]

	if len(fhs) == 0 {
		return nil
	}

	if _, ok := v.Interface().(*multipart.FileHeader); ok {
		v.Set(reflect.ValueOf(fhs[0]))
		return nil
	}

	if _, ok := v.Interface().([]*multipart.FileHeader); ok {
		vSlice := reflect.MakeSlice(v.Type(), 0, len(fhs))
		for _, fh := range fhs {
			vSlice = reflect.Append(vSlice, reflect.ValueOf(fh))
		}
		v.Set(vSlice)
		return nil
	}

	return nil
}