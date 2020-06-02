package main

//
// List case structure
type ResultInfo struct {
	Portrait     string `json:"portrait"`
	FullName     string `json:"fullName"`
	BirthDate    int64  `json:"birthDate"`
	CaseId       string `json:"caseId"`
	AgencyCode   string `json:"agencyCode"`
	Status       string `json:"status"`
	Public       bool   `json:"public"`
	Type         string `json:"type"`
	State        string `json:"state"`
	City         string `json:"city"`
	MissingSince int64  `json:"missingSince"`
	Country      string `json:"country"`
	OpenDate     int64  `json:"openDate"`
	CreateDate   int64  `json:"createDate"`
	LastUpdate   int64  `json:"lastUpdate"`
	ChildId      string `json:"childId"`
}

type CaseInfo struct {
	Total   int          `json:"total"`
	Results []ResultInfo `json:"results"`
}

type SearchCasesResult struct {
	Cases CaseInfo `json:"cases"`
}

//
// Detailed case structure
type DetailedChildImage struct {
	Portrait string `json:"portrait"`
}

type DetailedChildInfo struct {
	ChildId     string             `json:"childId"`
	FullName    string             `json:"fullName"`
	BirthDate   int64              `json:"birthDate"`
	Sex         string             `json:"sex"`
	EyeColor    string             `json:"eyeColor"`
	HairColor   string             `json:"hairColor"`
	Height      string             `json:"height"`
	HeightUnit  string             `json:"heightUnit"`
	Weight      string             `json:"weight"`
	WeightUnit  string             `json:"weightUnit"`
	MissingDate int64              `json:"missingDate"`
	Images      DetailedChildImage `json:"images"`
}

type CenterCaseInfo struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type DetailedCaseInfo struct {
	AgencyCode         string              `json:"agencyCode"`
	CaseId             string              `json:"caseId"`
	CaseType           string              `json:"caseType"`
	Children           []DetailedChildInfo `json:"children"`
	Circumstances      string              `json:"circumstances"`
	City               string              `json:"city"`
	ContactInformation string              `json:"contactInformation"`
	Country            string              `json:"country"`
	CreateDate         int64               `json:"createDate"`
	Miscellaneous      map[string]string   `json:"miscellaneous"`
	MissingDate        int64               `json:"missingDate"`
	OpenDate           int64               `json:"openDate"`
	Poster             string              `json:"poster"`
	Public             bool                `json:"public"`
	State              string              `json:"state"`
	Status             string              `json:"status"`
	Etl                bool                `json:"etl"`
	Center             CenterCaseInfo      `json:"center"`
	LastUpdate         int64               `json:"lastUpdate"`
}

type DetailedCaseResult struct {
	Case DetailedCaseInfo `json:"case"`
}
