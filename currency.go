package currency

import "strings"

type Currency struct {
	Alpha3  string
	Numeric string
	Symbol  string

	//Country string
	//BankURL string

	Name     string
	AltNames []string
	Tags     []string
	Accuracy int

	Users []string
}

func Get(v string) *Currency {
	for _, c := range List {
		if c.Alpha3 == v || c.Numeric == v {
			return &c
		}
	}

	return nil
}

func ByName(v string) *Currency {
	v = strings.ToLower(v)

	for _, c := range List {
		if strings.ToLower(c.Name) == v {
			return &c
		}

		for _, n := range c.AltNames {
			if strings.ToLower(n) == v {
				return &c
			}
		}

		for _, n := range c.Tags {
			if strings.ToLower(n) == v {
				return &c
			}
		}
	}

	return nil
}
