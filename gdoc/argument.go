package gdoc

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
)

const (
	tagJson     = "json"
	tagNote     = "note"
	tagRequired = "required"
)

type argument struct {
	Model

	Parent   *argument   `json:"-"`
	Children []*argument `json:"children"`
}

func (s *argument) ToModel() []*Type {
	models := make(TypeCollection, 0)

	ts := make(map[string]*Type, 0)
	s.toType(ts)
	for _, v := range ts {
		models = append(models, v)
	}
	sort.Stable(models)

	return models
}

func (s *argument) FromExample(example interface{}) *argument {
	if example == nil {
		return nil
	}

	model := &argument{Children: make([]*argument, 0)}

	exampleType := reflect.TypeOf(example)
	exampleTypeKind := exampleType.Kind()
	switch exampleTypeKind {
	case reflect.Ptr:
		{
			s.parseExample(reflect.ValueOf(example).Elem(), model)
			break
		}
	case reflect.Interface,
		reflect.Struct,
		reflect.Array,
		reflect.Slice,
		reflect.Bool,
		reflect.String,
		reflect.Float32, reflect.Float64,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		{
			s.parseExample(reflect.ValueOf(example), model)
			break
		}
	default:
		return nil
	}

	return model
}

func (s *argument) parseExample(v reflect.Value, a *argument) {
	if a == nil {
		return
	}
	if v.Kind() == reflect.Invalid {
		return
	}

	t := v.Type()
	k := t.Kind()
	switch k {
	case reflect.Ptr:
		{
			if a.Type == "" {
				a.Type = t.String()
			}
			s.parseExample(v.Elem(), a)
			break
		}
	case reflect.Interface:
		{
			if a.Type == "" {
				a.Type = k.String()
			}
			if v.CanInterface() {
				value := reflect.ValueOf(v.Interface())
				if value.Kind() != reflect.Invalid {
					s.parseExample(value, a)
				}
			}
			break
		}
	case reflect.Struct:
		{
			if a.Type == "" {
				a.Type = t.Name()
			}

			n := v.NumField()
			for i := 0; i < n; i++ {
				valueField := v.Field(i)
				if !valueField.CanInterface() {
					continue
				}

				typeField := t.Field(i)
				if typeField.Anonymous {
					if valueField.CanAddr() {
						s.parseExample(valueField.Addr().Elem(), a)
					} else {
						s.parseExample(valueField, a)
					}
				} else {
					child := &argument{Children: make([]*argument, 0)}
					child.Name = typeField.Tag.Get(tagJson)
					if child.Name == "" {
						child.Name = typeField.Name
					}
					cns := strings.Split(child.Name, ",")
					if len(cns) > 1 {
						child.Name = cns[0]
					}
					child.Type = valueField.Kind().String()
					//child.Type = valueField.Type().String()
					if typeField.Tag.Get(tagRequired) == "true" {
						child.Required = true
					}
					child.Note = typeField.Tag.Get(tagNote)
					child.Parent = a
					a.Children = append(a.Children, child)

					value := reflect.ValueOf(valueField.Interface())
					if value.Kind() != reflect.Invalid {
						child.Type = value.Type().Name()
						s.parseExample(value, child)
						if child.Type == "" {
							child.Type = valueField.Type().String()
						}
					}
				}
			}
			break
		}
	case reflect.Array:
		{
			break
		}
	case reflect.Slice:
		{
			st := t.Elem()
			stk := st.Kind()

			var ste reflect.Type = nil
			if stk == reflect.Ptr {
				ste = st.Elem()
			} else {
				ste = st
			}
			if ste != nil {
				a.Type = fmt.Sprintf("%s[]", ste.Name())

				if ste.Kind() == reflect.Struct && a.parentType() != ste.Name() {
					stet := reflect.New(ste)
					child := &argument{Children: make([]*argument, 0)}
					child.Type = ste.Name()
					a.Children = append(a.Children, child)
					s.parseExample(stet.Elem(), child)
				}
			} else {
				a.Type = fmt.Sprintf("%s[]", stk.String())
			}

			if a.Type == "[]" {
				a.Type = fmt.Sprintf("%s[]", stk.String())
			}

			break
		}
	case reflect.Bool,
		reflect.String,
		reflect.Float32, reflect.Float64,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		{
			a.Type = t.Name()
			break
		}
	default:
		{
			return
		}
	}
}

func (s *argument) parentType() string {
	if s.Parent == nil {
		return ""
	}

	return s.Parent.Type
}

func (s *argument) toType(ts map[string]*Type) {
	if ts == nil {
		return
	}
	n := len(s.Children)
	if n <= 0 {
		return
	}
	tn := strings.ReplaceAll(strings.ReplaceAll(s.Type, "*", ""), "[]", "")
	if _, ok := ts[tn]; ok {
		return
	}

	t := &Type{
		Name:   tn,
		Fields: make([]*Model, 0),
		index:  len(ts),
	}
	ts[tn] = t

	for i := 0; i < n; i++ {
		child := s.Children[i]
		if child == nil {
			continue
		}

		t.Fields = append(t.Fields, &Model{
			Name:     child.Name,
			Type:     child.Type,
			Note:     child.Note,
			Required: child.Required,
		})
		child.toType(ts)
	}

}
