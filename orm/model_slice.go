package orm

import (
	"reflect"

	"github.com/go-pg/pg/internal"
	"github.com/go-pg/pg/types"
)

type sliceModel struct {
	Discard
	slice    reflect.Value
	nextElem func() reflect.Value
	scan     func(reflect.Value, types.Reader, int) error
	currElem reflect.Value
}

var _ Model = (*sliceModel)(nil)

func newSliceModel(slice reflect.Value, elemType reflect.Type) *sliceModel {
	return &sliceModel{
		slice: slice,
		scan:  types.Scanner(elemType),
	}
}

func (m *sliceModel) Init() error {
	if m.slice.IsValid() && m.slice.Len() > 0 {
		m.slice.Set(m.slice.Slice(0, 0))
	}
	return nil
}

func (m *sliceModel) NewModel() ColumnScanner {
	return m
}

func (m *sliceModel) ScanColumn(colIdx int, colName string, colType uint32, rd types.Reader, n int) error {
	if m.nextElem == nil {
		m.nextElem = internal.MakeSliceNextElemFunc(m.slice)
	}
	if m.slice.Type().Elem().Kind() == reflect.Map && m.slice.Type().Elem().Elem().Kind() == reflect.Interface {
		if !m.currElem.IsValid() || m.currElem.IsNil() {
			m.currElem = m.nextElem()
		}
		currElemMap := m.currElem.Interface().(map[string]interface{})
		if _, ok := currElemMap[colName]; ok {
			m.currElem = m.nextElem()
			currElemMap = m.currElem.Interface().(map[string]interface{})
		}
		var err error
		switch colType {
		case 16:
			if boolString, err := types.ScanString(rd, n); err != nil {
				return err
			} else if len(boolString) == 1 && (boolString[0] == 't' || boolString[0] == '1') {
				currElemMap[colName] = true
			} else {
				currElemMap[colName] = false
			}
		case 20:
			fallthrough
		case 21:
			fallthrough
		case 22:
			fallthrough
		case 23:
			fallthrough
		case 26:
			fallthrough
		case 27:
			fallthrough
		case 28:
			fallthrough
		case 29:
			fallthrough
		case 30:
			fallthrough
		case 700:
			fallthrough
		case 701:
			fallthrough
		case 1186:
			fallthrough
		case 1700:
			currElemMap[colName], err = types.ScanInt64(rd, n)
		default:
			currElemMap[colName], err = types.ScanString(rd, n)
		}
		return err
	}
	v := m.nextElem()
	return m.scan(v, rd, n)
}
