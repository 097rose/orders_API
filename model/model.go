package model
import (
	//"bufio"
	"encoding/xml"
)
type RoomInfo struct {
	RoomType  string
	StartDate string
	Days      int
}

type RoomInfo_site struct {
	RoomType string
	Count    int
}

type RoomInfo_owl struct {
	RoomType string

	Count int
}

type RoomInfo_360 struct {
	RoomType string
	Count    int
	Sd 		string
	Ed 		string
}

//住宿商戶資料表取得旅館價格計劃表的連結
type HotelGsInfo struct {
	HotelId   string
	HotelName string
	GsId      string
	Source	  string
}

//旅館要素表取得cookie, token, hotelID
type Hotel struct {
    ID                        string
    HotelIdOne                string
    HotelIdTwo                string
    HotelName                 string
    Channel                   string
    InventoryCookie           string
    InventoryAuth             string
    ReservationCookie         string
    ReservationAuth           string
}

//資料庫結構
type ReservationsDB struct {
	Platform     string `gorm:"uniqueIndex:platform_booking_id" json:"platform"`
	BookingId    string `gorm:"uniqueIndex:platform_booking_id" json:"booking_id"`
	BookDate     string `json:"book_date"`
	GuestName    string `json:"guest_name"`
	NumOfGuests  int64  `json:"num_of_guests"`
	CheckInDate  string `json:"check_in_date"`
	CheckOutDate string `json:"check_out_date"`
	Price             float64 `json:"price"`
	Currency          string  `json:"currency"`
	ReservationStatus string  `json:"reservation_status"`
	NumOfRooms        int64   `json:"num_of_rooms"`
	RoomNights int64  `json:"room_nights"`
	HotelId    string `json:"hotel_id"`
	RoomName    string `json:"room_name"`
	Channelname string `json:"channel_name"`
}

//第一次request url拿到的資料，要用order serial再進一步得到詳細訂單資訊
type GetOwltingOrderResponseBody struct {
	Data []struct {
		Order_serial string `json:"order_serial"`
		Order_status string `json:"order_status"`
		Created_at   string `json:"created_at"`
		Fullname     string `json:"fullname"`
		Room_names   string `json:"room_name"`
		Canceled_at  string `json:"canceled_at"`
		Sdate        string `json:"sdate"`
		Edate        string `json:"edate"`
		Source       string `json:"source"`
		Total        string `json:"total"`
	} `json:"data"`
	Pagination struct {
		Total_pages int `json:"total_pages"`
	} `json:"pagination"`
}
//詳細訂單資料
type GetOwltingOrderResponseBody2 struct {
	Data struct {
		Info struct {
			Order_serial     string `json:"order_serial"`
			Order_status     bool `json:"is_cancelled"`
			Orderer_fullname string `json:"orderer_fullname"`
			Source2          string `json:"order_source"`
			Source           string `json:"order_ota_full_name"`
			Sdate            string `json:"order_start_date"`
			Edate            string `json:"order_end_date"`
			Order_stay_night int    `json:"order_stay_night"`
		} `json:"info"`
		Rooms []struct {
			//Date   string  `json:"date"`
			Room_name        string `json:"room_name"`
			Room_config_name string `json:"room_config_name"`
		} `json:"rooms"`
		Summary struct {
			Hotel struct {
				Receivable_total float64 `json:"paid_total"`
			} `json:"hotel"`
		} `json:"summary"`

		First_payment struct {
			//Total        string  `json:"total"`
			Created_at string `json:"created_at"`
		} `json:"first_payment"`
	} `json:"data"`
}
//雲掌櫃的資料結構，但目前沒有串接雲掌櫃
type Get360OrderResponseBody struct {
	Orderdata struct {
		ID           string `json:"id"`
		Guestname    string `json:"guestname"`
		Roomtypename string `json:"roomtypename"`
		Roomname string `json:"roomname"`
		
		Arrivedate   string `json:"arrivedate"`
		Enddate      string `json:"enddate"`
		Channelname  string `json:"channelname"`
		Orderstatus  string `json:"orderstatus"`
		Createon     string `json:"createon"`
		Remark2      string `json:"remark2"`

		Orderset struct {
			Id string `json:"id"`
			Totalprice   string `json:"totalprice"`
		} `json:"orderset"`
	} `json:"orderdata"`
}

//大師的資料結構
type GetMasterOrderResponseBody[] struct {
	ID           string `json:"number"`
	Guestname    string `json:"name"`
	Roomname string `json:"check_in_room"`
	Rooms int `json:"rooms"`
	Channelname  string `json:"channelname"`
	Orderstatus  string `json:"status"`
	Totalprice   string `json:"invoice_amount"`
	Created_at   string `json:"created_at"`
	Sdate        string `json:"check_in"`
	Edate        string `json:"check_out"`
	Source       string `json:"source"`
}
//traiwan的資料結構，注意data是xml
type GetTraiwanOrderResponseBody struct {
	XMLName xml.Name `xml:"response"`
	Orders  struct {
		Order []struct {
			ID     string `xml:"id"`
			Person struct {
				Name string `xml:"name"`
			} `xml:"person"`
			Source            string `xml:"source"`
			Transaction_price string `xml:"transaction_price"`
			Room_reservations []struct {
				Room_type struct {
					Id   string `xml:"id"`
					Name string `xml:"name"`
				} `xml:"room_type"`
				Date string `xml:"date"`
			} ` xml:"room_reservations>room_reservation"`
			Delete_status  int    `xml:"delete_status"`
			Generated_time string `xml:"generated_time"`
		} `xml:"order"`
	} `xml:"orders"`
}
//新版sitminder
type GetSiteOrderResponseBody struct {
	Data struct {
		Hotel struct {
			Spid                 string `json:"spid"`
			PlatformReservations struct {
				Results []struct {
					SourceId string `json:"sourceId"`
					Channel  struct {
						Name string `json:"name"`
					} `json:"channel"`
					FromDate     string `json:"fromDate"`
					CheckOutDate string `json:"checkOutDate"`
					Currency     string `json:"currency"`
					TotalAmount  struct {
						AmountAfterTax float64 `json:"amountAfterTax"`
					} `json:"totalAmount"`
					RoomStays []struct {
						RoomName string `json:"roomName"`
						Guests   []struct {
							FirstName string `json:"firstName"`
							LastName  string `json:"lastName"`
						} `json:"guests"`
					} `json:"roomStays"`
					Profiles []struct {
						FirstName string `json:"firstName"`
						LastName  string `json:"lastName"`
					} `json:"profiles"`
					ChannelCreatedAt string `json:"channelCreatedAt"`
					
					PmsLastSentAt string `json:"pmsLastSentAt"`
					Type          string `json:"type"`
				} `json:"results"`
				Total int `json:"total"`
			} `json:"platformReservations"`
		} `json:"hotel"`
	} `json:"data"`
}
