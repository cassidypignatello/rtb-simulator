package scenarios

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/cass/rtb-simulator/pkg/openrtb"
)

// MobileApp generates bid requests simulating mobile app inventory.
type MobileApp struct {
	rng *rand.Rand
	mu  sync.Mutex
}

// NewMobileApp creates a new mobile app scenario.
func NewMobileApp() *MobileApp {
	return &MobileApp{
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (m *MobileApp) Name() string {
	return "mobile_app"
}

func (m *MobileApp) Generate(requestID string) *openrtb.BidRequest {
	m.mu.Lock()
	defer m.mu.Unlock()

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
	size := bannerSizes[m.rng.Intn(len(bannerSizes))]
	return &openrtb.Banner{
		W:   size.W,
		H:   size.H,
		Pos: m.rng.Intn(3), // 0=unknown, 1=above fold, 2=below fold
	}
}

func (m *MobileApp) randomApp() *openrtb.App {
	app := apps[m.rng.Intn(len(apps))]
	return &openrtb.App{
		ID:     fmt.Sprintf("app-%06d", m.rng.Intn(1000000)),
		Name:   app.Name,
		Bundle: app.Bundle,
		Cat:    []string{app.Category},
		Ver:    fmt.Sprintf("%d.%d.%d", m.rng.Intn(10), m.rng.Intn(10), m.rng.Intn(10)),
	}
}

func (m *MobileApp) randomDevice() *openrtb.Device {
	device := devices[m.rng.Intn(len(devices))]
	return &openrtb.Device{
		UA:             device.UA,
		IP:             m.randomIP(),
		Make:           device.Make,
		Model:          device.Model,
		OS:             device.OS,
		OSV:            device.OSV,
		DeviceType:     openrtb.DeviceTypePhone,
		ConnectionType: m.randomConnectionType(),
		Language:       "en",
		Geo:            m.randomGeo(),
	}
}

func (m *MobileApp) randomGeo() *openrtb.Geo {
	geo := geoLocations[m.rng.Intn(len(geoLocations))]
	return &openrtb.Geo{
		Lat:     geo.Lat + (m.rng.Float64()-0.5)*0.1, // Add small variance
		Lon:     geo.Lon + (m.rng.Float64()-0.5)*0.1,
		Country: geo.Country,
		Region:  geo.Region,
		City:    geo.City,
	}
}

func (m *MobileApp) randomIP() string {
	// Generate realistic-looking IPs (avoiding reserved ranges)
	return fmt.Sprintf("%d.%d.%d.%d",
		m.rng.Intn(223)+1, // 1-223 for first octet
		m.rng.Intn(256),
		m.rng.Intn(256),
		m.rng.Intn(254)+1, // 1-254 for last octet
	)
}

func (m *MobileApp) randomUserID() string {
	const chars = "abcdef0123456789"
	b := make([]byte, 32)
	for i := range b {
		b[i] = chars[m.rng.Intn(len(chars))]
	}
	return string(b)
}

func (m *MobileApp) randomBidFloor() float64 {
	// Bid floor between $0.25 and $3.00
	return 0.25 + m.rng.Float64()*2.75
}

func (m *MobileApp) randomConnectionType() int {
	types := []int{
		openrtb.ConnectionWifi,
		openrtb.ConnectionCell4G,
		openrtb.ConnectionCell3G,
	}
	return types[m.rng.Intn(len(types))]
}

// Data pools for randomization

type bannerSize struct {
	W, H int
}

var bannerSizes = []bannerSize{
	{320, 50},   // Mobile leaderboard
	{300, 250},  // Medium rectangle
	{320, 480},  // Mobile interstitial
	{728, 90},   // Leaderboard (tablet)
	{300, 50},   // Mobile banner
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
