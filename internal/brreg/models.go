package brreg

import "encoding/json"

type Underenhet struct {
	Name          string `json:"navn"`
	Orgnummer     string `json:"organisasjonsnummer"`
	Parent        string `json:"overordnetEnhet"`
	Adresse       string
	Postnummer    string
	Sted          string
	Kommune       string
	Kommunenummer string
}

func (u *Underenhet) UnmarshalJSON(data []byte) error {
	var raw struct {
		Name                string `json:"navn"`
		Orgnummer           string `json:"organisasjonsnummer"`
		Parent              string `json:"overordnetEnhet"`
		Beliggenhetsadresse struct {
			Postnummer    string   `json:"postnummer"`
			Poststed      string   `json:"poststed"`
			Adresse       []string `json:"adresse"`
			Kommune       string   `json:"kommune"`
			Kommunenummer string   `json:"kommunenummer"`
		} `json:"beliggenhetsadresse"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	u.Name = raw.Name
	u.Orgnummer = raw.Orgnummer
	u.Parent = raw.Parent
	if len(raw.Beliggenhetsadresse.Adresse) > 0 {
		u.Adresse = raw.Beliggenhetsadresse.Adresse[0]
	}
	u.Postnummer = raw.Beliggenhetsadresse.Postnummer
	u.Sted = raw.Beliggenhetsadresse.Poststed
	u.Kommune = raw.Beliggenhetsadresse.Kommune
	u.Kommunenummer = raw.Beliggenhetsadresse.Kommunenummer

	return nil
}

type Hovedenhet struct {
	Name                 string `json:"navn"`
	Orgnummer            string `json:"organisasjonsnummer"`
	Adresse              string
	Postnummer           string
	Sted                 string
	Kommune              string
	Kommunenummer        string
	Konkurs              bool
	UnderAvvikling       bool
	UnderTvangsavvikling bool
	Stiftelsesdato       string
}

func (e *Hovedenhet) UnmarshalJSON(data []byte) error {
	var raw struct {
		Name               string `json:"navn"`
		Orgnummer          string `json:"organisasjonsnummer"`
		Forretningsadresse struct {
			Postnummer    string   `json:"postnummer"`
			Poststed      string   `json:"poststed"`
			Adresse       []string `json:"adresse"`
			Kommune       string   `json:"kommune"`
			Kommunenummer string   `json:"kommunenummer"`
		} `json:"forretningsadresse"`
		Konkurs              bool   `json:"konkurs"`
		UnderAvvikling       bool   `json:"underAvvikling"`
		UnderTvangsavvikling bool   `json:"underTvangsavviklingEllerTvangsopplosning"`
		Stiftelsesdato       string `json:"vedtektsdato"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	e.Name = raw.Name
	e.Orgnummer = raw.Orgnummer
	if len(raw.Forretningsadresse.Adresse) > 0 {
		e.Adresse = raw.Forretningsadresse.Adresse[0]
	}
	e.Postnummer = raw.Forretningsadresse.Postnummer
	e.Sted = raw.Forretningsadresse.Poststed
	e.Kommune = raw.Forretningsadresse.Kommune
	e.Kommunenummer = raw.Forretningsadresse.Kommunenummer
	e.Konkurs = raw.Konkurs
	e.UnderAvvikling = raw.UnderAvvikling
	e.UnderTvangsavvikling = raw.UnderTvangsavvikling

	return nil
}
