package endpoint

import (
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"

	"github.com/vterdunov/janna-api/pkg/service"
)

// Endpoints collects all of the endpoints that compose the Service.
type Endpoints struct {
	InfoEndpoint                  endpoint.Endpoint
	ReadyzEndpoint                endpoint.Endpoint
	HealthzEndpoint               endpoint.Endpoint
	VMListEndpoint                endpoint.Endpoint
	VMInfoEndpoint                endpoint.Endpoint
	VMFindEndpoint                endpoint.Endpoint
	VMDeployEndpoint              endpoint.Endpoint
	VMSnapshotsListEndpoint       endpoint.Endpoint
	VMSnapshotCreateEndpoint      endpoint.Endpoint
	VMSnapshotDeleteEndpoint      endpoint.Endpoint
	VMRestoreFromSnapshotEndpoint endpoint.Endpoint
}

// New returns an Endpoints struct where each endpoint invokes
// the corresponding method on the provided service.
func New(s service.Service, logger log.Logger) Endpoints {
	infoEndpoint := MakeInfoEndpoint(s)
	infoEndpoint = LoggingMiddleware(log.With(logger, "endpoint", "Info"))(infoEndpoint)

	healthzEndpoint := MakeHealthzEndpoint(s)

	readyzEndpoint := MakeReadyzEndpoint(s)

	vmListEndpoint := MakeVMListEndpoint(s)
	vmListEndpoint = LoggingMiddleware(log.With(logger, "endpoint", "VMList"))(vmListEndpoint)

	vmInfoEndpoint := MakeVMInfoEndpoint(s)
	vmInfoEndpoint = LoggingMiddleware(log.With(logger, "endpoint", "VMInfo"))(vmInfoEndpoint)

	vmFindEndpoint := MakeVMFindEndpoint(s)
	vmFindEndpoint = LoggingMiddleware(log.With(logger, "endpoint", "VMFind"))(vmFindEndpoint)

	vmDeployEndpoint := MakeVMDeployEndpoint(s, logger)
	vmDeployEndpoint = LoggingMiddleware(log.With(logger, "endpoint", "VMDeploy"))(vmDeployEndpoint)

	vmSnapshotsListEndpoint := MakeVMSnapshotsListEndpoint(s)
	vmSnapshotsListEndpoint = LoggingMiddleware(log.With(logger, "endpoint", "VMSnapshotsList"))(vmSnapshotsListEndpoint)

	vmSnapshotCreateEndpoint := MakeVMSnapshotCreateEndpoint(s)
	vmSnapshotCreateEndpoint = LoggingMiddleware(log.With(logger, "endpoint", "VMSnapshotCreate"))(vmSnapshotCreateEndpoint)

	vmRestoreFromSnapshotEndpoint := MakeVMRestoreFromSnapshotEndpoint(s)
	vmRestoreFromSnapshotEndpoint = LoggingMiddleware(log.With(logger, "endpoint", "VMRestoreFromSnapshot"))(vmRestoreFromSnapshotEndpoint)

	vmSnapshotDeleteEndpoint := MakeVMSnapshotDeleteEndpoint(s)
	vmSnapshotDeleteEndpoint = LoggingMiddleware(log.With(logger, "endpoint", "VMSnapshotDelete"))(vmSnapshotDeleteEndpoint)

	return Endpoints{
		InfoEndpoint:                  infoEndpoint,
		HealthzEndpoint:               healthzEndpoint,
		ReadyzEndpoint:                readyzEndpoint,
		VMListEndpoint:                vmListEndpoint,
		VMInfoEndpoint:                vmInfoEndpoint,
		VMFindEndpoint:                vmFindEndpoint,
		VMDeployEndpoint:              vmDeployEndpoint,
		VMSnapshotsListEndpoint:       vmSnapshotsListEndpoint,
		VMSnapshotCreateEndpoint:      vmSnapshotCreateEndpoint,
		VMSnapshotDeleteEndpoint:      vmSnapshotDeleteEndpoint,
		VMRestoreFromSnapshotEndpoint: vmRestoreFromSnapshotEndpoint,
	}
}

// Failer is an interface that should be implemented by response types.
// Response encoders can check if responses are Failer, and if so they've
// failed, and if so encode them using a separate write path based on the error.
type Failer interface {
	Failed() error
}
