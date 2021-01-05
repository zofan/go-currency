package currency

import (
	"fmt"
	"github.com/zofan/go-country"
	"github.com/zofan/go-fwrite"
	"github.com/zofan/go-req"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

func Update() error {
	var (
		httpClient = req.New(req.DefaultConfig)
		list       = make(map[string]*Currency)
	)

	var (
		ccyBlockRe   = regexp.MustCompile(`(?s)<CcyNtry[^>]*>(.*?)</CcyNtry>`)
		ccyNmRe      = regexp.MustCompile(`(?s)<CcyNm[^>]*>([^<]+)</`)
		ccyRe        = regexp.MustCompile(`(?s)<Ccy>([^<]+)</`)
		ccyNbrRe     = regexp.MustCompile(`(?s)<CcyNbr[^>]*>([^<]+)</`)
		ccyMnrUntsRe = regexp.MustCompile(`(?s)<CcyMnrUnts[^>]*>([^<]+)</`)
	)

	resp := httpClient.Get(`https://www.currency-iso.org/dam/downloads/lists/list_one.xml`)
	if resp.Error() != nil {
		return resp.Error()
	}

	body := string(resp.ReadAll())

	for _, ccyBlock := range ccyBlockRe.FindAllStringSubmatch(body, -1) {
		ccyNm := ccyNmRe.FindStringSubmatch(ccyBlock[1])
		ccy := ccyRe.FindStringSubmatch(ccyBlock[1])
		ccyNbr := ccyNbrRe.FindStringSubmatch(ccyBlock[1])
		ccyMnrUnts := ccyMnrUntsRe.FindStringSubmatch(ccyBlock[1])

		if len(ccyMnrUnts) == 0 || strings.Contains(ccyBlock[1], `N.A.`) {
			continue
		}

		acc, _ := strconv.Atoi(ccyMnrUnts[1])

		c := &Currency{
			Alpha3:   strings.TrimSpace(ccy[1]),
			Numeric:  strings.TrimSpace(ccyNbr[1]),
			Name:     strings.TrimSpace(ccyNm[1]),
			Accuracy: acc,
		}

		list[c.Alpha3] = c
	}

	// ---

	resp = httpClient.Get(`https://thefactfile.org/countries-currencies-symbols/`)
	if resp.Error() != nil {
		return resp.Error()
	}

	body = string(resp.ReadAll())
	body = strings.ReplaceAll(body, `&nbsp;`, ` `)

	symbolsRe := regexp.MustCompile(`(?s)<td class="column-4">(\w+)</td><td class="column-5">([^<]+)</td>`)

	for _, s := range symbolsRe.FindAllStringSubmatch(body, -1) {
		if c, ok := list[s[1]]; ok {
			c.Symbol = strings.TrimSpace(s[2])
		}
	}

	// ---

	for _, c := range country.List {
		for _, cc := range c.Currencies {
			if _, ok := list[cc]; ok {
				list[cc].Users = append(list[cc].Users, c.Alpha3)
			}
		}
	}

	// ---

	var tpl []string

	tpl = append(tpl, `package currency`)
	tpl = append(tpl, ``)
	tpl = append(tpl, `// Updated at: `+time.Now().String())
	tpl = append(tpl, `var List = []Currency{`)

	for _, c := range list {
		tpl = append(tpl, `	{`)
		tpl = append(tpl, `		Alpha3:    "`+c.Alpha3+`",`)
		tpl = append(tpl, `		Numeric:   "`+c.Numeric+`",`)
		tpl = append(tpl, `		Symbol:    "`+c.Symbol+`",`)
		tpl = append(tpl, `		Name:      "`+c.Name+`",`)
		//tpl = append(tpl, `		ShortName: "`+c.ShortName+`",`)
		//tpl = append(tpl, `		Country:   "`+c.Country+`",`)
		//tpl = append(tpl, `		BankURL:   "`+c.BankURL+`",`)
		tpl = append(tpl, `		Accuracy:  `+strconv.Itoa(c.Accuracy)+`,`)
		tpl = append(tpl, `		Users:     `+fmt.Sprintf(`%#v`, c.Users)+`,`)
		tpl = append(tpl, `	},`)
	}

	tpl = append(tpl, `}`)
	tpl = append(tpl, ``)

	_, file, _, _ := runtime.Caller(0)
	dir := filepath.Dir(file)

	return fwrite.WriteRaw(dir+`/currency_db.go`, []byte(strings.Join(tpl, "\n")))
}
