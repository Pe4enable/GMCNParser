package main

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
)

func getData() (content string, err error) {
	// Read from file
	if *InDataFile != "" {
		f, err := os.Open(*InDataFile)
		if err != nil {
			return "", fmt.Errorf("cannot read datafile [%s] with error: %v", *InDataFile, err)
		}
		defer f.Close()

		r, err := ioutil.ReadAll(f)
		if err != nil {
			return "", fmt.Errorf("error during reading datafile [%s] with error: %v", *InDataFile, err)
		}
		return string(r), nil
	}

	// Read from network
	req, err := http.NewRequest("POST", *URL, bytes.NewBuffer([]byte(*SearchString)))
	if err != nil {
		return "", fmt.Errorf("cannot init HTTP Request to %s with error: %v", *URL, err)
	}

	// Fill headers
	req.Header.Set("Content-type", "application/json;charset=utf-8")
	if *Origin != "" {
		req.Header.Set("Origin", *Origin)
	}
	if *Referer != "" {
		req.Header.Set("Referer", *Referer)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error during HTTP request to page [%s] with error: %v", *URL, err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	return string(body), nil
}

func saveData(fname string, content string) error {
	f, err := os.OpenFile(fname, os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		return err
	}
	_, err = f.WriteString(content)
	defer f.Close()

	return err
}

// Send data into processing queue
func sender(ctx context.Context, r []ResultInfo, out chan ResultInfo) {
	for _, item := range r {
		// Check for termination request
		select {
		case <-ctx.Done():
			return
		default:
		}

		// Send next item into processing channel
		out <- item
	}
}

// Process data and write output
func resolver(ctx context.Context, in chan ResultInfo, out chan []string, outerr chan error) {
	for {
		select {
		case r := <-in:
			fmt.Printf("> [%s] resolving case for child [%s]\n", r.CaseId, r.ChildId)
			data, err := getChildInfo(r.CaseId)
			if err != nil {
				outerr <- fmt.Errorf("error downloading case [%s] child [%s] info: %v", r.CaseId, r.ChildId, err)
				break
			}

			var resultData DetailedCaseResult
			err = json.Unmarshal([]byte(data), &resultData)
			if err != nil {
				outerr <- fmt.Errorf("error JSON unmarshal for case [%s] child [%s] info: %v", r.CaseId, r.ChildId, err)
				break
			}
			additionalPicURL := ""
			additionalPicBase64 := ""
			for k, _ := range resultData.Case.Miscellaneous {
				additionalPicURL = k
				break
			}

			// Case is closed
			if resultData.Case.Status == "closed" {
				outerr <- fmt.Errorf("case [%s] for child [%s] is closed", r.CaseId, r.ChildId)
				break
			}

			// Return an error if Children entry is not available
			if len(resultData.Case.Children) < 1 {
				outerr <- fmt.Errorf("no JSON children is available for case [%s] child [%s]", r.CaseId, r.ChildId)
				break
			}
			chld := resultData.Case.Children[0]
			picURL := chld.Images.Portrait
			picBase64 := ""
			if picURL != "" {
				_, picBase64, _ = downloadImage(*CacheDir, picURL)
			}
			if additionalPicURL != "" {
				_, additionalPicBase64, _ = downloadImage(*CacheDir, additionalPicURL)
			}

			output := []string{
				r.ChildId,                             // ID
				r.FullName,                            // Name
				strconv.FormatInt(r.MissingSince, 10), // DateOfCase
				fmt.Sprintf("%s,%s,%s", r.Country, r.State, r.City), // PlaceOfCase
				picURL,              // PicURL
				picBase64,           // PicBase64
				additionalPicURL,    // AdditionalPicURL
				additionalPicBase64, // AdditionalPicBase64
				strconv.FormatInt(chld.BirthDate.int64, 10), // DateOfBirth
				"-",            // PlaceOfBirth
				chld.HairColor, // Hair
				chld.EyeColor,  // Eyes
				fmt.Sprintf("%s %s", chld.Height, chld.HeightUnit), // Height
				fmt.Sprintf("%s %s", chld.Weight, chld.WeightUnit), // Weight
				chld.Sex,                                // Sex
				"",                                      // Race
				"",                                      // Nationality
				"",                                      // Reward
				"",                                      // Remarks
				"",                                      // Details
				"",                                      // Field Office
				"",                                      // Related Case
				fmt.Sprintf("%s%s", *URLCase, r.CaseId), // Source
			}
			out <- output
		case <-ctx.Done():
			return
		}
	}
}

//
func getChildInfo(caseID string) (string, error) {
	url := fmt.Sprintf("%s%s", *URLCase, caseID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("cannot init HTTP Request to %s with error: %v", url, err)
	}

	// Fill headers
	req.Header.Set("Content-type", "application/json;charset=utf-8")
	if *Origin != "" {
		req.Header.Set("Origin", *Origin)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error during HTTP request to page [%s] with error: %v", *URL, err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	return string(body), nil

}

func downloadImage(cacheDir, url string) (cacheFileName, dataBase64 string, err error) {
	// Check for cached image
	if cacheDir != "" {
		cacheFileName = fmt.Sprintf("%s/%x", cacheDir, sha1.Sum([]byte(url)))

		f, err := os.Open(cacheFileName)
		if err == nil {
			defer f.Close()
			r, err := ioutil.ReadAll(f)
			if err == nil {
				return cacheFileName, base64.StdEncoding.EncodeToString(r), nil
			}
		}
	}

	resp, err := http.Get(url)
	if err != nil {
		return "", "", fmt.Errorf("error downloading image: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("error during readall operation for image: %v", err)
	}

	if cacheDir != "" {
		if saveData(cacheFileName, string(body)) == nil {
			return cacheFileName, base64.StdEncoding.EncodeToString([]byte(body)), nil
		}
	}
	return "", base64.StdEncoding.EncodeToString([]byte(body)), nil
}
