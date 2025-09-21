package requests

type CreateGlobalMessage struct {
	Name    string `json:"name"`
	Message string `json:"message"`
}
