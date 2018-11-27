package getresponse

// GetResponseError holds an API error
type GetResponseError struct {
	HTTPStatus      int      `json:"httpStatus"`
	ErrorCode       int      `json:"code"`
	CodeDescription string   `json:"codeDescription"`
	Message         string   `json:"message"`
	MoreInfo        string   `json:"moreInfo"`
	Context         []string `json:"context"`
	UUID            string   `json:"uuid"`
}

func (g *GetResponseError) Error() string {
	return g.Message
}
