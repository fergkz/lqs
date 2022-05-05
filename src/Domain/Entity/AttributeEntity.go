package DomainEntity

type AttributeEntity struct {
	Key   string
	Value interface{}
}

func NewAttribute(
	Key string,
	Value interface{},
) (attribute *AttributeEntity) {
	attribute.Key = Key
	attribute.Value = Value
	return attribute
}
