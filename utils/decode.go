package utils

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

var TimeFormats = []string{"1/2/2006", "1/2/2006 15:4:5", "2006-1-2 15:4:5", "2006-1-2 15:4", "2006-1-2", "1-2", "15:4:5", "15:4", "15", "15:4:5 Jan 2, 2006 MST"}

// Decoder is the interface that wraps the basic Read method.
type Decoder interface {
	// Decode data type conversion
	Decode(interface{}) error
}

// NewDecoder Get Decoder interface
func NewDecoder(val interface{}) Decoder {
	return &decoder{input: val}
}

type decoder struct {
	input interface{}
}

func (d *decoder) error(errInfo string) error {
	return fmt.Errorf("[go-dal:utils:decode]%s", errInfo)
}

func (d *decoder) Decode(output interface{}) error {
	outputValue := reflect.Indirect(reflect.ValueOf(output))
	if !outputValue.CanAddr() {
		return d.error("output must be addressable (a pointer)")
	}
	return d.decode(d.input, outputValue)
}

func (d *decoder) getKind(val reflect.Value) reflect.Kind {
	kind := val.Kind()

	switch {
	case kind >= reflect.Int && kind <= reflect.Int64:
		return reflect.Int
	case kind >= reflect.Uint && kind <= reflect.Uint64:
		return reflect.Uint
	case kind >= reflect.Float32 && kind <= reflect.Float64:
		return reflect.Float32
	default:
		return kind
	}
}

func (d *decoder) decode(data interface{}, outputValue reflect.Value) (err error) {
	dataVal := reflect.Indirect(reflect.ValueOf(data))
	if !dataVal.IsValid() {
		outputValue.Set(reflect.Zero(outputValue.Type()))
		return
	}
	switch outputKind := d.getKind(outputValue); outputKind {
	case reflect.Bool:
		err = d.decodeBool(data, outputValue)
	case reflect.String:
		err = d.decodeString(data, outputValue)
	case reflect.Int:
		err = d.decodeInt(data, outputValue)
	case reflect.Uint:
		err = d.decodeUint(data, outputValue)
	case reflect.Float32:
		err = d.decodeFloat(data, outputValue)
	case reflect.Struct:
		switch outputType := outputValue.Type().String(); outputType {
		case "time.Time":
			err = d.decodeTime(data, outputValue)
		default:
			err = d.decodeStruct(data, outputValue)
		}
	case reflect.Map:
		err = d.decodeMap(data, outputValue)
	case reflect.Slice:
		err = d.decodeSlice(data, outputValue)
	case reflect.Interface:
		err = d.decodeBasic(data, outputValue)
	default:
		err = fmt.Errorf("Unsupported type: %s", outputKind)
	}
	return
}

func (d *decoder) decodeString(data interface{}, val reflect.Value) error {
	dataVal := reflect.ValueOf(data)
	dataKind := d.getKind(dataVal)
	switch {
	case dataKind == reflect.String:
		val.SetString(dataVal.String())
	case dataKind == reflect.Bool:
		if dataVal.Bool() {
			val.SetString("1")
		} else {
			val.SetString("0")
		}
	case dataKind == reflect.Int:
		val.SetString(strconv.FormatInt(dataVal.Int(), 10))
	case dataKind == reflect.Uint:
		val.SetString(strconv.FormatUint(dataVal.Uint(), 10))
	case dataKind == reflect.Float32:
		val.SetString(strconv.FormatFloat(dataVal.Float(), 'f', -1, 64))
	default:
		return fmt.Errorf("expected type '%s', got unconvertible type '%s'", val.Type(), dataVal.Type())
	}

	return nil
}

func (d *decoder) decodeInt(data interface{}, val reflect.Value) error {
	dataVal := reflect.ValueOf(data)
	dataKind := d.getKind(dataVal)

	switch {
	case dataKind == reflect.Int:
		val.SetInt(dataVal.Int())
	case dataKind == reflect.Uint:
		val.SetInt(int64(dataVal.Uint()))
	case dataKind == reflect.Float32:
		val.SetInt(int64(dataVal.Float()))
	case dataKind == reflect.Bool:
		if dataVal.Bool() {
			val.SetInt(1)
		} else {
			val.SetInt(0)
		}
	case dataKind == reflect.String:
		dVal := dataVal.String()
		if dVal == "" {
			dVal = "0"
		}
		i, err := strconv.ParseInt(dVal, 10, 64)
		if err == nil {
			val.SetInt(i)
		} else {
			return err
		}
	default:
		return fmt.Errorf("expected type '%s', got unconvertible type '%s'", val.Type(), dataVal.Type())
	}

	return nil
}

func (d *decoder) decodeUint(data interface{}, val reflect.Value) error {
	dataVal := reflect.ValueOf(data)
	dataKind := d.getKind(dataVal)

	switch {
	case dataKind == reflect.Int:
		val.SetUint(uint64(dataVal.Int()))
	case dataKind == reflect.Uint:
		val.SetUint(dataVal.Uint())
	case dataKind == reflect.Float32:
		val.SetUint(uint64(dataVal.Float()))
	case dataKind == reflect.Bool:
		if dataVal.Bool() {
			val.SetUint(1)
		} else {
			val.SetUint(0)
		}
	case dataKind == reflect.String:
		dVal := dataVal.String()
		if dVal == "" {
			dVal = "0"
		}
		i, err := strconv.ParseUint(dVal, 10, 64)
		if err == nil {
			val.SetUint(i)
		} else {
			return err
		}
	default:
		return fmt.Errorf("expected type '%s', got unconvertible type '%s'", val.Type(), dataVal.Type())
	}

	return nil
}

func (d *decoder) decodeBool(data interface{}, val reflect.Value) error {
	dataVal := reflect.ValueOf(data)
	dataKind := d.getKind(dataVal)

	switch {
	case dataKind == reflect.Bool:
		val.SetBool(dataVal.Bool())
	case dataKind == reflect.Int:
		val.SetBool(dataVal.Int() != 0)
	case dataKind == reflect.Uint:
		val.SetBool(dataVal.Uint() != 0)
	case dataKind == reflect.Float32:
		val.SetBool(dataVal.Float() != 0)
	case dataKind == reflect.String:
		b, err := strconv.ParseBool(dataVal.String())
		if err == nil {
			val.SetBool(b)
		} else if dataVal.String() == "" {
			val.SetBool(false)
		} else {
			return err
		}
	default:
		return fmt.Errorf("expected type '%s', got unconvertible type '%s'", val.Type(), dataVal.Type())
	}

	return nil
}

func (d *decoder) decodeFloat(data interface{}, val reflect.Value) error {
	dataVal := reflect.ValueOf(data)
	dataKind := d.getKind(dataVal)

	switch {
	case dataKind == reflect.Int:
		val.SetFloat(float64(dataVal.Int()))
	case dataKind == reflect.Uint:
		val.SetFloat(float64(dataVal.Uint()))
	case dataKind == reflect.Float32:
		val.SetFloat(float64(dataVal.Float()))
	case dataKind == reflect.Bool:
		if dataVal.Bool() {
			val.SetFloat(1)
		} else {
			val.SetFloat(0)
		}
	case dataKind == reflect.String:
		dVal := dataVal.String()
		if dVal == "" {
			dVal = "0"
		}
		f, err := strconv.ParseFloat(dVal, 64)
		if err == nil {
			val.SetFloat(f)
		} else {
			return err
		}
	default:
		return fmt.Errorf("expected type '%s', got unconvertible type '%s'", val.Type(), dataVal.Type())
	}

	return nil
}

func (d *decoder) decodeMap(data interface{}, val reflect.Value) error {
	valType := val.Type()
	valKeyType := valType.Key()
	valElemType := valType.Elem()

	valMap := val
	if val.IsNil() {
		mapType := reflect.MapOf(valKeyType, valElemType)
		valMap = reflect.MakeMap(mapType)
	}

	dataVal := reflect.Indirect(reflect.ValueOf(data))

	switch {
	case dataVal.Kind() == reflect.Map:
		for _, dataKey := range dataVal.MapKeys() {
			currentKey := reflect.Indirect(reflect.New(valKeyType))
			if err := d.decode(dataKey.Interface(), currentKey); err != nil {
				return err
			}
			currentValue := reflect.Indirect(reflect.New(valElemType))
			if err := d.decode(dataVal.MapIndex(dataKey).Interface(), currentValue); err != nil {
				return err
			}
			valMap.SetMapIndex(currentKey, currentValue)
		}
	case dataVal.Kind() == reflect.Struct:
		dataType := reflect.TypeOf(data)
		for i, l := 0, dataType.NumField(); i < l; i++ {
			field := dataType.Field(i)
			fieldValue := dataVal.FieldByName(field.Name).Interface()
			if reflect.DeepEqual(fieldValue, reflect.Zero(field.Type).Interface()) {
				// fieldValue = reflect.Zero(field.Type).Interface()
				continue
			}
			if field.Type.String() == "time.Time" && valElemType.Kind() == reflect.String {
				if !reflect.DeepEqual(reflect.Zero(field.Type).Interface(), fieldValue) {
					val.SetMapIndex(reflect.ValueOf(field.Name), reflect.ValueOf(fieldValue.(time.Time).Format(time.RFC3339Nano)))
				}
				continue
			}
			currentValue := reflect.Indirect(reflect.New(valElemType))
			if err := d.decode(fieldValue, currentValue); err != nil {
				return err
			}
			valMap.SetMapIndex(reflect.ValueOf(field.Name), currentValue)
		}
	default:
		return fmt.Errorf("expected type '%s', got unconvertible type '%s'", val.Type(), dataVal.Type())
	}

	val.Set(valMap)

	return nil
}

func (d *decoder) decodeSlice(data interface{}, val reflect.Value) error {
	dataVal := reflect.Indirect(reflect.ValueOf(data))
	if dataVal.Kind() != reflect.Slice {
		return fmt.Errorf("Expected type slice")
	}
	if dataVal.Type() == val.Type() {
		val.Set(dataVal)
		return nil
	}
	valSlice := reflect.MakeSlice(reflect.SliceOf(val.Type().Elem()), dataVal.Len(), dataVal.Len())
	for i, l := 0, dataVal.Len(); i < l; i++ {
		currentData := dataVal.Index(i).Interface()
		currentField := valSlice.Index(i)
		if err := d.decode(currentData, currentField); err != nil {
			return err
		}
	}
	val.Set(valSlice)

	return nil
}

func (d *decoder) decodeStruct(data interface{}, val reflect.Value) error {
	dataVal := reflect.Indirect(reflect.ValueOf(data))
	valType := val.Type()
	if dataVal.Type() == valType {
		val.Set(dataVal)
		return nil
	}
	if kind := dataVal.Kind(); kind != reflect.Map {
		return fmt.Errorf("Expected a map, got '%s'", kind.String())
	}
	for i, l := 0, valType.NumField(); i < l; i++ {
		fieldName := valType.Field(i).Name
		rawMapKey := reflect.ValueOf(fieldName)
		rawMapValue := dataVal.MapIndex(rawMapKey)
		if !rawMapValue.IsValid() {
			dataValKeys := dataVal.MapKeys()
			for j, jl := 0, len(dataValKeys); j < jl; j++ {
				rawMapKeyName, ok := dataValKeys[j].Interface().(string)
				if !ok {
					continue
				}
				if strings.EqualFold(fieldName, rawMapKeyName) {
					rawMapKey = dataValKeys[j]
					rawMapValue = dataVal.MapIndex(dataValKeys[j])
					break
				}
			}
			if !rawMapValue.IsValid() {
				continue
			}
		}
		field := val.Field(i)
		if !field.CanSet() {
			continue
		}
		if err := d.decode(rawMapValue.Interface(), field); err != nil {
			return err
		}
	}
	return nil
}

func (d *decoder) decodeTime(data interface{}, val reflect.Value) error {
	var tVal time.Time
	if v, ok := data.(string); ok && v != "" {
		var exist bool
		for i, l := 0, len(TimeFormats); i < l; i++ {
			t, err := time.Parse(TimeFormats[i], v)
			if err == nil {
				tVal = t
				exist = true
				break
			}
		}
		if !exist {
			return fmt.Errorf("Unknown time format.")
		}
	} else if v, ok := data.(time.Time); ok {
		tVal = v
	} else {
		tVal = time.Now()
	}
	val.Set(reflect.ValueOf(tVal))
	return nil
}

func (d *decoder) decodeBasic(data interface{}, val reflect.Value) error {
	dataVal := reflect.ValueOf(data)
	dataValType := dataVal.Type()
	if !dataValType.AssignableTo(val.Type()) {
		return fmt.Errorf("expected type '%s', got unconvertible type '%s'", val.Type(), dataValType)
	}

	val.Set(dataVal)
	return nil
}
