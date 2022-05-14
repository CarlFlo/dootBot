package structs

// used when a button needs to be updated in the context of interactions
type ButtonData struct {
	CustomID string
	Disabled bool
	Label    string
}

// Wrapper because an array cannot be referenced, but a struct can
type ButtonDataWrapper struct {
	ButtonData []ButtonData
}
