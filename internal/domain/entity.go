// domain contains the entities of the domain model
package domain

import (
	"context"
	"time"

	"github.com/vmware/govmomi/object"
)

// Finder provides access a to VirtualMachine inventory.
type Finder interface {
	FindVMByUUID(uuid string) (*VirtualMachine, error)
}

type VirtualMachine struct {
	vmareVM *object.VirtualMachine
}

func NewWithObjectVM(vmareVM *object.VirtualMachine) *VirtualMachine {
	return &VirtualMachine{
		vmareVM: vmareVM,
	}
}

func (vm *VirtualMachine) Rename(ctx context.Context, name string) error {
	task, err := vm.vmareVM.Rename(ctx, name)
	if err != nil {
		return err
	}

	return task.Wait(ctx)
}

// VMSummary stores some information about Virtual Machines
type VMSummary struct {
	BootTime            time.Time
	Name                string
	UUID                string
	GuestID             string
	Annotation          string
	PowerState          string
	ConnectionState     string
	NumCPU              int32
	NumEthernetCards    int32
	NumVirtualDisks     int32
	Paused              bool
	ConsolidationNeeded bool
	Template            bool
	VMGuestInfo
}

type VMGuestInfo struct {
	GuestID            string
	GuestFullName      string
	ToolsRunningStatus string
	HostName           string
	IPAddress          string
}

// VMUuid saves a VM uuid
type VMUuid struct {
	Name string
	UUID string
}

// Snapshot stores info about VM snapshot
type Snapshot struct {
	Name        string
	Description string
	ID          int32
	CreatedAt   time.Time
}

// Role represents ESXi role
type Role struct {
	Name        string
	ID          int32
	Description struct {
		Label   string
		Summary string
	}
}
