package modA // want package:"package modA"

// @packageonly modB
type InternalType struct {
	Value string
}

// @packageonly multimodule_packageonly/modB
func NewInternalType(value string) *InternalType {
	return &InternalType{Value: value}
}

// @packageonly modB
func (it *InternalType) Process() string {
	return "processed: " + it.Value
}

// @packageonly multimodule_packageonly/modB
var InternalGlobal = "internal"

// Regular type that can be used anywhere
type PublicType struct {
	Data string
}

// @packageonly modC
type RestrictedType struct {
	Secret string
}

// @packageonly modC
func NewRestrictedType(secret string) *RestrictedType {
	return &RestrictedType{Secret: secret}
}

// @packageonly modC
func (rt *RestrictedType) GetSecret() string {
	return rt.Secret
}

// Regular function that can be used anywhere
func NewPublicType(data string) *PublicType {
	return &PublicType{Data: data}
}
