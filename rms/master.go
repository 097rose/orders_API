package rms
import (
	//"bufio"
	"encoding/json"
	"fmt"	
	"io"
	"bytes"	
	"net/http"	
	"api/model"	
	"strings"
	"time"
	"strconv"
)

func Master_order(sd string, ed string, hotelId string){
	var result, url string
	url = "http://mrhost.xcodemy.com/api/vendor/getMasterOrders"
	//fmt.Println(hotelId)
	rawbody := `{"domain": "`+hotelId+`","date_type": "check_out","start_date": "`+sd+`" ,"end_date": "`+ed+`"}`
	if err := DoRequestAndGetResponse_master("POST", url, strings.NewReader(rawbody), &result); err != nil {
		fmt.Println("DoRequestAndGetResponse failed!")
		fmt.Println("err", err)
		return
	}
	if(result==`{"message":"\u53ea\u80fd\u67e5\u8a62\u524d\u5f8c\u4e09\u500b\u6708\u7684\u8cc7\u6599","error":"\u53ea\u80fd\u67e5\u8a62\u524d\u5f8c\u4e09\u500b\u6708\u7684\u8cc7\u6599"}`){
		fmt.Println("只能查詢前後三個月的資料")
	}
	
	var ordersData model.GetMasterOrderResponseBody
	err := json.Unmarshal([]byte(result), &ordersData)
	if err != nil {
		fmt.Println("JSON解码错误:", err)
		return
	}
	//fmt.Println(ordersData)
	var resultData []model.ReservationsDB
	var data model.ReservationsDB
	for _, reservation := range ordersData {
		data.BookingId = reservation.ID
		data.GuestName = reservation.Guestname

		
		data.NumOfRooms = int64(reservation.Rooms)
		data.RoomName = reservation.Roomname
		data.Channelname = "大師"

		parsedTime, err := time.Parse("2006-01-02 15:04:05", reservation.Created_at)
		if err != nil {
			fmt.Println("Error parsing time:", err)
			return
		}

		arrivalTime, err := time.Parse("2006-01-02", reservation.Sdate)
		if err != nil {
			fmt.Println("Error parsing arrival time:", err)
		}
		departureTime, err := time.Parse("2006-01-02", reservation.Edate)
		if err != nil {
			fmt.Println("Error parsing arrival time:", err)
		}
		duration := departureTime.Sub(arrivalTime)
		days := int(duration.Hours() / 24)
		data.RoomNights = int64(days)
		checkOutTime := departureTime
		checkInTime := arrivalTime
		data.CheckOutDate = checkOutTime.Format("2006-01-02")
		data.CheckInDate = checkInTime.Format("2006-01-02")
		resultTimeStr := parsedTime.Format("2006-01-02")
		data.BookDate = resultTimeStr
		cleanStr := strings.ReplaceAll(reservation.Totalprice, ",", "")
		floatNum, _ := strconv.ParseFloat(cleanStr, 64)
		data.Price = floatNum
		if reservation.Orderstatus == "CANCELED" {
			data.ReservationStatus = "已取消"
		} else {
			data.ReservationStatus = "已成立"
		}
		data.Platform = reservation.Source
		data.Currency = "TWD"
		data.HotelId = hotelId
		resultData = append(resultData, data)
	}
	fmt.Println("resultdata", resultData)
	fmt.Println("-----------------------------------------------------------------------------------------------------")
	fmt.Print("Total numbers of order：")
	fmt.Println(len(resultData))

	resultDataJSON, err := json.Marshal(resultData)
	if err != nil {
		fmt.Println("JSON 轉換錯誤:", err)
		return
	}
	// fmt.Println(resultDataJSON)
	// 將資料存入DB
	var resultDB string	
	cookie:="222"					
	apiurl := fmt.Sprintf("http://149.28.24.90:8893/revenue_booking/setDataAnalysisToDB")
	if err := DoRequestAndGetResponse("POST", apiurl, bytes.NewBuffer(resultDataJSON), cookie, &resultDB); err != nil {
		fmt.Println("setParseHtmlToDB failed!")
		return
	}
	fmt.Println("資料下載完成")


}


func DoRequestAndGetResponse_master(method, postUrl string, reqBody io.Reader, resBody any) error {
	req, err := http.NewRequest(method, postUrl, reqBody)
	if err != nil {
		return err
	}

	req.Header.Set("Cookie", "worktel_session=eyJpdiI6IkdzWDBNdTRTMUFjMitlTFp6UEV4eWc9PSIsInZhbHVlIjoiVGlJMFMyMHQ5MnorRU85bTArQWl5NmVkcUxVOU9mSEorWVBGVlpYNExGRVMvYjNWZU9zdnZCYnJQS3VMdFpBbXVaMGZmN3p1Wkg4L3BFR3B0OFBBY2pRSDRqWnRhd0Z3ZEYrUlpqTStBaHcrdEwrK3k1MU53SHpoRmVZc3BRRTUiLCJtYWMiOiI5OGZjODYwNTMyNWQwMGQxN2JlOTNlNGY4YjY4MGViNmVlY2Q0ODY2ZjM1ODhjMjc1MDJkNzcxNTgwZjdjYzkyIiwidGFnIjoiIn0%3D")
	req.Header.Set("Content-Type", "application/json")
	

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