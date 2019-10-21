package bwreport

import (
	context "context"
	"lproxy/server"

	log "github.com/sirupsen/logrus"
)

type myReportService struct {
}

func (s *myReportService) Report(c context.Context, r *BandwidthStatistics) (*ReportResult, error) {
	statistics := r.GetStatistics()
	for _, s := range statistics {
		log.Printf("gRPC report service, uuid:%s, send:%d, recv:%d",
			s.GetUuid(), s.GetSendBytes(), s.GetRecvBytes())
	}

	reply := &ReportResult{Code: 0}
	return reply, nil
}

func init() {
	server.InvokeAfterCfgLoaded(func() {
		s := server.GetGRPCServer()
		RegisterBandwidthReportServer(s, &myReportService{})
	})
}
