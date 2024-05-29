package rms
import(
	"bytes"
	"encoding/json"
	"fmt"
	
	"io"

	"net/http"
	"api/model"
	"strconv"
	"time"
)
func Owlting_order(sd string, ed string,hotelId string, batchid string, cookie string) {
	var result string
	var url string
	url = `https://www.owlting.com/booking/v2/admin/hotels/` + batchid + `/orders/calendar_list?lang=zh_TW&limit=50&page=1&during_checkout_date=` + sd + `,` + ed + `&order_by=id&sort_by=asc`
	fmt.Println("Start From Page 1")
	if err := DoRequestAndGetResponse_owl("GET", url, http.NoBody, cookie, &result); err != nil {
		fmt.Println("DoRequestAndGetResponse failed!")
		fmt.Println("err", err)
		return
	}

	var ordersData model.GetOwltingOrderResponseBody
	err := json.Unmarshal([]byte(result), &ordersData)
	if err != nil {
		fmt.Println("JSON解码错误:", err)
		return
	}
	pageCount := ordersData.Pagination.Total_pages
	var resultData []model.ReservationsDB
	var data model.ReservationsDB
	for _, reservation := range ordersData.Data {
		url = `https://www.owlting.com/booking/v2/admin/hotels/` + batchid + `/orders/` + reservation.Order_serial + `/detail?lang=zh_TW`
		if err := DoRequestAndGetResponse_owl("GET", url, http.NoBody, cookie, &result); err != nil {
			fmt.Println("DoRequestAndGetResponse failed!")
			fmt.Println("err", err)
			return
		}

		var orderData model.GetOwltingOrderResponseBody2
		err = json.Unmarshal([]byte(result), &orderData)
		if err != nil {
			fmt.Println("JSON解码错误:", err)
			return
		}

		data.RoomNights = int64(orderData.Data.Info.Order_stay_night)
		count:=0
		roomInfoData := make(map[string]*model.RoomInfo_owl)
		for _, roomReservation := range orderData.Data.Rooms {
			roomType := roomReservation.Room_name
			roomInfo, ok := roomInfoData[roomType]
			if !ok {
				// 如果房间信息不存在，创建新的 RoomInfo
				roomInfo = &model.RoomInfo_owl{
					RoomType: roomType,
					Count:    1,
				}
				roomInfoData[roomType] = roomInfo
			} else {
				roomInfo.Count++
				//roomInfo.StartDate = date
			}
		}							
		var combinedRoomInfo string
		for _, roomInfo := range roomInfoData {
			count += roomInfo.Count/int(orderData.Data.Info.Order_stay_night)
			//fmt.Println("num",count,roomInfo.RoomType)
			if combinedRoomInfo != "" {
				combinedRoomInfo += " + "
			}
			combinedRoomInfo += fmt.Sprintf("%s*%s", roomInfo.RoomType, strconv.Itoa(roomInfo.Count/int(orderData.Data.Info.Order_stay_night)))
		}
		data.RoomName = combinedRoomInfo
		data.BookingId = orderData.Data.Info.Order_serial
		data.NumOfRooms = int64(count)
		data.GuestName = orderData.Data.Info.Orderer_fullname

		arrivalTime, err := time.Parse("2006-01-02", orderData.Data.Info.Sdate)
		if err != nil {
			fmt.Println("Error parsing arrival time:", err)
		}
		departureTime, err := time.Parse("2006-01-02", orderData.Data.Info.Edate)
		if err != nil {
			fmt.Println("Error parsing arrival time:", err)
		}
		parsedTime, err := time.Parse(time.RFC3339, orderData.Data.First_payment.Created_at)
		if err != nil {
			fmt.Println("Error parsing time:", err)
			return
		}
		resultTimeStr := parsedTime.Format("2006-01-02")
		data.BookDate = resultTimeStr
		checkOutTime := departureTime
		checkInTime := arrivalTime
		data.CheckOutDate = checkOutTime.Format("2006-01-02")
		data.CheckInDate = checkInTime.Format("2006-01-02")
		data.Price = float64(orderData.Data.Summary.Hotel.Receivable_total)
		data.Price = float64(orderData.Data.Summary.Hotel.Receivable_total)
		if !orderData.Data.Info.Order_status {
			data.ReservationStatus = "已成立"

		}else{
			data.ReservationStatus = "已取消"
		}

		if orderData.Data.Info.Source == "" {
			data.Platform = orderData.Data.Info.Source2

		} else {
			data.Platform = orderData.Data.Info.Source
		}
		data.Channelname = "owlting"
		data.Currency = "TWD"
		data.HotelId = batchid
		resultData = append(resultData, data)
	}
	//頁數大於1的情況
	if pageCount > 1 {
		for i := 0; i < pageCount; i++ {
			url = `https://www.owlting.com/booking/v2/admin/hotels/` + batchid + `/orders/calendar_list?lang=zh_TW&limit=50&page=` + strconv.Itoa(i+2) + `&during_checkout_date=` + sd + `,` + ed + `&order_by=id&sort_by=asc&_=` + hotelId
			fmt.Println("page",i+2)
			if err := DoRequestAndGetResponse_owl("GET", url, http.NoBody, cookie, &result); err != nil {
				fmt.Println("DoRequestAndGetResponse failed!")
				fmt.Println("err", err)
				return
			}
			var ordersData model.GetOwltingOrderResponseBody
			err = json.Unmarshal([]byte(result), &ordersData)
			if err != nil {
				fmt.Println("JSON解码错误:", err)
				return
			}
			var data model.ReservationsDB
			for _, reservation := range ordersData.Data {
				url = `https://www.owlting.com/booking/v2/admin/hotels/` + batchid + `/orders/` + reservation.Order_serial + `/detail?lang=zh_TW`
				if err := DoRequestAndGetResponse_owl("GET", url, http.NoBody, cookie, &result); err != nil {
					fmt.Println("DoRequestAndGetResponse failed!")
					fmt.Println("err", err)
					return
				}

				var orderData model.GetOwltingOrderResponseBody2
				err = json.Unmarshal([]byte(result), &orderData)
				if err != nil {
					fmt.Println("JSON解码错误:", err)
					return
				}
			
				data.RoomNights = int64(orderData.Data.Info.Order_stay_night)
				roomInfoData := make(map[string]*model.RoomInfo_owl)
				count := 0
				for _, roomReservation := range orderData.Data.Rooms {										
					roomType := roomReservation.Room_name									
					roomInfo, ok := roomInfoData[roomType]
					if !ok {
						// 如果房间信息不存在，创建新的 RoomInfo
						roomInfo = &model.RoomInfo_owl{
							RoomType: roomType,
							Count:    1,
						}
						roomInfoData[roomType] = roomInfo
					} else {
						roomInfo.Count++
						//roomInfo.StartDate = date
					}
				}
				var combinedRoomInfo string
				for _, roomInfo := range roomInfoData {
					count += roomInfo.Count/int(orderData.Data.Info.Order_stay_night)
					if combinedRoomInfo != "" {
						combinedRoomInfo += " + "
					}
					//combinedRoomInfo +=  (roomInfo.RoomType + "*" +roomInfo.Count)
					combinedRoomInfo += fmt.Sprintf("%s*%s", roomInfo.RoomType, strconv.Itoa(roomInfo.Count/int(orderData.Data.Info.Order_stay_night)))
				}
				data.RoomName = combinedRoomInfo
				data.NumOfRooms = int64(count)
				data.BookingId = orderData.Data.Info.Order_serial
				data.GuestName = orderData.Data.Info.Orderer_fullname
				arrivalTime, err := time.Parse("2006-01-02", orderData.Data.Info.Sdate)
				if err != nil {
					fmt.Println("Error parsing arrival time:", err)
				}
				departureTime, err := time.Parse("2006-01-02", orderData.Data.Info.Edate)
				if err != nil {
					fmt.Println("Error parsing arrival time:", err)
				}
				parsedTime, err := time.Parse(time.RFC3339, orderData.Data.First_payment.Created_at)
				if err != nil {
					fmt.Println("Error parsing time:", err)
					return
				}
				resultTimeStr := parsedTime.Format("2006-01-02")
				data.BookDate = resultTimeStr
				checkOutTime := departureTime
				checkInTime := arrivalTime
				data.CheckOutDate = checkOutTime.Format("2006-01-02")
				data.CheckInDate = checkInTime.Format("2006-01-02")
				data.Channelname = "owlting"								
				data.Price = float64(orderData.Data.Summary.Hotel.Receivable_total)
				if !orderData.Data.Info.Order_status {
					data.ReservationStatus = "已成立"

				}else{
					data.ReservationStatus = "已取消"
				}									
				if orderData.Data.Info.Source == "" {
					data.Platform = orderData.Data.Info.Source2

				} else {
					data.Platform = orderData.Data.Info.Source
				}
				data.Currency = "TWD"
				data.HotelId = batchid
				resultData = append(resultData, data)
			}

		}
	}
	fmt.Print("Total numbers of order：")
	fmt.Println(len(resultData))

	resultDataJSON, err := json.Marshal(resultData)
	if err != nil {
		fmt.Println("JSON 轉換錯誤:", err)
		return
	}
	// 將資料存入DB
	var resultDB string						
	apiurl := fmt.Sprintf("http://149.28.24.90:8893/revenue_booking/setDataAnalysisToDB")
	if err := DoRequestAndGetResponse("POST", apiurl, bytes.NewBuffer(resultDataJSON), cookie, &resultDB); err != nil {
		fmt.Println("setParseHtmlToDB failed!")
		return
	}
	fmt.Println("資料下載完成")


}

func DoRequestAndGetResponse_owl(method, postUrl string, reqBody io.Reader, cookie string, resBody any) error {
	req, err := http.NewRequest(method, postUrl, reqBody)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer cec478d3bca0a16b9b95b85f43096913cc9253ff80ac805dc546b9426f55e885")
	req.Header.Set("Content-Type", "application/json")

	// req.Header.Set("Cookie", cookie)
	// switch resBody := resBody.(type) {
	// case *string:
	

	// 	req.Header.Set("Content-Type", "application/json")
	// default:
	// 	fmt.Println("not string")
	// 	req.Header.Set("Content-Type", "application/json")
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
func DoRequestAndGetResponse(method, postUrl string, reqBody io.Reader, cookie string, resBody any) error {
	req, err := http.NewRequest(method, postUrl, reqBody)
	if err != nil {
		return err
	}
	//req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	//req.Header.Set("x-xsrf-token","96145c60a1fb784a7ceee08a13f8c3904292c1d32c63ee2275d9bcf2ddd7ef7de3ff8f44c7fc153e2cb5de9e5f256f59")
	//req.Header.Set(" x-sm-trace-token","b6e084f6-1d63-4e31-915e-a4d73e487bdb")
	req.Header.Set("Authorization", "Bearer cec478d3bca0a16b9b95b85f43096913cc9253ff80ac805dc546b9426f55e885")

	req.Header.Set("Cookie", cookie)
	req.Header.Set("Content-Type", "text/html; charset=utf-8")
	// switch resBody := resBody.(type) {
	// case *string:
	// 	fmt.Println("string")
	// 	fmt.Println(resBody)

	// 	req.Header.Set("Content-Type", "text/html; charset=utf-8")
	// default:
	// 	fmt.Println("not string")
	// 	req.Header.Set("Content-Type", "text/html; charset=utf-8")
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