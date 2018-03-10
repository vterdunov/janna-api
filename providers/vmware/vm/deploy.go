package vm

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/pkg/errors"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/ovf"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/progress"
	"github.com/vmware/govmomi/vim25/soap"
	"github.com/vmware/govmomi/vim25/types"
	"github.com/vterdunov/janna-api/config"
)

type ovfx struct {
	Name string

	Client       *vim25.Client
	Datacenter   *object.Datacenter
	Datastore    *object.Datastore
	ResourcePool *object.ResourcePool
}

// Network represents VM network
type Network struct {
	Name    string
	Network string
}

type progressLogger struct {
	prefix string

	wg sync.WaitGroup

	sink chan chan progress.Report
	done chan struct{}
}

func (p *progressLogger) loopA() {
	var err error

	defer p.wg.Done()

	tick := time.NewTicker(100 * time.Millisecond)
	defer tick.Stop()

	called := false

	for stop := false; !stop; {
		select {
		case ch := <-p.sink:
			err = p.loopB(tick, ch)
			stop = true
			called = true
		case <-p.done:
			stop = true
		case <-tick.C:
			line := fmt.Sprintf("\r%s", p.prefix)
			fmt.Println(line)
		}
	}

	if err != nil && err != io.EOF {
		fmt.Println(fmt.Sprintf("\r%sError: %s\n", p.prefix, err))
	} else if called {
		fmt.Println(fmt.Sprintf("\r%sOK\n", p.prefix))
	}
}

// loopA runs after Sink() has been called.
func (p *progressLogger) loopB(tick *time.Ticker, ch <-chan progress.Report) error {
	var r progress.Report
	var ok bool
	var err error

	for ok = true; ok; {
		select {
		case r, ok = <-ch:
			if !ok {
				break
			}
			err = r.Error()
		case <-tick.C:
			line := fmt.Sprintf("\r%s", p.prefix)
			if r != nil {
				line += fmt.Sprintf("(%.0f%%", r.Percentage())
				detail := r.Detail()
				if detail != "" {
					line += fmt.Sprintf(", %s", detail)
				}
				line += ")"
			}
			fmt.Println(line)
		}
	}

	return err
}
func (p *progressLogger) Wait() {
	close(p.done)
	p.wg.Wait()
}

func (p *progressLogger) Sink() chan<- progress.Report {
	ch := make(chan progress.Report)
	p.sink <- ch
	return ch
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

	finder := find.NewFinder(c, true)

	dc, err := finder.DatacenterOrDefault(ctx, cfg.Vmware.DC)
	if err != nil {
		logger.Log("err", errors.Wrap(err, "Could not get Datacenter"))
		return jid, err
	}
	finder.SetDatacenter(dc)
	deployment.Datacenter = dc

	// TODO: try to use DatastoreCLuster instead of Datastore
	//   user can choose that want to use
	ds, err := finder.DatastoreOrDefault(ctx, cfg.Vmware.DS)
	if err != nil {
		logger.Log("err", err)
		return jid, err
	}

	deployment.Datastore = ds

	rp, err := finder.ResourcePoolOrDefault(ctx, cfg.Vmware.RP)
	if err != nil {
		logger.Log("err", err)
		return jid, err
	}

	deployment.ResourcePool = rp

	// ---------------------------------
	// TODO: OVF is work. Need to try work with OVA
	f, err := os.Open("vyacheslav.terdunov.test.ovf")
	if err != nil {
		logger.Log("err", err)
		return jid, err
	}

	readOVF, err := ioutil.ReadAll(f)
	if err != nil {
		logger.Log("err", err)
		return jid, err
	}
	f.Close()

	r := bytes.NewReader(readOVF)

	e, errUNM := ovf.Unmarshal(r)
	if errUNM != nil {
		logger.Log("err", errUNM)
		return jid, err
	}

	name := "Govc Virtual Appliance"
	if e.VirtualSystem != nil {
		name = e.VirtualSystem.ID
		if e.VirtualSystem.Name != nil {
			name = *e.VirtualSystem.Name
		}
	}

	var nm []types.OvfNetworkMapping
	networks := map[string]string{}

	if e.Network != nil {
		logger.Log("msg", "network is NOT null")
		for _, net := range e.Network.Networks {
			networks[net.Name] = net.Name
		}
	}
	// fmt.Println(networks)
	// spew.Dump(networks)
	// networks["dv-net-27"] = "dv-net-27"

	// net, errN := finder.Network(ctx, "dv-net-27")
	// if errN != nil {
	// 	logger.Log("errN", errN)
	// }
	// logger.Log("msg", "Found network")

	// nm = append(nm, types.OvfNetworkMapping{
	// 	Name:    "dv-net-27",
	// 	Network: net.Reference(),
	// })

	// spew.Dump(e.Network.Networks)

	// spew.Dump(nm)

	for src, dst := range networks {
		if net, errN := finder.Network(ctx, dst); errN == nil {
			nm = append(nm, types.OvfNetworkMapping{
				Name:    src,
				Network: net.Reference(),
			})
		}
	}
	// spew.Dump(nm)
	cisp := types.OvfCreateImportSpecParams{
		EntityName:     name,
		NetworkMapping: nm,
	}

	m := ovf.NewManager(c)
	spec, err := m.CreateImportSpec(ctx, string(readOVF), rp, ds, cisp)
	if err != nil {
		logger.Log("err", err)
		return jid, err
	}
	if spec.Error != nil {
		logger.Log("err", errors.New(spec.Error[0].LocalizedMessage))
		return jid, errors.New(spec.Error[0].LocalizedMessage)
	}
	if spec.Warning != nil {
		for _, w := range spec.Warning {
			logger.Log("Warning", w)
		}
	}

	host, err := finder.HostSystemOrDefault(ctx, "vi-devops-esx7.lab.vi.local")
	if err != nil {
		logger.Log("err", err)
		return jid, err
	}

	folder, err := finder.FolderOrDefault(ctx, "vagrant")
	if err != nil {
		logger.Log("err", err)
		return jid, err
	}

	lease, err := rp.ImportVApp(ctx, spec.ImportSpec, folder, host)
	if err != nil {
		logger.Log("err", err)
		return jid, err
	}

	info, err := lease.Wait(ctx, spec.FileItem)
	if err != nil {
		logger.Log("err", err)
		return jid, err
	}

	u := lease.StartUpdater(ctx, info)
	defer u.Done()

	for _, item := range info.Items {
		file := item.Path
		f, errOpen := os.Open(file)
		if errOpen != nil {
			logger.Log("err", err)
			return jid, err
		}
		defer f.Close()

		s, errStat := f.Stat()
		if errStat != nil {
			logger.Log("err", err)
			return jid, err

		}
		st := fmt.Sprintf("Uploading %s... ", path.Base(file))
		pl := newProgressLogger(st)
		defer pl.Wait()

		opts := soap.Upload{
			ContentLength: s.Size(),
			Progress:      pl,
		}
		lease.Upload(ctx, item, f, opts)

	}
	lease.Complete(ctx)

	moref := &info.Entity
	vm := object.NewVirtualMachine(c, *moref)

	// PowerON
	logger.Log("msg", "Powering on...")
	task, err := vm.PowerOn(ctx)
	if err != nil {
		logger.Log("err", errors.Wrap(err, "Could not power on VM"))
		return jid, err
	}
	if _, err = task.WaitForResult(ctx, nil); err != nil {
		logger.Log("err", errors.Wrap(err, "Failed to wait powering on task"))
	}

	// WaitForIP
	logger.Log("msg", "Waiting for ip")
	ctx, cancel := context.WithTimeout(ctx, 1*time.Minute)
	defer cancel()

	ip, err := vm.WaitForIP(ctx)
	if err != nil {
		logger.Log("err", errors.Wrap(err, "Could not get IP address"))
		return jid, err
	}
	logger.Log("msg", "Received IP address", "ip", ip)
	// end ReadOvf

	logger.Log(
		"msg", "Deploy OVA",
		"name", vmName,
		"ova_url", OVAURL,
	)
	return jid, nil
}

func newProgressLogger(prefix string) *progressLogger {
	p := &progressLogger{
		prefix: prefix,

		sink: make(chan chan progress.Report),
		done: make(chan struct{}),
	}

	p.wg.Add(1)

	go p.loopA()

	return p
}
