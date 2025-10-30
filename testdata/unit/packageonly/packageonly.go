package packageonly

// PackageOnlyHelper is a helper type that should only be used in testing and testutil packages
// @packageonly testing, testutil
type PackageOnlyHelper struct {
	data string
}

// NewPackageOnlyHelper creates a new PackageOnlyHelper
// This function should only be used in testing package
// @packageonly testing
func NewPackageOnlyHelper(data string) *PackageOnlyHelper {
	return &PackageOnlyHelper{data: data}
}

// Reset resets the helper state
// @packageonly testing, testutil
func (h *PackageOnlyHelper) Reset() {
	h.data = ""
}

// GetData returns the current data
// @packageonly testing
func (h *PackageOnlyHelper) GetData() string {
	return h.data
}

// ProcessPackageOnlyData processes package-only data
// @packageonly testing
func ProcessPackageOnlyData(helper *PackageOnlyHelper) string {
	return helper.GetData()
}

// PublicService is a regular type without package-only restrictions
type PublicService struct {
	name string
}

// NewPublicService creates a new PublicService - no restrictions
func NewPublicService(name string) *PublicService {
	return &PublicService{name: name}
}

// Reset method without package-only restriction
func (s *PublicService) Reset() {
	s.name = ""
}

// GetName returns the service name
func (s *PublicService) GetName() string {
	return s.name
}
