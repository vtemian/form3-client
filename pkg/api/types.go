package api

type Resource struct {
	Type    string `json:"type"`
	ID      string `json:"id"`
	Version int    `json:"version"`
}

type OrganisationResource struct {
	Resource

	OrganisationID string `json:"organisation_id"`
}

type AccountAttributes struct {
	Country                 string   `json:"country"`
	BaseCurrency            string   `json:"base_currency"`
	AccountNumber           string   `json:"account_number"`
	BankID                  string   `json:"bank_id"`
	BankIDCode              string   `json:"bank_id_code"`
	BIC                     string   `json:"bic"`
	IBAN                    string   `json:"iban"`
	Name                    []string `json:"name"`
	AlternativeNames        []string `json:"alternative_names"`
	AccountClassification   string   `json:"account_classification"`
	SecondaryIdentification string   `json:"secondary_identification"`
	Switched                bool     `json:"switched"`
	Status                  string   `json:"status"`
}

type Account struct {
	OrganisationResource
	AccountAttributes
}

func (a Account) GetID() string {
	return a.ID
}

type AccountList struct {
	Items []Account
}
