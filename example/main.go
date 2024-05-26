package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/justtaldevelops/playfab"
	"github.com/sandertv/gophertunnel/minecraft/auth"
	"golang.org/x/oauth2"
)

func main() {
	src := tokenSource()
	db, err := playfab.New(http.DefaultClient, src)
	if err != nil {
		panic(err)
	}

	// the api only permits getting 300 items per request so we need to split it up into multiple requests
	const limit = 300
	offset := 0

	var allItems []interface{}
	var basePayload map[string]interface{}
	const outputFile = "test.json"

	for {
		resp, err := db.Search(playfab.Filter{
			Count:   true,
			Filter:  "(contentType eq 'PersonaDurable' and displayProperties/pieceType eq 'persona_emote')",
			OrderBy: "creationDate desc",
			SCID:    "4fc10100-5f7a-4470-899b-280835760c07",
			Limit:   limit,
			Skip:    offset,
		})

		if err != nil {
			panic(err)
		}

		items, ok := resp["Items"].([]interface{})
		if !ok {
			fmt.Println("Error: Unable to extract items from response.")
			break
		}

		// check if the number of items is 0 to determine if we fetched all the items
		if len(items) == 0 {
			break
		}

		if basePayload == nil {
			basePayload = resp
		} else {
			allItems = append(allItems, items...)
		}

		offset += limit
	}

	// combine the items into the base payload
	if basePayload != nil {
		baseItems, ok := basePayload["Items"].([]interface{})
		if !ok {
			fmt.Println("Error: Unable to extract base items from payload.")
		} else {
			basePayload["Items"] = append(baseItems, allItems...)
		}

		b, err := json.MarshalIndent(basePayload, "", "\t")
		if err != nil {
			panic(err)
		}

		err = os.WriteFile(outputFile, b, 0644)
		if err != nil {
			panic(err)
		}

		fmt.Printf("All data successfully written to %s.\n", outputFile)
	} else {
		fmt.Println("No data was fetched.")
	}
}

func tokenSource() oauth2.TokenSource {
	token := new(oauth2.Token)
	data, err := os.ReadFile("token.tok")
	if err == nil {
		_ = json.Unmarshal(data, token)
	} else {
		token, err = auth.RequestLiveToken()
		if err != nil {
			panic(err)
		}
	}
	src := auth.RefreshTokenSource(token)
	_, err = src.Token()
	if err != nil {
		token, err = auth.RequestLiveToken()
		if err != nil {
			panic(err)
		}
		src = auth.RefreshTokenSource(token)
	}
	tok, _ := src.Token()
	b, _ := json.Marshal(tok)
	_ = os.WriteFile("token.tok", b, 0644)
	return src
}
