package main

import (
	"fmt"
	"flag"
	"time"
	"os"
	"io/fs"
	"strings"
	"errors"
	"path/filepath"
	"encoding/json"
	fstring "github.com/Fipaan/lib.go/string"
	nob "github.com/Fipaan/nob.go"
)

const (
	BUILD_FOLDER string = ".build"
)

type Service struct {
	Name     string
	DB_Host  string
	DB_Port  uint
	DB_Name  string 
	DB_User  string
	WithPass bool
}
func (s Service) Path() string {
	return filepath.Join(nob.ProgramDir(), fmt.Sprintf("%v-service", s.Name))
}
func (s Service) TmpName() string {
	return fmt.Sprintf("%v-cleaned", s.Name)
}
func (s Service) TmpPath() string {
	return filepath.Join(nob.ProgramDir(), BUILD_FOLDER, s.TmpName())
}
func (s Service) DB_ActualName() string {
	if s.DB_Name != "" { return s.DB_Name }
	return fmt.Sprintf("%v_db", s.Name)
}
func (s Service) MigPath() string {
	return filepath.Join(s.Path(), "migrations")
}

func prepareService(cmd *nob.Cmd, service Service) {
	cmd.Push("psql")
	if service.DB_Host != "" { cmd.Push("-h", service.DB_Host) }
	if service.DB_Port != 0  {
		cmd.Push("-p", fmt.Sprintf("%v", service.DB_Port))
	}
	cmd.Push("-U", service.DB_User)
	if service.WithPass { cmd.Push("-W") }
}

func cleanService(cmd *nob.Cmd, service Service) bool {
	// re-initialize database
	prepareService(cmd, service)
	dbName := service.DB_ActualName()
	cmd.Push("-c", fmt.Sprintf("DROP   DATABASE %v;", dbName))
	cmd.Push("-c", fmt.Sprintf("CREATE DATABASE %v;", dbName))
	cmd.Push("-c", "\\q")
	if !cmd.Run() { return false }
	// apply migrations
	prepareService(cmd, service)
	cmd.Push("-d", dbName)
	filepath.WalkDir(service.MigPath(), func(migPath string, d fs.DirEntry, err error) error {
		if err != nil { return err }
		if !strings.HasSuffix(migPath, ".sql") { return nil }
		cmd.Push("-f", migPath)
		return nil
	})
	if !cmd.Run() { return false }
	return true
}

func cleanServices(cmd *nob.Cmd) bool {
	for i := 0; i < len(SERVICES); i++ {
		service := SERVICES[i]
		if !cleanService(cmd, service) { return false }
	}
	return true
}

func startService(cmd *nob.Cmd, serviceName string) bool {
	for i := 0; i < len(SERVICES); i++ {
		service := SERVICES[i]
		if serviceName != service.Name { continue }
		savedDir := cmd.WalkIn(service.Path())
		// add/remove deps
		cmd.Push("go", "mod", "tidy")
		if !cmd.Run() { return false }
		// start service
		cmd.Push("go", "run", "./cmd")
		if !cmd.Run() { return false }

		cmd.Dir = savedDir
		return true
	}
	fmt.Printf("ERROR: unknown service %v. For list of services provide -l flag\n", fstring.Stringify(serviceName))
	return false
}

var ErrSrcFound = errors.New("src file found")
var ErrServiceClean = errors.New("couldn't clean service")

func preClean(cmd *nob.Cmd) (err error) {
	err = nob.MkdirIfNotExists(BUILD_FOLDER)
	if err != nil { return }
	for i := 0; i < len(SERVICES); i++ {
		service := SERVICES[i]
		tmpPath := service.TmpPath()
		var fileExist bool
		fileExist, err = nob.FileExist(tmpPath)
		if err != nil { return }
		if !fileExist {
			err = nob.Touch(tmpPath)
			if err != nil { return }
			if !cleanService(cmd, service) {
				err = ErrServiceClean
				return
			}
			continue
		}
		var tmpTime time.Time
		tmpTime, err = nob.GetModTime(tmpPath)
		if err != nil { return }
		err = filepath.WalkDir(service.Path(), func(srcPath string, d fs.DirEntry, dirErr error) error {
			if dirErr != nil { return dirErr }
			if !strings.HasSuffix(srcPath, ".go") { return nil }
			var srcTime time.Time
			srcTime, dirErr = nob.GetModTime(srcPath)
			if dirErr != nil { return dirErr }
			if srcTime.After(tmpTime) { return ErrSrcFound }
			return nil
		})
		if err != nil {
			if !errors.Is(err, ErrSrcFound) { return }
			err = nob.Touch(tmpPath)
			if err != nil { return }
			cleanService(cmd, service)
			continue
		}
	}
	return
}

var SERVICES = []Service{
	Service{Name: "order",   DB_User: "postgres"},
	Service{Name: "payment", DB_User: "postgres"},
}

func listServices() {
	fmt.Printf("Services:\n")
	for i := 0; i < len(SERVICES); i++ {
		service := SERVICES[i]
		fmt.Printf("  %v\n", service.Name)
	}
}

func main() {
	nob.GoRebuildUrself("go", "build", "-o", "nob")

	var isClean bool
	var service string
	var isList  bool 
	var isCurl  bool 
	flag.BoolVar  (&isClean, "clean", false, "cleans databases")
	flag.StringVar(&service, "s",     "",    "start service")
	flag.BoolVar  (&isList,  "l",     false, "list of all services")
	flag.BoolVar  (&isCurl,  "curl",  false, "run curl test")
	flag.Parse()
	cmd := nob.CmdInit()
	if isList {
		listServices()
		return
	}
	if isCurl {
		cmd.Push("curl", "http://localhost:8000/orders")
		cmd.Push("-X", "POST")
		cmd.Push("-H", "Content-Type: application/json")
		type order struct {
			CustomerId   string `json:"customer_id"`
  		  	ItemName     string `json:"item_name"`
  		  	Amount       int    `json:"amount"`
		}
		orderBytes, err := json.Marshal(order{
			CustomerId: "cust-1",
  			ItemName: "server",
  		  	Amount: 200000,
		})
		if err != nil {
			fmt.Printf("ERROR: couldn't curl test: %v", err)
			os.Exit(1)
		}
		cmd.Push("-d", string(orderBytes))
		cmd.Stderr = nil
		cmd.Pipe = nob.CmdInit("jq")
		if !cmd.Run() {
			os.Exit(1)
		}
		return
	}
	if isClean {
		if !cleanServices(cmd) {
			os.Exit(1)
		}
	} else {
		err := preClean(cmd)
		if err != nil {
			fmt.Printf("ERROR: %v\n", err)
			os.Exit(1)
		}
	}
	if service != "" {
		if !startService(cmd, service) {
			os.Exit(1)
		}
	}
}
