package structs

// Globals is a structure containing template data.
type (
	Globals struct {
		PageData   interface{} //@TODO Put some kind of structure here
		PageParams PageParams
	}

	PageParams struct {
		Interface interface{}
	}
)
