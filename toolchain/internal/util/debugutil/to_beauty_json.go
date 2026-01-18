package debugutil

import "encoding/json"

// ToBeautyJSON converts a data structure to a well formatted JSON.
// Useful for debugging.
//
// If error occurs, it returns the error as a string.
func ToBeautyJSON(data any) string {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(jsonData)
}
