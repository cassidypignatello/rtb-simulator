package openrtb

// BidRequest represents an OpenRTB 2.5 bid request.
type BidRequest struct {
	ID     string   `json:"id"`
	Imp    []Imp    `json:"imp"`
	App    *App     `json:"app,omitempty"`
	Site   *Site    `json:"site,omitempty"`
	Device *Device  `json:"device,omitempty"`
	User   *User    `json:"user,omitempty"`
	At     int      `json:"at"`
	Tmax   int      `json:"tmax"`
	Cur    []string `json:"cur,omitempty"`
	Bcat   []string `json:"bcat,omitempty"`
}

// Imp represents an impression object.
type Imp struct {
	ID       string  `json:"id"`
	Banner   *Banner `json:"banner,omitempty"`
	Video    *Video  `json:"video,omitempty"`
	BidFloor float64 `json:"bidfloor"`
	Secure   int     `json:"secure,omitempty"`
	Tagid    string  `json:"tagid,omitempty"`
}

// Banner represents a banner impression.
type Banner struct {
	W     int   `json:"w,omitempty"`
	H     int   `json:"h,omitempty"`
	Wmax  int   `json:"wmax,omitempty"`
	Hmax  int   `json:"hmax,omitempty"`
	Wmin  int   `json:"wmin,omitempty"`
	Hmin  int   `json:"hmin,omitempty"`
	Btype []int `json:"btype,omitempty"`
	Battr []int `json:"battr,omitempty"`
	Pos   int   `json:"pos,omitempty"`
}

// Video represents a video impression (placeholder for future use).
type Video struct {
	Mimes       []string `json:"mimes,omitempty"`
	Minduration int      `json:"minduration,omitempty"`
	Maxduration int      `json:"maxduration,omitempty"`
	W           int      `json:"w,omitempty"`
	H           int      `json:"h,omitempty"`
}

// App represents an application object.
type App struct {
	ID       string   `json:"id,omitempty"`
	Name     string   `json:"name,omitempty"`
	Bundle   string   `json:"bundle,omitempty"`
	Domain   string   `json:"domain,omitempty"`
	StoreURL string   `json:"storeurl,omitempty"`
	Cat      []string `json:"cat,omitempty"`
	Ver      string   `json:"ver,omitempty"`
	Paid     int      `json:"paid,omitempty"`
}

// Site represents a website object (placeholder for future use).
type Site struct {
	ID     string   `json:"id,omitempty"`
	Name   string   `json:"name,omitempty"`
	Domain string   `json:"domain,omitempty"`
	Page   string   `json:"page,omitempty"`
	Cat    []string `json:"cat,omitempty"`
}

// Device represents device information.
type Device struct {
	UA           string `json:"ua,omitempty"`
	IP           string `json:"ip,omitempty"`
	Geo          *Geo   `json:"geo,omitempty"`
	Make         string `json:"make,omitempty"`
	Model        string `json:"model,omitempty"`
	OS           string `json:"os,omitempty"`
	OSV          string `json:"osv,omitempty"`
	DeviceType   int    `json:"devicetype,omitempty"`
	Carrier      string `json:"carrier,omitempty"`
	Language     string `json:"language,omitempty"`
	IFA          string `json:"ifa,omitempty"`
	ConnectionType int  `json:"connectiontype,omitempty"`
}

// Geo represents geographic location.
type Geo struct {
	Lat     float64 `json:"lat,omitempty"`
	Lon     float64 `json:"lon,omitempty"`
	Country string  `json:"country,omitempty"`
	Region  string  `json:"region,omitempty"`
	City    string  `json:"city,omitempty"`
	ZIP     string  `json:"zip,omitempty"`
	Type    int     `json:"type,omitempty"`
}

// User represents user information.
type User struct {
	ID       string `json:"id,omitempty"`
	BuyerUID string `json:"buyeruid,omitempty"`
	Gender   string `json:"gender,omitempty"`
	Yob      int    `json:"yob,omitempty"`
}

// Auction types
const (
	AuctionFirstPrice  = 1
	AuctionSecondPrice = 2
)

// Device types
const (
	DeviceTypeMobile  = 1
	DeviceTypePC      = 2
	DeviceTypeTV      = 3
	DeviceTypePhone   = 4
	DeviceTypeTablet  = 5
	DeviceTypeWatch   = 6
)

// Connection types
const (
	ConnectionUnknown  = 0
	ConnectionEthernet = 1
	ConnectionWifi     = 2
	ConnectionCell     = 3
	ConnectionCell2G   = 4
	ConnectionCell3G   = 5
	ConnectionCell4G   = 6
)
