package danskebank

// Account represents a danskebank account
type Account struct {
	SortValue int

	AccessToCredit             bool
	AccessToDebit              bool
	AccessToQuery              bool
	IsBreadcrumbAccountProduct bool
	IsFixedTermDeposit         bool
	IsInLimitGroup             bool
	IsJisaAccountProduct       bool
	IsLoanAccount              bool
	IsSavingGoalAccountProduct bool
	ShowAvailable              bool

	AccountName     string
	AccountNoExt    string
	AccountNoInt    string
	AccountProduct  string
	AccountRegNoExt string
	AccountType     string
	CardType        string
	Currency        string
	InvIDOwner      string
	MandateAccMk    string
	ShowCategory    string

	// LanguageCode - not entirely sure what datatype this is, as i have only seen it as null
	// when not null, its properly just a string with something like "DA" in it
	LanguageCode string

	Balance          float64
	BalanceAvailable float64
}

// AccountListResponse is the struct for accountlist responses
type AccountListResponse struct {
	Accounts []Account

	// These two strings seems to always be null
	EupToken        string
	ResponseMessage string

	LanguageCode string
	LastUpdated  string
	ResponseCode int
}
