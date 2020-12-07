// generated by directives_generate.go; DO NOT EDIT

package dnsserver

// Directives are registered in the order they should be
// executed.
//
// Ordering is VERY important. Every plugin will
// feel the effects of all other plugin below
// (after) them during a request, but they must not
// care what plugin above them are doing.
var Directives = []string{
	"metadata",
	"cancel",
	"tls",
	"reload",
	"nsid",
	"root",
	"bind",
	"debug",
	"trace",
	"ready",
	"health",
	"pprof",
	"prometheus",
	"errors",
	"log",
	"dnstap",
	"chaos",
	"iplookup",
	"loadbalance",
	"cache",
	"domainreport",
	"iplookup",
	"rewrite",
	"dnssec",
	"autopath",
	"template",
	"hosts",
	"route53",
	"federation",
	"k8s_external",
	"kubernetes",
	"file",
	"auto",
	"secondary",
	"etcd",
	"loop",
	"alternate",
	"forward",
	"grpc",
	"erratic",
	"whoami",
	"on",
}
