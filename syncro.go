// syncro.go
//
// The file contains data synchronization functions
// between clients using test systems and Captura.
//
package main

import (
	"github.com/jinzhu/gorm"
)

//?Maybe then it will not be necessary to pass the tester type, since the function will be common to any system?
func runSync(db *gorm.DB, api tester, interval int64) {
	return
}
