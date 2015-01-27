// Copyright (c) 2015, b3log.org
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"sort"
	"strings"
)

// Locale.
type locale struct {
	Name     string
	Langs    map[string]interface{}
	TimeZone string
}

// All locales.
var locales = map[string]locale{}

// loadLocales loads i18n message configurations.
func loadLocales() {
	f, _ := os.Open("i18n")
	names, _ := f.Readdirnames(-1)
	f.Close()

	for _, name := range names {
		if !strings.HasSuffix(name, ".json") {
			continue
		}

		loc := name[:strings.LastIndex(name, ".")]
		loadLocale(loc)
	}
}

func loadLocale(localeStr string) {
	bytes, err := ioutil.ReadFile("i18n/" + localeStr + ".json")
	if nil != err {
		logger.Error(err)

		os.Exit(-1)
	}

	l := locale{Name: localeStr}

	err = json.Unmarshal(bytes, &l.Langs)
	if nil != err {
		logger.Error(err)

		os.Exit(-1)
	}

	locales[localeStr] = l
}

// getMsg gets a message with the specified locale and key.
func getMsg(locale, key string) interface{} {
	return locales[locale].Langs[key]
}



// getMsgs gets all messages with the specified locale.
func getMsgs(locale string) map[string]interface{} {
	return locales[locale].Langs
}

// getLocalesNames gets names of all locales. Returns ["zh_CN", "en_US"] for example.
func getLocalesNames() []string {
	ret := []string{}

	for name := range locales {
		ret = append(ret, name)
	}

	sort.Strings(ret)

	return ret
}
