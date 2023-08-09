package collector

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/antchfx/xmlquery"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace = "nginx_rtmp"
)

func newServerMetric(metricName string, docString string, varLabels []string, constLabels prometheus.Labels) *prometheus.Desc {
	return prometheus.NewDesc(prometheus.BuildFQName(namespace, "server", metricName), docString, varLabels, constLabels)
}

func newStreamMetric(metricName string, docString string, varLabels []string, constLabels prometheus.Labels) *prometheus.Desc {
	return prometheus.NewDesc(prometheus.BuildFQName(namespace, "stream", metricName), docString, varLabels, constLabels)
}

type metrics map[string]*prometheus.Desc

var (
	serverMetrics = metrics{
		"bytesIn":        newServerMetric("incoming_bytes_total", "Current total of incoming bytes", nil, nil),
		"bytesOut":       newServerMetric("outgoing_bytes_total", "Current total of outgoing bytes", nil, nil),
		"bandwidthIn":    newServerMetric("receive_bytes", "Current bandwidth in per second", nil, nil),
		"bandwidthOut":   newServerMetric("transmit_bytes", "Current bandwidth out per second", nil, nil),
		"currentStreams": newServerMetric("current_streams", "Current number of active streams", nil, nil),
		"uptime":         newServerMetric("uptime_seconds_total", "Number of seconds NGINX-RTMP started", nil, nil),
	}
	streamMetrics = metrics{
		"bytesIn":      newStreamMetric("incoming_bytes_total", "Current total of incoming bytes", []string{"stream"}, nil),
		"bytesOut":     newStreamMetric("outgoing_bytes_total", "Current total of outgoing bytes", []string{"stream"}, nil),
		"bandwidthIn":  newStreamMetric("receive_bytes", "Current bandwidth in per second", []string{"stream"}, nil),
		"bandwidthOut": newStreamMetric("transmit_bytes", "Current bandwidth out per second", []string{"stream"}, nil),
		"uptime":       newStreamMetric("uptime_seconds_total", "Number of seconds since the stream started", []string{"stream"}, nil),
	}
)

// Exporter collects NGINX-RTMP stats from the status page URI
// using the prometheus metrics package
type RTMPExporter struct {
	URI                  string
	TIMEOUT              time.Duration
	mutex                sync.RWMutex
	fetch                func() (io.ReadCloser, error)
	streamNameNormalizer *regexp.Regexp
	logger               log.Logger

	serverMetrics map[string]*prometheus.Desc
	streamMetrics map[string]*prometheus.Desc
}

// ServerInfo characteristics of the RTMP server
type ServerInfo struct {
	BytesIn     float64
	BytesOut    float64
	BandwidthIn float64
	BandwidhOut float64
	Uptime      float64
}

// StreamInfo characteristics of a stream
type StreamInfo struct {
	Name        string
	BytesIn     float64
	BytesOut    float64
	BandwidthIn float64
	BandwidhOut float64
	Uptime      float64
}

// NewServerInfo builds a ServerInfo struct from string values
func (e *RTMPExporter) NewServerInfo(bytesIn, bytesOut, bandwidthIn, bandwidthOut, uptime string) ServerInfo {
	var bytesInNum, bytesOutNum, bandwidthInNum, bandwidthOutNum, uptimeNum float64
	if n, err := strconv.ParseFloat(bytesIn, 64); err == nil {
		bytesInNum = n
	}
	if n, err := strconv.ParseFloat(bytesOut, 64); err == nil {
		bytesOutNum = n
	}
	if n, err := strconv.ParseFloat(bandwidthIn, 64); err == nil {
		bandwidthInNum = n / 1048576 // bandwidth is in bits
	}
	if n, err := strconv.ParseFloat(bandwidthOut, 64); err == nil {
		bandwidthOutNum = n / 1048576 // bandwidth is in bits
	}
	if n, err := strconv.ParseFloat(uptime, 64); err == nil {
		uptimeNum = n
	}
	return ServerInfo{
		BytesIn:     bytesInNum,
		BytesOut:    bytesOutNum,
		BandwidthIn: bandwidthInNum,
		BandwidhOut: bandwidthOutNum,
		Uptime:      uptimeNum,
	}
}

// NewStreamInfo builds a StreamInfo struct from string values
func (e *RTMPExporter) NewStreamInfo(name, bytesIn, bytesOut, bandwidthIn, bandwidthOut, uptime string) StreamInfo {
	var bytesInNum, bytesOutNum, bandwidthInNum, bandwidthOutNum, uptimeNum float64
	if n, err := strconv.ParseFloat(bytesIn, 64); err == nil {
		bytesInNum = n
	}
	if n, err := strconv.ParseFloat(bytesOut, 64); err == nil {
		bytesOutNum = n
	}
	if n, err := strconv.ParseFloat(bandwidthIn, 64); err == nil {
		bandwidthInNum = n / 1048576 // bandwidth is in bits
	}
	if n, err := strconv.ParseFloat(bandwidthOut, 64); err == nil {
		bandwidthOutNum = n / 1048576 // bandwidth is in bits
	}
	if n, err := strconv.ParseFloat(uptime, 64); err == nil {
		uptimeNum = n / 1000 // it is in miliseconds
	}
	return StreamInfo{
		Name:        name,
		BytesIn:     bytesInNum,
		BytesOut:    bytesOutNum,
		BandwidthIn: bandwidthInNum,
		BandwidhOut: bandwidthOutNum,
		Uptime:      uptimeNum,
	}
}

func (e *RTMPExporter) NewRTMPExporter(uri string, timeout time.Duration, streamNameNormalizer *regexp.Regexp, logger log.Logger) (*RTMPExporter, error) {
	return &RTMPExporter{
		URI:                  uri,
		fetch:                e.fetchStats(uri, timeout),
		streamNameNormalizer: streamNameNormalizer,
		logger:               logger,

		serverMetrics: serverMetrics,
		streamMetrics: streamMetrics,
	}, nil
}

func (e *RTMPExporter) fetchStats(uri string, timeout time.Duration) func() (io.ReadCloser, error) {
	client := http.Client{
		Timeout: timeout,
	}

	return func() (io.ReadCloser, error) {
		resp, err := client.Get(uri)
		if err != nil {
			return nil, err
		}
		if !(resp.StatusCode >= 200 && resp.StatusCode < 300) {
			resp.Body.Close()
			return nil, fmt.Errorf("HTTP status %d", resp.StatusCode)
		}
		return resp.Body, nil
	}
}

func (e *RTMPExporter) Collect(ch chan<- prometheus.Metric) {
	e.mutex.Lock() // To protect from concurrent collects
	defer e.mutex.Unlock()

	e.scrape(ch)
}

func (e *RTMPExporter) parseServerStats(doc *xmlquery.Node) (ServerInfo, error) {
	data := xmlquery.FindOne(doc, "//rtmp")

	bytesIn := data.SelectElement("bytes_in").InnerText()
	bytesOut := data.SelectElement("bytes_out").InnerText()
	receiveBytes := data.SelectElement("bw_in").InnerText()
	transmitBytes := data.SelectElement("bw_out").InnerText()
	uptime := data.SelectElement("uptime").InnerText()

	return e.NewServerInfo(bytesIn, bytesOut, receiveBytes, transmitBytes, uptime), nil
}

func (e *RTMPExporter) parseStreamsStats(doc *xmlquery.Node, streamNameNormalizer *regexp.Regexp) ([]StreamInfo, error) {
	streams := make([]StreamInfo, 0)
	data := xmlquery.Find(doc, "//stream")

	for _, stream := range data {
		name := streamNameNormalizer.FindString(stream.SelectElement("name").InnerText())
		// adding the app name here to ensure that the metrics are unique
		app := ""
		if stream.Parent != nil && stream.Parent.Parent != nil {
			appName := stream.Parent.Parent.SelectElement("name")
			if appName != nil {
				app = appName.InnerText() + "-" // dash separator between app and stream names
			}
		}
		bytesIn := stream.SelectElement("bytes_in").InnerText()
		bytesOut := stream.SelectElement("bytes_out").InnerText()
		receiveBytes := stream.SelectElement("bw_in").InnerText()
		transmitBytes := stream.SelectElement("bw_out").InnerText()
		uptime := stream.SelectElement("time").InnerText()
		streams = append(streams, e.NewStreamInfo(app+name, bytesIn, bytesOut, receiveBytes, transmitBytes, uptime))
	}
	return streams, nil
}

func (e *RTMPExporter) scrape(ch chan<- prometheus.Metric) {
	data, err := e.fetch()
	if err != nil {
		level.Error(e.logger).Log("msg", "Can't scrape NGINX-RTMP", "err", err)
		return
	}
	defer data.Close()

	doc, err := xmlquery.Parse(data)
	if err != nil {
		return
	}

	server, err := e.parseServerStats(doc)
	if err != nil {
		level.Error(e.logger).Log("msg", "Can't parse XML", "err", err)
		return
	}
	ch <- prometheus.MustNewConstMetric(e.serverMetrics["bytesIn"], prometheus.CounterValue, server.BytesIn)
	ch <- prometheus.MustNewConstMetric(e.serverMetrics["bytesOut"], prometheus.CounterValue, server.BytesOut)
	ch <- prometheus.MustNewConstMetric(e.serverMetrics["bandwidthIn"], prometheus.GaugeValue, server.BandwidthIn)
	ch <- prometheus.MustNewConstMetric(e.serverMetrics["bandwidthOut"], prometheus.GaugeValue, server.BandwidhOut)
	ch <- prometheus.MustNewConstMetric(e.serverMetrics["uptime"], prometheus.CounterValue, server.Uptime)

	streams, err := e.parseStreamsStats(doc, e.streamNameNormalizer)
	if err != nil {
		level.Error(e.logger).Log("msg", "Can't parse XML", "err", err)
		return
	}

	for _, stream := range streams {
		ch <- prometheus.MustNewConstMetric(e.streamMetrics["bytesIn"], prometheus.CounterValue, stream.BytesIn, stream.Name)
		ch <- prometheus.MustNewConstMetric(e.streamMetrics["bytesOut"], prometheus.CounterValue, stream.BytesOut, stream.Name)
		ch <- prometheus.MustNewConstMetric(e.streamMetrics["bandwidthIn"], prometheus.GaugeValue, stream.BandwidthIn, stream.Name)
		ch <- prometheus.MustNewConstMetric(e.streamMetrics["bandwidthOut"], prometheus.GaugeValue, stream.BandwidhOut, stream.Name)
		ch <- prometheus.MustNewConstMetric(e.streamMetrics["uptime"], prometheus.CounterValue, stream.Uptime, stream.Name)
	}

	ch <- prometheus.MustNewConstMetric(e.serverMetrics["currentStreams"], prometheus.GaugeValue, float64(len(streams)))
}

// Describe describes all metrics to be exported to Prometheus
func (e *RTMPExporter) Describe(ch chan<- *prometheus.Desc) {
	for _, metric := range e.serverMetrics {
		ch <- metric
	}

	for _, metric := range e.streamMetrics {
		ch <- metric
	}
}
