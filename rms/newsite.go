package rms
import (
	//"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"api/model"
	"strconv"
	"strings"
	"time"
)

func Newsite_order(sd string, ed string,hotelId string, x_xsrf_token string, cookie string){
	var result string
	//var url string
	url := `https://platform.siteminder.com/api/cm-beef/graphql`
	requestJSON := `{"operationName":"getReservationsSearch","variables":{"spid":"` + hotelId + `","filters":{"checkOutDateRange":{"startDate":"` + sd + `","endDate":"` + ed + `"}},"pagination":{"page":1,"pageSize":100,"sortBy":"checkOutDate","sortOrder":"asc"}},"query":"query getReservationsSearch($spid: ID!, $filters: PlatformReservationsFilterInput, $pagination: PlatformReservationsPaginationInput) {\n  hotel(spid: $spid) {\n    spid\n    platformReservations(filters: $filters, pagination: $pagination) {\n      results {\n        uuid\n        sourceId\n        smPlatformReservationId\n        channel {\n          code\n          name\n          __typename\n        }\n        fromDate\n        toDate\n        checkOutDate\n        channelCreatedAt\n        currency\n        totalAmount {\n          amountAfterTax\n          amountBeforeTax\n          __typename\n        }\n        roomStays {\n          cmRoomRateUuid\n          cmChannelRoomRateUuid\n          roomName\n          numberOfAdults\n          numberOfChildren\n          numberOfInfants\n          guests {\n            companyName\n            firstName\n            middleName\n            lastName\n            __typename\n          }\n          __typename\n        }\n        guests {\n          companyName\n          firstName\n          middleName\n          lastName\n          __typename\n        }\n        profiles {\n          companyName\n          firstName\n          middleName\n          lastName\n          __typename\n        }\n        type\n        pmsDeliveryStatus\n        pmsLastSentAt\n        __typename\n      }\n      sortBy\n      sortOrder\n      page\n      pageSize\n      total\n      __typename\n    }\n    __typename\n  }\n}\n"}`
	//requestJSON :=`{"operationName":"getReservationsSearch","variables":{"spid":"b4f66c65-4d1c-4c14-97ea-410615a65158","filters":{"checkOutDateRange":{"startDate":"2024-01-01","endDate":"2024-04-01"}},"pagination":{"page":1,"pageSize":10,"sortBy":"checkOutDate","sortOrder":"asc"}},"query":"query getReservationsSearch($spid: ID!, $filters: PlatformReservationsFilterInput, $pagination: PlatformReservationsPaginationInput) {\n  hotel(spid: $spid) {\n    spid\n    platformReservations(filters: $filters, pagination: $pagination) {\n      results {\n        uuid\n        sourceId\n        smPlatformReservationId\n        channel {\n          code\n          name\n          __typename\n        }\n        fromDate\n        toDate\n        checkOutDate\n        channelCreatedAt\n        currency\n        totalAmount {\n          amountAfterTax\n          amountBeforeTax\n          __typename\n        }\n        roomStays {\n          cmRoomRateUuid\n          cmChannelRoomRateUuid\n          roomName\n          numberOfAdults\n          numberOfChildren\n          numberOfInfants\n          guests {\n            companyName\n            firstName\n            middleName\n            lastName\n            __typename\n          }\n          __typename\n        }\n        guests {\n          companyName\n          firstName\n          middleName\n          lastName\n          __typename\n        }\n        profiles {\n          companyName\n          firstName\n          middleName\n          lastName\n          __typename\n        }\n        type\n        pmsDeliveryStatus\n        pmsLastSentAt\n        __typename\n      }\n      sortBy\n      sortOrder\n      page\n      pageSize\n      total\n      __typename\n    }\n    __typename\n  }\n}\n"}`
	if err := DoRequestAndGetResponse_sit("POST", url, strings.NewReader(requestJSON), cookie, x_xsrf_token, &result); err != nil {
		fmt.Println("DoRequestAndGetResponse failed!")
		fmt.Println("err", err)
		return
	}
	
	var ordersData model.GetSiteOrderResponseBody
	err := json.Unmarshal([]byte(result), &ordersData)
	if err != nil {
		fmt.Println("JSON解码错误:", err)
		return
	}
	orderCount := ordersData.Data.Hotel.PlatformReservations.Total
	fmt.Println("number of order",orderCount)
	var resultData []model.ReservationsDB
	var data model.ReservationsDB
	for _, reservation := range ordersData.Data.Hotel.PlatformReservations.Results {
		data.BookingId = reservation.SourceId
		if len(reservation.Profiles) > 0 {
			data.GuestName = reservation.Profiles[0].FirstName + " " + reservation.Profiles[0].LastName
		} else {
			//data.GuestName = reservation.RoomStays[0].Guests[0].FirstName + " " + reservation.RoomStays[0].Guests[0].LastName
		}

		arrivalTime, err := time.Parse("2006-01-02", reservation.FromDate)
		if err != nil {
			fmt.Println("Error parsing arrival time:", err)
		}

		departureTime, err := time.Parse("2006-01-02", reservation.CheckOutDate)
		if err != nil {
			fmt.Println("Error parsing arrival time:", err)
		}
		
		
		if reservation.PmsLastSentAt!=""{
			parsedTime, err := time.Parse(time.RFC3339, reservation.PmsLastSentAt)
			if err != nil {
				fmt.Println("Error parsing booktime:", err)
				return
			}
			resultTimeStr := parsedTime.Format("2006-01-02")
			data.BookDate = resultTimeStr
		} else if reservation.ChannelCreatedAt!=""{
			parsedTime, err := time.Parse(time.RFC3339, reservation.ChannelCreatedAt)
			if err != nil {
				fmt.Println("Error parsing booktime:", err)
				return
			}
			resultTimeStr := parsedTime.Format("2006-01-02")
			data.BookDate = resultTimeStr
		}else{
			data.BookDate = "err"
		}
		

		
		roomStayCount := len(reservation.RoomStays)
		roomInfoData := make(map[string]*model.RoomInfo_site)
		for _, roomReservation := range reservation.RoomStays {
			roomType := roomReservation.RoomName
			roomInfo, ok := roomInfoData[roomType]
			if !ok {
				// 如果房间信息不存在，创建新的 RoomInfo
				roomInfo = &model.RoomInfo_site{
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
			if combinedRoomInfo != "" {
				combinedRoomInfo += " + "
			}
			combinedRoomInfo += fmt.Sprintf("%s*%s", roomInfo.RoomType, strconv.Itoa(roomInfo.Count))
		}
		data.RoomName = combinedRoomInfo
		
		checkOutTime := departureTime
		checkInTime := arrivalTime
		data.CheckOutDate = checkOutTime.Format("2006-01-02")
		data.CheckInDate = checkInTime.Format("2006-01-02")

		duration := checkOutTime.Sub(checkInTime)
		roomNights := int64(duration.Hours() / 24)

		data.RoomNights = roomNights					
		data.Price = reservation.TotalAmount.AmountAfterTax
		if reservation.Type == "Cancellation"{
			data.ReservationStatus = "已取消"
		}else{
			data.ReservationStatus = "已成立"
		}
		//data.ReservationStatus = reservation.Type
		data.Platform = reservation.Channel.Name
		data.Currency = reservation.Currency
		data.HotelId = hotelId
		data.NumOfRooms = int64(roomStayCount)
		data.Channelname = "siteminder"
		resultData = append(resultData, data)
	}

	//頁數大於1的情況
	if orderCount > 10 {
		page := orderCount / 10
		for i := 0; i < page; i++ {
			//requestJSON :=`{"operationName":"getReservationsSearch","variables":{"spid":"b4f66c65-4d1c-4c14-97ea-410615a65158","filters":{"bookedOnDateRange":{"startDate":"2024-01-01","endDate":"2024-04-01"}},"pagination":{"page":` + strconv.Itoa(i+2) + `,"pageSize":10,"sortBy":"checkOutDate","sortOrder":"asc"}},"query":"query getReservationsSearch($spid: ID!, $filters: PlatformReservationsFilterInput, $pagination: PlatformReservationsPaginationInput) {\n  hotel(spid: $spid) {\n    spid\n    platformReservations(filters: $filters, pagination: $pagination) {\n      results {\n        uuid\n        sourceId\n        smPlatformReservationId\n        channel {\n          code\n          name\n          __typename\n        }\n        fromDate\n        toDate\n        checkOutDate\n        channelCreatedAt\n        currency\n        totalAmount {\n          amountAfterTax\n          amountBeforeTax\n          __typename\n        }\n        roomStays {\n          cmRoomRateUuid\n          cmChannelRoomRateUuid\n          roomName\n          numberOfAdults\n          numberOfChildren\n          numberOfInfants\n          guests {\n            companyName\n            firstName\n            middleName\n            lastName\n            __typename\n          }\n          __typename\n        }\n        guests {\n          companyName\n          firstName\n          middleName\n          lastName\n          __typename\n        }\n        profiles {\n          companyName\n          firstName\n          middleName\n          lastName\n          __typename\n        }\n        type\n        pmsDeliveryStatus\n        pmsLastSentAt\n        __typename\n      }\n      sortBy\n      sortOrder\n      page\n      pageSize\n      total\n      __typename\n    }\n    __typename\n  }\n}\n"}`
			requestJSON := `{"operationName":"getReservationsSearch","variables":{"spid":"` + hotelId + `","filters":{"checkOutDateRange":{"startDate":"` + sd + `","endDate":"` + ed + `"}},"pagination":{"page":` + strconv.Itoa(i+2) + `,"pageSize":100,"sortBy":"checkOutDate","sortOrder":"asc"}},"query":"query getReservationsSearch($spid: ID!, $filters: PlatformReservationsFilterInput, $pagination: PlatformReservationsPaginationInput) {\n  hotel(spid: $spid) {\n    spid\n    platformReservations(filters: $filters, pagination: $pagination) {\n      results {\n        uuid\n        sourceId\n        smPlatformReservationId\n        channel {\n          code\n          name\n          __typename\n        }\n        fromDate\n        toDate\n        checkOutDate\n        channelCreatedAt\n        currency\n        totalAmount {\n          amountAfterTax\n          amountBeforeTax\n          __typename\n        }\n        roomStays {\n          cmRoomRateUuid\n          cmChannelRoomRateUuid\n          roomName\n          numberOfAdults\n          numberOfChildren\n          numberOfInfants\n          guests {\n            companyName\n            firstName\n            middleName\n            lastName\n            __typename\n          }\n          __typename\n        }\n        guests {\n          companyName\n          firstName\n          middleName\n          lastName\n          __typename\n        }\n        profiles {\n          companyName\n          firstName\n          middleName\n          lastName\n          __typename\n        }\n        type\n        pmsDeliveryStatus\n        pmsLastSentAt\n        __typename\n      }\n      sortBy\n      sortOrder\n      page\n      pageSize\n      total\n      __typename\n    }\n    __typename\n  }\n}\n"}`
			if err := DoRequestAndGetResponse_sit("POST", url, strings.NewReader(requestJSON), cookie, x_xsrf_token, &result); err != nil {
				fmt.Println("DoRequestAndGetResponse failed!")
				fmt.Println("err", err)
				return
			}

			var ordersData model.GetSiteOrderResponseBody
			err = json.Unmarshal([]byte(result), &ordersData)
			if err != nil {
				fmt.Println("JSON解码错误:", err)
				return
			}
			//var resultData []ReservationsDB
			var data model.ReservationsDB
			for _, reservation := range ordersData.Data.Hotel.PlatformReservations.Results {
				data.BookingId = reservation.SourceId
				if len(reservation.Profiles) > 0 {
					data.GuestName = reservation.Profiles[0].FirstName + " " + reservation.Profiles[0].LastName
				} else {
					//data.GuestName = reservation.RoomStays[0].Guests[0].FirstName + " " + reservation.RoomStays[0].Guests[0].LastName
				}

				arrivalTime, err := time.Parse("2006-01-02", reservation.FromDate)
				if err != nil {
					fmt.Println("Error parsing arrival time:", err)
				}

				departureTime, err := time.Parse("2006-01-02", reservation.CheckOutDate)
				if err != nil {
					fmt.Println("Error parsing arrival time:", err)
				}

				if reservation.PmsLastSentAt!=""{
					parsedTime, err := time.Parse(time.RFC3339, reservation.PmsLastSentAt)
					if err != nil {
						fmt.Println("Error parsing booktime:", err)
						return
					}
					resultTimeStr := parsedTime.Format("2006-01-02")
					data.BookDate = resultTimeStr
				} else if reservation.ChannelCreatedAt!=""{
					parsedTime, err := time.Parse(time.RFC3339, reservation.ChannelCreatedAt)
					if err != nil {
						fmt.Println("Error parsing booktime:", err)
						return
					}
					resultTimeStr := parsedTime.Format("2006-01-02")
					data.BookDate = resultTimeStr
				}else{
					data.BookDate = "err"
				}
				roomStayCount := len(reservation.RoomStays)								
				roomInfoData := make(map[string]*model.RoomInfo_site)
				for _, roomReservation := range reservation.RoomStays {
					roomType := roomReservation.RoomName
					roomInfo, ok := roomInfoData[roomType]
					if !ok {
						// 如果房间信息不存在，创建新的 RoomInfo
						roomInfo = &model.RoomInfo_site{
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
					if combinedRoomInfo != "" {
						combinedRoomInfo += " + "
					}
					combinedRoomInfo += fmt.Sprintf("%s*%s", roomInfo.RoomType, strconv.Itoa(roomInfo.Count))
				}
				data.RoomName = combinedRoomInfo

			
				checkOutTime := departureTime
				checkInTime := arrivalTime
				data.CheckOutDate = checkOutTime.Format("2006-01-02")
				data.CheckInDate = checkInTime.Format("2006-01-02")
				duration := checkOutTime.Sub(checkInTime)
				roomNights := int64(duration.Hours() / 24)
				data.Channelname = "siteminder"
				data.RoomNights = roomNights
				data.Price = reservation.TotalAmount.AmountAfterTax
				if reservation.Type == "Cancellation"{
					data.ReservationStatus = "已取消"
				}else{
					data.ReservationStatus = "已成立"
				}
				//data.ReservationStatus = reservation.Type
				data.Platform = reservation.Channel.Name
				data.Currency = reservation.Currency
				data.HotelId = hotelId
				data.NumOfRooms = int64(roomStayCount)
				resultData = append(resultData, data)
				//fmt.Println("checkout",data.CheckOutDate)
			}

		}

	}
	
	//fmt.Println("resultdata", resultData)

	resultDataJSON, err := json.Marshal(resultData)
	
	if err != nil {
		fmt.Println("JSON 轉換錯誤:", err)
		return
	}
	fmt.Println(resultData)
	fmt.Println(len(resultData))
	// 將資料存入DB
	var resultDB string
	apiurl := `http://149.28.24.90:8893/revenue_booking/setDataAnalysisToDB`
	if err := DoRequestAndGetResponse("POST", apiurl, bytes.NewBuffer(resultDataJSON), cookie, &resultDB); err != nil {
		fmt.Println("setParseHtmlToDB failed!")
		return
	}
	fmt.Println("資料下載完成")

}

func DoRequestAndGetResponse_sit(method, postUrl string, reqBody io.Reader, cookie string, x_xsrf_token string, resBody any) error {
	req, err := http.NewRequest(method, postUrl, reqBody)
	if err != nil {
		return err
	}
	req.Header.Set("x-xsrf-token", x_xsrf_token)
	//req.Header.Set("x-sm-trace-token","4fbf17b4-4c98-4337-98d7-c2f32395afd1")
	//req.Header.Set("user-agent","Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36")
	req.Header.Set("Cookie", cookie)
	req.Header.Set("Content-Type", "application/json")
	// switch resBody := resBody.(type) {
	// case *string:
	// 	fmt.Println("string")
	// 	fmt.Println(resBody)

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