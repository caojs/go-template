package binding

import (
	"github.com/pkg/errors"
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

func setValue(v reflect.Value, f reflect.StructField, r *http.Request) error {
	tag := f.Tag.Get("json")
	if tag == "" {
		tag = f.Name
	}

	if tag == "" {
		return nil
	}

	if err := r.ParseForm(); err != nil {
		return nil
	}

	if v.Kind() == reflect.Slice {
		if vs, ok := r.PostForm[tag]; ok {
			vSlide := reflect.MakeSlice(v.Type(), 0, len(vs))

			for _, value := range vs {
				if convertValue, err := convert(v.Type().Elem().Kind(), value); err != nil {
					return err
				} else {
					vSlide = reflect.Append(vSlide, convertValue)
				}
			}

			v.Set(vSlide)
		}
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
