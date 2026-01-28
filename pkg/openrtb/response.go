// Package openrtb provides OpenRTB 2.5 bid request and response types.
// It defines the core domain models for real-time bidding operations.
package openrtb

// BidResponse represents an OpenRTB 2.5 bid response.
type BidResponse struct {
	ID      string    `json:"id"`
	SeatBid []SeatBid `json:"seatbid,omitempty"`
	BidID   string    `json:"bidid,omitempty"`
	Cur     string    `json:"cur,omitempty"`
	NBR     int       `json:"nbr,omitempty"`
}

// SeatBid represents a collection of bids from a single seat.
type SeatBid struct {
	Bid  []Bid  `json:"bid"`
	Seat string `json:"seat,omitempty"`
}

// Bid represents a single bid.
type Bid struct {
	ID      string   `json:"id"`
	ImpID   string   `json:"impid"`
	Price   float64  `json:"price"`
	AdID    string   `json:"adid,omitempty"`
	NURL    string   `json:"nurl,omitempty"`
	AdM     string   `json:"adm,omitempty"`
	ADomain []string `json:"adomain,omitempty"`
	CID     string   `json:"cid,omitempty"`
	CrID    string   `json:"crid,omitempty"`
	Cat     []string `json:"cat,omitempty"`
	W       int      `json:"w,omitempty"`
	H       int      `json:"h,omitempty"`
}

// NoBidReason codes
const (
	NBRUnknown           = 0
	NBRTechnicalError    = 1
	NBRInvalidRequest    = 2
	NBRKnownSpider       = 3
	NBRSuspectedNonHuman = 4
	NBRCloudIP           = 5
	NBRUnsupportedDevice = 6
	NBRBlockedPublisher  = 7
	NBRUnmatchedUser     = 8
)

// IsNoBid returns true if the response contains no bids.
func (r *BidResponse) IsNoBid() bool {
	if len(r.SeatBid) == 0 {
		return true
	}
	for _, sb := range r.SeatBid {
		if len(sb.Bid) > 0 {
			return false
		}
	}
	return true
}

// AllBids returns a flattened slice of all bids across all seats.
func (r *BidResponse) AllBids() []Bid {
	// Pre-calculate total capacity to avoid reallocations
	totalBids := 0
	for _, sb := range r.SeatBid {
		totalBids += len(sb.Bid)
	}
	bids := make([]Bid, 0, totalBids)
	for _, sb := range r.SeatBid {
		bids = append(bids, sb.Bid...)
	}
	return bids
}

// HighestBid returns the bid with the highest price, or nil if no bids.
func (r *BidResponse) HighestBid() *Bid {
	var highest *Bid
	for i := range r.SeatBid {
		for j := range r.SeatBid[i].Bid {
			bid := &r.SeatBid[i].Bid[j]
			if highest == nil || bid.Price > highest.Price {
				highest = bid
			}
		}
	}
	return highest
}
