package util

import (
	"slices"
)

// AttachmentsMap is a special data structure that allows attaching arrays of strings
// within packages to types, methods, functions, and fields
type AttachmentsMap struct {
	packageAttachments map[string]PackageAttachments
}

// HasPkgAttachment checks if package has the attachment
func (t *AttachmentsMap) HasPkgAttachment(pkg string, attachment string) bool {
	if t == nil {
		return false
	}

	found, ok := t.packageAttachments[pkg]
	return ok && found.HasAttachment(attachment)
}

// AddPkgAttachment adds an attachment to package
func (t *AttachmentsMap) AddPkgAttachment(pkg string, attachment string) {
	if t.packageAttachments == nil {
		t.packageAttachments = make(map[string]PackageAttachments)
	}

	found, ok := t.packageAttachments[pkg]
	if !ok {
		found = PackageAttachments{}
		t.packageAttachments[pkg] = found
	}
	found.AddAttachment(attachment)
	t.packageAttachments[pkg] = found
}

// HasPkgFunctionAttachment checks if package function has the attachment
func (t *AttachmentsMap) HasPkgFunctionAttachment(pkg string, funcname string, attachment string) bool {
	if t == nil {
		return false
	}

	found, ok := t.packageAttachments[pkg]
	return ok && found.HasFunctionAttachment(funcname, attachment)
}

// AddPkgFunctionAttachment adds an attachment to package function
func (t *AttachmentsMap) AddPkgFunctionAttachment(pkg string, funcname string, attachment string) {
	if t.packageAttachments == nil {
		t.packageAttachments = make(map[string]PackageAttachments)
	}

	found, ok := t.packageAttachments[pkg]
	if !ok {
		found = PackageAttachments{}
		t.packageAttachments[pkg] = found
	}
	found.AddFunctionAttachment(funcname, attachment)
	t.packageAttachments[pkg] = found
}

// HasPkgTypeAttachment checks if package type has the attachment
func (t *AttachmentsMap) HasPkgTypeAttachment(pkg string, typename string, attachment string) bool {
	if t == nil {
		return false
	}

	found, ok := t.packageAttachments[pkg]
	return ok && found.HasTypeAttachment(typename, attachment)
}

// AddPkgTypeAttachment adds an attachment to package type
func (t *AttachmentsMap) AddPkgTypeAttachment(pkg string, typename string, attachment string) {
	if t.packageAttachments == nil {
		t.packageAttachments = make(map[string]PackageAttachments)
	}

	found, ok := t.packageAttachments[pkg]
	if !ok {
		found = PackageAttachments{}
		t.packageAttachments[pkg] = found
	}
	found.AddTypeAttachment(typename, attachment)
	t.packageAttachments[pkg] = found
}

// HasPkgTypeFieldAttachment checks if package type field has the attachment
func (t *AttachmentsMap) HasPkgTypeFieldAttachment(pkg string, typename string, field string, attachment string) bool {
	if t == nil {
		return false
	}

	found, ok := t.packageAttachments[pkg]
	return ok && found.HasTypeFieldAttachment(typename, field, attachment)
}

// AddPkgTypeFieldAttachment adds an attachment to package type field
func (t *AttachmentsMap) AddPkgTypeFieldAttachment(pkg string, typename string, field string, attachment string) {
	if t.packageAttachments == nil {
		t.packageAttachments = make(map[string]PackageAttachments)
	}

	found, ok := t.packageAttachments[pkg]
	if !ok {
		found = PackageAttachments{}
		t.packageAttachments[pkg] = found
	}
	found.AddTypeFieldAttachment(typename, field, attachment)
	t.packageAttachments[pkg] = found
}

// HasPkgTypeMethodAttachment checks if package type method has the attachment
func (t *AttachmentsMap) HasPkgTypeMethodAttachment(pkg string, typename string, method string, attachment string) bool {
	if t == nil {
		return false
	}

	found, ok := t.packageAttachments[pkg]
	return ok && found.HasTypeMethodAttachment(typename, method, attachment)
}

// AddPkgTypeMethodAttachment adds an attachment to package type method
func (t *AttachmentsMap) AddPkgTypeMethodAttachment(pkg string, typename string, method string, attachment string) {
	if t.packageAttachments == nil {
		t.packageAttachments = make(map[string]PackageAttachments)
	}

	found, ok := t.packageAttachments[pkg]
	if !ok {
		found = PackageAttachments{}
		t.packageAttachments[pkg] = found
	}
	found.AddTypeMethodAttachment(typename, method, attachment)
	t.packageAttachments[pkg] = found
}

// PackageAttachments stores all attachments within a single package
type PackageAttachments struct {
	LocalAttachments     []string
	FunctionsAttachments map[string][]string
	TypesAttachments     map[string]TypeAttachments
}

// HasAttachment checks if package has the attachment
func (t *PackageAttachments) HasAttachment(attachment string) bool {
	if t == nil {
		return false
	}

	return slices.Contains(t.LocalAttachments, attachment)
}

// AddAttachment adds an attachment to package
func (t *PackageAttachments) AddAttachment(attachment string) {
	t.LocalAttachments = append(t.LocalAttachments, attachment)
}

// HasFunctionAttachment checks if function has the attachment
func (t *PackageAttachments) HasFunctionAttachment(funcname string, attachment string) bool {
	if t == nil {
		return false
	}

	if t.FunctionsAttachments == nil {
		return false
	}

	attachments, ok := t.FunctionsAttachments[funcname]

	return ok && slices.Contains(attachments, attachment)
}

// AddFunctionAttachment adds an attachment to function
func (t *PackageAttachments) AddFunctionAttachment(funcname string, attachment string) {
	if t.FunctionsAttachments == nil {
		t.FunctionsAttachments = make(map[string][]string)
	}

	t.FunctionsAttachments[funcname] = append(t.FunctionsAttachments[funcname], attachment)
}

// HasTypeAttachment checks if type has the attachment
func (t *PackageAttachments) HasTypeAttachment(typename string, attachment string) bool {
	if t == nil {
		return false
	}

	if t.TypesAttachments == nil {
		return false
	}

	attachments, ok := t.TypesAttachments[typename]

	return ok && attachments.HasAttachment(attachment)
}

// AddTypeAttachment adds an attachment to type
func (t *PackageAttachments) AddTypeAttachment(typename string, attachment string) {
	if t.TypesAttachments == nil {
		t.TypesAttachments = make(map[string]TypeAttachments)
	}

	found, ok := t.TypesAttachments[typename]
	if !ok {
		found = TypeAttachments{}
		t.TypesAttachments[typename] = found
	}
	found.AddAttachment(attachment)
	t.TypesAttachments[typename] = found
}

// HasTypeFieldAttachment checks if type field has the attachment
func (t *PackageAttachments) HasTypeFieldAttachment(typename string, field string, attachment string) bool {
	if t == nil {
		return false
	}

	if t.TypesAttachments == nil {
		return false
	}

	attachments, ok := t.TypesAttachments[typename]

	return ok && attachments.HasFieldAttachment(field, attachment)
}

// AddTypeFieldAttachment adds an attachment to type field
func (t *PackageAttachments) AddTypeFieldAttachment(typename string, field string, attachment string) {
	if t.TypesAttachments == nil {
		t.TypesAttachments = make(map[string]TypeAttachments)
	}

	found, ok := t.TypesAttachments[typename]
	if !ok {
		found = TypeAttachments{}
		t.TypesAttachments[typename] = found
	}
	found.AddFieldAttachment(field, attachment)
	t.TypesAttachments[typename] = found
}

// HasTypeMethodAttachment checks if type method has the attachment
func (t *PackageAttachments) HasTypeMethodAttachment(typename string, method string, attachment string) bool {
	if t == nil {
		return false
	}

	if t.TypesAttachments == nil {
		return false
	}

	attachments, ok := t.TypesAttachments[typename]

	return ok && attachments.HasMethodAttachment(method, attachment)
}

// AddTypeMethodAttachment adds an attachment to type method
func (t *PackageAttachments) AddTypeMethodAttachment(typename string, method string, attachment string) {
	if t.TypesAttachments == nil {
		t.TypesAttachments = make(map[string]TypeAttachments)
	}

	found, ok := t.TypesAttachments[typename]
	if !ok {
		found = TypeAttachments{}
		t.TypesAttachments[typename] = found
	}
	found.AddMethodAttachment(method, attachment)
	t.TypesAttachments[typename] = found
}

// TypeAttachments stores all attachments for a specific type
type TypeAttachments struct {
	LocalAttachments   []string
	FieldsAttachments  map[string][]string
	MethodsAttachments map[string][]string
}

// HasAttachment checks if type has the attachment
func (t *TypeAttachments) HasAttachment(attachment string) bool {
	if t == nil {
		return false
	}
	return slices.Contains(t.LocalAttachments, attachment)
}

// AddAttachment adds an attachment to type
func (t *TypeAttachments) AddAttachment(attachment string) {
	t.LocalAttachments = append(t.LocalAttachments, attachment)
}

// HasFieldAttachment checks if field has the attachment
func (t *TypeAttachments) HasFieldAttachment(field string, attachment string) bool {
	if t == nil {
		return false
	}
	if t.FieldsAttachments == nil {
		return false
	}
	found, ok := t.FieldsAttachments[field]
	return ok && slices.Contains(found, attachment)
}

// AddFieldAttachment adds an attachment to field
func (t *TypeAttachments) AddFieldAttachment(field string, attachment string) {
	if t.FieldsAttachments == nil {
		t.FieldsAttachments = make(map[string][]string)
	}

	t.FieldsAttachments[field] = append(t.FieldsAttachments[field], attachment)
}

// HasMethodAttachment checks if method has the attachment
func (t *TypeAttachments) HasMethodAttachment(method string, attachment string) bool {
	if t == nil {
		return false
	}
	if t.MethodsAttachments == nil {
		return false
	}
	found, ok := t.MethodsAttachments[method]
	return ok && slices.Contains(found, attachment)
}

// AddMethodAttachment adds an attachment to method
func (t *TypeAttachments) AddMethodAttachment(method string, attachment string) {
	if t.MethodsAttachments == nil {
		t.MethodsAttachments = make(map[string][]string)
	}

	t.MethodsAttachments[method] = append(t.MethodsAttachments[method], attachment)
}

// GetPackageAttachments returns PackageAttachments for a given package path
func (am *AttachmentsMap) GetPackageAttachments(pkgPath string) *PackageAttachments {
	if am == nil {
		return nil
	}

	if pkg, exists := am.packageAttachments[pkgPath]; exists {
		return &pkg
	}
	return nil
}

// GetTypeAttachments returns TypeAttachments for a given type in a package
func (pa *PackageAttachments) GetTypeAttachments(typeName string) *TypeAttachments {
	if pa == nil || pa.TypesAttachments == nil {
		return nil
	}

	if typeAnn, exists := pa.TypesAttachments[typeName]; exists {
		return &typeAnn
	}
	return nil
}

// GetAttachmentsForType returns all attachments for a type, excluding specified values
func (am *AttachmentsMap) GetAttachmentsForType(pkgPath string, typeName string, excludes ...string) []string {
	pkgAttachments := am.GetPackageAttachments(pkgPath)
	if pkgAttachments == nil {
		return nil
	}

	typeAttachments := pkgAttachments.GetTypeAttachments(typeName)
	if typeAttachments == nil {
		return nil
	}

	var attachments []string
	for _, attachment := range typeAttachments.LocalAttachments {
		// Skip excluded values
		if !slices.Contains(excludes, attachment) {
			attachments = append(attachments, attachment)
		}
	}

	return attachments
}

// GetAttachmentsForFunction returns all attachments for a function, excluding specified values
func (am *AttachmentsMap) GetAttachmentsForFunction(pkgPath string, funcName string, excludes ...string) []string {
	pkgAttachments := am.GetPackageAttachments(pkgPath)
	if pkgAttachments == nil || pkgAttachments.FunctionsAttachments == nil {
		return nil
	}

	attachments, exists := pkgAttachments.FunctionsAttachments[funcName]
	if !exists {
		return nil
	}

	var result []string
	for _, attachment := range attachments {
		// Skip excluded values
		if !slices.Contains(excludes, attachment) {
			result = append(result, attachment)
		}
	}

	return result
}

// GetAttachmentsForMethod returns all attachments for a method, excluding specified values
func (am *AttachmentsMap) GetAttachmentsForMethod(pkgPath string, typeName string, methodName string, excludes ...string) []string {
	pkgAttachments := am.GetPackageAttachments(pkgPath)
	if pkgAttachments == nil {
		return nil
	}

	typeAttachments := pkgAttachments.GetTypeAttachments(typeName)
	if typeAttachments == nil || typeAttachments.MethodsAttachments == nil {
		return nil
	}

	attachments, exists := typeAttachments.MethodsAttachments[methodName]
	if !exists {
		return nil
	}

	var result []string
	for _, attachment := range attachments {
		// Skip excluded values
		if !slices.Contains(excludes, attachment) {
			result = append(result, attachment)
		}
	}

	return result
}

// HasAnyTypeAttachments checks if a type has any attachments at all
func (am *AttachmentsMap) HasAnyTypeAttachments(pkgPath string, typeName string) bool {
	if am == nil {
		return false
	}

	pkgAttachments := am.GetPackageAttachments(pkgPath)
	if pkgAttachments == nil {
		return false
	}

	typeAttachments := pkgAttachments.GetTypeAttachments(typeName)
	if typeAttachments == nil {
		return false
	}

	return len(typeAttachments.LocalAttachments) > 0
}

// HasAnyFunctionAttachments checks if a function has any attachments at all
func (am *AttachmentsMap) HasAnyFunctionAttachments(pkgPath string, funcName string) bool {
	if am == nil {
		return false
	}

	pkgAttachments := am.GetPackageAttachments(pkgPath)
	if pkgAttachments == nil || pkgAttachments.FunctionsAttachments == nil {
		return false
	}

	attachments, exists := pkgAttachments.FunctionsAttachments[funcName]
	return exists && len(attachments) > 0
}

// HasAnyMethodAttachments checks if a method has any attachments at all
func (am *AttachmentsMap) HasAnyMethodAttachments(pkgPath string, typeName string, methodName string) bool {
	if am == nil {
		return false
	}

	pkgAttachments := am.GetPackageAttachments(pkgPath)
	if pkgAttachments == nil {
		return false
	}

	typeAttachments := pkgAttachments.GetTypeAttachments(typeName)
	if typeAttachments == nil || typeAttachments.MethodsAttachments == nil {
		return false
	}

	attachments, exists := typeAttachments.MethodsAttachments[methodName]
	return exists && len(attachments) > 0
}
