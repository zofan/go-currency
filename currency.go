package currency

type Currency struct {
	Alpha3  string
	Numeric string
	Symbol  string

	//Country string
	//BankURL string

	Name string
	//ShortName string
	Accuracy int

	Users []string
}
