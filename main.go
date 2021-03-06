package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"
)

const sleep = 5
const bssleep = 120

var scriptpath string

func check() {
	log.Print("K8S | Info: Checking state")
	resp, err := http.Get("http://localhost:8081")

	if err != nil {
		time.Sleep(sleep * time.Second)
		check()
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)

		if err != nil {
			log.Fatal(err)
			return
		}

		bodyString := string(bodyBytes)

		if bodyString == "Galera Cluster Node status: synced" {
			log.Print("K8S | Info: Synced state reached")

			log.Print("K8S | Info: SST bootstrap cooldown")
			time.Sleep(bssleep * time.Second)

			cmd := exec.Command(scriptpath)

			if err := cmd.Start(); err != nil {
				log.Fatalf("K8S | Error: %v", err)
				return
			}

			log.Print("K8S | Info: Running input script")

			if err := cmd.Wait(); err != nil {
				log.Fatal(err)
				return
			}

			log.Print("K8S | Info: Backup done")
			os.Exit(0)
		} else {
			log.Print("K8S | Info: Body string has no match")
			time.Sleep(sleep * time.Second)
			check()
			return
		}
	} else {
		time.Sleep(sleep * time.Second)
		check()
		return
	}
}

func main() {
	args := os.Args[1:]

	if len(args) != 1 {
		log.Fatal("Missing backup script path")
	}

	if _, err := os.Stat(args[0]); os.IsNotExist(err) {
		log.Fatal("Backup script path incorrect")
	}

	scriptpath = args[0]

	log.Print("K8S | Info: Launching galera script in background")

	cmd := exec.Command("start.sh", "node", fmt.Sprintf("seed,%s", os.Getenv("SEED_ADDRESS")))

	if err := cmd.Start(); err != nil {
		log.Fatalf("K8S | Error: %v", err)
	}

	log.Print("K8S | Info: State checker active")
	check()

	if err := cmd.Wait(); err != nil {
		log.Printf("K8S | Error: %v", err)
	}
}
