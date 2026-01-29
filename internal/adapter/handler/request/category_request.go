package request

type CategoryRequest struct {
	Title string `json:"title" validate:"required"`
}
