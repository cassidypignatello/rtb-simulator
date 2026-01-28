package scenarios

import (
	"math/rand/v2"
	"strconv"

	"github.com/cass/rtb-simulator/pkg/openrtb"
)

// Pre-computed version strings (1000 combinations: 0.0.0 to 9.9.9)
var versionStrings []string

// Connection types as a fixed-size array (zero allocation on access)
var connectionTypes = [...]int{
	openrtb.ConnectionWifi,
	openrtb.ConnectionCell4G,
	openrtb.ConnectionCell3G,
}

// Hex characters for user ID generation
const hexChars = "0123456789abcdef"

func init() {
	// Pre-compute all version strings at startup
	versionStrings = make([]string, 0, 1000)
	for major := 0; major < 10; major++ {
		for minor := 0; minor < 10; minor++ {
			for patch := 0; patch < 10; patch++ {
				versionStrings = append(versionStrings,
					strconv.Itoa(major)+"."+strconv.Itoa(minor)+"."+strconv.Itoa(patch))
			}
		}
	}
}

// MobileApp generates bid requests simulating mobile app inventory.
// Thread-safe: uses math/rand/v2 top-level functions which have per-OS-thread state.
type MobileApp struct{}

// NewMobileApp creates a new mobile app scenario.
func NewMobileApp() *MobileApp {
	return &MobileApp{}
}

func (m *MobileApp) Name() string {
	return "mobile_app"
}

func (m *MobileApp) Generate(requestID string) *openrtb.BidRequest {
	// No mutex needed - rand/v2 top-level functions are thread-safe
	device := m.randomDevice()
	app := m.randomApp()

	return &openrtb.BidRequest{
		ID: requestID,
		Imp: []openrtb.Imp{
			{
				ID:       "imp-1",
				Banner:   m.randomBanner(),
				BidFloor: m.randomBidFloor(),
				Secure:   1,
			},
		},
		App:    app,
		Device: device,
		User: &openrtb.User{
			ID: m.randomUserID(),
		},
		At:   openrtb.AuctionFirstPrice,
		Tmax: 100,
		Cur:  []string{"USD"},
	}
}

func (m *MobileApp) randomBanner() *openrtb.Banner {
	size := bannerSizes[rand.IntN(len(bannerSizes))]
	return &openrtb.Banner{
		W:   size.W,
		H:   size.H,
		Pos: rand.IntN(3), // 0=unknown, 1=above fold, 2=below fold
	}
}

func (m *MobileApp) randomApp() *openrtb.App {
	app := apps[rand.IntN(len(apps))]
	return &openrtb.App{
		ID:     m.randomAppID(),
		Name:   app.Name,
		Bundle: app.Bundle,
		Cat:    []string{app.Category},
		Ver:    versionStrings[rand.IntN(len(versionStrings))],
	}
}

func (m *MobileApp) randomDevice() *openrtb.Device {
	device := devices[rand.IntN(len(devices))]
	return &openrtb.Device{
		UA:             device.UA,
		IP:             m.randomIP(),
		Make:           device.Make,
		Model:          device.Model,
		OS:             device.OS,
		OSV:            device.OSV,
		DeviceType:     openrtb.DeviceTypePhone,
		ConnectionType: connectionTypes[rand.IntN(len(connectionTypes))],
		Language:       "en",
		Geo:            m.randomGeo(),
	}
}

func (m *MobileApp) randomGeo() *openrtb.Geo {
	geo := geoLocations[rand.IntN(len(geoLocations))]
	return &openrtb.Geo{
		Lat:     geo.Lat + (rand.Float64()-0.5)*0.1, // Add small variance
		Lon:     geo.Lon + (rand.Float64()-0.5)*0.1,
		Country: geo.Country,
		Region:  geo.Region,
		City:    geo.City,
	}
}

// randomIP generates a realistic-looking IP address using direct byte manipulation.
// Avoids fmt.Sprintf overhead.
func (m *MobileApp) randomIP() string {
	var buf [15]byte // Max: "223.255.255.254"
	n := 0

	// First octet: 1-223
	n += writeUint8(buf[n:], uint8(rand.IntN(223)+1))
	buf[n] = '.'
	n++

	// Second octet: 0-255
	n += writeUint8(buf[n:], uint8(rand.IntN(256)))
	buf[n] = '.'
	n++

	// Third octet: 0-255
	n += writeUint8(buf[n:], uint8(rand.IntN(256)))
	buf[n] = '.'
	n++

	// Fourth octet: 1-254
	n += writeUint8(buf[n:], uint8(rand.IntN(254)+1))

	return string(buf[:n])
}

// randomUserID generates a 32-character hex string without fmt.Sprintf.
func (m *MobileApp) randomUserID() string {
	var buf [32]byte
	for i := range buf {
		buf[i] = hexChars[rand.IntN(16)]
	}
	return string(buf[:])
}

// randomAppID generates an app ID like "app-123456" without fmt.Sprintf.
func (m *MobileApp) randomAppID() string {
	var buf [10]byte // "app-" + 6 digits
	copy(buf[:4], "app-")
	n := rand.IntN(1000000)
	for i := 9; i >= 4; i-- {
		buf[i] = '0' + byte(n%10)
		n /= 10
	}
	return string(buf[:])
}

func (m *MobileApp) randomBidFloor() float64 {
	// Bid floor between $0.25 and $3.00
	return 0.25 + rand.Float64()*2.75
}

// writeUint8 writes a uint8 to buf and returns the number of bytes written.
// This is faster than strconv.Itoa for small numbers.
func writeUint8(buf []byte, n uint8) int {
	if n >= 100 {
		buf[0] = '0' + n/100
		buf[1] = '0' + (n/10)%10
		buf[2] = '0' + n%10
		return 3
	} else if n >= 10 {
		buf[0] = '0' + n/10
		buf[1] = '0' + n%10
		return 2
	}
	buf[0] = '0' + n
	return 1
}

// Data pools for randomization

type bannerSize struct {
	W, H int
}

var bannerSizes = []bannerSize{
	{320, 50},  // Mobile leaderboard
	{300, 250}, // Medium rectangle
	{320, 480}, // Mobile interstitial
	{728, 90},  // Leaderboard (tablet)
	{300, 50},  // Mobile banner
}

type appInfo struct {
	Name     string
	Bundle   string
	Category string
}

var apps = []appInfo{
	{"Puzzle Quest", "com.games.puzzlequest", "IAB9-30"},
	{"Daily News", "com.news.dailynews", "IAB12"},
	{"Weather Pro", "com.weather.weatherpro", "IAB15"},
	{"Fitness Tracker", "com.health.fitnesstracker", "IAB7"},
	{"Social Chat", "com.social.chatapp", "IAB14"},
	{"Music Stream", "com.music.streamapp", "IAB1"},
	{"Photo Editor", "com.photo.editorpro", "IAB9"},
	{"Recipe Book", "com.food.recipebook", "IAB8"},
	{"Travel Guide", "com.travel.guidebook", "IAB20"},
	{"Finance Manager", "com.finance.manager", "IAB13"},
}

type deviceInfo struct {
	Make  string
	Model string
	OS    string
	OSV   string
	UA    string
}

var devices = []deviceInfo{
	{
		Make:  "Apple",
		Model: "iPhone14,2",
		OS:    "iOS",
		OSV:   "16.0",
		UA:    "Mozilla/5.0 (iPhone; CPU iPhone OS 16_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.0 Mobile/15E148 Safari/604.1",
	},
	{
		Make:  "Apple",
		Model: "iPhone15,2",
		OS:    "iOS",
		OSV:   "17.0",
		UA:    "Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1",
	},
	{
		Make:  "Samsung",
		Model: "SM-G998B",
		OS:    "Android",
		OSV:   "13",
		UA:    "Mozilla/5.0 (Linux; Android 13; SM-G998B) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/112.0.0.0 Mobile Safari/537.36",
	},
	{
		Make:  "Samsung",
		Model: "SM-S908B",
		OS:    "Android",
		OSV:   "14",
		UA:    "Mozilla/5.0 (Linux; Android 14; SM-S908B) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Mobile Safari/537.36",
	},
	{
		Make:  "Google",
		Model: "Pixel 7",
		OS:    "Android",
		OSV:   "14",
		UA:    "Mozilla/5.0 (Linux; Android 14; Pixel 7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Mobile Safari/537.36",
	},
	{
		Make:  "Xiaomi",
		Model: "2201116SG",
		OS:    "Android",
		OSV:   "13",
		UA:    "Mozilla/5.0 (Linux; Android 13; 2201116SG) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/112.0.0.0 Mobile Safari/537.36",
	},
}

type geoInfo struct {
	Lat     float64
	Lon     float64
	Country string
	Region  string
	City    string
}

var geoLocations = []geoInfo{
	{37.7749, -122.4194, "USA", "CA", "San Francisco"},
	{40.7128, -74.0060, "USA", "NY", "New York"},
	{34.0522, -118.2437, "USA", "CA", "Los Angeles"},
	{41.8781, -87.6298, "USA", "IL", "Chicago"},
	{29.7604, -95.3698, "USA", "TX", "Houston"},
	{33.4484, -112.0740, "USA", "AZ", "Phoenix"},
	{39.7392, -104.9903, "USA", "CO", "Denver"},
	{47.6062, -122.3321, "USA", "WA", "Seattle"},
	{25.7617, -80.1918, "USA", "FL", "Miami"},
	{42.3601, -71.0589, "USA", "MA", "Boston"},
}
