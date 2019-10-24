package dv

import (
	context "context"
	"lproxy/server"
	"lproxy/servercfg"

	log "github.com/sirupsen/logrus"
)

type myReportService struct {
}
type myDvImportService struct{}

func (s *myReportService) Report(c context.Context, r *BandwidthStatistics) (*ReportResult, error) {
	statistics := r.GetStatistics()
	for _, s := range statistics {
		log.Printf("gRPC report service, uuid:%s, send:%d, recv:%d",
			s.GetUuid(), s.GetSendBytes(), s.GetRecvBytes())
	}

	reply := &ReportResult{Code: 0}
	return reply, nil
}

func (s *myDvImportService) PullCfg(ctx context.Context, req *CfgPullRequest) (*CfgPullResult, error) {
	log.Println("gRPC PullCfg called, uuid:", req.GetUuid())

	reply := &CfgPullResult{Code: 0, BandwidthLimitKbs: uint64(servercfg.BandwidthKbs)}

	return reply, nil
}

func init() {
	server.InvokeAfterCfgLoaded(func() {
		s := server.GetGRPCServer()
		RegisterBandwidthReportServer(s, &myReportService{})
		RegisterDeviceCfgPullServer(s, &myDvImportService{})
	})
}
