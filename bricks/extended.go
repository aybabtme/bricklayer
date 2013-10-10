package bricks

import (
	"encoding/xml"
	"fmt"
	"net/http"
)

const (
	fullBrickApiUrl = "http://parts.igem.org/cgi/xml/part.cgi?part=%s"
)

type ExtendedBiobrick struct {
	ID                int            `xml:"part_id" json:"id"`
	Name              string         `xml:"part_name" json:"name"`
	ShortName         string         `xml:"part_short_name" json:"short_name"`
	Desc              string         `xml:"part_short_desc" json:"short_desc"`
	FullType          string         `xml:"part_type" json:"type"`
	ReleaseStatus     string         `xml:"release_status" json:"release_status"`
	SampleStatus      string         `xml:"sample_status" json:"sample_status"`
	Results           string         `xml:"part_results" json:"results"`
	Nickname          string         `xml:"part_nickname" json:"nickname"`
	Rating            int            `xml:"part_rating" json:"rating"`
	URL               string         `xml:"part_url" json:"url"`
	Entered           string         `xml:"part_entered" json:"entered"`
	Author            string         `xml:"part_author" json:"author"`
	DeepSubParts      []string       `xml:"deep_subparts" json:"deep_subparts"`
	SpecifiedSubParts []string       `xml:"specified_subparts" json:"specified_subparts"`
	SpecifiedSubScars []string       `xml:"specified_subscars" json:"specified_subscars"`
	Features          []BioFeature   `xml:"features>feature" json:"features"`
	Parameters        []BioParameter `xml:"parameters>parameter" json:"parameters"`
	Categories        []string       `xml:"categories>category" json:"categories"`
	Twins             []string       `xml:"twins>twin" json:"twins"`
}

type BioFeature struct {
	ID        int    `xml:"id" json:"id"`
	Title     string `xml:"title" json:"title"`
	Type      string `xml:"type" json:"type"`
	Direction string `xml:"direction" json:"direction"`
	StartPos  int    `xml:"startpos" json:"startpos"`
	EndPos    int    `xml:"endpos" json:"endpos"`
}

type BioParameter struct {
	ID       int    `xml:"id" json:"id"`
	Name     string `xml:"name" json:"name"`
	Value    string `xml:"value" json:"value"`
	Units    string `xml:"units" json:"units"`
	URL      string `xml:"url" json:"url"`
	Date     string `xml:"m_date" json:"m_date"`
	UserID   int    `xml:"user_id" json:"user_id"`
	Username string `xml:"user_name" json:"user_name"`
}

func QueryExtendedBiobricks(partname string) ([]ExtendedBiobrick, error) {
	resp, err := http.Get(fmt.Sprintf(fullBrickApiUrl, partname))
	if err != nil {
		return nil, fmt.Errorf("getting part '%s', %v", partname, err)
	}
	defer resp.Body.Close()

	xmlDec := xml.NewDecoder(resp.Body)

	var rsbpml struct {
		XMLName xml.Name           `xml:"rsbpml"`
		Parts   []ExtendedBiobrick `xml:"part_list>part"`
	}

	err = xmlDec.Decode(&rsbpml)
	if err != nil {
		return nil, fmt.Errorf("decoding XML stream, %v", err)
	}

	return rsbpml.Parts, nil
}
