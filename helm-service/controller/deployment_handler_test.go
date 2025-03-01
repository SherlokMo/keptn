package controller

import (
	"testing"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/golang/mock/gomock"
	keptn "github.com/keptn/go-utils/pkg/lib"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	"github.com/keptn/keptn/helm-service/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
)

func TestHandleEventWithDeploymentURLAndUserManagedDeploymentStrategy(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockedBaseHandler := NewMockedHandler(createKeptn(), "")
	mockedOnboarder := mocks.NewMockOnboarder(ctrl)
	mockedChartGenerator := mocks.NewMockChartGenerator(ctrl)

	deploymentHandler := DeploymentHandler{
		Handler:               mockedBaseHandler,
		mesh:                  mocks.NewMockMesh(ctrl),
		generatedChartHandler: mockedChartGenerator,
		onboarder:             mockedOnboarder,
	}

	deploymentTriggeredEventData := keptnv2.DeploymentTriggeredEventData{
		EventData: keptnv2.EventData{
			Project: "my-project",
			Stage:   "production",
			Service: "my-service",
		},
		ConfigurationChange: keptnv2.ConfigurationChange{},
		Deployment: keptnv2.DeploymentTriggeredData{
			DeploymentStrategy:   keptn.UserManaged.String(),
			DeploymentURIsPublic: []string{"https://myurl"},
		},
	}

	ce := cloudevents.NewEvent()
	_ = ce.SetData(cloudevents.ApplicationJSON, deploymentTriggeredEventData)
	ce.SetExtension("gitcommitid", "USER_CHART_GIT_ID")
	deploymentHandler.HandleEvent(ce)

	expectedDeploymentFinishedEvent := cloudevents.NewEvent()
	expectedDeploymentFinishedEvent.SetType("sh.keptn.event.deployment.finished")
	expectedDeploymentFinishedEvent.SetSource("helm-service")
	expectedDeploymentFinishedEvent.SetDataContentType(cloudevents.ApplicationJSON)
	expectedDeploymentFinishedEvent.SetExtension("triggeredid", "")
	expectedDeploymentFinishedEvent.SetExtension("shkeptncontext", "")
	expectedDeploymentFinishedEvent.SetData(cloudevents.ApplicationJSON, keptnv2.DeploymentFinishedEventData{
		EventData: keptnv2.EventData{
			Project: "my-project",
			Stage:   "production",
			Service: "my-service",
			Status:  keptnv2.StatusSucceeded,
			Result:  keptnv2.ResultPass,
			Message: "Successfully deployed",
		},
		Deployment: keptnv2.DeploymentFinishedData{
			DeploymentStrategy:   "user_managed",
			DeploymentURIsPublic: []string{"https://myurl"},
			DeploymentNames:      []string{"user_managed"},
			GitCommit:            "USER_CHART_GIT_ID",
		},
	})

	require.Equal(t, 2, len(mockedBaseHandler.sentCloudEvents))
	assert.Equal(t, expectedDeploymentFinishedEvent, mockedBaseHandler.sentCloudEvents[1])
	require.Equal(t, 2, len(mockedBaseHandler.upgradeChartInvocations))
	assert.Equal(t, "carts", mockedBaseHandler.upgradeChartInvocations[0].ch.Metadata.Name)
	assert.Equal(t, deploymentTriggeredEventData.EventData, mockedBaseHandler.upgradeChartInvocations[0].event)
	assert.Equal(t, keptn.UserManaged, mockedBaseHandler.upgradeChartInvocations[0].strategy)
}

func TestHandleEventWithDeploymentURLAndDirectDeploymentStrategy(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockedBaseHandler := NewMockedHandler(createKeptn(), "")
	mockedOnboarder := mocks.NewMockOnboarder(ctrl)
	mockedChartGenerator := mocks.NewMockChartGenerator(ctrl)

	deploymentHandler := DeploymentHandler{
		Handler:               mockedBaseHandler,
		mesh:                  mocks.NewMockMesh(ctrl),
		generatedChartHandler: mockedChartGenerator,
		onboarder:             mockedOnboarder,
	}

	deploymentTriggeredEventData := keptnv2.DeploymentTriggeredEventData{
		EventData: keptnv2.EventData{
			Project: "my-project",
			Stage:   "production",
			Service: "my-service",
		},
		ConfigurationChange: keptnv2.ConfigurationChange{},
		Deployment: keptnv2.DeploymentTriggeredData{
			DeploymentStrategy:  keptn.Direct.String(),
			DeploymentURIsLocal: []string{"http://my-service.my-project-dev:80"},
		},
	}

	ce := cloudevents.NewEvent()
	_ = ce.SetData(cloudevents.ApplicationJSON, deploymentTriggeredEventData)
	ce.SetExtension("gitcommitid", "USER_CHART_GIT_ID")
	deploymentHandler.HandleEvent(ce)

	expectedDeploymentFinishedEvent := cloudevents.NewEvent()
	expectedDeploymentFinishedEvent.SetType("sh.keptn.event.deployment.finished")
	expectedDeploymentFinishedEvent.SetSource("helm-service")
	expectedDeploymentFinishedEvent.SetDataContentType(cloudevents.ApplicationJSON)
	expectedDeploymentFinishedEvent.SetExtension("triggeredid", "")
	expectedDeploymentFinishedEvent.SetExtension("shkeptncontext", "")
	expectedDeploymentFinishedEvent.SetData(cloudevents.ApplicationJSON, keptnv2.DeploymentFinishedEventData{
		EventData: keptnv2.EventData{
			Project: "my-project",
			Stage:   "production",
			Service: "my-service",
			Status:  keptnv2.StatusSucceeded,
			Result:  keptnv2.ResultPass,
			Message: "Successfully deployed",
		},
		Deployment: keptnv2.DeploymentFinishedData{
			DeploymentStrategy:   "direct",
			DeploymentURIsLocal:  []string{"http://my-service.my-project-production:80"},
			DeploymentURIsPublic: []string{"http://my-service.my-project-production.svc.cluster.local:80"},
			DeploymentNames:      []string{"direct"},
			GitCommit:            "USER_CHART_GIT_ID",
		},
	})

	require.Equal(t, 2, len(mockedBaseHandler.sentCloudEvents))
	assert.Equal(t, expectedDeploymentFinishedEvent, mockedBaseHandler.sentCloudEvents[1])
	require.Equal(t, 2, len(mockedBaseHandler.upgradeChartInvocations))
	assert.Equal(t, "carts", mockedBaseHandler.upgradeChartInvocations[0].ch.Metadata.Name)
	assert.Equal(t, deploymentTriggeredEventData.EventData, mockedBaseHandler.upgradeChartInvocations[0].event)
	assert.Equal(t, keptn.Direct, mockedBaseHandler.upgradeChartInvocations[0].strategy)
	assert.Equal(t, "carts-generated", mockedBaseHandler.upgradeChartInvocations[1].ch.Metadata.Name)
	assert.Equal(t, deploymentTriggeredEventData.EventData, mockedBaseHandler.upgradeChartInvocations[1].event)
	assert.Equal(t, keptn.Direct, mockedBaseHandler.upgradeChartInvocations[1].strategy)
}

func TestHandleEventWithNoConfigurationChangeAndDirectDeploymentStrategy(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockedBaseHandler := NewMockedHandler(createKeptn(), "")
	mockedOnboarder := mocks.NewMockOnboarder(ctrl)
	mockedChartGenerator := mocks.NewMockChartGenerator(ctrl)

	deploymentHandler := DeploymentHandler{
		Handler:               mockedBaseHandler,
		mesh:                  mocks.NewMockMesh(ctrl),
		generatedChartHandler: mockedChartGenerator,
		onboarder:             mockedOnboarder,
	}

	deploymentTriggeredEventData := keptnv2.DeploymentTriggeredEventData{
		EventData: keptnv2.EventData{
			Project: "my-project",
			Stage:   "my-stage",
			Service: "my-service",
		},
		ConfigurationChange: keptnv2.ConfigurationChange{},
		Deployment: keptnv2.DeploymentTriggeredData{
			DeploymentStrategy: keptn.Direct.String(),
		},
	}

	ce := cloudevents.NewEvent()
	_ = ce.SetData(cloudevents.ApplicationJSON, deploymentTriggeredEventData)
	ce.SetExtension("gitcommitid", "USER_CHART_GIT_ID")
	deploymentHandler.HandleEvent(ce)

	expectedDeploymentFinishedEvent := cloudevents.NewEvent()
	expectedDeploymentFinishedEvent.SetType("sh.keptn.event.deployment.finished")
	expectedDeploymentFinishedEvent.SetSource("helm-service")
	expectedDeploymentFinishedEvent.SetDataContentType(cloudevents.ApplicationJSON)
	expectedDeploymentFinishedEvent.SetExtension("triggeredid", "")
	expectedDeploymentFinishedEvent.SetExtension("shkeptncontext", "")
	expectedDeploymentFinishedEvent.SetData(cloudevents.ApplicationJSON, keptnv2.DeploymentFinishedEventData{
		EventData: keptnv2.EventData{
			Project: "my-project",
			Stage:   "my-stage",
			Service: "my-service",
			Status:  keptnv2.StatusSucceeded,
			Result:  keptnv2.ResultPass,
			Message: "Successfully deployed",
		},
		Deployment: keptnv2.DeploymentFinishedData{
			DeploymentStrategy:   "direct",
			DeploymentURIsLocal:  []string{"http://my-service.my-project-my-stage:80"},
			DeploymentURIsPublic: []string{"http://my-service.my-project-my-stage.svc.cluster.local:80"},
			DeploymentNames:      []string{"direct"},
			GitCommit:            "USER_CHART_GIT_ID",
		},
	})

	require.Equal(t, 2, len(mockedBaseHandler.sentCloudEvents))
	assert.Equal(t, expectedDeploymentFinishedEvent, mockedBaseHandler.sentCloudEvents[1])
	require.Equal(t, 2, len(mockedBaseHandler.upgradeChartInvocations))
	assert.Equal(t, "carts", mockedBaseHandler.upgradeChartInvocations[0].ch.Metadata.Name)
	assert.Equal(t, deploymentTriggeredEventData.EventData, mockedBaseHandler.upgradeChartInvocations[0].event)
	assert.Equal(t, keptn.Direct, mockedBaseHandler.upgradeChartInvocations[0].strategy)
	assert.Equal(t, "carts-generated", mockedBaseHandler.upgradeChartInvocations[1].ch.Metadata.Name)
	assert.Equal(t, deploymentTriggeredEventData.EventData, mockedBaseHandler.upgradeChartInvocations[1].event)
	assert.Equal(t, keptn.Direct, mockedBaseHandler.upgradeChartInvocations[1].strategy)
}

func TestHandleEventWithDeploymentURLResourceAndUserManagedDeploymentStrategy(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockedBaseHandler := NewMockedHandler(createKeptn(), "", func(options *MockedHandlerOptions) {
		options.GetUserManagedEndpointsBehavior = func(event keptnv2.EventData) (*keptnv2.Endpoints, error) {
			return &keptnv2.Endpoints{
				DeploymentURIsLocal:  []string{"https://my-local-url"},
				DeploymentURIsPublic: []string{"https://myurl"},
			}, nil
		}
	})
	mockedOnboarder := mocks.NewMockOnboarder(ctrl)
	mockedChartGenerator := mocks.NewMockChartGenerator(ctrl)

	deploymentHandler := DeploymentHandler{
		Handler:               mockedBaseHandler,
		mesh:                  mocks.NewMockMesh(ctrl),
		generatedChartHandler: mockedChartGenerator,
		onboarder:             mockedOnboarder,
	}

	deploymentTriggeredEventData := keptnv2.DeploymentTriggeredEventData{
		EventData: keptnv2.EventData{
			Project: "my-project",
			Stage:   "production",
			Service: "my-service",
		},
		ConfigurationChange: keptnv2.ConfigurationChange{},
		Deployment: keptnv2.DeploymentTriggeredData{
			DeploymentStrategy: keptn.UserManaged.String(),
		},
	}

	ce := cloudevents.NewEvent()
	_ = ce.SetData(cloudevents.ApplicationJSON, deploymentTriggeredEventData)
	ce.SetExtension("gitcommitid", "USER_CHART_GIT_ID")
	deploymentHandler.HandleEvent(ce)

	expectedDeploymentFinishedEvent := cloudevents.NewEvent()
	expectedDeploymentFinishedEvent.SetType("sh.keptn.event.deployment.finished")
	expectedDeploymentFinishedEvent.SetSource("helm-service")
	expectedDeploymentFinishedEvent.SetDataContentType(cloudevents.ApplicationJSON)
	expectedDeploymentFinishedEvent.SetExtension("triggeredid", "")
	expectedDeploymentFinishedEvent.SetExtension("shkeptncontext", "")
	expectedDeploymentFinishedEvent.SetData(cloudevents.ApplicationJSON, keptnv2.DeploymentFinishedEventData{
		EventData: keptnv2.EventData{
			Project: "my-project",
			Stage:   "production",
			Service: "my-service",
			Status:  keptnv2.StatusSucceeded,
			Result:  keptnv2.ResultPass,
			Message: "Successfully deployed",
		},
		Deployment: keptnv2.DeploymentFinishedData{
			DeploymentStrategy:   "user_managed",
			DeploymentURIsLocal:  []string{"https://my-local-url"},
			DeploymentURIsPublic: []string{"https://myurl"},
			DeploymentNames:      []string{"user_managed"},
			GitCommit:            "USER_CHART_GIT_ID",
		},
	})

	require.Equal(t, 2, len(mockedBaseHandler.sentCloudEvents))
	assert.Equal(t, expectedDeploymentFinishedEvent, mockedBaseHandler.sentCloudEvents[1])
	require.Equal(t, 2, len(mockedBaseHandler.upgradeChartInvocations))
	assert.Equal(t, "carts", mockedBaseHandler.upgradeChartInvocations[0].ch.Metadata.Name)
	assert.Equal(t, deploymentTriggeredEventData.EventData, mockedBaseHandler.upgradeChartInvocations[0].event)
	assert.Equal(t, keptn.UserManaged, mockedBaseHandler.upgradeChartInvocations[0].strategy)
}

func TestHandleEventWithoutDeploymentURLAndUserManagedDeploymentStrategy(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockedBaseHandler := NewMockedHandler(createKeptn(), "", func(options *MockedHandlerOptions) {
		options.GetUserManagedEndpointsBehavior = func(event keptnv2.EventData) (*keptnv2.Endpoints, error) {
			return nil, nil
		}
	})
	mockedOnboarder := mocks.NewMockOnboarder(ctrl)
	mockedChartGenerator := mocks.NewMockChartGenerator(ctrl)

	deploymentHandler := DeploymentHandler{
		Handler:               mockedBaseHandler,
		mesh:                  mocks.NewMockMesh(ctrl),
		generatedChartHandler: mockedChartGenerator,
		onboarder:             mockedOnboarder,
	}

	deploymentTriggeredEventData := keptnv2.DeploymentTriggeredEventData{
		EventData: keptnv2.EventData{
			Project: "my-project",
			Stage:   "production",
			Service: "my-service",
		},
		ConfigurationChange: keptnv2.ConfigurationChange{},
		Deployment: keptnv2.DeploymentTriggeredData{
			DeploymentStrategy: keptn.UserManaged.String(),
		},
	}

	ce := cloudevents.NewEvent()
	_ = ce.SetData(cloudevents.ApplicationJSON, deploymentTriggeredEventData)
	ce.SetExtension("gitcommitid", "USER_CHART_GIT_ID")
	deploymentHandler.HandleEvent(ce)

	expectedDeploymentFinishedEvent := cloudevents.NewEvent()
	expectedDeploymentFinishedEvent.SetType("sh.keptn.event.deployment.finished")
	expectedDeploymentFinishedEvent.SetSource("helm-service")
	expectedDeploymentFinishedEvent.SetDataContentType(cloudevents.ApplicationJSON)
	expectedDeploymentFinishedEvent.SetExtension("triggeredid", "")
	expectedDeploymentFinishedEvent.SetExtension("shkeptncontext", "")
	expectedDeploymentFinishedEvent.SetData(cloudevents.ApplicationJSON, keptnv2.DeploymentFinishedEventData{
		EventData: keptnv2.EventData{
			Project: "my-project",
			Stage:   "production",
			Service: "my-service",
			Status:  keptnv2.StatusSucceeded,
			Result:  keptnv2.ResultPass,
			Message: "Successfully deployed",
		},
		Deployment: keptnv2.DeploymentFinishedData{
			DeploymentStrategy: "user_managed",
			DeploymentNames:    []string{"user_managed"},
			GitCommit:          "USER_CHART_GIT_ID",
		},
	})

	require.Equal(t, 2, len(mockedBaseHandler.sentCloudEvents))
	assert.Equal(t, expectedDeploymentFinishedEvent, mockedBaseHandler.sentCloudEvents[1])
	require.Equal(t, 2, len(mockedBaseHandler.upgradeChartInvocations))
	assert.Equal(t, "carts", mockedBaseHandler.upgradeChartInvocations[0].ch.Metadata.Name)
	assert.Equal(t, deploymentTriggeredEventData.EventData, mockedBaseHandler.upgradeChartInvocations[0].event)
	assert.Equal(t, keptn.UserManaged, mockedBaseHandler.upgradeChartInvocations[0].strategy)
}

func TestHandleEventWithoutResources(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockedBaseHandler := NewMockedHandler(createKeptn(), "", func(options *MockedHandlerOptions) {
		options.GetUserManagedEndpointsBehavior = func(event keptnv2.EventData) (*keptnv2.Endpoints, error) {
			return nil, nil
		}
		options.GetUserChartsErrBehaviour = true

	})

	mockedOnboarder := mocks.NewMockOnboarder(ctrl)
	mockedChartGenerator := mocks.NewMockChartGenerator(ctrl)

	deploymentHandler := DeploymentHandler{
		Handler:               mockedBaseHandler,
		mesh:                  mocks.NewMockMesh(ctrl),
		generatedChartHandler: mockedChartGenerator,
		onboarder:             mockedOnboarder,
	}

	deploymentTriggeredEventData := keptnv2.DeploymentTriggeredEventData{
		EventData: keptnv2.EventData{
			Project: "my-project",
			Stage:   "production",
			Service: "my-service",
		},
		ConfigurationChange: keptnv2.ConfigurationChange{
			Values: map[string]interface{}{
				"image": "nexus-registry.apps.ocp4-test.it/krateo-20220308:latest",
			},
		},
		Deployment: keptnv2.DeploymentTriggeredData{
			DeploymentStrategy: keptn.UserManaged.String(),
		},
	}

	ce := cloudevents.NewEvent()
	_ = ce.SetData(cloudevents.ApplicationJSON, deploymentTriggeredEventData)
	ce.SetExtension("gitcommitid", "silly-one")
	deploymentHandler.HandleEvent(ce)

	expectedDeploymentFinishedEvent := keptnv2.DeploymentFinishedEventData{
		EventData: keptnv2.EventData{
			Project: "my-project",
			Stage:   "production",
			Service: "my-service",
			Status:  keptnv2.StatusErrored,
			Result:  keptnv2.ResultFailed,
			Message: "failed to load chart: some error",
		},
	}

	assert.Equal(t, 1, len(mockedBaseHandler.handledErrorEvents))
	assert.Equal(t, expectedDeploymentFinishedEvent, mockedBaseHandler.handledErrorEvents[0])
}

func TestHandleUnparsableDeploymentEvent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockedBaseHandler := NewMockedHandler(createKeptn(), "")

	instance := DeploymentHandler{
		Handler: mockedBaseHandler,
		mesh:    mocks.NewMockMesh(ctrl),
	}

	expectedTriggerEventData := keptnv2.DeploymentFinishedEventData{
		EventData: keptnv2.EventData{
			Status:  "errored",
			Result:  "fail",
			Message: "Failed to unmarshal data: unable to convert json data from cloudEvent to deployment event",
		},
	}

	instance.HandleEvent(createUnparsableEvent())
	assert.Equal(t, 1, len(mockedBaseHandler.handledErrorEvents))
	assert.Equal(t, expectedTriggerEventData, mockedBaseHandler.handledErrorEvents[0])
}

func Test_getPortOfService(t *testing.T) {
	type args struct {
		service *corev1.Service
	}
	tests := []struct {
		name  string
		args  args
		want  int32
		want1 bool
	}{
		{
			name: "get tcp port 80",
			args: args{
				service: &corev1.Service{
					Spec: corev1.ServiceSpec{
						Ports: []corev1.ServicePort{
							{
								Name:     "http",
								Protocol: corev1.ProtocolTCP,
								Port:     80,
							},
						},
					},
				},
			},
			want:  80,
			want1: true,
		},
		{
			name: "multiple tcp ports: get lowest (80)",
			args: args{
				service: &corev1.Service{
					Spec: corev1.ServiceSpec{
						Ports: []corev1.ServicePort{
							{
								Name:     "http",
								Protocol: corev1.ProtocolTCP,
								Port:     80,
							},
							{
								Name:     "http2",
								Protocol: corev1.ProtocolTCP,
								Port:     8080,
							},
						},
					},
				},
			},
			want:  80,
			want1: true,
		},
		{
			name: "no port marked explicitly as tcp found - get port 80",
			args: args{
				service: &corev1.Service{
					Spec: corev1.ServiceSpec{
						Ports: []corev1.ServicePort{
							{
								Name: "http",
								Port: 80,
							},
						},
					},
				},
			},
			want:  80,
			want1: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := getPortOfService(tt.args.service)
			if got != tt.want {
				t.Errorf("getPortOfService() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("getPortOfService() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
