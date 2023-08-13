package retry

type AttributeAccessorSupport interface {
	SetAttribute(name string, value interface{})
	GetAttribute(name string) interface{}
	RemoveAttribute(name string) interface{}
	HasAttribute(name string) bool
	AttributeName() []string
}

// SimpleAttributeAccessorSupport 属性访问方法
type SimpleAttributeAccessorSupport struct {
	// 用于存储属性对象的map
	attributes map[string]interface{}
}

func (a *SimpleAttributeAccessorSupport) SetAttribute(name string, value interface{}) {
	if a.attributes == nil {
		a.attributes = make(map[string]interface{})
	}
	if value != nil {
		a.attributes[name] = value
	} else {
		a.RemoveAttribute(name)
	}
}

func (a *SimpleAttributeAccessorSupport) GetAttribute(name string) interface{} {
	if a.attributes == nil {
		return nil
	}
	return a.attributes[name]
}

func (a *SimpleAttributeAccessorSupport) RemoveAttribute(name string) interface{} {
	if a.attributes == nil {
		return nil
	}
	var prev = a.attributes[name]
	a.attributes[name] = nil
	return prev
}

func (a *SimpleAttributeAccessorSupport) HasAttribute(name string) bool {
	if a.attributes == nil {
		return false
	}
	return a.attributes[name] != nil
}

func (a *SimpleAttributeAccessorSupport) AttributeName() []string {
	if a.attributes == nil {
		return nil
	}
	var attributes []string
	for k := range a.attributes {
		attributes = append(attributes, k)
	}
	return attributes
}
