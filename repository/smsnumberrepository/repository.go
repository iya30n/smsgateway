package smsnumberrepository

type SMSNumberRepository interface {
	// IsActive reports whether the number exists and is currently usable as a
	// source number. A number that is missing is reported as inactive, so the
	// caller does not have to tell the two cases apart.
	IsActive(number string) (bool, error)
}
