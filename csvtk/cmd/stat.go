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
	"runtime"

	"github.com/brentp/xopen"
	"github.com/spf13/cobra"

	"github.com/tatsushid/go-prettytable"
	"github.com/dustin/go-humanize"
)

// statCmd represents the seq command
var statCmd = &cobra.Command{
	Use:   "stat",
	Short: "summary of CSV file",
	Long: `summary of CSV file

`,
	Run: func(cmd *cobra.Command, args []string) {
		config := getConfigs(cmd)
		files := getFileList(args)
		runtime.GOMAXPROCS(config.NumCPUs)

		outfh, err := xopen.Wopen(config.OutFile)
		checkError(err)
		defer outfh.Close()


		tbl, err := prettytable.NewTable([]prettytable.Column{
		    {Header: "file"},
		    {Header: "num_cols", AlignRight: true},
			{Header: "num_rows", AlignRight: true}}...)
		checkError(err)
		tbl.Separator = "   "

		for _, file := range files {
			csvReader, err := newCSVReaderByConfig(config, file)
			checkError(err)
			csvReader.Run()

			var numCols, numRows uint64
			once := true
			for chunk := range csvReader.Ch {
				checkError(chunk.Err)

				numRows += uint64(len(chunk.Data))
				if once {
					for _, record := range chunk.Data {
						numCols = uint64(len(record))
						break
					}
					once = false
				}
			}
			if !config.NoHeaderRow {
				numRows--
			}
			tbl.AddRow(
				file,
				humanize.Comma(int64(numCols)),
				humanize.Comma(int64(numRows)))

		}
		outfh.Write(tbl.Bytes())
	},
}

func init() {
	RootCmd.AddCommand(statCmd)
}
