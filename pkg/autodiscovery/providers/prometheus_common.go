// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-2020 Datadog, Inc.

package providers

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	// Default openmetrics check configuration values
	openmetricsCheckName   = "openmetrics"
	openmetricsInitConfig  = "{}"
	openmetricsURLPrefix   = "http://%%host%%:"
	openmetricsDefaultPort = "%%port%%"
	openmetricsDefaultPath = "/metrics"
	openmetricsDefaultNS   = ""

	// Prometheus annnotation keys
	prometheusScrapeAnnotation = "prometheus.io/scrape"
	prometheusPathAnnotation   = "prometheus.io/path"
	prometheusPortAnnotation   = "prometheus.io/port"
)

var (
	openmetricsDefaultMetrics = []string{"*"}
)

// PrometheusCheck represents the openmetrics check instances and the corresponding autodiscovery rules
type PrometheusCheck struct {
	Instances []*OpenmetricsInstance `mapstructure:"configurations"`
	AD        *ADConfig              `mapstructure:"autodiscovery"`
}

// OpenmetricsInstance contains the openmetrics check instance fields
type OpenmetricsInstance struct {
	URL                           string                      `mapstructure:"prometheus_url" json:"prometheus_url,omitempty"`
	Namespace                     string                      `mapstructure:"namespace" json:"namespace"`
	Metrics                       []string                    `mapstructure:"metrics" json:"metrics,omitempty"`
	Prefix                        string                      `mapstructure:"prometheus_metrics_prefix" json:"prometheus_metrics_prefix,omitempty"`
	HealthCheck                   bool                        `mapstructure:"health_service_check" json:"health_service_check,omitempty"`
	LabelToHostname               bool                        `mapstructure:"label_to_hostname" json:"label_to_hostname,omitempty"`
	LabelJoins                    map[string]LabelJoinsConfig `mapstructure:"label_joins" json:"label_joins,omitempty"`
	LabelsMapper                  map[string]string           `mapstructure:"labels_mapper" json:"labels_mapper,omitempty"`
	TypeOverride                  map[string]string           `mapstructure:"type_overrides" json:"type_overrides,omitempty"`
	HistogramBuckets              bool                        `mapstructure:"send_histograms_buckets" json:"send_histograms_buckets,omitempty"`
	DistribuitionBuckets          bool                        `mapstructure:"send_distribution_buckets" json:"send_distribution_buckets,omitempty"`
	MonotonicCounter              bool                        `mapstructure:"send_monotonic_counter" json:"send_monotonic_counter,omitempty"`
	DistributionCountsAsMonotonic bool                        `mapstructure:"send_distribution_counts_as_monotonic" json:"send_distribution_counts_as_monotonic,omitempty"`
	DistributionSumsAsMonotonic   bool                        `mapstructure:"send_distribution_sums_as_monotonic" json:"send_distribution_sums_as_monotonic,omitempty"`
	ExcludeLabels                 []string                    `mapstructure:"exclude_labels" json:"exclude_labels,omitempty"`
	BearerTokenAuth               bool                        `mapstructure:"bearer_token_auth" json:"bearer_token_auth,omitempty"`
	BearerTokenPath               string                      `mapstructure:"bearer_token_path" json:"bearer_token_path,omitempty"`
	IgnoreMetrics                 []string                    `mapstructure:"ignore_metrics" json:"ignore_metrics,omitempty"`
	Proxy                         map[string]string           `mapstructure:"proxy" json:"proxy,omitempty"`
	SkipProxy                     bool                        `mapstructure:"skip_proxy" json:"skip_proxy,omitempty"`
	Username                      string                      `mapstructure:"username" json:"username,omitempty"`
	Password                      string                      `mapstructure:"password" json:"password,omitempty"`
	TLSVerify                     bool                        `mapstructure:"tls_verify" json:"tls_verify,omitempty"`
	TLSHostHeader                 bool                        `mapstructure:"tls_use_host_header" json:"tls_use_host_header,omitempty"`
	TLSIgnoreWarn                 bool                        `mapstructure:"tls_ignore_warning" json:"tls_ignore_warning,omitempty"`
	TLSCert                       string                      `mapstructure:"tls_cert" json:"tls_cert,omitempty"`
	TLSPrivateKey                 string                      `mapstructure:"tls_private_key" json:"tls_private_key,omitempty"`
	TLSCACert                     string                      `mapstructure:"tls_ca_cert" json:"tls_ca_cert,omitempty"`
	Headers                       map[string]string           `mapstructure:"headers" json:"headers,omitempty"`
	ExtraHeaders                  map[string]string           `mapstructure:"extra_headers" json:"extra_headers,omitempty"`
	Timeout                       int                         `mapstructure:"timeout" json:"timeout,omitempty"`
	Tags                          []string                    `mapstructure:"tags" json:"tags,omitempty"`
	Service                       string                      `mapstructure:"service" json:"service,omitempty"`
	MinCollectInterval            int                         `mapstructure:"min_collection_interval" json:"min_collection_interval,omitempty"`
	EmptyDefaultHost              bool                        `mapstructure:"empty_default_hostname" json:"empty_default_hostname,omitempty"`
}

// LabelJoinsConfig contains the label join configuration fields
type LabelJoinsConfig struct {
	LabelsToMatch []string `mapstructure:"labels_to_match" json:"labels_to_match"`
	LabelsToGet   []string `mapstructure:"labels_to_get" json:"labels_to_get"`
}

// ADConfig contains the autodiscovery configuration data for a PrometheusCheck
type ADConfig struct {
	ExcludeAutoconf    *bool     `mapstructure:"exclude_autoconfig_files"`
	KubeAnnotations    *InclExcl `mapstructure:"kubernetes_annotations"`
	KubeContainerNames []string  `mapstructure:"kubernetes_container_names"`
	containersRe       *regexp.Regexp
}

// InclExcl contains the include/exclude data structure
type InclExcl struct {
	Incl map[string]string `mapstructure:"include"`
	Excl map[string]string `mapstructure:"exclude"`
}

// defaultCheck has the default openmetrics check values
// To be used when the checks configuration is empty
var defaultCheck = &PrometheusCheck{
	Instances: []*OpenmetricsInstance{
		{
			Metrics:   openmetricsDefaultMetrics,
			Namespace: openmetricsDefaultNS,
		},
	},
	AD: &ADConfig{
		ExcludeAutoconf: boolPointer(true),
		KubeAnnotations: &InclExcl{
			Excl: map[string]string{prometheusScrapeAnnotation: "false"},
			Incl: map[string]string{prometheusScrapeAnnotation: "true"},
		},
		KubeContainerNames: []string{},
	},
}

// init prepares the PrometheusCheck structure and defaults its values
// init must be called only once
func (pc *PrometheusCheck) init() error {
	pc.initInstances()
	return pc.initAD()
}

// initInstances defaults the Instances field in PrometheusCheck
func (pc *PrometheusCheck) initInstances() {
	if len(pc.Instances) == 0 {
		// Put a default config
		pc.Instances = append(pc.Instances, &OpenmetricsInstance{
			Metrics:   openmetricsDefaultMetrics,
			Namespace: openmetricsDefaultNS,
		})
		return
	}

	for _, instance := range pc.Instances {
		// Default the required config values if not set
		if len(instance.Metrics) == 0 {
			instance.Metrics = openmetricsDefaultMetrics
		}
	}
}

// initAD defaults the AD field in PrometheusCheck
// It also prepares the regex to match the containers by name
func (pc *PrometheusCheck) initAD() error {
	if pc.AD == nil {
		pc.AD = &ADConfig{}
	}

	pc.AD.defaultAD()
	return pc.AD.setContainersRegex()
}

// defaultAD defaults the values of the autodiscovery structure
func (ad *ADConfig) defaultAD() {
	if ad.ExcludeAutoconf == nil {
		// TODO: Implement OOTB autoconf exclusion
		ad.ExcludeAutoconf = boolPointer(true)
	}

	if ad.KubeContainerNames == nil {
		ad.KubeContainerNames = []string{}
	}

	if ad.KubeAnnotations == nil {
		ad.KubeAnnotations = &InclExcl{
			Excl: map[string]string{prometheusScrapeAnnotation: "false"},
			Incl: map[string]string{prometheusScrapeAnnotation: "true"},
		}
		return
	}

	if ad.KubeAnnotations.Excl == nil {
		ad.KubeAnnotations.Excl = map[string]string{prometheusScrapeAnnotation: "false"}
	}

	if ad.KubeAnnotations.Incl == nil {
		ad.KubeAnnotations.Incl = map[string]string{prometheusScrapeAnnotation: "true"}
	}
}

// setContainersRegex precompiles the regex to match the container names for autodiscovery
// returns an error if the container names cannot be converted to a valid regex
func (ad *ADConfig) setContainersRegex() error {
	ad.containersRe = nil
	if len(ad.KubeContainerNames) == 0 {
		return nil
	}

	regexString := strings.Join(ad.KubeContainerNames, "|")
	re, err := regexp.Compile(regexString)
	if err != nil {
		return fmt.Errorf("Invalid container names - regex: '%s': %v", regexString, err)
	}

	ad.containersRe = re
	return nil
}

// matchContainer returns whether a container name matches the 'kubernetes_container_names' configuration
func (ad *ADConfig) matchContainer(name string) bool {
	if ad.containersRe == nil {
		return true
	}
	return ad.containersRe.MatchString(name)
}

// buildURL returns the 'prometheus_url' based on the default values
// and the prometheus path and port annotations
func buildURL(annotations map[string]string) string {
	port := openmetricsDefaultPort
	if portFromAnnotation, found := annotations[prometheusPortAnnotation]; found {
		port = portFromAnnotation
	}

	path := openmetricsDefaultPath
	if pathFromAnnotation, found := annotations[prometheusPathAnnotation]; found {
		path = pathFromAnnotation
	}

	return openmetricsURLPrefix + port + path
}

func boolPointer(b bool) *bool {
	return &b
}
