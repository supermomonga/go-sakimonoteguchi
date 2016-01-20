package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/codegangsta/cli"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"time"
)

type Data struct {
	InfoDate  string `json:"info_date"`
	Company   string `json:"company"`
	N225Sell  string `json:"n225_sell"`
	N225Buy   string `json:"n225_buy"`
	N225Net   string `json:"n225_net"`
	TopixSell string `json:"topix_sell"`
	TopixBuy  string `json:"topix_buy"`
	TopixNet  string `json:"topix_net"`
	NetTotal  string `json:"net_total"`
}

func isFileExist(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func getDates() []time.Time {
	dates := []time.Time{}
	indexURL := "http://j2funds.info/invest/contents/futures/daily.php"
	doc, err := goquery.NewDocument(indexURL)
	if err != nil {
		log.Fatal(err)
	}

	doc.Find("select[name='search_key'] option").Each(func(_ int, s *goquery.Selection) {
		dateString, _ := s.Attr("value")
		date, _ := time.Parse("2006-01-02", dateString)
		dates = append(dates, date)
	})

	return dates
}

func dataDir() string {
	pwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	dataDir := pwd + "/data"
	return dataDir
}

func dataFileName(date time.Time) string {
	dateString := date.Format("2006-01-02")
	return dateString + ".csv"
}

func dataFilePath(date time.Time) string {
	fileName := dataFileName(date)
	return dataDir() + "/" + fileName
}

func isDataFileExist(date time.Time) bool {
	return isFileExist(dataFilePath(date))
}

func getData(date time.Time) []Data {
	dataURL := "http://j2funds.info/invest/func/futures/_get_futures_daily_data.php"
	// get json string
	resp, err := http.PostForm(
		dataURL,
		url.Values{
			"file_name":  {"daily.php"},
			"search_key": {date.Format("2006-01-02")},
		},
	)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	// parse as json
	raw, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var dataset []Data
	err = json.Unmarshal([]byte(raw), &dataset)
	if err != nil {
		log.Fatal(err)
	}
	return dataset
}

func saveDatasetAsCsv(dataset []Data, filePath string) {
	f, err := os.Create(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	var w *csv.Writer
	if runtime.GOOS == "windows" {
		// on Windows, use Shift-JIS to open csv file via Microsoft Excel.
		converter := bufio.NewWriter(transform.NewWriter(f, japanese.ShiftJIS.NewEncoder()))
		w = csv.NewWriter(converter)
	} else {
		w = csv.NewWriter(f)
	}
	defer w.Flush()

	// Write header first
	header := []string{
		// "日付",
		"証券会社名",
		"n225_sell",
		"n225_buy",
		"n225_net",
		"topix_sell",
		"topix_buy",
		"topix_net",
		"net_total",
	}
	w.Write(header)

	// Write dataset
	for _, data := range dataset {
		var record []string
		// record = append(record, obj.InfoDate)
		record = append(record, data.Company)
		record = append(record, data.N225Sell)
		record = append(record, data.N225Buy)
		record = append(record, data.N225Net)
		record = append(record, data.TopixSell)
		record = append(record, data.TopixBuy)
		record = append(record, data.TopixNet)
		record = append(record, data.NetTotal)
		w.Write(record)
	}
}

func main() {
	app := cli.NewApp()
	app.Name = "sakimonoteguchi"
	app.Usage = "KIAI"
	app.Action = func(c *cli.Context) {

		// Create data dir
		dataDir := dataDir()
		if !isFileExist(dataDir) {
			fmt.Println("Data dir doesn't exist. create: " + dataDir)
			if err := os.Mkdir(dataDir, 0755); err != nil {
				log.Fatal(err)
			}
		}

		dates := getDates()

		for _, date := range dates {
			fileName := dataFileName(date)
			filePath := dataFilePath(date)

			if isDataFileExist(date) {
				fmt.Println(fileName + " already exists. skip it.")
			} else {
				time.Sleep(2000 * time.Millisecond)
				dataset := getData(date)
				saveDatasetAsCsv(dataset, filePath)
				fmt.Println(fileName + " saved.")
			}
		}
	}
	app.Run(os.Args)
}
