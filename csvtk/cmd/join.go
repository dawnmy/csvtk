// Copyright © 2016 Wei Shen <shenwei356@gmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

import (
	"encoding/csv"
	"fmt"
	"runtime"
	"strings"

	"github.com/brentp/xopen"
	"github.com/spf13/cobra"
)

// joinCmd represents the seq command
var joinCmd = &cobra.Command{
	Use:   "join",
	Short: "join multiple CSV files by selected fields",
	Long: ` join 2nd and later files to the first file by selected fields.
	
Multiple keys supported, but the orders are ignored.

`,
	Run: func(cmd *cobra.Command, args []string) {
		config := getConfigs(cmd)
		files := getFileList(args)
		runtime.GOMAXPROCS(config.NumCPUs)

		if len(files) < 2 {
			checkError(fmt.Errorf("two or more files needed"))
		}

		allFields := getFlagSemicolonSeparatedStrings(cmd, "fields")
		if len(allFields) == 1 {
			s := make([]string, len(files))
			for i := range files {
				s[i] = allFields[0]
			}
			allFields = s
		} else if len(allFields) != len(files) {
			checkError(fmt.Errorf("number of fields (%d) should be equal to number of files (%d)", len(allFields), len(files)))
		}
		// ignoreCase := getFlagBool(cmd, "ignore-case")

		fuzzyFields := getFlagBool(cmd, "fuzzy-fields")
		keepUnmatched := getFlagBool(cmd, "keep-unmatched")

		outfh, err := xopen.Wopen(config.OutFile)
		checkError(err)
		defer outfh.Close()

		writer := csv.NewWriter(outfh)
		if config.OutTabs || config.Tabs {
			writer.Comma = '\t'
		} else {
			writer.Comma = config.OutDelimiter
		}

		var HeaderRow []string
		var Data [][]string
		var Fields []int
		firstFile := true
		var withHeaderRow bool

		var key string
		var items []string
		for i, file := range files {
			headerRow, data, fields := parseCSVfile(cmd, config,
				file, allFields[i], fuzzyFields)
			if firstFile {
				HeaderRow, Data, Fields = headerRow, data, fields
				firstFile = false
				if len(HeaderRow) > 0 {
					withHeaderRow = true
				}
				continue
			}

			// fieldsMap
			fieldsMap := make(map[int]struct{}, len(fields))
			for _, f := range fields {
				fieldsMap[f] = struct{}{}
			}

			// csv to map
			keysMaps := make(map[string][]string)
			items = make([]string, len(fields))
			for _, record := range data {
				for i, f := range fields {
					items[i] = record[f-1]
				}
				key = strings.Join(items, "_shenwei356_")
				keysMaps[key] = record
			}

			Data2 := [][]string{}
			if withHeaderRow {
				newHeaderRow := HeaderRow
				for f, colname := range headerRow {
					if _, ok := fieldsMap[f+1]; !ok {
						newHeaderRow = append(newHeaderRow, colname)
					}
				}
				HeaderRow = newHeaderRow
			}
			items = make([]string, len(Fields))
			for _, record := range Data {
				for i, f := range Fields {
					items[i] = record[f-1]
				}
				key = strings.Join(items, "_shenwei356_")
				if record2, ok := keysMaps[key]; ok {
					for f, v := range record2 {
						if _, ok := fieldsMap[f+1]; !ok {
							record = append(record, v)
						}
					}
					Data2 = append(Data2, record)
				} else {
					if keepUnmatched {
						for i := 1; i <= len(data[0])-len(fieldsMap); i++ {
							record = append(record, "")
						}
						Data2 = append(Data2, record)
					}
				}
			}
			Data = Data2
		}

		if withHeaderRow {
			checkError(writer.Write(HeaderRow))
		}
		for _, record := range Data {
			checkError(writer.Write(record))
		}

		writer.Flush()
		checkError(writer.Error())
	},
}

func init() {
	RootCmd.AddCommand(joinCmd)
	joinCmd.Flags().StringP("fields", "f", "1", "Semicolon seperated key fields of all files, "+
		"if given one, we think all the files have the same key columns. e.g -f 1;2 or -f A,B;C,D or -f id")
	joinCmd.Flags().BoolP("ignore-case", "i", false, `ignore case`)
	joinCmd.Flags().BoolP("fuzzy-fields", "F", false, `using fuzzy fileds, e.g. *name or id123*`)
	joinCmd.Flags().BoolP("keep-unmatched", "k", false, `keep unmatched data of the first file`)
}
