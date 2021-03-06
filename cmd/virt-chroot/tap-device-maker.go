package main

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
)

func createTapDevice(name string, owner uint, group uint, queueNumber int, mtu int) error {
	var err error = nil
	tapDevice := &netlink.Tuntap{
		LinkAttrs:  netlink.LinkAttrs{Name: name},
		Mode:       unix.IFF_TAP,
		NonPersist: false,
		Queues:     queueNumber,
		Owner:      uint32(owner),
		Group:      uint32(group),
	}

	// when netlink receives a request for a tap device with 1 queue, it uses
	// the MULTI_QUEUE flag, which differs from libvirt; as such, we need to
	// manually request the single queue flags, enabling libvirt to consume
	// the tap device.
	// See https://github.com/vishvananda/netlink/issues/574
	if queueNumber == 1 {
		tapDevice.Flags = netlink.TUNTAP_DEFAULTS
	}
	if err := netlink.LinkAdd(tapDevice); err != nil {
		return fmt.Errorf("failed to create tap device named %s. Reason: %v", name, err)
	}

	if err := netlink.LinkSetMTU(tapDevice, mtu); err != nil {
		return fmt.Errorf("failed to set MTU on tap device named %s. Reason: %v", name, err)
	}

	return err
}

func NewCreateTapCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "create-tap",
		Short: "create a tap device in a given PID net ns",
		RunE: func(cmd *cobra.Command, args []string) error {
			tapName := cmd.Flag("tap-name").Value.String()
			uidStr := cmd.Flag("uid").Value.String()
			gidStr := cmd.Flag("gid").Value.String()
			queueNumber, err := cmd.Flags().GetUint32("queue-number")
			if err != nil {
				return fmt.Errorf("could not access queue-number parameter: %v", err)
			}
			mtu, err := cmd.Flags().GetUint32("mtu")
			if err != nil {
				return fmt.Errorf("could not access mtu parameter: %v", err)
			}

			uid, err := strconv.ParseUint(uidStr, 10, 32)
			if err != nil {
				return fmt.Errorf("could not parse tap device owner: %v", err)
			}
			gid, err := strconv.ParseUint(gidStr, 10, 32)
			if err != nil {
				return fmt.Errorf("could not parse tap device group: %v", err)
			}

			if err := createTapDevice(tapName, uint(uid), uint(gid), int(queueNumber), int(mtu)); err != nil {
				return fmt.Errorf("failed to create tap device named %s. Reason: %v", tapName, err)
			}

			return nil
		},
	}
}
