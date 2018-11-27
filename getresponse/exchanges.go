package getresponse

type (
	CreateContactRequest struct {
		Name         *string       `json:"name,omitempty"`
		Email        string        `json:"email"`
		DayOfCycle   *int32        `json:"dayOfCycle,omitempty"`
		Campaign     Campaign      `json:"campaign"`
		CustomFields []CustomField `json:"customFieldValues,omitempty"`
		IPAddress    *string       `json:"ipAddress,omitempty"`
	}
	UpdateContactResponse struct {
		Contact Contact
	}
	UpdateContactRequest struct {
		ID      string
		NewData Contact
	}
	GetContactResponse struct {
		Contact Contact
	}
	GetContactRequest struct {
		ID     string
		Fields []string
	}
	GetContactsResponse struct {
		Contacts []Contact
	}
	GetContactsRequest struct {
		QueryHash       map[string]string
		Fields          []string
		SortHash        map[string]string
		Page            int32
		PerPage         int32
		AdditionalFlags *string
	}
	UpdateContactCustomFieldsRequest struct {
		ID           string        `json:"-"`
		CustomFields []CustomField `json:"customFieldValues"`
	}
	UpdateContactCustomFieldsResponse struct {
		Contact Contact
	}
	DeleteContactRequest struct {
		ID        string
		MessageID string
		IpAddress string
	}
)
