package smsnumberrepository

type SMSNumberRepository interface {
	IsActive(number string) (bool, error)
	GetActiveNumberByOperator(operatorID uint) (string, bool, error)
}
