package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
	"golang.org/x/net/context"
)

var authKey string

func main() {
	authKey = getAuthKey()

	http.HandleFunc("/services", handleListServices)
	http.HandleFunc("/services/update", handleUpdateService)

	port := os.Getenv("PORT")
	if port == "" {
		port = "80"
	}

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}

func handleListServices(w http.ResponseWriter, r *http.Request) {
	if checkAuth(r) == false {
		http.Error(w, "Not Authorized", http.StatusUnauthorized)
		return
	}

	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	services, err := cli.ServiceList(context.Background(), types.ServiceListOptions{})
	if err != nil {
		fmt.Fprintf(w, err.Error())
		log.Print(err.Error())
		return
	}

	for _, service := range services {
		fmt.Fprintf(w, "id: %q, name: %q, image: %q, version: %d\n",
			service.ID, service.Spec.Name, service.Spec.TaskTemplate.ContainerSpec.Image,
			service.Version.Index)
	}
}

func handleUpdateService(w http.ResponseWriter, r *http.Request) {
	if checkAuth(r) == false {
		http.Error(w, "Not Authorized", http.StatusUnauthorized)
		return
	}

	// get params
	name := r.FormValue("name")
	image := r.FormValue("image")
	commit := r.FormValue("commit")

	if r.Method != "POST" || name == "" || image == "" {
		http.Error(w, "POST attributes 'name' and 'image' are requried", http.StatusUnprocessableEntity)
		return
	}

	if checkWhitelist(name) == false {
		http.Error(w, "Service Not Whitelisted", http.StatusForbidden)
		return
	}

	// create docker client
	cli, err := client.NewEnvClient()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	services := fetchServiceInfo(cli, name)
	if len(services) != 1 {
		http.Error(w, fmt.Sprintf("Service %q not found.", name), http.StatusNotFound)
		return
	}

	// update service spec
	services[0].Spec.TaskTemplate.ContainerSpec.Image = image
	services[0].Spec.TaskTemplate.ContainerSpec.Labels["last_deploy"] = time.Now().Format(time.RFC3339)
	services[0].Spec.TaskTemplate.ContainerSpec.Labels["commit_hash"] = commit
	services[0].Spec.Labels["commit_hash"] = commit
	services[0].Spec.Labels["last_deploy"] = time.Now().Format(time.RFC3339)

	response, err := cli.ServiceUpdate(context.Background(),
		services[0].ID,
		services[0].Version,
		services[0].Spec,
		types.ServiceUpdateOptions{
			QueryRegistry: true,
		})
	if err != nil {
		panic(err)
	}
	for _, warn := range response.Warnings {
		fmt.Fprintf(w, "Warning: %s\n", warn)
	}

	fmt.Fprintf(w, "Updating service %s image to %s", name, image)
}

func fetchServiceInfo(cli *client.Client, name string) []swarm.Service {
	filters := filters.NewArgs()
	filters.Add("name", name)
	services, err := cli.ServiceList(context.Background(), types.ServiceListOptions{
		Filters: filters,
	})
	if err != nil {
		panic(err)
	}

	return services
}

// Gets the auth key value from file or env var
func getAuthKey() string {
	fileEnv := os.Getenv("AUTH_KEY_FILE")

	if fileEnv != "" {
		data, err := ioutil.ReadFile(fileEnv)
		if err != nil {
			panic(err)
		}
		log.Println("AUTH: ", string(data))
		return strings.TrimRight(string(data), "\r\n")
	}

	keyEnv := os.Getenv("AUTH_KEY")

	if keyEnv == "" {
		log.Println("Warning: AUTH_KEY is not set. Public access is allowed.")
	}

	return keyEnv
}

func checkAuth(r *http.Request) bool {
	if authKey == "" {
		return true
	}

	reqToken := r.Header.Get("Authorization")
	if reqToken == "" {
		return false
	}

	splitToken := strings.Split(reqToken, "Bearer ")
	if len(splitToken) != 2 {
		return false
	}

	return authKey == splitToken[1]
}

func checkWhitelist(name string) bool {
	whitelist := os.Getenv("WHITELIST")

	if whitelist == "" {
		log.Println("Warning: Whitelist is disabled. Any service can be updated.")
		return true
	}

	for _, service := range strings.Split(whitelist, ",") {
		if service == name {
			return true
		}
	}

	return false
}
