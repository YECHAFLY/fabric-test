/*
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"encoding/json"
	"fmt"
	flogging "github.com/Hnampk/fabric-flogging"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type BidRes struct {
	Units  int     `json:"units"`
	Prices float64 `json:"prices"`
}

type AuctionContract struct {
	contractapi.Contract
}

func (c *AuctionContract) AuctionExists(ctx contractapi.TransactionContextInterface, auctionID string) (bool, error) {
	data, err := ctx.GetStub().GetState(auctionID)
	if err != nil {
		return false, err
	}
	return data != nil, nil
}

func (c *AuctionContract) InitFeedbackSystem(ctx contractapi.TransactionContextInterface, addr string) error {
	if addr != "auctioneer" {
		return fmt.Errorf("illegal user %s tries to build feedback system", addr)
	}
	accounts := new(Accounts)
	bytes, _ := json.Marshal(accounts)
	return ctx.GetStub().PutState("acc", bytes)
}

func (c *AuctionContract) RegisterAccount(ctx contractapi.TransactionContextInterface, addr string, balance float64) error {
	bytes, _ := ctx.GetStub().GetState("acc")
	accounts := new(Accounts)
	err := json.Unmarshal(bytes, accounts)
	if err != nil {
		return fmt.Errorf("could not unmarshal world state data to type Auction")
	}
	// The element to search for
	element := Hash(addr)
	// Iterate over the elements of the list, it can be finished in a simpler way
	for _, e := range accounts.Address {
		if e == element {
			return fmt.Errorf("we have already registered this account. %s", addr)
		}
	}
	accounts.Address = append(accounts.Address, Hash(addr))
	accounts.Balance = append(accounts.Balance, balance)
	bytes, _ = json.Marshal(accounts)
	err = ctx.GetStub().PutState("acc", bytes)
	return err
}

// CreateAuction creates a new instance of Auction
func (c *AuctionContract) CreateAuction(ctx contractapi.TransactionContextInterface, auctionID string) error {
	exists, err := c.AuctionExists(ctx, auctionID)
	if err != nil {
		return fmt.Errorf("could not read from world state. %s", err)
	} else if exists {
		return fmt.Errorf("the auction %s already exists", auctionID)
	}

	auction := new(Auction)
	auction.Closed = false
	bytes, _ := json.Marshal(auction)
	return ctx.GetStub().PutState(auctionID, bytes)
}

// QueryAuction retrieves an instance of Auction from the world state
func (c *AuctionContract) QueryAuction(ctx contractapi.TransactionContextInterface, auctionID string) (string, error) {
	exists, err := c.AuctionExists(ctx, auctionID)
	if err != nil {
		return "", fmt.Errorf("could not read from world state. %s", err)
	} else if !exists {
		return "", fmt.Errorf("the auction %s does not exist", auctionID)
	}

	bytes, _ := ctx.GetStub().GetState(auctionID)
	auction := new(Auction)
	err = json.Unmarshal(bytes, auction)

	if err != nil {
		return "", fmt.Errorf("could not unmarshal world state data to type Auction")
	}

	return string(bytes), nil
}

// QueryAuction retrieves an instance of Auction from the world state
func (c *AuctionContract) QueryAccounts(ctx contractapi.TransactionContextInterface) (string, error) {
	var logger = flogging.MustGetLogger("fabric-double-auction")
	bytes, _ := ctx.GetStub().GetState("acc")
	accounts := new(Accounts)
	err := json.Unmarshal(bytes, accounts)
	if err != nil {
		return "", fmt.Errorf("could not unmarshal world state data to type Auction")
	}
	arrayLength := len(accounts.Address)
	logger.Error("arrayLength: ", arrayLength)
	return string(arrayLength), nil
}

// DeleteAuction deletes an instance of Auction from the world state
func (c *AuctionContract) CloseAuction(ctx contractapi.TransactionContextInterface, auctionID string) error {
	exists, err := c.AuctionExists(ctx, auctionID)
	if err != nil {
		return fmt.Errorf("could not read from world state. %s", err)
	} else if !exists {
		return fmt.Errorf("the auction %s does not exist", auctionID)
	}

	return ctx.GetStub().DelState(auctionID)
}

func (c *AuctionContract) Bid(ctx contractapi.TransactionContextInterface, auctionID string, prices string, quantities string, addr string) error {
	bytes, _ := ctx.GetStub().GetState("acc")
	accounts := new(Accounts)
	err := json.Unmarshal(bytes, accounts)
	if err != nil {
		return fmt.Errorf("could not unmarshal world state data to type Auction")
	}

	exists, err := c.AuctionExists(ctx, auctionID)
	if err != nil {
		return fmt.Errorf("could not read from world state. %s", err)
	} else if !exists {
		return fmt.Errorf("no %s auction", auctionID)
	}

	bytes, _ = ctx.GetStub().GetState(auctionID)
	auction := new(Auction)
	json.Unmarshal(bytes, auction)
	found := false
	// The element to search for
	element := Hash(addr)
	// Iterate over the elements of the list, it can be finished in a simpler way
	for _, e := range accounts.Address {
		if e == element {
			found = true
			break
		}
	}
	if found == true {
		AddBid(Hash(addr), StrToFloatArr(prices), StrToIntArr(quantities), auction)
		newBytes, _ := json.Marshal(auction)
		return ctx.GetStub().PutState(auctionID, newBytes)
	} else {
		return fmt.Errorf("Account cannot found in the system")
	}
}

// Claer all bids
//func (c *AuctionContract) ClearBids(ctx contractapi.TransactionContextInterface, auctionID string) error {
//	exists, err := c.AuctionExists(ctx, auctionID)
//	if err != nil {
//		return fmt.Errorf("could not read from world state. %s", err)
//	} else if !exists {
//		return fmt.Errorf("no %s auction", auctionID)
//	}
//
//	err = ctx.GetStub().DelState(auctionID)
//	return err
//}

// Claer all bids
func (c *AuctionContract) ClearBids(ctx contractapi.TransactionContextInterface, auctionID string) error {
	exists, err := c.AuctionExists(ctx, auctionID)
	if err != nil {
		return fmt.Errorf("could not read from world state. %s", err)
	} else if !exists {
		return fmt.Errorf("no %s auction", auctionID)
	}
	bytes, _ := ctx.GetStub().GetState(auctionID)
	auction := new(Auction)
	json.Unmarshal(bytes, auction)
	auction.Buyers = []BuyerBid{}
	auction.Sellers = []SellerBid{}
	auction.BuyersPay = [100]float64{}
	auction.SellersPay = [100]float64{}
	bytes, _ = json.Marshal(auction)
	ctx.GetStub().PutState(auctionID, bytes)
	return err
}

func (c *AuctionContract) Withdraw(ctx contractapi.TransactionContextInterface, auctionID string, addr string) (string, error) {
	var logger = flogging.MustGetLogger("fabric-double-auction")
	var res string
	exists, err := c.AuctionExists(ctx, auctionID)
	if err != nil {
		return res, fmt.Errorf("could not read from world state. %s", err)
	} else if !exists {
		return res, fmt.Errorf("no %s auction", auctionID)
	}

	bytes, _ := ctx.GetStub().GetState(auctionID)
	auction := new(Auction)
	json.Unmarshal(bytes, auction)
	if !auction.Closed {
		// winners := Allocate(auction.Buyers, auction.Sellers)
		buyersPay, sellersPay, clearedUnits, clearedPrice := DeterminePayment(auction.Buyers, auction.Sellers)
		bytes, _ = ctx.GetStub().GetState("acc")
		accounts := new(Accounts)
		err = json.Unmarshal(bytes, accounts)
		if err != nil {
			return res, fmt.Errorf("could not unmarshal world state data to type Auction")
		}
		logger.Error("-----------------------------------------")
		logger.Error("Buyer", auction.Buyers)
		logger.Error("-----------------------------------------")
		logger.Error("-----------------------------------------")
		logger.Error("Seller", auction.Sellers)
		logger.Error("-----------------------------------------")
		for i := 0; i < len(auction.Buyers); i++ {
			baddr := auction.Buyers[i].Address
			auction.BuyersPay[i] = buyersPay[i]
			ChangeBalance(baddr, -buyersPay[i], accounts)
		}
		for i := 0; i < len(auction.Sellers); i++ {
			saddr := auction.Sellers[i].Address
			auction.SellersPay[i] = sellersPay[i]
			ChangeBalance(saddr, sellersPay[i], accounts)
		}
		bytes, _ = json.Marshal(accounts)
		ctx.GetStub().PutState("acc", bytes)
		bidRes := &BidRes{
			Units:  clearedUnits,
			Prices: clearedPrice,
		}
		// Encode the data as a pretty-printed JSON string
		data, err := json.MarshalIndent(bidRes, "", "    ")
		if err != nil {
			panic(err)
		}
		return string(data), err
	}

	return res, err
}
