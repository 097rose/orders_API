package main

import (
	"encoding/json"
	"fmt"	
	"io"
	// "log"
	"net/http"
	"os"
	"bufio"
	"io/ioutil"
	"strings"
	"api/rms"
	"api/model"
	// "github.com/spf13/viper"
	"reflect"
	"strconv"
	"time"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"gopkg.in/Iwark/spreadsheet.v2"
)

func main() {
	//input from google sheet
	var hotelId string
	fmt.Print("Give me the hotel ID: ")
	fmt.Scanln(&hotelId)
	dbGsId := "1q77JBoTzquzf2WJw1IvUYyHiDRltAqN82Fc0N2TmiBg"

	data, err := ioutil.ReadFile("client_secret.json")
	checkError(err)

	//conf: Google 認證
	//spreadsheet.Scope: 作用範圍，指定認證配置可以訪問 Google Sheets 的哪些資源
	conf, err := google.JWTConfigFromJSON(data, spreadsheet.Scope)
	checkError(err)

	client := conf.Client(context.TODO())
	service := spreadsheet.NewServiceWithClient(client)

	//從指定的 Google Sheets 擷取資料
	//dbGsId: 電子表格的識別符
	spreadsheetDB, err := service.FetchSpreadsheet(dbGsId)
	checkError(err)

	//Get Google Sheet Id of the hotel
	sheetHotelGs, _ := spreadsheetDB.SheetByTitle("2. 住宿商戶資料")
	var hotelGsList []model.HotelGsInfo

	//idx: 索引，row: 那一行的內容
	for idx, row := range sheetHotelGs.Rows {
		//如果第一行，跳過 (標題行)
		if idx == 0 {
			continue
		}

		if row[0].Value != hotelId {
			// 如果不相等，則跳過這一行資料
			continue
		}

		hotel := model.HotelGsInfo{
			HotelId:   row[0].Value,
			HotelName: row[4].Value,
			GsId:      row[23].Value,
			Source:      row[93].Value,
		}

		hotel.GsId = strings.ReplaceAll(hotel.GsId, "https://docs.google.com/spreadsheets/d/", "")
		slashIndex := strings.Index(hotel.GsId, "/")
		if slashIndex != -1 {
			// 如果找到斜槓，則將斜槓之後的部分替換為空格
			hotel.GsId = hotel.GsId[:slashIndex] + ""
		}
		// 輸出 GsId
		hotelGsList = append(hotelGsList, hotel)
	}

	//找到符合 HOTEL ID 的物件
	var selectedGs *model.HotelGsInfo
	for _, hotel := range hotelGsList {
		if hotel.HotelId == hotelId {

			//&: 取址
			selectedGs = &hotel
			break
		}
	}
	//fmt.Println(selectedGs.GsId)
	spreadsheet, err := service.FetchSpreadsheet(selectedGs.GsId)
	checkError(err)

		//Get hotel info
	sheetHotelInfo, err := spreadsheet.SheetByTitle("旅館要素表")

	var hotels []model.Hotel

	if err != nil {
		fmt.Printf("無法獲取旅館要素表：%v\n", err)
		return
	}
	
	for idx, row := range sheetHotelInfo.Rows {
		if idx != 1 { // 只处理索引为1的行，即第二行
			continue
		}

		// 建立一個新的 Hotel 結構體
		var hotel model.Hotel

		// 用 reflect 轉型欄位值
		v := reflect.ValueOf(&hotel).Elem() // 取得值

		for i := 0; i < v.NumField(); i++ {
			fieldValue := row[i].Value
			field := v.Field(i)
			fieldTag := v.Type().Field(i).Tag.Get("defaultDiscount")

			if fieldTag == "true" {
				if fieldValue != "" {
					floatValue, err := strconv.ParseFloat(strings.TrimSuffix(fieldValue, "%"), 64)
					if err != nil {
						fmt.Println("Error converting percentage to float:", err)
					} else {
						field.SetFloat(floatValue / 100.0)
					}
				} 
			} else if field.Type().Name() == "bool" && field.CanSet() {

				switch fieldValue {
				case "官網(ON)":
					field.SetBool(true)
				case "官網(OFF)":
					field.SetBool(false)
				}
				
			} else {
				processField(field, fieldValue)
			}
		}

		hotels = append(hotels, hotel)
	}


	for _, hotel := range hotels {
		fmt.Println("-----------------------------------------------------------------------------")
		fmt.Printf("Hotel ID: %s\n", hotel.ID)
		fmt.Printf("Hotel Name: %s\n", hotel.HotelName)
		fmt.Printf("channel: %s\n", hotel.Channel)
		fmt.Printf("HotelOne: %s\n", hotel.HotelIdOne)
		fmt.Printf("HotelTwo: %s\n", hotel.HotelIdTwo)
		fmt.Printf("Coookie: %s\n", hotel.ReservationCookie)
		fmt.Printf("auth: %s\n", hotel.ReservationAuth)
		fmt.Println("-----------------------------------------------------------------------------")

		
		//fmt.Printf("Hotel Name: %s, Hotel ID: %s\n", hotel.HotelName, hotel.ID)
		var sd,ed string
		fmt.Print("Give me the start date: ")
		fmt.Scanln(&sd)
		fmt.Print("Give me the end date  : ")
		fmt.Scanln(&ed)
		fmt.Println("From", sd, "to", ed, "   Start Downloading")

		if(selectedGs.Source=="奧丁丁"){
			rms.Owlting_order(sd, ed, hotel.HotelIdTwo, hotel.HotelIdOne, hotel.ReservationCookie)
		}else if(selectedGs.Source=="Traiwan"){
			rms.Traiwan_order(sd,ed,hotel.ReservationCookie,hotel.HotelIdOne)
		}else if(selectedGs.Source=="新版Siteminder"){
			rms.Newsite_order(sd,ed,hotel.HotelIdOne,hotel.ReservationAuth,hotel.ReservationCookie)
		}else if(selectedGs.Source=="大師"){
			rms.Master_order(sd,ed,hotel.HotelIdOne)
		}else{
			fmt.Print("此旅館無法使用golang數據分析")
		}

		fmt.Println("Press 'Enter' to exit...")
		bufio.NewReader(os.Stdin).ReadBytes('\n')
		
	}

	
}

func DoRequestAndGetResponse(method, postUrl string, reqBody io.Reader, cookie string, resBody any) error {
	req, err := http.NewRequest(method, postUrl, reqBody)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer cec478d3bca0a16b9b95b85f43096913cc9253ff80ac805dc546b9426f55e885")

	req.Header.Set("Cookie", cookie)
	switch resBody := resBody.(type) {
	case *string:
		fmt.Println("string")
		fmt.Println(resBody)

		req.Header.Set("Content-Type", "text/html; charset=utf-8")
	default:
		fmt.Println("not string")
		req.Header.Set("Content-Type", "text/html; charset=utf-8")
	}

	client := &http.Client{Timeout: 40 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	// resBody of type *string is for html
	switch resBody := resBody.(type) {
	case *string:
		// If resBody is a string
		resBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		*resBody = string(resBytes)
	default:
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(data, resBody); err != nil {
			return err
		}
	}

	defer resp.Body.Close()

	return nil
}
func DoRequestAndGetResponse_360(method, postUrl string, reqBody io.Reader, cookie string, resBody any) error {
	req, err := http.NewRequest(method, postUrl, reqBody)
	if err != nil {
		return err
	}
	req.Header.Set("sec-ch-ua-platform", "Windows")
	req.Header.Set("sec-ch-ua-Mobile", "?0")
	req.Header.Set("Cookie", cookie)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded;charset=UTF-8")
	// switch resBody := resBody.(type) {
	// case *string:
	// 	fmt.Println("string")
	// 	fmt.Println(resBody)

	// 	req.Header.Set("Content-Type", "application/x-www-form-urlencoded;charset=UTF-8")
	// default:
	// 	fmt.Println("not string")
	// 	req.Header.Set("Content-Type", "application/x-www-form-urlencoded;charset=UTF-8")
	// }

	client := &http.Client{Timeout: 40 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	//fmt.Println(resp)

	// resBody of type *string is for html
	switch resBody := resBody.(type) {
	case *string:
		// If resBody is a string
		resBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		*resBody = string(resBytes)
	default:
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(data, resBody); err != nil {
			return err
		}
	}

	defer resp.Body.Close()

	return nil
}

func checkError(err error) {
	if err != nil {
		panic(err.Error())
	}
}
func parseFloat(str string) (float64, error) {
	return strconv.ParseFloat(str, 64)
}
func parseInt(str string) (int, error) {
	return strconv.Atoi(str)
}
func parseBool(str string) (bool, error) {
	return strconv.ParseBool(str)
}
func processField(field reflect.Value, fieldValue string) {
	switch field.Kind() {
	case reflect.Float64:
		if floatValue, err := parseFloat(fieldValue); err == nil {
			field.SetFloat(floatValue)
		}
	case reflect.Int:
		if intValue, err := parseInt(fieldValue); err == nil {
			field.SetInt(int64(intValue))
		}
	case reflect.Bool:
		if boolValue, err := parseBool(fieldValue); err == nil {
			field.SetBool(boolValue)
		}
	case reflect.String:
		field.SetString(fieldValue)
	}
}
