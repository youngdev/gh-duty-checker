// main.go
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/olekukonko/tablewriter"
)

var isDebug bool

const (
	baseURL              = "https://external.unipassghana.com"
	searchMakePopupURL   = baseURL + "/co/popup/selectCommonVehicleMakePopup.do"
	searchModelPopupURL  = baseURL + "/co/popup/selectCommonVehicleModelPopup.do"
	usedVehicleSearchURL = baseURL + "/cl/tm/tax/selectUsedVehicleTaxCalculate.do?decorator=popup&MENU_ID=IIM01S03V02"
	usedVehicleDetailURL = baseURL + "/cl/tm/tax/selectUsedVehicleDetails.do"
	userAgent            = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/137.0.0.0 Safari/537.36"
)

type VehicleResult struct {
	No             string
	TrimLevel      string
	Year           string
	Make           string
	Model          string
	ExchangeRate   string
	ReceiptDate    string
	AssessmentDate string
	TotalTax       string
	DetailParams   DetailParams
}
type DetailParams struct {
	CustomsOfficeCd  string
	DeclarationYear  string
	DeclarationSeqNo string
	AssessmentSeqNo  string
	ItemNo           string
}
type TaxItem struct {
	No          string
	TaxCode     string
	TaxCodeName string
	TaxRate     string
	TaxNCY      string
}
type Client struct{ httpClient *http.Client }

func NewClient() *Client {
	return &Client{httpClient: &http.Client{Timeout: time.Second * 60}}
}

func (c *Client) GetMakeCode(makeName string) (string, error) {
	formData := url.Values{
		"codeObjId":      {"searchMakeCd"},
		"codeNameObjId":  {"searchMakeNm"},
		"callBackNm":     {"usedVehicleTax.setVehicleMakeCallBack"},
		"searchCodeName": {makeName},
	}
	doc, err := c.postFormAndGetDoc(searchMakePopupURL, formData, "")
	if err != nil {
		return "", err
	}
	var makeCode string
	var found bool
	doc.Find("table.g-table tbody tr").EachWithBreak(func(i int, s *goquery.Selection) bool {
		if strings.EqualFold(strings.TrimSpace(s.Find("td:nth-child(3)").Text()), makeName) {
			onclick, _ := s.Attr("onclick")
			re := regexp.MustCompile(`selectCommonCode\('([^']*)'`)
			matches := re.FindStringSubmatch(onclick)
			if len(matches) > 1 {
				makeCode = matches[1]
				found = true
				return false
			}
		}
		return true
	})
	if !found {
		return "", fmt.Errorf("could not find make code for '%s'", makeName)
	}
	return makeCode, nil
}
func (c *Client) GetModelCode(makeName, makeCode, modelName string) (string, error) {
	formData := url.Values{
		"codeObjId":              {"searchModelTypeCd"},
		"codeNameObjId":          {"searchModelTypeNm"},
		"callBackNm":             {"usedVehicleTax.setVehicleModelCallBack"},
		"searchModelDescription": {modelName},
	}
	doc, err := c.postFormAndGetDoc(searchModelPopupURL, formData, "")
	if err != nil {
		return "", err
	}
	var modelCode string
	var found bool
	doc.Find("table.g-table tbody tr").EachWithBreak(func(i int, s *goquery.Selection) bool {
		rowModel := strings.TrimSpace(s.Find("td:nth-child(3)").Text())
		rowMake := strings.TrimSpace(s.Find("td:nth-child(5)").Text())
		if strings.EqualFold(rowModel, modelName) && strings.EqualFold(rowMake, makeName) {
			onclick, _ := s.Attr("onclick")
			re := regexp.MustCompile(`selectCommonCode\('([^']*)'`)
			matches := re.FindStringSubmatch(onclick)
			if len(matches) > 1 {
				modelCode = matches[1]
				found = true
				return false
			}
		}
		return true
	})
	if !found {
		return "", fmt.Errorf("could not find model code for '%s' by make '%s'", modelName, makeName)
	}
	return modelCode, nil
}

func (c *Client) SearchVehicles(makeName, makeCode, modelName, modelCode, year, startDate, endDate string) ([]VehicleResult, error) {
	// FINAL FIX: Manually build the request body to enforce the exact parameter order
	// observed in the HAR file. The server is fragile and requires this specific order.
	listOpJSON := `{"searchEndApprovalDate":null,"miv_pageNo":"1","searchChassisNo":null,"searchStartApprovalDate":null,"searchType":null,"miv_start_index":"0","searchMakeCd":null,"searchMakeNm":null,"searchManufactureYear":null,"miv_end_index":"10","searchModelTypeNm":null,"miv_sort":"","miv_pageSize":"10","searchModelTypeCd":null}`
	encodedListOp := url.QueryEscape(url.QueryEscape(listOpJSON))

	orderedParams := []string{
		"screenType=" + url.QueryEscape("S"),
		"MENU_ID=" + url.QueryEscape("IIM01S03V02"),
		"LISTOP=" + encodedListOp, // This value is already encoded, no need to escape it again
		"searchType=" + url.QueryEscape("02"),
		"searchChassisNo=" + url.QueryEscape(""),
		"searchMakeCd=" + url.QueryEscape(makeCode),
		"searchMakeNm=" + url.QueryEscape(makeName),
		"searchModelTypeCd=" + url.QueryEscape(modelCode),
		"searchModelTypeNm=" + url.QueryEscape(modelName),
		"searchManufactureYear=" + url.QueryEscape(year),
		"searchStartApprovalDate=" + url.QueryEscape(startDate),
		"searchEndApprovalDate=" + url.QueryEscape(endDate),
		"miv_pageNo=" + url.QueryEscape(""),
		"miv_pageSize=" + url.QueryEscape("10"),
	}
	bodyString := strings.Join(orderedParams, "&")

	doc, err := c.postFormAndGetDoc(usedVehicleSearchURL, nil, bodyString)
	if err != nil {
		return nil, err
	}

	var results []VehicleResult
	detailRe := regexp.MustCompile(`goDetail\('([^']*)', '([^']*)', '([^']*)', '([^']*)', '([^']*)'`)

	doc.Find("table[data-table='rwd'] tbody tr").Each(func(i int, s *goquery.Selection) {
		if s.Find("td").Length() < 10 {
			return
		}
		if strings.Contains(s.Text(), "No data found") {
			return
		}

		href, _ := s.Find("td:nth-child(2) a").Attr("href")
		detailMatches := detailRe.FindStringSubmatch(href)
		if len(detailMatches) < 6 {
			return
		}

		results = append(results, VehicleResult{
			No:             strings.TrimSpace(s.Find("td:nth-child(1)").Text()),
			TrimLevel:      strings.TrimSpace(s.Find("td:nth-child(2)").Text()),
			Year:           strings.TrimSpace(s.Find("td:nth-child(3)").Text()),
			Make:           strings.TrimSpace(s.Find("td:nth-child(4)").Text()),
			Model:          strings.TrimSpace(s.Find("td:nth-child(5)").Text()),
			ExchangeRate:   strings.TrimSpace(s.Find("td:nth-child(10)").Text()),
			ReceiptDate:    strings.TrimSpace(s.Find("td:nth-child(11)").Text()),
			AssessmentDate: strings.TrimSpace(s.Find("td:nth-child(12)").Text()),
			TotalTax:       strings.TrimSpace(s.Find("td:nth-child(14)").Text()),
			DetailParams: DetailParams{
				CustomsOfficeCd:  detailMatches[1],
				DeclarationYear:  detailMatches[2],
				DeclarationSeqNo: detailMatches[3],
				AssessmentSeqNo:  detailMatches[4],
				ItemNo:           detailMatches[5],
			},
		})
	})
	return results, nil
}

func (c *Client) GetTaxDetails(params DetailParams) ([]TaxItem, error) {
	formData := url.Values{
		"customsOfficeCd":  {params.CustomsOfficeCd},
		"declarationYear":  {params.DeclarationYear},
		"declarationSeqNo": {params.DeclarationSeqNo},
		"assessmentSeqNo":  {params.AssessmentSeqNo},
		"itemNo":           {params.ItemNo},
	}
	doc, err := c.postFormAndGetDoc(usedVehicleDetailURL, formData, "")
	if err != nil {
		return nil, err
	}

	var taxItems []TaxItem
	doc.Find("tbody tr").Each(func(i int, s *goquery.Selection) {
		if s.Find("td").Length() < 5 {
			return
		}
		taxItems = append(taxItems, TaxItem{
			No:          strings.TrimSpace(s.Find("td:nth-child(1)").Text()),
			TaxCode:     strings.TrimSpace(s.Find("td:nth-child(2)").Text()),
			TaxCodeName: strings.TrimSpace(s.Find("td").Eq(2).Text()),
			TaxRate:     strings.TrimSpace(s.Find("td:nth-child(4)").Text()),
			TaxNCY:      strings.TrimSpace(s.Find("td:nth-child(5)").Text()),
		})
	})
	return taxItems, nil
}

func (c *Client) postFormAndGetDoc(reqURL string, data url.Values, rawBody string) (*goquery.Document, error) {
	var bodyReader io.Reader
	bodyToLog := ""

	if rawBody != "" {
		bodyReader = strings.NewReader(rawBody)
		bodyToLog = rawBody
	} else {
		encodedData := data.Encode()
		bodyReader = strings.NewReader(encodedData)
		bodyToLog = encodedData
	}

	if isDebug {
		fmt.Println("------------------------------------------------------")
		fmt.Printf("DEBUG: Sending POST request to: %s\n", reqURL)
		fmt.Println("DEBUG: Request Body:")
		if len(bodyToLog) > 500 {
			fmt.Printf("  %s...\n", bodyToLog[:500])
		} else {
			fmt.Printf("  %s\n", bodyToLog)
		}
		fmt.Println("------------------------------------------------------")
	}

	req, err := http.NewRequest("POST", reqURL, bodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Origin", baseURL)
	req.Header.Set("Referer", usedVehicleSearchURL)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("bad status: %s, Body: %s", res.Status, string(body))
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}
	return doc, nil
}

func parseAssessment(durationStr string) (string, string, error) {
	if durationStr == "" {
		durationStr = "3m"
	}
	endDate := time.Now()
	if durationStr == "1d" {
		dateStr := endDate.Format("02/01/2006")
		return dateStr, dateStr, nil
	}
	endDateStr := endDate.Format("02/01/2006")
	re := regexp.MustCompile(`^(\d+)([dwmy])$`)
	matches := re.FindStringSubmatch(strings.ToLower(durationStr))
	if len(matches) != 3 {
		return "", "", fmt.Errorf("invalid duration format: '%s'. Use format like 4d, 1w, 3m, 1y", durationStr)
	}
	val, _ := strconv.Atoi(matches[1])
	unit := matches[2]
	var startDate time.Time
	switch unit {
	case "d":
		startDate = endDate.AddDate(0, 0, -val)
	case "w":
		startDate = endDate.AddDate(0, 0, -val*7)
	case "m":
		startDate = endDate.AddDate(0, -val, 0)
	case "y":
		startDate = endDate.AddDate(-val, 0, 0)
	}
	startDateStr := startDate.Format("02/01/2006")
	return startDateStr, endDateStr, nil
}

func displayVehicleList(results []VehicleResult) {
	table := tablewriter.NewWriter(os.Stdout)
	table.Header([]string{"Make", "Model", "Year", "Exchange Rate", "Receipt Date", "Assessment Date", "Total Tax"})
	for _, r := range results {
		table.Append([]string{r.Make, r.Model, r.Year, r.ExchangeRate, r.ReceiptDate, r.AssessmentDate, r.TotalTax})
	}
	table.Render()
}
func displayTaxList(items []TaxItem) {
	table := tablewriter.NewWriter(os.Stdout)
	table.Header([]string{"No.", "Tax Code", "Tax Code Name", "Tax Rate", "Tax NCY"})
	for _, item := range items {
		table.Append([]string{item.No, item.TaxCode, item.TaxCodeName, item.TaxRate, item.TaxNCY})
	}
	table.Render()
}

func main() {
	makeFlag := flag.String("make", "", "Make of the car (e.g., 'Tesla')")
	modelFlag := flag.String("model", "", "Model of the car (e.g., 'Model X')")
	yearFlag := flag.String("year", "2024", "Year of manufacture")
	assessmentFlag := flag.String("assessment", "3m", "Assessment date range (e.g., 4d, 2w, 3m, 1y). '1d' for today only.")
	listFlag := flag.Bool("list", false, "Display the list of matching vehicles and their total tax.")
	taxListFlag := flag.Bool("tax-list", false, "Display the detailed tax breakdown for the most recent vehicle found.")
	debugFlag := flag.Bool("debug", false, "Enable debug logging to print request details.")

	flag.Parse()
	isDebug = *debugFlag

	if *makeFlag == "" || *modelFlag == "" {
		fmt.Println("Error: -make and -model flags are required.")
		flag.Usage()
		return
	}
	if !*listFlag && !*taxListFlag {
		fmt.Println("Error: You must specify either -list or -tax-list.")
		flag.Usage()
		return
	}

	client := NewClient()

	fmt.Printf("Fetching codes for %s %s...\n", *makeFlag, *modelFlag)
	makeCode, err := client.GetMakeCode(*makeFlag)
	if err != nil {
		log.Fatalf("Error getting make code: %v", err)
	}
	modelCode, err := client.GetModelCode(*makeFlag, makeCode, *modelFlag)
	if err != nil {
		log.Fatalf("Error getting model code: %v", err)
	}
	fmt.Printf("-> Found Make Code: %s, Model Code: %s\n", makeCode, modelCode)

	startDate, endDate, err := parseAssessment(*assessmentFlag)
	if err != nil {
		log.Fatalf("Error parsing assessment date: %v", err)
	}

	fmt.Printf("Searching for vehicles from %s to %s...\n", startDate, endDate)
	results, err := client.SearchVehicles(*makeFlag, makeCode, *modelFlag, modelCode, *yearFlag, startDate, endDate)
	if err != nil {
		log.Fatalf("Error searching for vehicles: %v", err)
	}

	if len(results) == 0 {
		fmt.Println("No data found for the specified criteria.")
		return
	}
	fmt.Printf("Found %d result(s).\n\n", len(results))

	if *listFlag {
		fmt.Println("--- Vehicle List ---")
		displayVehicleList(results)
		fmt.Println()
	}

	if *taxListFlag {
		fmt.Println("--- Tax Breakdown (Most Recent) ---")
		mostRecent := results[0]
		fmt.Printf("Fetching tax details for %s %s (Trim: %s)...\n", mostRecent.Make, mostRecent.Model, mostRecent.TrimLevel)
		taxItems, err := client.GetTaxDetails(mostRecent.DetailParams)
		if err != nil {
			log.Fatalf("Error getting tax details: %v", err)
		}
		if len(taxItems) == 0 {
			fmt.Println("Could not retrieve detailed tax breakdown.")
		} else {
			displayTaxList(taxItems)
		}
	}
}
