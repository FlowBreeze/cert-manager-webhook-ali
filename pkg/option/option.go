package option

import (
	"fmt"
	"github.com/jetstack/cert-manager/pkg/acme/webhook"
	whapi "github.com/jetstack/cert-manager/pkg/acme/webhook/apis/acme/v1alpha1"
	"github.com/jetstack/cert-manager/pkg/acme/webhook/apiserver"
	"io"
	"k8s.io/apiserver/pkg/features"
	"k8s.io/apiserver/pkg/server"
	"k8s.io/apiserver/pkg/server/options"
	"k8s.io/apiserver/pkg/util/feature"
	"net"
)

const defaultEtcdPathPrefix = "/registry/acme.cert-manager.io"

type Options struct {
	RecommendedOptions *options.RecommendedOptions

	SolverGroup string
	Solvers     []webhook.Solver

	StdOut io.Writer
	StdErr io.Writer
}

// applyTo patched from "k8s.io/apiserver/pkg/server/options/recommended.go"
func applyTo(o *options.RecommendedOptions, config *server.RecommendedConfig) error {
	if err := o.Etcd.ApplyTo(&config.Config); err != nil {
		return err
	}
	if err := o.EgressSelector.ApplyTo(&config.Config); err != nil {
		return err
	}
	if feature.DefaultFeatureGate.Enabled(features.APIServerTracing) {
		if err := o.Traces.ApplyTo(config.Config.EgressSelector, &config.Config); err != nil {
			return err
		}
	}
	if err := o.SecureServing.ApplyTo(&config.Config.SecureServing, &config.Config.LoopbackClientConfig); err != nil {
		return err
	}
	if err := o.Authentication.ApplyTo(&config.Config.Authentication, config.SecureServing, config.OpenAPIConfig); err != nil {
		return err
	}
	if err := o.Authorization.ApplyTo(&config.Config.Authorization); err != nil {
		return err
	}
	if err := o.Audit.ApplyTo(&config.Config); err != nil {
		return err
	}
	if err := o.Features.ApplyTo(&config.Config); err != nil {
		return err
	}
	if err := o.CoreAPI.ApplyTo(config); err != nil {
		return err
	}
	if initializers, err := o.ExtraAdmissionInitializers(config); err != nil {
		return err
	} else if err := o.Admission.ApplyTo(&config.Config, config.SharedInformerFactory, config.ClientConfig, o.FeatureGate, initializers...); err != nil {
		return err
	}

	// remove priority and fairness ,so we don't need to create a role

	return nil
}

func (o Options) Config() (*apiserver.Config, error) {
	// TODO have a "real" external address
	if err := o.RecommendedOptions.SecureServing.MaybeDefaultWithSelfSignedCerts("localhost", nil, []net.IP{net.ParseIP("127.0.0.1")}); err != nil {
		return nil, fmt.Errorf("error creating self-signed certificates: %v", err)
	}

	serverConfig := server.NewRecommendedConfig(apiserver.Codecs)

	if err := applyTo(o.RecommendedOptions, serverConfig); err != nil {
		return nil, err
	}

	config := &apiserver.Config{
		GenericConfig: serverConfig,
		ExtraConfig: apiserver.ExtraConfig{
			SolverGroup: o.SolverGroup,
			Solvers:     o.Solvers,
		},
	}
	return config, nil
}

func NewOptions(out, errOut io.Writer, groupName string, solvers ...webhook.Solver) *Options {
	o := &Options{
		// TODO we will nil out the etcd storage options.  This requires a later level of k8s.io/apiserver
		RecommendedOptions: options.NewRecommendedOptions(
			defaultEtcdPathPrefix,
			apiserver.Codecs.LegacyCodec(whapi.SchemeGroupVersion),
		),

		SolverGroup: groupName,
		Solvers:     solvers,

		StdOut: out,
		StdErr: errOut,
	}
	o.RecommendedOptions.Etcd = nil
	o.RecommendedOptions.Admission = nil

	return o
}
