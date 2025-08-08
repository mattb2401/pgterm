package pgterm

import (
	"fmt"
	"strings"
)

// confirmIfNoWhere inspects the SQL and prompts if UPDATE or DELETE has no WHERE
func confirmIfNoWhere(sql string) (bool, error) {
	sqlUpper := strings.ToUpper(sql)
	sqlTrim := strings.TrimSpace(sqlUpper)

	isUpdate := strings.HasPrefix(sqlTrim, "UPDATE")
	isDelete := strings.HasPrefix(sqlTrim, "DELETE")

	if !isUpdate && !isDelete {
		// Not update/delete, allow directly
		return true, nil
	}

	// crude check: does it have " WHERE " (with spaces) after the first word?
	// This simple approach avoids false positives like column names
	hasWhere := false
	words := strings.Fields(sqlUpper)
	for i, w := range words {
		if (isUpdate && w == "UPDATE") || (isDelete && w == "DELETE") {
			// look for WHERE after this position
			for _, w2 := range words[i+1:] {
				if w2 == "WHERE" {
					hasWhere = true
					break
				}
			}
			break
		}
	}
	if hasWhere {
		return true, nil
	}
	if isUpdate {
		if askConfirmation("UPDATE") {
			return true, nil
		}
	} else if isDelete {
		if askConfirmation("DELETE") {
			return true, nil
		}
	}
	return false, fmt.Errorf("Safe choice. Query cancelled")
}
