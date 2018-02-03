package vm

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/url"

	"github.com/vmware/govmomi/ovf"
	"github.com/vmware/govmomi/vim25/soap"

	"github.com/go-kit/kit/log"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25"
	"github.com/vterdunov/janna-api/config"
)

type ovfx struct {
	Name string

	Client       *vim25.Client
	Datacenter   *object.Datacenter
	Datastore    *object.Datastore
	ResourcePool *object.ResourcePool
}

type Section struct {
	Required *bool  `xml:"required,attr"`
	Info     string `xml:"Info"`
}

type AnnotationSection struct {
	Section

	Annotation string `xml:"Annotation"`
}

// Deploy returns summary information about Virtual Machines
func Deploy(ctx context.Context, vmName string, OVAURL string, logger log.Logger, cfg *config.Config, c *vim25.Client, opts ...string) (int, error) {
	// TODO: make up a metod to check deploy progress.
	// Job ID and endpoint with status?
	// keep HTTP connection with client and poll it?
	var jid int

	deployment := &ovfx{}
	deployment.Name = vmName
	deployment.Client = c

	f := find.NewFinder(c, true)

	dc, err := f.DatacenterOrDefault(ctx, cfg.Vmware.DC)
	if err != nil {
		logger.Log("err", err)
		return jid, err
	}
	f.SetDatacenter(dc)
	deployment.Datacenter = dc

	// TODO: try to use DatastoreCLuster instead of Datastore
	//   user can choose that want to use
	ds, err := f.DatastoreOrDefault(ctx, cfg.Vmware.DS)
	if err != nil {
		logger.Log("err", err)
		return jid, err
	}

	deployment.Datastore = ds

	rp, err := f.ResourcePoolOrDefault(ctx, cfg.Vmware.RP)
	if err != nil {
		logger.Log("err", err)
		return jid, err
	}

	deployment.ResourcePool = rp

	// -----------------------------------------
	// ReadOvf
	u, err := url.Parse(OVAURL)
	if err != nil {
		logger.Log("err", err)
		return jid, err
	}

	//  f, _, err := t.OpenFile(t.path). Parse and download -> ReadCloser
	fil, _, err := c.Download(u, &soap.DefaultDownload)
	if err != nil {
		logger.Log("err", err)
		return jid, err
	}

	// func (t *TapeArchive) Open(name string) (io.ReadCloser, int64, error) {
	// rdr := tar.NewReader(fil)
	// r := &TapeArchiveEntry{rdr, fil}
	// fmt.Println(fil)
	// file.Close()

	o, err := ioutil.ReadAll(fil)
	if err != nil {
		logger.Log("err", "err1")
	}
	// end ReadOvf
	// -----------------------------------------

	// -----------------------------------------
	// ReadEnvelope
	fmt.Println("HERE")
	r := bytes.NewReader(o)
	e := &ovf.Envelope{}
	dec := xml.NewDecoder(r)
	err = dec.Decode(&e)
	if err != nil {
		logger.Log("err", "err")
	}

	// br := bytes.NewReader(o)
	// e, err = ovf.Unmarshal(br)
	// if err != nil {
	// 	logger.Log("err", err)
	// 	return jid, err
	// }

	// e, err := ovf.Unmarshal(re)
	// if err != nil {
	// 	logger.Log("err", err)
	// 	return jid, err
	// }

	fmt.Println("++++++")
	// fmt.Println(e.Annotation)
	fmt.Println(e.VirtualSystem.Name)
	fmt.Println("++++++")
	// +1) create empty struct that represents a deploy object

	// 2) Run a chain of calls:
	// 3) "Prepare":
	// 	 - validate OVA URL (try to use vim25.Client as Opener it has OpemRemote method)
	//   - fil OVA struct with: vim25.Client, Datacenter, Datastore, ResourcePool
	// 4) Download OVA OR use vim25.Client as Opener (see govc importx/archive.go:143)
	// 5) Import OVA (see govc importx)/ovf.go:212) it returns *types.ManagedObjectReference
	//   - Read file to []byte and unmarshal it to ovf.Envelope. The type allows to access to fileds of OVF strucure
	//   - Create a struct that represents a pert of OVF spec: name, nwtwork, etc.
	// 6) Create OVF Manager (ovf.NewManager)
	// 7) Fil/overrride the spec into current OVF (ovf.CreateImportSpec)
	// ?
	// 9) Get VM Folder.
	// 10) Crate Lease object (optional)
	// 11) Upload OVF
	// 12) ... Start

	logger.Log(
		"msg", "Deploy OVA",
		"name", vmName,
		"ova_url", OVAURL,
	)
	return jid, nil
}
