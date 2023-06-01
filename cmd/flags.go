package main

import "flag"

var (
	MetricsAddr          string
	EnableLeaderElection bool
	ProbeAddr            string
	AssociationNamespace string
)

func init() {
	flag.StringVar(&MetricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&ProbeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&EnableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.StringVar(&AssociationNamespace, "association-namespace", "",
		"The namespace where the EksPodEipAssociation CR is created. "+
			"If not specified, the CR will be created in the same namespace as the Pod.")
}
