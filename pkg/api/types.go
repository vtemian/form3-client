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

type AccountClassification string

const (
	AccountClassificationPersonal AccountClassification = "Personal"
	AccountClassificationBusiness AccountClassification = "Business"
)

type AccountAttributes struct {
	Country                 string                `json:"country"`
	BaseCurrency            string                `json:"base_currency"`
	AccountNumber           string                `json:"account_number"`
	BankID                  string                `json:"bank_id"`
	BankIDCode              string                `json:"bank_id_code"`
	BIC                     string                `json:"bic"`
	IBAN                    string                `json:"iban"`
	Name                    []string              `json:"name"`
	AlternativeNames        []string              `json:"alternative_names"`
	AccountClassification   AccountClassification `json:"account_classification"`
	SecondaryIdentification string                `json:"secondary_identification"`
	Switched                bool                  `json:"switched"`
	Status                  string                `json:"status"`
}

type Account struct {
	OrganisationResource

	Attributes AccountAttributes `json:"attributes"`
}

func (a Account) GetID() string { // nolint: gocritic
	return a.ID
}

func (a Account) GetVersion() int { // nolint: gocritic
	return a.Version
}

func NewAccount(id string, version int) *Account {
	return &Account{
		OrganisationResource: OrganisationResource{
			Resource: Resource{
				ID:      id,
				Version: version,
			},
		},
	}
}

type AccountList struct {
	Items []Account
}

func (a AccountList) GetID() string {
	return ""
}

func (a AccountList) GetVersion() int {
	return 0
}
