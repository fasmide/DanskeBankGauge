package danskebank

// LogonPackage represents the result of the javascript sealer
type LogonPackage struct {
	UserID       string
	LogonPackage string
}

// Valid just checks struct for payload
func (l *LogonPackage) Valid() bool {
	if l.UserID == "" {
		return false
	}
	if l.LogonPackage == "" {
		return false
	}
	return true
}
