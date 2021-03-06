package ddcloud

import (
	"fmt"
	"log"
	"time"

	"github.com/DimensionDataResearch/go-dd-cloud-compute/compute"
	"github.com/hashicorp/terraform/helper/schema"
)

const (
	resourceKeyServerName              = "name"
	resourceKeyServerDescription       = "description"
	resourceKeyServerAdminPassword     = "admin_password"
	resourceKeyServerNetworkDomainID   = "networkdomain"
	resourceKeyServerMemoryGB          = "memory_gb"
	resourceKeyServerCPUCount          = "cpu_count"
	resourceKeyServerCPUCoreCount      = "cores_per_cpu"
	resourceKeyServerCPUSpeed          = "cpu_speed"
	resourceKeyServerOSImageID         = "os_image_id"
	resourceKeyServerOSImageName       = "os_image_name"
	resourceKeyServerCustomerImageID   = "customer_image_id"
	resourceKeyServerCustomerImageName = "customer_image_name"
	resourceKeyServerPrimaryVLAN       = "primary_adapter_vlan"
	resourceKeyServerPrimaryIPv4       = "primary_adapter_ipv4"
	resourceKeyServerPrimaryIPv6       = "primary_adapter_ipv6"
	resourceKeyServerPublicIPv4        = "public_ipv4"
	resourceKeyServerPrimaryDNS        = "dns_primary"
	resourceKeyServerSecondaryDNS      = "dns_secondary"
	resourceKeyServerAutoStart         = "auto_start"

	resourceCreateTimeoutServer = 30 * time.Minute
	resourceUpdateTimeoutServer = 10 * time.Minute
	resourceDeleteTimeoutServer = 15 * time.Minute
	serverShutdownTimeout       = 5 * time.Minute
)

func resourceServer() *schema.Resource {
	return &schema.Resource{
		Create: resourceServerCreate,
		Read:   resourceServerRead,
		Update: resourceServerUpdate,
		Delete: resourceServerDelete,

		Schema: map[string]*schema.Schema{
			resourceKeyServerName: &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "A name for the server",
			},
			resourceKeyServerDescription: &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "A description for the server",
			},
			resourceKeyServerAdminPassword: &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				Description: "The initial administrative password for the deployed server",
			},
			resourceKeyServerMemoryGB: &schema.Schema{
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Default:     nil,
				Description: "The amount of memory (in GB) allocated to the server",
			},
			resourceKeyServerCPUCount: &schema.Schema{
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Default:     nil,
				Description: "The number of CPUs allocated to the server",
			},
			resourceKeyServerCPUCoreCount: &schema.Schema{
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Default:     nil,
				Description: "The number of cores per CPU allocated to the server",
			},
			resourceKeyServerCPUSpeed: &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Default:     nil,
				Description: "The speed (quality-of-service) for CPUs allocated to the server",
			},
			resourceKeyServerDisk: schemaServerDisk(),
			resourceKeyServerNetworkDomainID: &schema.Schema{
				Type:        schema.TypeString,
				ForceNew:    true,
				Required:    true,
				Description: "The Id of the network domain in which the server is deployed",
			},
			resourceKeyServerPrimaryVLAN: &schema.Schema{
				Type:        schema.TypeString,
				ForceNew:    true,
				Optional:    true,
				Computed:    true,
				Default:     nil,
				Description: "The Id of the VLAN to which the server's primary network adapter will be attached (the first available IPv4 address will be allocated)",
			},
			resourceKeyServerPrimaryIPv4: &schema.Schema{
				Type:        schema.TypeString,
				ForceNew:    true,
				Optional:    true,
				Computed:    true,
				Default:     nil,
				Description: "The IPv4 address for the server's primary network adapter",
			},
			resourceKeyServerPublicIPv4: &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Default:     nil,
				Description: "The server's public IPv4 address (if any)",
			},
			resourceKeyServerPrimaryIPv6: &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The IPv6 address of the server's primary network adapter",
			},
			resourceKeyServerPrimaryDNS: &schema.Schema{
				Type:        schema.TypeString,
				ForceNew:    true,
				Optional:    true,
				Default:     "",
				Description: "The IP address of the server's primary DNS server",
			},
			resourceKeyServerSecondaryDNS: &schema.Schema{
				Type:        schema.TypeString,
				ForceNew:    true,
				Optional:    true,
				Default:     "",
				Description: "The IP address of the server's secondary DNS server",
			},
			resourceKeyServerOSImageID: &schema.Schema{
				Type:        schema.TypeString,
				ForceNew:    true,
				Optional:    true,
				Computed:    true,
				Default:     nil,
				Description: "The Id of the OS (built-in) image from which the server is created",
			},
			resourceKeyServerOSImageName: &schema.Schema{
				Type:        schema.TypeString,
				ForceNew:    true,
				Optional:    true,
				Computed:    true,
				Default:     nil,
				Description: "The name of the OS (built-in) image from which the server is created",
			},
			resourceKeyServerCustomerImageID: &schema.Schema{
				Type:        schema.TypeString,
				ForceNew:    true,
				Optional:    true,
				Computed:    true,
				Default:     nil,
				Description: "The Id of the customer (custom) image from which the server is created",
			},
			resourceKeyServerCustomerImageName: &schema.Schema{
				Type:        schema.TypeString,
				ForceNew:    true,
				Optional:    true,
				Computed:    true,
				Default:     nil,
				Description: "The name of the customer (custom) image from which the server is created",
			},
			resourceKeyServerAutoStart: &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Should the server be started automatically once it has been deployed",
			},
			resourceKeyServerTag: schemaServerTag(),
		},
	}
}

// Create a server resource.
func resourceServerCreate(data *schema.ResourceData, provider interface{}) error {
	name := data.Get(resourceKeyServerName).(string)
	description := data.Get(resourceKeyServerDescription).(string)
	adminPassword := data.Get(resourceKeyServerAdminPassword).(string)
	networkDomainID := data.Get(resourceKeyServerNetworkDomainID).(string)
	primaryDNS := data.Get(resourceKeyServerPrimaryDNS).(string)
	secondaryDNS := data.Get(resourceKeyServerSecondaryDNS).(string)
	autoStart := data.Get(resourceKeyServerAutoStart).(bool)

	log.Printf("Create server '%s' in network domain '%s' (description = '%s').", name, networkDomainID, description)

	providerState := provider.(*providerState)
	apiClient := providerState.Client()

	networkDomain, err := apiClient.GetNetworkDomain(networkDomainID)
	if err != nil {
		return err
	}

	if networkDomain == nil {
		return fmt.Errorf("No network domain was found with Id '%s'.", networkDomainID)
	}

	dataCenterID := networkDomain.DatacenterID
	log.Printf("Server will be deployed in data centre '%s'.", dataCenterID)

	deploymentConfiguration := compute.ServerDeploymentConfiguration{
		Name:                  name,
		Description:           description,
		AdministratorPassword: adminPassword,
		Start: autoStart,
	}

	propertyHelper := propertyHelper(data)

	// Retrieve image details.
	osImageID := propertyHelper.GetOptionalString(resourceKeyServerOSImageID, false)
	osImageName := propertyHelper.GetOptionalString(resourceKeyServerOSImageName, false)
	customerImageID := propertyHelper.GetOptionalString(resourceKeyServerCustomerImageID, false)
	customerImageName := propertyHelper.GetOptionalString(resourceKeyServerCustomerImageName, false)

	var (
		osImage       *compute.OSImage
		customerImage *compute.CustomerImage
	)
	if osImageID != nil {
		log.Printf("Looking up OS image '%s' by Id...", *osImageID)

		osImage, err = apiClient.GetOSImage(*osImageID)
		if err != nil {
			return err
		}

		log.Printf("Server will be deployed from OS image named '%s' (Id = '%s').", osImage.Name, osImage.ID)
		data.Set(resourceKeyServerOSImageName, osImage.Name)
	} else if osImageName != nil {
		log.Printf("Looking up OS image '%s' by name...", *osImageName)

		osImage, err = apiClient.FindOSImage(*osImageName, dataCenterID)
		if err != nil {
			return err
		}

		if osImage == nil {
			log.Printf("Warning - unable to find an OS image named '%s' in data centre '%s' (which is where the target network domain, '%s', is located).", *osImageName, dataCenterID, networkDomainID)

			return fmt.Errorf("Unable to find an OS image named '%s' in data centre '%s' (which is where the target network domain, '%s', is located).", *osImageName, dataCenterID, networkDomainID)
		}

		log.Printf("Server will be deployed from OS image named '%s' (Id = '%s').", osImage.Name, osImage.ID)
		data.Set(resourceKeyServerOSImageID, osImage.ID)
	} else if customerImageID != nil {
		log.Printf("Looking up customer image '%s' by Id...", *customerImageID)

		customerImage, err = apiClient.GetCustomerImage(*customerImageID)
		if err != nil {
			return err
		}

		log.Printf("Server will be deployed from customer image named '%s' (Id = '%s').", customerImage.Name, customerImage.ID)
		data.Set(resourceKeyServerOSImageName, osImage.Name)
	} else if customerImageName != nil {
		log.Printf("Looking up customer image '%s' by name...", *customerImageName)

		osImage, err = apiClient.FindOSImage(*customerImageName, dataCenterID)
		if err != nil {
			return err
		}

		if osImage == nil {
			log.Printf("Warning - unable to find a customer image named '%s' in data centre '%s' (which is where the target network domain, '%s', is located).", *customerImageName, dataCenterID, networkDomainID)

			return fmt.Errorf("Unable to find a customer image named '%s' in data centre '%s' (which is where the target network domain, '%s', is located).", *customerImageName, dataCenterID, networkDomainID)
		}

		log.Printf("Server will be deployed from OS image named '%s' (Id = '%s').", osImage.Name, osImage.ID)
		data.Set(resourceKeyServerOSImageID, osImage.ID)
	}

	if osImage != nil {
		err = deploymentConfiguration.ApplyOSImage(osImage)
		if err != nil {
			return err
		}
	} else if customerImage != nil {
		err = deploymentConfiguration.ApplyCustomerImage(customerImage)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("Must specify either os_image_id, os_image_name, customer_image_id, or customer_image_name.")
	}

	// Memory and CPU
	memoryGB := propertyHelper.GetOptionalInt(resourceKeyServerMemoryGB, false)
	if memoryGB != nil {
		deploymentConfiguration.MemoryGB = *memoryGB
	} else {
		data.Set(resourceKeyServerMemoryGB, deploymentConfiguration.MemoryGB)
	}

	cpuCount := propertyHelper.GetOptionalInt(resourceKeyServerCPUCount, false)
	if cpuCount != nil {
		deploymentConfiguration.CPU.Count = *cpuCount
	} else {
		data.Set(resourceKeyServerCPUCount, deploymentConfiguration.CPU.Count)
	}

	cpuCoreCount := propertyHelper.GetOptionalInt(resourceKeyServerCPUCoreCount, false)
	if cpuCoreCount != nil {
		deploymentConfiguration.CPU.CoresPerSocket = *cpuCoreCount
	} else {
		data.Set(resourceKeyServerCPUCoreCount, deploymentConfiguration.CPU.CoresPerSocket)
	}

	cpuSpeed := propertyHelper.GetOptionalString(resourceKeyServerCPUSpeed, false)
	if cpuSpeed != nil {
		deploymentConfiguration.CPU.Speed = *cpuSpeed
	} else {
		data.Set(resourceKeyServerCPUSpeed, deploymentConfiguration.CPU.Speed)
	}

	// Network
	primaryVLANID := propertyHelper.GetOptionalString(resourceKeyServerPrimaryVLAN, false)
	primaryIPv4Address := propertyHelper.GetOptionalString(resourceKeyServerPrimaryIPv4, false)

	deploymentConfiguration.Network = compute.VirtualMachineNetwork{
		NetworkDomainID: networkDomainID,
		PrimaryAdapter: compute.VirtualMachineNetworkAdapter{
			VLANID:             primaryVLANID,
			PrivateIPv4Address: primaryIPv4Address,
		},
	}
	deploymentConfiguration.PrimaryDNS = primaryDNS
	deploymentConfiguration.SecondaryDNS = secondaryDNS

	log.Printf("Server deployment configuration: %+v", deploymentConfiguration)
	log.Printf("Server CPU deployment configuration: %+v", deploymentConfiguration.CPU)

	serverID, err := apiClient.DeployServer(deploymentConfiguration)
	if err != nil {
		return err
	}

	data.SetId(serverID)

	log.Printf("Server '%s' is being provisioned...", name)

	resource, err := apiClient.WaitForDeploy(compute.ResourceTypeServer, serverID, resourceCreateTimeoutServer)
	if err != nil {
		return err
	}

	// Capture additional properties that may only be available after deployment.
	server := resource.(*compute.Server)
	captureServerNetworkConfiguration(server, data, false)

	data.Partial(true)

	var publicIPv4Address string
	publicIPv4Address, err = findPublicIPv4Address(apiClient,
		networkDomainID,
		*server.Network.PrimaryAdapter.PrivateIPv4Address,
	)
	if err != nil {
		return err
	}
	if !isEmpty(publicIPv4Address) {
		data.Set(resourceKeyServerPublicIPv4, publicIPv4Address)
	} else {
		data.Set(resourceKeyServerPublicIPv4, nil)
	}

	data.SetPartial(resourceKeyServerPublicIPv4)

	err = applyServerTags(data, apiClient, providerState.Settings())
	if err != nil {
		return err
	}
	data.SetPartial(resourceKeyServerTag)

	err = createDisks(server.Disks, data, apiClient)
	if err != nil {
		return err
	}

	data.Partial(false)

	return nil
}

// Read a server resource.
func resourceServerRead(data *schema.ResourceData, provider interface{}) error {
	id := data.Id()
	name := data.Get(resourceKeyServerName).(string)
	description := data.Get(resourceKeyServerDescription).(string)
	networkDomainID := data.Get(resourceKeyServerNetworkDomainID).(string)

	log.Printf("Read server '%s' (Id = '%s') in network domain '%s' (description = '%s').", name, id, networkDomainID, description)

	apiClient := provider.(*providerState).Client()
	server, err := apiClient.GetServer(id)
	if err != nil {
		return err
	}

	if server == nil {
		log.Printf("Server '%s' has been deleted.", id)

		// Mark as deleted.
		data.SetId("")

		return nil
	}
	data.Set(resourceKeyServerName, server.Name)
	data.Set(resourceKeyServerDescription, server.Description)
	data.Set(resourceKeyServerOSImageID, server.SourceImageID)
	data.Set(resourceKeyServerMemoryGB, server.MemoryGB)
	data.Set(resourceKeyServerCPUCount, server.CPU.Count)
	data.Set(resourceKeyServerCPUCoreCount, server.CPU.CoresPerSocket)
	data.Set(resourceKeyServerCPUSpeed, server.CPU.Speed)

	captureServerNetworkConfiguration(server, data, false)

	var publicIPv4Address string
	publicIPv4Address, err = findPublicIPv4Address(apiClient,
		networkDomainID,
		*server.Network.PrimaryAdapter.PrivateIPv4Address,
	)
	if err != nil {
		return err
	}
	if !isEmpty(publicIPv4Address) {
		data.Set(resourceKeyServerPublicIPv4, publicIPv4Address)
	} else {
		data.Set(resourceKeyServerPublicIPv4, nil)
	}

	err = readServerTags(data, apiClient)
	if err != nil {
		return err
	}

	propertyHelper := propertyHelper(data)
	propertyHelper.SetServerDisks(server.Disks)

	return nil
}

// Update a server resource.
func resourceServerUpdate(data *schema.ResourceData, provider interface{}) error {
	serverID := data.Id()

	// These changes can only be made through the V1 API (we're mostly using V2).
	// Later, we can come back and implement the required functionality in the compute API client.
	if data.HasChange(resourceKeyServerName) {
		return fmt.Errorf("Changing the 'name' property of a 'ddcloud_server' resource type is not yet implemented.")
	}

	if data.HasChange(resourceKeyServerDescription) {
		return fmt.Errorf("Changing the 'description' property of a 'ddcloud_server' resource type is not yet implemented.")
	}

	log.Printf("Update server '%s'.", serverID)

	providerState := provider.(*providerState)

	apiClient := providerState.Client()
	server, err := apiClient.GetServer(serverID)
	if err != nil {
		return err
	}

	if server == nil {
		log.Printf("Server '%s' has been deleted.", serverID)
		data.SetId("")

		return nil
	}

	serverLock := providerState.GetServerLock(serverID, "resourceServerUpdate(id = '%s')", serverID)
	serverLock.Lock()
	defer serverLock.Unlock()

	data.Partial(true)

	propertyHelper := propertyHelper(data)

	var memoryGB, cpuCount, cpuCoreCount *int
	var cpuSpeed *string
	if data.HasChange(resourceKeyServerMemoryGB) {
		memoryGB = propertyHelper.GetOptionalInt(resourceKeyServerMemoryGB, false)
	}
	if data.HasChange(resourceKeyServerCPUCount) {
		cpuCount = propertyHelper.GetOptionalInt(resourceKeyServerCPUCount, false)
	}
	if data.HasChange(resourceKeyServerCPUCoreCount) {
		cpuCoreCount = propertyHelper.GetOptionalInt(resourceKeyServerCPUCoreCount, false)
	}
	if data.HasChange(resourceKeyServerCPUSpeed) {
		cpuSpeed = propertyHelper.GetOptionalString(resourceKeyServerCPUSpeed, false)
	}

	if memoryGB != nil || cpuCount != nil || cpuCoreCount != nil || cpuSpeed != nil {
		log.Printf("Server CPU / memory configuration change detected.")

		err = updateServerConfiguration(apiClient, server, memoryGB, cpuCount, cpuCoreCount, cpuSpeed)
		if err != nil {
			return err
		}

		if data.HasChange(resourceKeyServerMemoryGB) {
			data.SetPartial(resourceKeyServerMemoryGB)
		}

		if data.HasChange(resourceKeyServerCPUCount) {
			data.SetPartial(resourceKeyServerCPUCount)
		}
	}

	var primaryIPv4, primaryIPv6 *string
	if data.HasChange(resourceKeyServerPrimaryIPv4) {
		primaryIPv4 = propertyHelper.GetOptionalString(resourceKeyServerPrimaryIPv4, false)
	}
	if data.HasChange(resourceKeyServerPrimaryIPv6) {
		primaryIPv6 = propertyHelper.GetOptionalString(resourceKeyServerPrimaryIPv6, false)
	}

	if primaryIPv4 != nil || primaryIPv6 != nil {
		log.Printf("Server network configuration change detected.")

		err = updateServerIPAddresses(apiClient, server, primaryIPv4, primaryIPv6)
		if err != nil {
			return err
		}

		if data.HasChange(resourceKeyServerPrimaryIPv4) {
			data.SetPartial(resourceKeyServerPrimaryIPv4)
		}

		if data.HasChange(resourceKeyServerPrimaryIPv6) {
			data.SetPartial(resourceKeyServerPrimaryIPv6)
		}

		var publicIPv4Address string
		publicIPv4Address, err = findPublicIPv4Address(apiClient,
			server.Network.NetworkDomainID,
			*server.Network.PrimaryAdapter.PrivateIPv4Address,
		)
		if err != nil {
			return err
		}
		if !isEmpty(publicIPv4Address) {
			data.Set(resourceKeyServerPublicIPv4, publicIPv4Address)
		} else {
			data.Set(resourceKeyServerPublicIPv4, nil)
		}
	}

	if data.HasChange(resourceKeyServerTag) {
		err = applyServerTags(data, apiClient, providerState.Settings())
		if err != nil {
			return err
		}

		data.SetPartial(resourceKeyServerTag)
	}

	if data.HasChange(resourceKeyServerDisk) {
		err = updateDisks(data, apiClient)
		if err != nil {
			return err
		}
	}

	data.Partial(false)

	return nil
}

// Delete a server resource.
func resourceServerDelete(data *schema.ResourceData, provider interface{}) error {
	var id, name, networkDomainID string

	id = data.Id()
	name = data.Get(resourceKeyServerName).(string)
	networkDomainID = data.Get(resourceKeyServerNetworkDomainID).(string)

	log.Printf("Delete server '%s' ('%s') in network domain '%s'.", id, name, networkDomainID)

	providerState := provider.(*providerState)

	apiClient := providerState.Client()
	server, err := apiClient.GetServer(id)
	if err != nil {
		return err
	}

	if server == nil {
		log.Printf("Server '%s' not found; will treat the server as having already been deleted.", id)

		return nil
	}

	serverLock := providerState.GetServerLock(id, "resourceServerDelete(id = '%s')", id)
	serverLock.Lock()
	defer serverLock.Unlock()

	if server.Started {
		log.Printf("Server '%s' is currently running. The server will be powered off.", id)

		err = apiClient.PowerOffServer(id)
		if err != nil {
			return err
		}

		_, err = apiClient.WaitForChange(compute.ResourceTypeServer, id, "Power off server", serverShutdownTimeout)
		if err != nil {
			return err
		}
	}

	log.Printf("Server '%s' is being deleted...", id)

	err = apiClient.DeleteServer(id)
	if err != nil {
		return err
	}

	return apiClient.WaitForDelete(compute.ResourceTypeServer, id, resourceDeleteTimeoutServer)
}

func findPublicIPv4Address(apiClient *compute.Client, networkDomainID string, privateIPv4Address string) (publicIPv4Address string, err error) {
	page := compute.DefaultPaging()
	for {
		var natRules *compute.NATRules
		natRules, err = apiClient.ListNATRules(networkDomainID, page)
		if err != nil {
			return
		}
		if natRules.IsEmpty() {
			break // We're done
		}

		for _, natRule := range natRules.Rules {
			if natRule.InternalIPAddress == privateIPv4Address {
				return natRule.ExternalIPAddress, nil
			}
		}

		page.Next()
	}

	return
}
