package auction

import (
	"github.com/cass/rtb-simulator/internal/dispatcher"
	"github.com/cass/rtb-simulator/pkg/openrtb"
)

// Outcome represents the result of an auction.
type Outcome struct {
	RequestID     string
	Winner        *openrtb.Bid
	WinningDSP    string
	ClearingPrice float64
	AllBids       []BidWithDSP
}

// BidWithDSP associates a bid with its originating DSP.
type BidWithDSP struct {
	Bid     openrtb.Bid
	DSPName string
}

// Auction defines the interface for auction implementations.
type Auction interface {
	Run(requestID string, bidFloor float64, results []dispatcher.Result) Outcome
}

// FirstPrice implements a first-price auction where the highest bidder wins
// and pays their bid price.
type FirstPrice struct{}

// NewFirstPrice creates a new first-price auction.
func NewFirstPrice() *FirstPrice {
	return &FirstPrice{}
}

// Run executes the first-price auction on the given results.
func (a *FirstPrice) Run(requestID string, bidFloor float64, results []dispatcher.Result) Outcome {
	outcome := Outcome{RequestID: requestID}

	// Collect all eligible bids (above floor, no errors)
	var eligibleBids []BidWithDSP

	for _, r := range results {
		if r.Error != nil || r.Response == nil {
			continue
		}

		for _, sb := range r.Response.SeatBid {
			for _, bid := range sb.Bid {
				if bid.Price >= bidFloor {
					eligibleBids = append(eligibleBids, BidWithDSP{
						Bid:     bid,
						DSPName: r.DSPName,
					})
				}
			}
		}
	}

	outcome.AllBids = eligibleBids

	if len(eligibleBids) == 0 {
		return outcome
	}

	// Find the highest bid
	var highestIdx int
	for i, b := range eligibleBids {
		if b.Bid.Price > eligibleBids[highestIdx].Bid.Price {
			highestIdx = i
		}
	}

	winner := eligibleBids[highestIdx]
	outcome.Winner = &winner.Bid
	outcome.WinningDSP = winner.DSPName
	outcome.ClearingPrice = winner.Bid.Price // First-price: pay what you bid

	return outcome
}
