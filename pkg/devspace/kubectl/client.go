package kubectl

import (
	"fmt"
	"net"
	"net/url"
	"time"

	"github.com/devspace-cloud/devspace/pkg/devspace/config/generated"
	"github.com/devspace-cloud/devspace/pkg/util/kubeconfig"
	"github.com/devspace-cloud/devspace/pkg/util/log"
	"github.com/devspace-cloud/devspace/pkg/util/survey"

	"github.com/mgutz/ansi"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Client holds all important information for kubernetes
type Client struct {
	Client       kubernetes.Interface
	ClientConfig clientcmd.ClientConfig
	RestConfig   *rest.Config

	CurrentContext string
	Namespace      string
}

// NewDefaultClient creates the new default kube client from the active context
func NewDefaultClient() (*Client, error) {
	return NewClientFromContext("", "", false)
}

// NewClientFromContext creates a new kubernetes client from given context
func NewClientFromContext(context, namespace string, switchContext bool) (*Client, error) {
	// Load new raw config
	kubeConfig, err := kubeconfig.NewConfig().RawConfig()
	if err != nil {
		return nil, err
	}

	// If we should use a certain kube context use that
	activeContext := kubeConfig.CurrentContext
	if context != "" && activeContext != context {
		activeContext = context
		if switchContext {
			kubeConfig.CurrentContext = activeContext

			err = kubeconfig.SaveConfig(&kubeConfig)
			if err != nil {
				return nil, fmt.Errorf("Error saving kube config: %v", err)
			}
		}
	}

	// Change context namespace
	activeNamespace := metav1.NamespaceDefault
	if kubeConfig.Contexts[activeContext] != nil && kubeConfig.Contexts[activeContext].Namespace != "" {
		activeNamespace = kubeConfig.Contexts[activeContext].Namespace
	}
	if kubeConfig.Contexts[activeContext] != nil && namespace != "" && activeNamespace != namespace {
		activeNamespace = namespace
		kubeConfig.Contexts[activeContext].Namespace = namespace
	}

	clientConfig := clientcmd.NewNonInteractiveClientConfig(kubeConfig, activeContext, &clientcmd.ConfigOverrides{}, clientcmd.NewDefaultClientConfigLoadingRules())
	if kubeConfig.Contexts[activeContext] == nil {
		return nil, fmt.Errorf("Error loading kube config, context '%s' doesn't exist", activeContext)
	}

	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, err
	}

	client, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrap(err, "new client")
	}

	return &Client{
		Client:       client,
		ClientConfig: clientConfig,
		RestConfig:   restConfig,

		Namespace:      activeNamespace,
		CurrentContext: activeContext,
	}, nil
}

// NewClientBySelect creates a new kubernetes client by user select
func NewClientBySelect(allowPrivate bool, switchContext bool) (*Client, error) {
	kubeConfig, err := kubeconfig.LoadRawConfig()
	if err != nil {
		return nil, err
	}

	// Get all kube contexts
	options := make([]string, 0, len(kubeConfig.Contexts))
	for context := range kubeConfig.Contexts {
		options = append(options, context)
	}
	if len(options) == 0 {
		return nil, errors.New("No kubectl context found. Make sure kubectl is installed and you have a working kubernetes context configured")
	}

	for true {
		kubeContext := survey.Question(&survey.QuestionOptions{
			Question:     "Which kube context do you want to use",
			DefaultValue: kubeConfig.CurrentContext,
			Options:      options,
		})

		// Check if cluster is in private network
		if allowPrivate == false {
			context := kubeConfig.Contexts[kubeContext]
			cluster := kubeConfig.Clusters[context.Cluster]

			url, err := url.Parse(cluster.Server)
			if err != nil {
				return nil, errors.Wrap(err, "url parse")
			}

			ip := net.ParseIP(url.Hostname())
			if ip != nil {
				if IsPrivateIP(ip) {
					log.Infof("Clusters with private ips (%s) cannot be used", url.Hostname())
					continue
				}
			}
		}

		return NewClientFromContext(kubeContext, "", switchContext)
	}

	return nil, errors.New("We should not reach this point")
}

// PrintWarning prints a warning if the last kube context is different than this one
func (client *Client) PrintWarning(updateGenerated bool, log log.Logger) error {
	generatedConfig, err := generated.LoadConfig()
	if err == nil {
		// print warning if context or namespace has changed since last deployment process (expect if explicitly provided as flags)
		if generatedConfig.GetActive().LastContext != nil {
			wait := false

			if generatedConfig.GetActive().LastContext.Context != "" && generatedConfig.GetActive().LastContext.Context != client.CurrentContext {
				log.WriteString("\n")
				log.Warnf(ansi.Color("Are you using the correct kube context?", "white+b"))
				log.Warnf("Current kube context: '%s'", ansi.Color(client.CurrentContext, "white+b"))
				log.Warnf("Last    kube context: '%s'", ansi.Color(generatedConfig.GetActive().LastContext.Context, "white+b"))
				log.WriteString("\n")

				log.Infof("Run '%s' to change to the previous context", ansi.Color("devspace use context "+generatedConfig.GetActive().LastContext.Context, "white+b"))
				wait = true
			} else if generatedConfig.GetActive().LastContext.Namespace != "" && generatedConfig.GetActive().LastContext.Namespace != client.Namespace {
				log.WriteString("\n")
				log.Warnf(ansi.Color("Are you using the correct namespace?", "white+b"))
				log.Warnf("Current namespace: '%s'", ansi.Color(client.Namespace, "white+b"))
				log.Warnf("Last    namespace: '%s'", ansi.Color(generatedConfig.GetActive().LastContext.Namespace, "white+b"))
				log.WriteString("\n")

				log.Infof("Run '%s' to change to the previous namespace", ansi.Color("devspace use namespace "+generatedConfig.GetActive().LastContext.Namespace, "white+b"))
				wait = true
			}

			if wait && updateGenerated {
				log.StartWait("Will continue in 10 seconds...")
				time.Sleep(10 * time.Second)
				log.StopWait()
				log.WriteString("\n")
			}
		} else if updateGenerated && client.Namespace == metav1.NamespaceDefault {
			log.WriteString("\n")
			log.Warn("Deploying into the 'default' namespace is usually not a good idea as this namespace cannot be deleted")
			log.StartWait("Will continue in 5 seconds...")
			time.Sleep(5 * time.Second)
			log.StopWait()
			log.WriteString("\n")
		}

		// Update generated if we deploy the application
		if updateGenerated {
			generatedConfig.GetActive().LastContext = &generated.LastContextConfig{
				Context:   client.CurrentContext,
				Namespace: client.Namespace,
			}

			err = generated.SaveConfig(generatedConfig)
			if err != nil {
				return errors.Wrap(err, "save generated")
			}
		}
	}

	// Info messages
	log.Infof("Using kube context '%s'", ansi.Color(client.CurrentContext, "white+b"))
	log.Infof("Using namespace '%s'", ansi.Color(client.Namespace, "white+b"))

	return nil
}
