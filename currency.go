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

func ByAlpha3(v string) *Currency {
	for _, c := range List {
		if c.Alpha3 == v {
			return &c
		}
	}

	return nil
}

func ByNumeric(v string) *Currency {
	for _, c := range List {
		if c.Numeric == v {
			return &c
		}
	}

	return nil
}
