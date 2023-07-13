package misc

import "reflect"

func IsNil(obj interface{}) bool {
	if obj == nil {
		return true
	}

	objValue := reflect.ValueOf(obj)
	// проверяем, что тип значения ссылочный, то есть в принципе может быть равен nil
	if objValue.Kind() != reflect.Ptr {
		return false
	}
	// проверяем, что значение равно nil
	//  важно, что IsNil() вызывает панику, если value не является ссылочным типом. Поэтому всегда проверяйте на Kind()
	if objValue.IsNil() {
		return true
	}

	return false
}
