package main

//	"bytes"
//	"database/sql"
//	"fmt"
//	"net/http"

//"strings"
//	_ "github.com/lib/pq"
//	"github.com/vulcand/oxy/forward"
//	"log"
//"runtime"
//	"time"

func main() {

	configure()

	HTTPServe()
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
