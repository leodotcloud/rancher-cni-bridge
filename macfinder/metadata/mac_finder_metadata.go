package metadata

import (
	"fmt"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/rancher/go-rancher-metadata/metadata"
)

const (
	metadataURLTemplate = "http://%v/2015-12-19"
	multiplierForTwoMin = 240
	emptyMACAddress     = ""

	// DefaultMetadataAddress specifies the default value to use if nothing is specified
	DefaultMetadataAddress = "169.254.169.250"
)

// MACFinderFromMetadata is used to hold information related to
// Metadata client and other stuff.
type MACFinderFromMetadata struct {
	m metadata.Client
}

// NewMACFinderFromMetadata returns a new instance of the MACFinderFromMetadata
func NewMACFinderFromMetadata(metadataAddress string) (*MACFinderFromMetadata, error) {
	if metadataAddress == "" {
		metadataAddress = DefaultMetadataAddress
	}
	metadataURL := fmt.Sprintf(metadataURLTemplate, metadataAddress)
	m := metadata.NewClient(metadataURL)
	return &MACFinderFromMetadata{m}, nil
}

// GetMACAddress returns the IP address for the given container id, return an empty string
// if not found
func (mf *MACFinderFromMetadata) GetMACAddress(cid, rancherid string) string {
	for i := 0; i < multiplierForTwoMin; i++ {
		containers, err := mf.m.GetContainers()
		if err != nil {
			logrus.Errorf("rancher-cni-bridge: Error getting metadata containers: %v", err)
			return emptyMACAddress
		}

		for _, container := range containers {
			if container.ExternalId == cid && container.PrimaryMacAddress != "" {
				logrus.Infof("rancher-cni-bridge: got MAC address: %v for container: %v", container.PrimaryMacAddress, container.ExternalId)
				return container.PrimaryMacAddress
			}
			if rancherid != "" && container.UUID == rancherid && container.PrimaryMacAddress != "" {
				logrus.Infof("rancher-cni-bridge: got MAC address from rancherid: %v for container: %v", container.PrimaryMacAddress, container.UUID)
				return container.PrimaryMacAddress
			}
		}
		logrus.Infof("Waiting to find MAC address for container: %s, %s", cid, rancherid)
		time.Sleep(500 * time.Millisecond)
	}
	logrus.Infof("MAC address not found for cid: %v", cid)
	return emptyMACAddress
}
