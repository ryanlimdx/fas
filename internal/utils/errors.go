// Contains common utils.
package utils

import (
	"fmt"
	"net/http"
    
	"github.com/go-sql-driver/mysql"
)

// HandleInsertError handles any errors from insertion of entries into the database.
func HandleInsertError(w http.ResponseWriter, err error, entity string) {
    if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1062 {
        http.Error(w, fmt.Sprintf("An entry for the %s already exists", entity), http.StatusConflict)
        return
    }
    
    http.Error(w, fmt.Sprintf("Failed to insert %s", entity), http.StatusInternalServerError)
}
