package currency

import (
	"fmt"
	"github.com/zofan/go-country"
	"github.com/zofan/go-fwrite"
	"github.com/zofan/go-req"
	"github.com/zofan/go-xmlre"
	"golang.org/x/net/html"
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
		ccyBlockRe   = xmlre.Compile(`<CcyNtry>(.*?)</CcyNtry>`)
		ccyNmRe      = xmlre.Compile(`<CcyNm>([^<]+)</`)
		ccyRe        = xmlre.Compile(`<Ccy>([^<]+)</`)
		ccyNbrRe     = xmlre.Compile(`<CcyNbr>([^<]+)</`)
		ccyMnrUntsRe = xmlre.Compile(`<CcyMnrUnts>([^<]+)</`)
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
	body = html.UnescapeString(body)

	symbolsRe := xmlre.Compile(`<td class="column-4">(\w+)</td><td class="column-5">([^<]+)</td>`)

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

	updateTags(list)

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
		tpl = append(tpl, `		AltNames:  `+fmt.Sprintf(`%#v`, c.AltNames)+`,`)
		tpl = append(tpl, `		Tags:      `+fmt.Sprintf(`%#v`, c.Tags)+`,`)
		tpl = append(tpl, `	},`)
	}

	tpl = append(tpl, `}`)
	tpl = append(tpl, ``)

	_, file, _, _ := runtime.Caller(0)
	dir := filepath.Dir(file)

	return fwrite.WriteRaw(dir+`/currency_db.go`, []byte(strings.Join(tpl, "\n")))
}

func updateTags(list map[string]*Currency) {
	wordSplitRe := regexp.MustCompile(`[^\p{L}\p{N}]+`)
	wordMap := map[string][]*Currency{}

	for _, c := range list {
		name := strings.ToLower(c.Name + ` ` + strings.Join(c.AltNames, ` `))
		words := wordSplitRe.Split(name, -1)
		for _, w := range words {
			if len(w) > 0 {
				wordMap[w] = append(wordMap[w], c)
			}
		}
		c.Tags = []string{}
	}

	for w, cs := range wordMap {
		if len(cs) == 1 {
			cs[0].Tags = append(cs[0].Tags, w)
		}
	}
}
