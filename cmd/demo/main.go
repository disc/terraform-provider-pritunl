package main

import (
	"fmt"
	"pritunl-terraform/internal/pritunl"
)

func main() {
	apiToken := "rv2xqPtDiszTLN7IUsMooDXbpYZ7AAiC"
	apiSecret := "Oq3FeJCa7hBSVD13We39GnVEty86toTI"
	baseUrl := "https://connect.cydriver.com"
	client := pritunl.NewClient(baseUrl, apiToken, apiSecret)

	orgName := "disc-org-test"
	organization, err := client.CreateOrganization(orgName)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	serverName := "disc-server-test"
	protocol := "udp"
	cipher := "aes128"
	hash := "sha1"
	server, err := client.CreateServer(serverName, protocol, cipher, hash)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	err = client.AttachOrganizationToServer(organization.ID, server.ID)
	if err != nil {
		fmt.Println(err.Error())
	}

	//serverId := server.ID
	serverId := "610339fb4bce2ca96a757a91"
	err = client.AddRouteToServer(serverId, "5.12.45.222/32")
	if err != nil {
		fmt.Println(err.Error())
	}
}
