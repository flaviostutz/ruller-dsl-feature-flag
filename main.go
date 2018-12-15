package main

import (
	_ "errors"
	_ "flag"
	"fmt"
	_ "os"
	_ "os/signal"
	_ "path/filepath"
	_ "syscall"
	_ "text/template"

	_ "github.com/Sirupsen/logrus"
	_ "github.com/gorilla/mux"

	_ "github.com/flaviostutz/ruller/ruller"
)

func main() {
	fmt.Println("This is used for build caching purposes. Should be replaced.")
}
