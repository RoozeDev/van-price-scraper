package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/RoozeDev/van-price-scraper/internal/tools"
	"github.com/andybalholm/brotli"
	"go.mongodb.org/mongo-driver/bson"
)

type BookingBody struct {
	CheckInCity      string   `json:"checkin_city"`
	CheckInDateTime  string   `json:"checkin_datetime"`
	CheckOutCity     string   `json:"checkout_city"`
	CheckOutDateTime string   `json:"checkout_datetime"`
	VanCategories    []string `json:"van_categories"`
	VanCategory      string   `json:"van_category"`
	LegacySearch     bool     `json:"legacy_search"`
	Page             int      `json:"page"`
	Locale           string   `json:"locale"`
}

type RequestMeta struct {
	CurrentRoute string `json:"current_route"`
}

type RequestBody struct {
	Booking BookingBody `json:"booking"`
	Meta    RequestMeta `json:"meta"`
}

type ResponsePrices struct {
	TotalAmount float64 `json:"total_amount"`
}

type ResponseLocation struct {
	OriginCity      string `json:"origin_city"`
	DestinationCity string `json:"destination_city"`
}

type Availability struct {
	Available  bool    `json:"available"`
	TotalPrice float64 `json:"total_cost"`
	Location   string  `json:"location_name"`
	StartDate  string  `json:"checkin_date"`
	EndDate    string  `json:"checkout_date"`
	VanType    string  `json:"van_category"`
}

type Price struct {
	Supplier   string
	TotalPrice float64
	Location   string
	StartDate  string
	EndDate    string
	VanType    string
}

type ResponseBody struct {
	Data struct {
		Availability []Availability `json:"availability"`
	} `json:"data"`
}

var RequestHeaders = map[string]string{
	"Content-Type":    "application/json",
	"Content-Length":  "570",
	"Accept":          "application/json, text/plain, */*",
	"Accept-Language": "en-GB,en-US;q=0.9,en;q=0.8",
	"Accept-Encoding": "gzip, deflate, br",
	"Origin":          "https://indiecampers.co.uk",
	"Referer":         "https://indiecampers.co.uk/",
	"User-Agent":      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/116.0.0.0 Safari/537.36",
}

var VanCategories []string = []string{
	"active-long-2",
	"active-plus",
	"active-plus-2",
	"adventure-truck-2",
	"adventure-truck-3",
	"applause",
	"atlas",
	"atlas-5",
	"atlas-5-auto",
	"atlas-5-new",
	"california",
	"carado",
	"comfort-family-5",
	"comfort-long-4",
	"comfort-space-4",
	"comfort-space-8",
	"conquest",
	"conquest-plus",
	"duster",
	"etrusco",
	"etrusco-auto",
	"four-winds",
	"four-winds-plus",
	"id-buzz",
	"jimny",
	"joa",
	"jogger",
	"marco-polo",
	"maverick",
	"metris",
	"nomad",
	"nomad-auto",
	"nomad-ivy",
	"nomad-new",
	"nomad-pop-top",
	"nomad-pop-top-auto",
	"odyssey",
	"odyssey-pop-top",
	"odyssey-pop-top-auto",
	"outback",
	"ovation",
	"quest",
	"quest-auto",
	"rebel",
	"seeker",
	"sierra",
	"solis",
	"solis-air",
	"sporty",
	"tellaro",
	"trailer-towable",
	"trekker",
	"vw-california-premium",
	"vw-grand-california",
	"wander",
	"wrangler",
}

var BaseURL string = "https://indiecampers.com/api/v3/availability"

func main() {
	var bookingBody BookingBody = BookingBody{
		CheckInCity:      "anchorage",
		CheckInDateTime:  "2025-06-30T16:30:00+00:00",
		CheckOutCity:     "anchorage",
		CheckOutDateTime: "2025-07-07T11:00:00+00:00",
		VanCategories:    VanCategories,
		VanCategory:      "",
		LegacySearch:     false,
		Page:             1,
		Locale:           "en",
	}
	var meta RequestMeta = RequestMeta{
		CurrentRoute: "rent-an-rv-search",
	}
	var requestBody RequestBody = RequestBody{
		Booking: bookingBody,
		Meta:    meta,
	}
	body, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Println("Error decoding body")
		return
	}
	req, err := http.NewRequest(http.MethodPost, BaseURL, bytes.NewBuffer(body))
	if err != nil {
		fmt.Println("Error creating request", err)
		return
	}

	for k, v := range RequestHeaders {
		req.Header.Set(k, v)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Error response received", err)
	}
	defer resp.Body.Close()

	var responseData ResponseBody

	if resp.Header.Get("Content-Encoding") == "br" {
		// Wrap the response body with Brotli decompressor
		decompressedBody := brotli.NewReader(resp.Body)
		err = json.NewDecoder(decompressedBody).Decode(&responseData)
		if err != nil {
			fmt.Println("Error decoding decompressed body", err)
		}

	} else {
		err = json.NewDecoder(resp.Body).Decode(&responseData)
		fmt.Println("Error decoding body", err)
	}

	var results []Price

	for _, v := range responseData.Data.Availability {
		if v.Available {
			var result Price = Price{
				Supplier:   "IndieCampers",
				StartDate:  v.StartDate,
				EndDate:    v.EndDate,
				Location:   v.Location,
				TotalPrice: v.TotalPrice,
				VanType:    v.VanType,
			}
			results = append(results, result)
		}
	}

	databaseName := tools.GetDatabaseName()
	uri := tools.GetDatabaseConnectionURL()
	client := tools.GetDatabaseClient(uri)
	defer tools.CloseConnection(client)

	databaseConnection := tools.GetDatabaseConnection(client, databaseName)

	coll := databaseConnection.Collection("prices")

	dbresult, err := coll.InsertMany(context.TODO(), results)
	tools.Check(err)

	fmt.Println(dbresult.InsertedIDs...)

	filter := bson.M{}

	cursor, err := coll.Find(context.TODO(), filter)
	tools.Check(err)

	var dbResults []Price
	if err = cursor.All(context.TODO(), &dbResults); err != nil {
		panic(err)
	}
	for _, result := range dbResults {
		res, _ := bson.MarshalExtJSON(result, false, false)
		fmt.Println(string(res))
	}

}
